package main

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"testing"
)

func TestHasPragma(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   string
	}{
		{name: "pure", source: "//bona:pure", want: "pure"},
		{name: "deterministic", source: "//bona:deterministic", want: "deterministic"},
		{name: "other package", source: "//other:pure", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			file, err := parser.ParseFile(token.NewFileSet(), "", "package p\n"+test.source+"\nfunc f() {}", parser.ParseComments)
			if err != nil {
				t.Fatal(err)
			}

			got, pragma := hasPragma(file.Decls[0].(*ast.FuncDecl).Doc)
			if want := test.want != ""; got != want || pragma != test.want {
				t.Fatalf("hasPragma() = (%t, %q), want (%t, %q)", got, pragma, want, test.want)
			}
		})
	}
}

func TestParameterNames(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "", "package p\nfunc f(first, second int, third string) {}", 0)
	if err != nil {
		t.Fatal(err)
	}

	got := parameterNames(file.Decls[0].(*ast.FuncDecl).Type.Params)
	want := []string{"first", "second", "third"}
	if len(got) != len(want) {
		t.Fatalf("parameterNames() = %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("parameterNames() = %v, want %v", got, want)
		}
	}
}

func TestResultName(t *testing.T) {
	named := &ast.Field{Names: []*ast.Ident{ast.NewIdent("value")}}
	if got := resultName(named, 2); got != "value" {
		t.Fatalf("resultName(named) = %q, want %q", got, "value")
	}

	if got := resultName(&ast.Field{}, 2); got != "_bona_ret2" {
		t.Fatalf("resultName(unnamed) = %q, want %q", got, "_bona_ret2")
	}
}

func TestBuildHintInitFunc(t *testing.T) {
	decl := buildHintInitFunc([]string{"pure|utils.Sum", "deterministic|utils.Read"})
	var output bytes.Buffer
	if err := format.Node(&output, token.NewFileSet(), decl); err != nil {
		t.Fatal(err)
	}

	want := "func init() {\n\t__bona_LogHint(\"pure\", \"utils.Sum\")\n\t__bona_LogHint(\"deterministic\", \"utils.Read\")\n}"
	if output.String() != want {
		t.Fatalf("buildHintInitFunc() = %q, want %q", output.String(), want)
	}
}
