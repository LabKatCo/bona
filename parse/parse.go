package parse

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/labkatco/bona/constants"
)

//bona:pure
func HasPragma(doc *ast.CommentGroup) (bool, string) {
	if doc == nil {
		return false, ""
	}

	for _, c := range doc.List {
		if !strings.Contains(c.Text, "//"+constants.LibName) {
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

//bona:pure
func ParameterNames(params *ast.FieldList) []string {
	if params == nil {
		return nil
	}

	var names []string
	for _, field := range params.List {
		for _, name := range field.Names {
			names = append(names, name.Name)
		}
	}
	return names
}

//bona:pure
func ResultName(field *ast.Field, index int) string {
	if len(field.Names) > 0 {
		return field.Names[0].Name
	}
	return fmt.Sprintf("_%v_ret%d", constants.LibName, index)
}

//bona:pure
func AssignmentLog(name, funcName string) ast.Stmt {
	stmtStr := fmt.Sprintf(`__%v_LogAssign("%s", "%s", %s)`, constants.LibName, funcName, name, name)
	return ParseStmt(stmtStr)
}

// BuildMapAssignmentLog records an assignment to a map entry using the evaluated key.
func BuildMapAssignmentLog(mapName, keyName, mapExpr, funcName string) ast.Stmt {
	stmtStr := fmt.Sprintf(`__%v_LogMapAssign("%s", "%s", %s, %s)`, constants.LibName, funcName, mapName, keyName, mapExpr)
	return ParseStmt(stmtStr)
}

// ParseStmt is a robust trick to generate valid AST statements without manually constructing
// a dozen nested ast.*Type structures.
//
//bona:pure
func ParseStmt(stmtStr string) ast.Stmt {
	src := "package p\nfunc f() {\n" + stmtStr + "\n}"
	f, err := parser.ParseFile(token.NewFileSet(), "", src, 0)
	if err != nil {
		panic(fmt.Errorf("failed to parse injected stmt: %s\n%v", stmtStr, err))
	}
	return f.Decls[0].(*ast.FuncDecl).Body.List[0]
}

//bona:pure
func BuildHintInitFunc(hints []string) *ast.FuncDecl {
	var stmts []string
	for _, h := range hints {
		parts := strings.SplitN(h, "|", 2)
		stmts = append(stmts, fmt.Sprintf(`__%v_LogHint("%s", "%s")`, constants.LibName, parts[0], parts[1]))
	}
	src := fmt.Sprintf("package p\nfunc init() {\n%s\n}", strings.Join(stmts, "\n"))
	f, _ := parser.ParseFile(token.NewFileSet(), "", src, 0)
	return f.Decls[0].(*ast.FuncDecl)
}
