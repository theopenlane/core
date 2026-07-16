package genhelpers

import (
	"fmt"
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/packages"
)

// LoadCommentMap parses the packages matched by the given patterns and returns their Go doc
// comments keyed by "pkgpath.TypeName" for type declarations and "pkgpath.TypeName.FieldName"
// for exported struct fields; the key format is shared by the jsonschema reflector comment map
// and the OpenAPI schema customizer so both generators source descriptions the same way
func LoadCommentMap(patterns ...string) (map[string]string, error) {
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedSyntax | packages.NeedCompiledGoFiles}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoadPackages, err)
	}

	comments := make(map[string]string)

	for _, pkg := range pkgs {
		for _, file := range pkg.Syntax {
			addFileComments(comments, pkg.PkgPath, file)
		}
	}

	return comments, nil
}

// addFileComments records the type and struct field doc comments of one parsed file
func addFileComments(comments map[string]string, pkgPath string, file *ast.File) {
	for _, decl := range file.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			typeKey := pkgPath + "." + typeSpec.Name.Name

			// a type inside a declaration group may carry its comment on the group
			txt := typeSpec.Doc.Text()
			if txt == "" {
				txt = genDecl.Doc.Text()
			}

			if txt = strings.TrimSpace(txt); txt != "" {
				comments[typeKey] = txt
			}

			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				addFieldComments(comments, typeKey, structType)
			}
		}
	}
}

// addFieldComments records the doc or trailing line comments of a struct's exported fields
func addFieldComments(comments map[string]string, typeKey string, structType *ast.StructType) {
	for _, field := range structType.Fields.List {
		txt := field.Doc.Text()
		if txt == "" {
			txt = field.Comment.Text()
		}

		txt = strings.TrimSpace(txt)
		if txt == "" {
			continue
		}

		for _, name := range field.Names {
			if name.IsExported() {
				comments[typeKey+"."+name.Name] = txt
			}
		}
	}
}
