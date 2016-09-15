package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"os"
)

func main() {
	var (
		write = flag.Bool("w", false, "write result to (source) file instead of stdout")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: structfix [flags] [path ...]\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	for _, path := range flag.Args() {
		switch dir, err := os.Stat(path); {
		case err != nil:
			panic(err)
		case dir.IsDir():
			processDir(path, printResult(*write, os.Stdout))
		default:
			processFile(path, printResult(*write, os.Stdout))
		}
	}

}

func printResult(write bool, out io.Writer) func(string, *token.FileSet, *ast.File) {
	return func(filename string, fset *token.FileSet, f *ast.File) {
		if write {
			fi, err := os.OpenFile(filename, os.O_TRUNC|os.O_RDWR, 0644)
			if err != nil {
				panic(err)
			}
			defer fi.Close()
			out = fi
		}
		printer.Fprint(out, fset, f)
	}
}

func processDir(path string, fn func(string, *token.FileSet, *ast.File)) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, path, nil, parser.AllErrors)
	if err != nil {
		panic(err)
	}

	for _, pkg := range pkgs {
		for fname, f := range pkg.Files {
			ast.Walk(&V{fset: fset}, f)
			fn(fname, fset, f)
		}
	}
}

func processFile(filename string, fn func(string, *token.FileSet, *ast.File)) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filename, nil, parser.AllErrors)
	if err != nil {
		panic(err)
	}

	ast.Walk(&V{fset: fset}, f)
	fn(filename, fset, f)
}

type V struct {
	fset *token.FileSet
}

func (v *V) Visit(n ast.Node) ast.Visitor {
	if n == nil {
		return v
	}
	switch n := n.(type) {
	case *ast.CompositeLit:
		id, ok := n.Type.(*ast.Ident)
		if !ok || id.Obj == nil || id.Obj.Decl == nil {
			return v
		}
		ts, ok := id.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			return v
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			return v
		}
		for _, el := range n.Elts {
			if kv, ok := el.(*ast.KeyValueExpr); ok {
				checkAndCorrect(kv, st, v.fset)
			}
		}
	}
	return v
}

func checkAndCorrect(kv *ast.KeyValueExpr, st *ast.StructType, fset *token.FileSet) {
	ki, ok := kv.Key.(*ast.Ident)
	if !ok {
		panic("key is not an Ident")
	}

	vcl, ok := kv.Value.(*ast.CompositeLit)
	if !ok {
		return
	}

	fld, ok := getField(st, ki.Name)
	if !ok {
		return
	}
	fst, ok := fld.Type.(*ast.StructType)
	if !ok {
		return
	}

	vcl.Type = fld.Type

	for _, el := range vcl.Elts {
		if subkv, ok := el.(*ast.KeyValueExpr); ok {
			checkAndCorrect(subkv, fst, fset)
		}
	}
}

func getField(st *ast.StructType, name string) (*ast.Field, bool) {
	if st.Fields == nil {
		return nil, false
	}

	for _, fld := range st.Fields.List {
		for _, fldName := range fld.Names {
			if fldName.Name == name {
				return fld, true
			}
		}
	}
	return nil, false
}
