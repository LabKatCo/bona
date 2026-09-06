package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/labkatco/bona/constants"
	"github.com/labkatco/bona/parse"
)

var mirrorDirName = fmt.Sprintf("__%v_mirror", constants.LibName)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run %v_watcher.go <project_directory>", constants.LibName)
	}

	srcDir, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatalf("Could not resolve project directory: %v", err)
	}
	mirrorDir := filepath.Join(srcDir, mirrorDirName)

	fmt.Printf("Starting %v dev tool...\n", constants.LibName)
	fmt.Printf("Watching: %s\n", srcDir)
	fmt.Printf("Mirror:   %s\n", mirrorDir)

	// 1. First Pass: Mirror existing files
	err = mirrorTree(srcDir, mirrorDir)
	if err != nil {
		log.Fatalf("Initial mirror failed: %v", err)
	}

	// 2. Setup Watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Watch all subdirectories except the mirror and hidden folders
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if info.Name() == mirrorDirName || strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// 3. Listen for events
	fmt.Println("Listening for file changes...")
	lastProcessedContent := make(map[string][]byte)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				// Ignore any file ending in test.go (e.g. _test.go)
				if strings.HasSuffix(event.Name, ".go") && !strings.HasSuffix(event.Name, "test.go") {
					// Add a small debounce to allow IDE saves to complete
					time.Sleep(50 * time.Millisecond)
					content, err := os.ReadFile(event.Name)
					if err != nil {
						log.Printf("Read Error on %s: %v\n", event.Name, err)
						continue
					}
					if previous, ok := lastProcessedContent[event.Name]; ok && bytes.Equal(previous, content) {
						continue
					}

					relPath, _ := filepath.Rel(srcDir, event.Name)
					dstPath := filepath.Join(mirrorDir, relPath)
					fmt.Printf("Change detected: %s\n", relPath)

					os.MkdirAll(filepath.Dir(dstPath), 0755)

					// Catch and report AST parsing errors
					if err := processGoFile(event.Name, dstPath); err != nil {
						log.Printf("Parse Error on %s: %v\n", relPath, err)
						continue
					}
					lastProcessedContent[event.Name] = append([]byte(nil), content...)

					// Trigger a build in the mirror to surface type and compiler errors
					start := time.Now()

					cmd := exec.Command("go", "build", "./...")
					cmd.Dir = mirrorDir
					out, err := cmd.CombinedOutput()

					duration := time.Since(start)

					if err != nil {
						log.Printf("Compiler Error (took %v):\n%s\n", duration, string(out))
					} else {
						log.Printf("Successfully compiled %s in %v\n", relPath, duration)
					}
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}

func mirrorTree(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip mirroring the mirror dir itself and hidden dirs
		if info.IsDir() && (info.Name() == mirrorDirName || strings.HasPrefix(info.Name(), ".")) {
			return filepath.SkipDir
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		// Explicitly ignore test.go files, but process other .go files
		if strings.HasSuffix(info.Name(), ".go") {
			if strings.HasSuffix(info.Name(), "test.go") {
				return nil // skip test files entirely
			}
			return processGoFile(path, dstPath)
		}

		// Copy raw file for everything else (go.mod, etc)
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(dstPath, b, info.Mode())
	})
}

func processGoFile(srcPath, dstPath string) error {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, srcPath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse err on %s: %w", srcPath, err)
	}

	pkgName := f.Name.Name
	var hints []string
	modified := false

	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			isTarget, pragma := parse.HasPragma(fn.Doc)
			if isTarget {
				funcName := pkgName + "." + fn.Name.Name
				hints = append(hints, pragma+"|"+funcName)
				transformFunc(fn, funcName)
				modified = true
			}
		}
	}

	// If we injected logging into any functions, add an init() for the hints
	if modified {
		if len(hints) > 0 {
			f.Decls = append(f.Decls, parse.BuildHintInitFunc(hints))
		}
		writeRuntimeLogger(filepath.Dir(dstPath), pkgName, filepath.Dir(srcPath), filepath.Dir(dstPath))
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return fmt.Errorf("format err on %s: %w", srcPath, err)
	}

	return os.WriteFile(dstPath, buf.Bytes(), 0644)
}

