// Package osexit checks if os.Exit is called in the main function
package osexit

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

// OSExitAnalyzer analyzer
var OSExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "reports use of os.Exit in main function",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// находим функцию main
	mainFunc := findMainFunc(pass)
	if mainFunc == nil {
		return nil, nil
	}

	// проверяем каждый statement в функции main
	for _, stmt := range mainFunc.Body.List {
		// если это вызов os.Exit, то создаем отчет об ошибке
		if callExpr, ok := stmt.(*ast.ExprStmt); ok {
			if call, ok := callExpr.X.(*ast.CallExpr); ok {
				if fun, ok := call.Fun.(*ast.SelectorExpr); ok {
					if ident, ok := fun.X.(*ast.Ident); ok {
						if ident.Name == "os" && fun.Sel.Name == "Exit" {
							pass.Reportf(call.Pos(), "do not use os.Exit in main function")
						}
					}
				}
			}
		}
	}

	return nil, nil
}

func findMainFunc(pass *analysis.Pass) *ast.FuncDecl {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok {
				if fn.Name.Name == "main" {
					return fn
				}
			}
		}
	}
	return nil
}
