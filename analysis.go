package intertype

import (
	"flag"
	"fmt"
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/astutil"
)

const Doc = `intertype: type annotations for empty interface{}s
Author: Sina Siadat github.com/siadat
Homepage: github.com/siadat/intertype
`

var MyAnalyzer = &analysis.Analyzer{
	Name:             "intertype",
	Doc:              Doc,
	Requires:         []*analysis.Analyzer{inspect.Analyzer},
	RunDespiteErrors: true,
	Run:              run,
	Flags:            *flags,
}

var flags = flag.NewFlagSet("flags", flag.ExitOnError)
var debugMode = flags.Bool("d", false, "enable debug mode")

func run(pass *analysis.Pass) (interface{}, error) {
	analyzer := NewAnalyzer(pass)

	for _, f := range pass.Files {
		for _, cg := range f.Comments {
			for _, comment := range cg.List {
				path, _ := astutil.PathEnclosingInterval(f, comment.Slash, comment.Slash)
				if len(path) < 3 {
					continue
				}
				_, ok1 := path[0].(*ast.FieldList)
				ifaceNode, ok2 := path[1].(*ast.InterfaceType)
				typeSpecNode, ok3 := path[2].(*ast.TypeSpec)
				if !(ok1 && ok2 && ok3) {
					continue
				}

				if ifaceNode.Methods.NumFields() != 0 {
					continue
				}

				typeSpecType := pass.TypesInfo.Defs[typeSpecNode.Name].Type()
				// analyzer.Add(typeSpecType, comment.Text)
				analyzer.AddAsExt(typeSpecType, comment.Text)
			}
		}
	}

	// fmt.Println(analyzer)
	// fmt.Println("--------------")

	// DONE: use Underlying for the container? No.
	for _, f := range pass.Files {
		ast.Inspect(f, func(n ast.Node) bool {

			if n != nil {
				verbose := false
				if verbose {
					var pos token.Pos = n.Pos()
					fmt.Printf("ALL %v %T %#+v\n",
						Path(pass.Fset.Position(pos)),
						n,
						n,
					)
				}
			}

			for i := range analyzer.Passes {
				analyzer.Passes[i].Pass(analyzer, pass.TypesInfo, pass.Fset, n, f)
			}

			return true
		})
	}

	return nil, nil
}