func transformFunc(fn *ast.FuncDecl, funcName string) {
	var outNames []string
	if fn.Type.Results != nil {
		idx := 0
		for _, field := range fn.Type.Results.List {
			if len(field.Names) == 0 {
				// Name unnamed returns so we can capture them with defer
				name := ast.NewIdent(parse.ResultName(field, idx))
				field.Names = []*ast.Ident{name}
				outNames = append(outNames, name.Name)
				idx++
			} else {
				for _, name := range field.Names {
					outNames = append(outNames, name.Name)
					idx++
				}
			}
		}
	}

	inNames := parse.ParameterNames(fn.Type.Params)

	// 1. Rewrite assignments in the body
	rewriteBlocks(fn.Body, funcName)

	// 2. Prepend Input and Output loggers to body
	var newStmts []ast.Stmt

	// Recover only to log the panic; re-panic immediately so the original behavior is preserved.
	deferStr := fmt.Sprintf(`defer func() { if __%v_panic := recover(); __%v_panic != nil { __%v_LogPanic("%s", __%v_panic); panic(__%v_panic) }; __%v_LogOutput("%s"`,
		constants.LibName, constants.LibName, constants.LibName, funcName, constants.LibName, constants.LibName, constants.LibName, funcName)
	if len(outNames) > 0 {
		deferStr += ", " + strings.Join(outNames, ", ")
	}
	deferStr += ") }() "
	newStmts = append(newStmts, parse.ParseStmt(deferStr))

	// Input (Immediate)
	inputStr := fmt.Sprintf(`__%v_LogInput("%s"`, constants.LibName, funcName)
	if len(inNames) > 0 {
		inputStr += ", " + strings.Join(inNames, ", ")
	}
	inputStr += ")"
	newStmts = append(newStmts, parse.ParseStmt(inputStr))

	fn.Body.List = append(newStmts, fn.Body.List...)
}

func rewriteBlocks(node ast.Node, funcName string) {
	ast.Inspect(node, func(n ast.Node) bool {
		if rangeStmt, ok := n.(*ast.RangeStmt); ok {
			var logs []ast.Stmt
			for _, expr := range []ast.Expr{rangeStmt.Key, rangeStmt.Value} {
				if id, ok := expr.(*ast.Ident); ok && id.Name != "_" {
					logs = append(logs, parse.AssignmentLog(id.Name, funcName))
				}
			}
			rangeStmt.Body.List = append(logs, rangeStmt.Body.List...)
		}
		if block, ok := n.(*ast.BlockStmt); ok {
			var newList []ast.Stmt
			for _, stmt := range block.List {
				newList = append(newList, stmt)
				if as, ok := stmt.(*ast.AssignStmt); ok {
					for _, lhs := range as.Lhs {
						if id, ok := lhs.(*ast.Ident); ok && id.Name != "_" {
							newList = append(newList, parse.AssignmentLog(id.Name, funcName))
						}
					}
				}
				if incDec, ok := stmt.(*ast.IncDecStmt); ok {
					if id, ok := incDec.X.(*ast.Ident); ok && id.Name != "_" {
						newList = append(newList, parse.AssignmentLog(id.Name, funcName))
					}
				}
			}
			block.List = newList
		}
		return true
	})
}

// writeRuntimeLogger drops an isolated runtime logger file into the target package so that
// transformed files can log inputs and outputs without cyclic dependency or import issues.
func writeRuntimeLogger(dir, pkgName, sourceDir, mirrorDir string) {
	sourceDir, _ = filepath.Abs(sourceDir)
	mirrorDir, _ = filepath.Abs(mirrorDir)
	sourceDir = filepath.ToSlash(sourceDir)
	mirrorDir = filepath.ToSlash(mirrorDir)
	content := strings.ReplaceAll(fmt.Sprintf(loggerTmpl, pkgName, sourceDir, mirrorDir), "{{libName}}", constants.LibName)
	// REMOVED the "__" prefix so the compiler includes this file
	_ = os.WriteFile(filepath.Join(dir, constants.LibName+"_logger_gen.go"), []byte(content), 0644)
}

