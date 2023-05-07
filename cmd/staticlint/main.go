// Static analyzer.
//
// Usage:
//   ./cmd/shortener/main.go
//  ./cmd/staticlint/staticlint ./...

package main

import (
	"github.com/itksb/go-url-shortener/cmd/staticlint/docexist"
	"github.com/itksb/go-url-shortener/cmd/staticlint/osexit"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"honnef.co/go/tools/staticcheck"
	"strings"
)

func main() {
	analyzers := []*analysis.Analyzer{
		// стандартные статические анализаторы из пакета golang.org/x/tools/go/analysis/passes:
		buildtag.Analyzer,
		errorsas.Analyzer,
		printf.Analyzer,
		shift.Analyzer,
		shadow.Analyzer,
		bools.Analyzer,
		assign.Analyzer,

		// noexit analyzer
		osexit.OSExitAnalyzer,

		// custom check based on yandex check
		docexist.DocExistAnalyzer,
	}

	for _, analyzer := range staticcheck.Analyzers {
		if strings.HasPrefix(analyzer.Analyzer.Name, "CA") || strings.HasPrefix(analyzer.Analyzer.Name, "SA") {
			analyzers = append(analyzers, analyzer.Analyzer)
		}
	}

	multichecker.Main(analyzers...)
}
