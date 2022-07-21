package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/wule61/macro"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/packages"
	"os"
	"regexp"
	"strings"
	"text/template"
)

func main() {
	// src is the input for which we create the AST that we
	// are going to manipulate.

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedSyntax,
		Dir:  "./",
	}, "./...")
	if err != nil {
		panic(err)
	}

	for _, pkg := range pkgs {

		goFiles := pkg.GoFiles
		spew.Dump("go files: ", goFiles)
		for i, fl := range pkg.Syntax {
			spew.Dump(fl.Imports)
			if strings.Contains(goFiles[i], "_macro") || strings.Contains(goFiles[i], "main.go") {
				continue
			}
			pkgName := fl.Name.Name
			for _, d := range fl.Decls {
				switch spec := d.(type) {
				case *ast.FuncDecl:
					spew.Println("FuncDecl", spec.Name.Name)
				case *ast.GenDecl:
					if spec.Tok == token.TYPE {

						spew.Println("len:", len(spec.Specs))

						for _, t := range spec.Specs {
							tp := t.(*ast.TypeSpec)

							switch typo := tp.Type.(type) {
							case *ast.Ident:
								spew.Printf("typo %#v\n", typo)
							case *ast.StructType:
								parserStruct(tp, typo.Fields)
							default:
							}

							if spec.Doc != nil {
								lines := spec.Doc.List
								for _, c := range lines {
									fmt.Println("  token type:", c.Text)
								}
							}
						}
					}
				}
			}
			t, _ := template.New("test").Parse(macro.Tpl)
			f := createMacroFile(goFiles[i])
			t.Execute(f, map[string]interface{}{"Pkg": pkgName})
			f.Close()
		}
	}
}

type Visitor struct {
}

func (v *Visitor) Visit(node ast.Node) ast.Visitor {

	switch spec := node.(type) {
	case *ast.File:
		fmt.Println("node: ", spec.Scope.String())
	case *ast.GenDecl:
		fmt.Println("GenDecl: ")
		if spec.Tok == token.VAR {
			// lines := spec.Doc.List
			// for _, c := range lines {
			//  fmt.Println("  token var:", c.Text)
			// }
		}

		if spec.Tok == token.CONST {
			// lines := spec.Doc.List
			// for _, c := range lines {
			//  fmt.Println("  token const:", c.Text)
			// }
		}

		if spec.Tok == token.TYPE {

			fmt.Println("len:", len(spec.Specs))
			i := spec.Specs[0].(*ast.TypeSpec)
			fmt.Printf("%+v\n", i)
			switch typo := i.Type.(type) {
			case *ast.Ident:
				fmt.Printf("typo %#v\n", typo)
			case *ast.StructType:

				for _, field := range typo.Fields.List {
					fmt.Printf("%#v\n", field.Type)
					fmt.Println("len(field.Names) =", len(field.Names))
					fmt.Printf("typo %#v\n", field.Names[0].Name)
				}

			default:
			}

			lines := spec.Doc.List
			for _, c := range lines {
				fmt.Println("  token type:", c.Text)
			}
		}

	case *ast.InterfaceType:
		//fmt.Println("InterfaceType: ", spec.Methods.List)
	case *ast.TypeSpec:
		//fmt.Println("TypeSpec: ", spec.Name)
	}

	return v
}

func createMacroFile(filePath string) *os.File {

	arr := strings.Split(filePath, ".")
	fileName := strings.Join([]string{arr[0], "_macro.", arr[1]}, "")
	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}

	return f
}

type Field struct {
	FieldName string
	FieldType string
	FieldTag  string
}

type Struct struct {
	Name        string
	Fields      []Field
	Annotations []string
}

func parserStruct(typeSpec *ast.TypeSpec, fields *ast.FieldList) *Struct {

	s := &Struct{
		Name: typeSpec.Name.Name,
	}

	var buf strings.Builder
	if typeSpec.Doc != nil {
		for _, c := range typeSpec.Doc.List {
			buf.WriteString(c.Text)
			buf.WriteByte(' ')
		}
	}

	reg1 := regexp.MustCompile(`@[a-z/A-Z]+`)
	if reg1 == nil {
		fmt.Println("regexp err")
		return nil
	}

	result1 := reg1.FindAllStringSubmatch(buf.String(), -1)
	for _, v := range result1 {
		s.Annotations = append(s.Annotations, v...)
	}

	for _, field := range fields.List {
		fType := field.Type
		for _, fieldName := range field.Names {
			var fieldType string
			switch ft := fType.(type) {
			case *ast.Ident:
				fieldType = ft.Name
			case *ast.StarExpr:
				se := ft.X.(*ast.SelectorExpr)
				fieldType = strings.Join([]string{"*", se.X.(*ast.Ident).Name, ".", se.Sel.Name}, "")
			case *ast.SelectorExpr:
				fieldType = strings.Join([]string{ft.X.(*ast.Ident).Name, ".", ft.Sel.Name}, "")
			case *ast.MapType:
				fieldType = fmt.Sprintf("map[%s]%s", ft.Key, ft.Value)
			case *ast.ArrayType:
				fieldType = fmt.Sprintf("[]%s", ft.Elt.(*ast.Ident).Name)
			default:
				expr := fType.(*ast.StarExpr)
				fmt.Println(expr)
			}

			var fieldTag string
			if field.Tag != nil {
				fieldTag = field.Tag.Value
			}

			s.Fields = append(s.Fields, Field{
				FieldName: fieldName.Name,
				FieldType: fieldType,
				FieldTag:  fieldTag,
			})
		}
	}
	spew.Dump(s)
	return s
}