const loggerTmpl = `// Code generated by {{libName}} dev tool. DO NOT EDIT.
package %s

import (
	"fmt"
	"runtime/debug"
	"os"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
	"sync"
)

var (
	__{{libName}}_out  *os.File
	__{{libName}}_once sync.Once
)

const (
	__{{libName}}_sourceDir = %q
	__{{libName}}_mirrorDir = %q
)

func __{{libName}}_initLogger() {
	__{{libName}}_once.Do(func() {
		toFile := false
		outPath := "{{libName}}_events.txt"
		for _, arg := range os.Args {
			if arg == "--to_file" {
				toFile = true
			} else if strings.HasPrefix(arg, "--output=") {
				outPath = strings.TrimPrefix(arg, "--output=")
			}
		}
		if toFile {
			f, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err == nil {
				__{{libName}}_out = f
			}
		}
	})
}

func __{{libName}}_Log(format string, args ...any) {
	__{{libName}}_initLogger()
	msg := fmt.Sprintf(format, args...)
	fmt.Println(msg)
	if __{{libName}}_out != nil {
		__{{libName}}_out.WriteString(msg + "\n")
	}
}

func __{{libName}}_LogHint(pragma, funcName string) {
	__{{libName}}_Log("hint|%%s|%%s", pragma, funcName)
}

func __{{libName}}_LogInput(funcName string, args ...any) {
	parts := []string{"input", funcName}
	for _, arg := range args {
		parts = append(parts, __{{libName}}_FormatValue(arg))
	}
	__{{libName}}_Log("%%s", strings.Join(parts, "|"))
}

func __{{libName}}_LogOutput(funcName string, args ...any) {
	parts := []string{"output", funcName}
	for _, arg := range args {
		parts = append(parts, __{{libName}}_FormatValue(arg))
	}
	__{{libName}}_Log("%%s", strings.Join(parts, "|"))
}

func __{{libName}}_LogAssign(funcName, varName string, val any) {
	__{{libName}}_Log("assign|%%s|%%s|%%s", funcName, varName, __{{libName}}_FormatValue(val))
}

func __{{libName}}_LogPanic(funcName string, recovered any) {
	stack := string(debug.Stack())
	stack = strings.ReplaceAll(stack, __{{libName}}_mirrorDir, __{{libName}}_sourceDir)
	__{{libName}}_Log("panic|%%s|%%s\n%%s", funcName, __{{libName}}_FormatValue(recovered), stack)
}

func __{{libName}}_FormatValue(value any) string {
	rendered := fmt.Sprintf("%%#v", value)
	expr, err := parser.ParseExpr(rendered)
	if err != nil {
		return rendered
	}

	__{{libName}}_ElideLiteralTypes(expr, false)
	var buf strings.Builder
	if err := format.Node(&buf, token.NewFileSet(), expr); err != nil {
		return rendered
	}
	return buf.String()
}

func __{{libName}}_ElideLiteralTypes(node ast.Node, elide bool) {
	literal, ok := node.(*ast.CompositeLit)
	if !ok {
		ast.Inspect(node, func(node ast.Node) bool {
			if nested, ok := node.(*ast.CompositeLit); ok {
				__{{libName}}_ElideLiteralTypes(nested, false)
				return false
			}
			return true
		})
		return
	}

	literalType := literal.Type
	if elide {
		literal.Type = nil
	}
	allowsElision := false
	switch literalType.(type) {
	case *ast.ArrayType, *ast.MapType:
		allowsElision = true
	}
	for _, element := range literal.Elts {
		if keyValue, ok := element.(*ast.KeyValueExpr); ok {
			__{{libName}}_ElideLiteralTypes(keyValue.Value, allowsElision)
			continue
		}
		__{{libName}}_ElideLiteralTypes(element, allowsElision)
	}
}
`
