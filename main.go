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
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

const libName = "bona"

var mirrorDirName = fmt.Sprintf("__%v_mirror", libName)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: go run %v_watcher.go <project_directory>", libName)
	}

	srcDir := filepath.Clean(os.Args[1])
	mirrorDir := filepath.Join(srcDir, mirrorDirName)

	fmt.Printf("Starting %v dev tool...\n", libName)
	fmt.Printf("Watching: %s\n", srcDir)
	fmt.Printf("Mirror:   %s\n", mirrorDir)

	// 1. First Pass: Mirror existing files
	err := mirrorTree(srcDir, mirrorDir)
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
					relPath, _ := filepath.Rel(srcDir, event.Name)
					dstPath := filepath.Join(mirrorDir, relPath)
					fmt.Printf("Change detected: %s\n", relPath)

					os.MkdirAll(filepath.Dir(dstPath), 0755)
					processGoFile(event.Name, dstPath)
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
			isTarget, pragma := hasPragma(fn.Doc)
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
			f.Decls = append(f.Decls, buildHintInitFunc(hints))
		}
		writeRuntimeLogger(filepath.Dir(dstPath), pkgName)
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return fmt.Errorf("format err on %s: %w", srcPath, err)
	}

	return os.WriteFile(dstPath, buf.Bytes(), 0644)
}

func hasPragma(doc *ast.CommentGroup) (bool, string) {
	if doc == nil {
		return false, ""
	}

	for _, c := range doc.List {
		if !strings.Contains(c.Text, "//"+libName) {
			continue
		}

		if strings.Contains(c.Text, ":pure") {
			return true, "pure"
		}
		if strings.Contains(c.Text, ":deterministic") {
			return true, "deterministic"
		}
	}

	return false, ""
}

func transformFunc(fn *ast.FuncDecl, funcName string) {
	var outNames []string
	if fn.Type.Results != nil {
		idx := 0
		for _, field := range fn.Type.Results.List {
			if len(field.Names) == 0 {
				// Name unnamed returns so we can capture them with defer
				name := ast.NewIdent(fmt.Sprintf("_%v_ret%d", libName, idx))
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

	var inNames []string
	if fn.Type.Params != nil {
		for _, field := range fn.Type.Params.List {
			for _, name := range field.Names {
				inNames = append(inNames, name.Name)
			}
		}
	}

	// 1. Rewrite assignments in the body
	rewriteBlocks(fn.Body, funcName)

	// 2. Prepend Input and Output loggers to body
	var newStmts []ast.Stmt

	// Output (Deferred inside a closure so it evaluates variables upon return, not immediately)
	deferStr := fmt.Sprintf(`defer func() { __%v_LogOutput("%s"`, libName, funcName)
	if len(outNames) > 0 {
		deferStr += ", " + strings.Join(outNames, ", ")
	}
	deferStr += ") }()"
	newStmts = append(newStmts, parseStmt(deferStr))

	// Input (Immediate)
	inputStr := fmt.Sprintf(`__%v_LogInput("%s"`, libName, funcName)
	if len(inNames) > 0 {
		inputStr += ", " + strings.Join(inNames, ", ")
	}
	inputStr += ")"
	newStmts = append(newStmts, parseStmt(inputStr))

	fn.Body.List = append(newStmts, fn.Body.List...)
}

func rewriteBlocks(node ast.Node, funcName string) {
	ast.Inspect(node, func(n ast.Node) bool {
		if block, ok := n.(*ast.BlockStmt); ok {
			var newList []ast.Stmt
			for _, stmt := range block.List {
				newList = append(newList, stmt)
				if as, ok := stmt.(*ast.AssignStmt); ok {
					for _, lhs := range as.Lhs {
						if id, ok := lhs.(*ast.Ident); ok && id.Name != "_" {
							stmtStr := fmt.Sprintf(`__%v_LogAssign("%s", "%s", %s)`, libName, funcName, id.Name, id.Name)
							newList = append(newList, parseStmt(stmtStr))
						}
					}
				}
			}
			block.List = newList
		}
		return true
	})
}

// parseStmt is a robust trick to generate valid AST statements without manually constructing
// a dozen nested ast.*Type structures.
func parseStmt(stmtStr string) ast.Stmt {
	src := "package p\nfunc f() {\n" + stmtStr + "\n}"
	f, err := parser.ParseFile(token.NewFileSet(), "", src, 0)
	if err != nil {
		panic(fmt.Errorf("failed to parse injected stmt: %s\n%v", stmtStr, err))
	}
	return f.Decls[0].(*ast.FuncDecl).Body.List[0]
}

func buildHintInitFunc(hints []string) *ast.FuncDecl {
	var stmts []string
	for _, h := range hints {
		parts := strings.SplitN(h, "|", 2)
		stmts = append(stmts, fmt.Sprintf(`__%v_LogHint("%s", "%s")`, libName, parts[0], parts[1]))
	}
	src := fmt.Sprintf("package p\nfunc init() {\n%s\n}", strings.Join(stmts, "\n"))
	f, _ := parser.ParseFile(token.NewFileSet(), "", src, 0)
	return f.Decls[0].(*ast.FuncDecl)
}

// writeRuntimeLogger drops an isolated runtime logger file into the target package so that
// transformed files can log inputs and outputs without cyclic dependency or import issues.
func writeRuntimeLogger(dir, pkgName string) {
	content := strings.ReplaceAll(fmt.Sprintf(loggerTmpl, pkgName), "{{libName}}", libName)
	// REMOVED the "__" prefix so the compiler includes this file
	_ = os.WriteFile(filepath.Join(dir, libName+"_logger_gen.go"), []byte(content), 0644)
}

const loggerTmpl = `// Code generated by {{libName}} dev tool. DO NOT EDIT.
package %s

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

var (
	__{{libName}}_out  *os.File
	__{{libName}}_once sync.Once
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
		parts = append(parts, fmt.Sprintf("%%#v", arg)) 
	}
	__{{libName}}_Log("%%s", strings.Join(parts, "|"))
}

func __{{libName}}_LogOutput(funcName string, args ...any) {
	parts := []string{"output", funcName}
	for _, arg := range args {
		parts = append(parts, fmt.Sprintf("%%#v", arg))
	}
	__{{libName}}_Log("%%s", strings.Join(parts, "|"))
}

func __{{libName}}_LogAssign(funcName, varName string, val any) {
	__{{libName}}_Log("assign|%%s|%%s|%%#v", funcName, varName, val)
}
`
