package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"log"
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
			log.Fatalf("Error checking path %q: %v", path, err)
		case dir.IsDir():
			processDir(path, printResult(*write, os.Stdout))
		default:
			processFile(path, printResult(*write, os.Stdout))
		}
	}
}

type nopWriteCloser struct {
	io.Writer
}

func (n *nopWriteCloser) Close() error {
	return nil
}

func printResult(write bool, w io.Writer) func(string) io.WriteCloser {
	return func(filename string) io.WriteCloser {
		if write {
			f, err := os.OpenFile(filename, os.O_TRUNC|os.O_WRONLY, 0644)
			if err != nil {
				log.Fatalf("Error opening file %q: %v", filename, err)
			}
			return f
		}
		return &nopWriteCloser{w}
	}
}

func processDir(dir string, outFn func(string) io.WriteCloser) {
	fset := token.NewFileSet()
	pkgs, err := parser.ParseDir(fset, dir, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Error parsing dir %q: %v", dir, err)
	}

	for _, pkg := range pkgs {
		pkg, _ := ast.NewPackage(fset, pkg.Files, nil, nil)

		ast.Walk(&V{fset: fset}, pkg)
		for filename, file := range pkg.Files {
			func() {
				wc := outFn(filename)
				defer wc.Close()
				printer.Fprint(wc, fset, file)
			}()

		}
	}

}

func processFile(filename string, outFn func(string) io.WriteCloser) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		log.Fatalf("Error parsing file %q: %v", filename, err)
	}

	ast.Walk(&V{fset: fset}, file)

	wc := outFn(filename)
	defer wc.Close()
	printer.Fprint(wc, fset, file)
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

	fst, ok := getFieldStructType(st, ki.Name)
	if !ok {
		return
	}

	vcl.Type = fst

	for _, el := range vcl.Elts {
		if subkv, ok := el.(*ast.KeyValueExpr); ok {
			checkAndCorrect(subkv, fst, fset)
		}
	}
}

func getFieldStructType(st *ast.StructType, name string) (*ast.StructType, bool) {
	if st.Fields == nil {
		return nil, false
	}

	for _, fld := range st.Fields.List {
		for _, fldName := range fld.Names {
			if fldName.Name == name {
				if fst, ok := fld.Type.(*ast.StructType); ok {
					return fst, true
				}
				return nil, false
			}
		}
	}
	return nil, false
}
