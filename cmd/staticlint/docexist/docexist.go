// Package docexist checks if lack of documentation
package docexist

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

// DocExistAnalyzer analyzer
var DocExistAnalyzer = &analysis.Analyzer{
	Name: "undocumentedExportedDecl",
	Doc:  "reports any exported declaration without documentation",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			switch decl := n.(type) {
			case *ast.GenDecl:
				if undocumentedGenDecl(decl) {
					pass.Reportf(decl.Pos(), "undocumented exported declaration")
				}
			}
			return true
		})
	}
	return nil, nil
}

// undocumentedGenDecl проверяет, что экспортированная декларация является недокументированной
func undocumentedGenDecl(decl *ast.GenDecl) bool {
	for _, spec := range decl.Specs {
		switch st := spec.(type) {
		case *ast.TypeSpec:
			if st.Name.IsExported() && decl.Doc == nil {
				return true
			}
		case *ast.ValueSpec:
			for _, name := range st.Names {
				if name.IsExported() && decl.Doc == nil {
					return true
				}
			}
		}
	}
	return false
}
