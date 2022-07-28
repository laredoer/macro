package main

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/wule61/macro"
	"go/ast"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/packages"
	"os"
	"reflect"
	"regexp"
	"strings"
)

// macroInterface holds the reflect.Type of macro.Annotator.
var macroInterface = reflect.TypeOf(struct{ macro.Annotator }{}).Field(0).Type

func main() {
	// src is the input for which we create the AST that we
	// are going to manipulate.

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedModule,
	}, "./src", macroInterface.PkgPath())
	if err != nil {
		panic(err)
	}

	macroPkg, pkg := pkgs[0], pkgs[1]
	if len(pkg.Errors) != 0 {
		panic(pkg.Errors[0])
	}
	if pkgs[0].PkgPath != macroInterface.PkgPath() {
		macroPkg, pkg = pkgs[1], pkgs[0]
	}
	var names []string
	iface := macroPkg.Types.Scope().Lookup(macroInterface.Name()).Type().Underlying().(*types.Interface)
	for k, v := range pkg.TypesInfo.Defs {
		typ, ok := v.(*types.TypeName)
		if !ok || !k.IsExported() || !types.Implements(typ.Type(), iface) {
			continue
		}
		spec, ok := k.Obj.Decl.(*ast.TypeSpec)
		if !ok {
			panic(fmt.Errorf("invalid declaration %T for %s", k.Obj.Decl, k.Name))
		}
		if _, ok := spec.Type.(*ast.StructType); !ok {
			panic(fmt.Errorf("invalid spec type %T for %s", spec.Type, k.Name))
		}
		names = append(names, k.Name)
	}

	fmt.Println(names)
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

func parserStruct(typeSpec *ast.TypeSpec, fields *ast.FieldList) *macro.Struct {

	s := &macro.Struct{
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

			s.Fields = append(s.Fields, macro.Field{
				FieldName: fieldName.Name,
				FieldType: fieldType,
				FieldTag:  fieldTag,
			})
		}
	}
	spew.Dump(s)
	return s
}
