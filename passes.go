package intertype

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/types/typeutil"
)

type Passer interface {
	Pass(*Analyzer, *types.Info, *token.FileSet, ast.Node, *ast.File)
}

var DefaultPasses = []Passer{
	ExtCallExpr{},
	ExtAssignStmt{},
	ExtValueSpec{},
	ExtReturnStmt{},
	ExtSwitchStmt{},
	ExtCompositeLitStruct{},
	ExtCompositeLitMap{},
	ExtIndexExpr{},
	ExtSendStmt{},
}

type ExtCallExpr struct{}
type ExtAssignStmt struct{}
type ExtValueSpec struct{}
type ExtReturnStmt struct{}
type ExtSwitchStmt struct{}
type ExtCompositeLitStruct struct{}
type ExtCompositeLitMap struct{}
type ExtIndexExpr struct{}
type ExtSendStmt struct{}

func (an *Analyzer) logError(fset *token.FileSet, pos token.Pos, err error) {
	an.AnalysisPass.Reportf(pos, "%v", err)
	// fmt.Printf("%v %v\n",
	// 	Path(fset.Position(pos)),
	// 	err,
	// )
}

func (ExtCompositeLitStruct) Pass(analyzer *Analyzer, typesInfo *types.Info, fset *token.FileSet, node ast.Node, f *ast.File) {
	switch node := node.(type) {
	case *ast.CompositeLit:
		typ := typesInfo.TypeOf(node.Type).Underlying()

		switch typ := typ.(type) {
		case *types.Struct:
			if len(node.Elts) == 0 {
				return
			}

			for i := 0; i < typ.NumFields(); i++ {
				field := typ.Field(i)

				rhsType := valueTypeForStructLit(typesInfo, field, i, node.Elts)
				if rhsType == nil {
					continue
				}

				// matcher := fmt.Sprintf("[] (%s).%s %s",
				// 	typesInfo.TypeOf(node.Type),
				// 	field.Name(),
				// 	field.Type(),
				// )
				matcher := fmt.Sprintf("[] (%s).%s",
					typesInfo.TypeOf(node.Type),
					field.Name(),
					// field.Type(),
				)

				if err := analyzer.CheckMatcher(matcher, field.Type(), rhsType); err != nil {
					analyzer.logError(fset, node.Pos(), err)
				}

			}
		}
	}
}

func (ExtIndexExpr) Pass(analyzer *Analyzer, typesInfo *types.Info, fset *token.FileSet, node ast.Node, f *ast.File) {
	switch node := node.(type) {
	case *ast.IndexExpr:
		// An index may be intertyped in two ways:
		//
		//   * the key is intertyped
		//   * the map's key is intertyped
		//
		// These two have different matchers.
		//
		//     [] pkg.TName interface{}
		//     [Key] pkg.TName map[interface{}]xxx
		//
		// However, both of these the check the type
		// of the index's Key

		wholeType := typesInfo.TypeOf(node.X).Underlying()
		rhsType := typesInfo.TypeOf(node.Index)
		lhsType := wholeType.(*types.Map).Key()

		{
			// matcher := fmt.Sprintf("[] %s %s",
			// 	lhsType,
			// 	lhsType.Underlying(),
			// )
			matcher := fmt.Sprintf("[] %s",
				lhsType,
				// lhsType.Underlying(),
			)
			if err := analyzer.CheckMatcher(matcher, lhsType, rhsType); err != nil {
				analyzer.logError(fset, node.Index.Pos(), err)
			}
		}

		{
			// matcher := fmt.Sprintf("[Key] %s %s",
			// 	typesInfo.TypeOf(node.X),
			// 	typesInfo.TypeOf(node.X).Underlying(),
			// )
			matcher := fmt.Sprintf("[Key] %s",
				typesInfo.TypeOf(node.X),
				// typesInfo.TypeOf(node.X).Underlying(),
			)
			if err := analyzer.CheckMatcher(matcher, lhsType, rhsType); err != nil {
				analyzer.logError(fset, node.Index.Pos(), err)
			}
		}
	}
}

func (ExtCompositeLitMap) Pass(analyzer *Analyzer, typesInfo *types.Info, fset *token.FileSet, node ast.Node, f *ast.File) {
	switch node := node.(type) {
	case *ast.CompositeLit:
		typ := typesInfo.TypeOf(node.Type)

		switch typp := typ.Underlying().(type) {
		case *types.Map:
			if len(node.Elts) == 0 {
				return
			}

			for jj := range node.Elts {
				{
					rhs := node.Elts[jj].(*ast.KeyValueExpr).Key
					rhsTyp := typesInfo.TypeOf(rhs)

					// matcher := fmt.Sprintf("[Key] %s %s", typ, typ.Underlying())
					matcher := fmt.Sprintf("[Key] %s", typ)
					if err := analyzer.CheckMatcher(matcher, typp.Key(), rhsTyp); err != nil {
						analyzer.logError(fset, rhs.Pos(), err)
					}
				}

				{
					rhs := node.Elts[jj].(*ast.KeyValueExpr).Value
					rhsTyp := typesInfo.TypeOf(rhs)

					// matcher := fmt.Sprintf("[Elem] %s %s", typ, typ.Underlying())
					matcher := fmt.Sprintf("[Elem] %s", typ)
					if err := analyzer.CheckMatcher(matcher, typp.Key(), rhsTyp); err != nil {
						analyzer.logError(fset, rhs.Pos(), err)
					}
				}
			}
		}
	}
}

func (ExtSwitchStmt) Pass(analyzer *Analyzer, typesInfo *types.Info, fset *token.FileSet, node ast.Node, file *ast.File) {
	switch node := node.(type) {
	case *ast.TypeSwitchStmt:
		expr := TypeSwitchExpr(node)
		lhsTyp := typesInfo.TypeOf(expr)
		rhsTyps, hasDefaultCase := TypesAssertedInSwitch(typesInfo, node) // TODO hasDefaultCase

		// for i := range rhsTyps {
		// matcher := fmt.Sprintf("[] %s %s", lhsTyp, lhsTyp.Underlying())
		matcher := fmt.Sprintf("[] %s", lhsTyp)
		// if err := analyzer.CheckMatcher(matcher, lhsTyp, rhsTyps[i]); err != nil {
		// 	fmt.Printf("%v %v\n",
		// 		Path(fset.Position(node.Pos())),
		// 		err,
		// 	)
		// }
		if err := analyzer.CheckSwitchStmt(matcher, lhsTyp, rhsTyps, hasDefaultCase); err != nil {
			analyzer.logError(fset, node.Pos(), err)
		}
		// }

	case *ast.TypeAssertExpr:
		rhs := node.Type
		if rhs == nil {
			// it is a switch case, which is already handled above
			break
		}
		rhsTyp := typesInfo.TypeOf(rhs)

		expr := node.X
		lhsTyp := typesInfo.TypeOf(expr)

		// matcher := fmt.Sprintf("[] %s %s", lhsTyp, lhsTyp.Underlying())
		matcher := fmt.Sprintf("[] %s", lhsTyp)
		if err := analyzer.CheckMatcher(matcher, lhsTyp, rhsTyp); err != nil {
			analyzer.logError(fset, node.Pos(), err)
		}
	}
}

func (ExtReturnStmt) Pass(analyzer *Analyzer, typesInfo *types.Info, fset *token.FileSet, node ast.Node, file *ast.File) {
	switch node := node.(type) {
	case *ast.ReturnStmt:

		if node.Results == nil {
			break
		}

		var rhsTyps []types.Type
		var lhsTyps []types.Type

		for i := range node.Results {
			typ := typesInfo.TypeOf(node.Results[i])
			rhsTyps = append(rhsTyps, typ)
		}

		path, _ := astutil.PathEnclosingInterval(file, node.Pos(), node.Pos())
		var ft *ast.FuncType
		var ftt *types.Func
	Q:
		for i := range path {
			switch funcDeclOrLit := path[i].(type) {
			case *ast.FuncLit:
				ft = funcDeclOrLit.Type
			case *ast.FuncDecl:
				ft = funcDeclOrLit.Type
				ftt = typesInfo.Defs[funcDeclOrLit.Name].(*types.Func)
			}

			if ft == nil {
				continue
			}

			for i := range ft.Results.List {
				typ := typesInfo.TypeOf(ft.Results.List[i].Type)
				lhsTyps = append(lhsTyps, typ)
			}
			break Q
		}

		if len(rhsTyps) == 1 {
			tuple, ok := rhsTyps[0].(*types.Tuple)
			if ok {
				rhsTyps = make([]types.Type, tuple.Len())
				for i := 0; i < tuple.Len(); i++ {
					rhsTyps[i] = tuple.At(i).Type()
				}
			}
		}

		for i := range rhsTyps {
			// TODO what if returning a function call that returns a tuple

			// matcher := fmt.Sprintf("[] %s %s", lhsTyps[i], lhsTyps[i].Underlying())
			matcher := fmt.Sprintf("[] %s", lhsTyps[i])
			if err := analyzer.CheckMatcher(matcher, lhsTyps[i], rhsTyps[i]); err != nil {
				analyzer.logError(fset, node.Pos(), err)
			}

			if ftt != nil {
				// matcher := fmt.Sprintf("[Returns, %d] %s %+v",
				// 	i,
				// 	ftt.FullName(),
				// 	ftt.Type(),
				// )
				matcher := fmt.Sprintf("[Returns, %d] %s",
					i,
					ftt.FullName(),
					// ftt.Type(),
				)
				if err := analyzer.CheckMatcher(matcher, lhsTyps[i], rhsTyps[i]); err != nil {
					analyzer.logError(fset, node.Pos(), err)
				}
			}
		}
	}
}

func (ExtValueSpec) Pass(analyzer *Analyzer, typesInfo *types.Info, fset *token.FileSet, node ast.Node, f *ast.File) {
	switch node := node.(type) {
	case *ast.ValueSpec:
		lhsTyp := typesInfo.TypeOf(node.Type)
		if lhsTyp == nil {
			return
		}
		var rhsTyps []types.Type
		for i := range node.Values {
			rhsTyps = append(rhsTyps, typesInfo.TypeOf(node.Values[i]))
		}

		// matcher := fmt.Sprintf("[] %s %s", lhsTyp, lhsTyp.Underlying())
		matcher := fmt.Sprintf("[] %s", lhsTyp)

		for i := range rhsTyps {
			if err := analyzer.CheckMatcher(matcher, lhsTyp, rhsTyps[i]); err != nil {
				analyzer.logError(fset, node.Pos(), err)
			}
		}
	}
}

func (ExtAssignStmt) Pass(analyzer *Analyzer, typesInfo *types.Info, fset *token.FileSet, node ast.Node, f *ast.File) {
	switch node := node.(type) {
	case *ast.AssignStmt:

		lhsTypes, valueTypes := expandAssignStmt(typesInfo, node)

		for i := range lhsTypes {
			lhsTyp := lhsTypes[i]
			if lhsTyp == nil {
				continue
			}

			switch xxLhsi := node.Lhs[i].(type) {
			case *ast.SelectorExpr:
				{
					lhsTyp := typesInfo.TypeOf(xxLhsi.Sel)

					// matcher := fmt.Sprintf("[] %s %s",
					// 	lhsTyp,
					// 	lhsTyp.Underlying(),
					// )
					matcher := fmt.Sprintf("[] %s",
						lhsTyp,
						// lhsTyp.Underlying(),
					)

					if err := analyzer.CheckMatcher(matcher, lhsTyp, valueTypes[i]); err != nil {
						analyzer.logError(fset, node.Pos(), err)
					}
				}
				{
					lhsTyp := typesInfo.TypeOf(xxLhsi.X)

					// matcher := fmt.Sprintf("[] (%s).%s %s",
					// 	lhsTyp,
					// 	xxLhsi.Sel,
					// 	typesInfo.TypeOf(xxLhsi.Sel),
					// )
					matcher := fmt.Sprintf("[] (%s).%s",
						lhsTyp,
						xxLhsi.Sel,
						// typesInfo.TypeOf(xxLhsi.Sel),
					)
					if err := analyzer.CheckMatcher(matcher, lhsTyp, valueTypes[i]); err != nil {
						analyzer.logError(fset, node.Pos(), err)
					}
				}
			default:
				// matcher := fmt.Sprintf("[] %s %s",
				// 	lhsTyp,
				// 	lhsTyp.Underlying(),
				// )
				matcher := fmt.Sprintf("[] %s",
					lhsTyp,
					// lhsTyp.Underlying(),
				)

				if err := analyzer.CheckMatcher(matcher, lhsTyp, valueTypes[i]); err != nil {
					analyzer.logError(fset, node.Pos(), err)
				}
			}
		}
	}
}

func (ExtCallExpr) Pass(analyzer *Analyzer, typesInfo *types.Info, fset *token.FileSet, node ast.Node, f *ast.File) {

	switch node := node.(type) {
	case *ast.CallExpr:
		funType := typesInfo.TypeOf(node.Fun)
		sig, isSig := funType.(*types.Signature)
		if !isSig {
			// it is a conversion like T1(expr)
			// it is not a signature,
			// eg it might be a []byte(...)
			// which is a ast.CallExpr but not a types.Signature
			// If T1 is annotated, we need to check

			// TODO: handle conversions of types in analyzer.Annots

			// matcher := fmt.Sprintf("[] %s %s", funType, funType.Underlying())
			matcher := fmt.Sprintf("[] %s", funType)
			if err := analyzer.CheckMatcher(matcher, funType, typesInfo.TypeOf(node.Args[0])); err != nil {
				analyzer.logError(fset, node.Lparen, err)
			}

			return
		}

		{
			params := sig.Params()
			var paramsVars []*types.Var

			for i := 0; i < params.Len(); i++ {
				paramsVars = append(paramsVars, params.At(i))
			}

			var variadicTyp types.Type
			var variadicIdx int

			if sig.Variadic() {
				variadicIdx = params.Len() - 1
				variadicTyp = params.At(variadicIdx).Type().(*types.Slice).Elem()
			}

			var rhsTyps []types.Type
			for i := range node.Args {
				rhsTyps = append(rhsTyps, typesInfo.Types[node.Args[i]].Type)
			}

			// DONE: what if a function that returns a tuple is passed to a
			//       a) non-variadic function
			//       b) variadic function

			if len(rhsTyps) == 1 {
				tuple, ok := rhsTyps[0].(*types.Tuple)
				if ok {
					rhsTyps = make([]types.Type, tuple.Len())
					for i := 0; i < tuple.Len(); i++ {
						rhsTyps[i] = tuple.At(i).Type()
					}
				}
			}

			for i := range rhsTyps {
				var lhsTyp types.Type
				if !sig.Variadic() {
					lhsTyp = paramsVars[i].Type()
				} else {
					if i >= variadicIdx {
						lhsTyp = variadicTyp
					} else {
						continue // no variadic rhs
					}
				}

				// matcher := fmt.Sprintf("[] %s %s", lhsTyp, lhsTyp.Underlying())
				matcher := fmt.Sprintf("[] %s", lhsTyp)
				if err := analyzer.CheckMatcher(matcher, lhsTyp, rhsTyps[i]); err != nil {
					analyzer.logError(fset, node.Lparen, err)
				}
			}

			fn, hasCallee := typeutil.Callee(typesInfo, node).(*types.Func)
			if fn != nil && hasCallee {

				var lhsTyps []types.Type

				for i := range rhsTyps {
					var lhsTyp types.Type
					if !sig.Variadic() {
						lhsTyp = paramsVars[i].Type()
					} else {
						if i >= variadicIdx {
							lhsTyp = variadicTyp
						}
					}
					lhsTyps = append(lhsTyps, lhsTyp)
				}

				// matcher := fmt.Sprintf("[] %s %s",
				// 	fn.FullName(),
				// 	fn.Type(),
				// )
				matcher := fmt.Sprintf("[] %s",
					fn.FullName(),
					// fn.Type(),
				)
				if err := analyzer.CheckMatcherMultiple(matcher, lhsTyps, rhsTyps); err != nil {
					analyzer.logError(fset, node.Lparen, err)
				}

				for i := range rhsTyps {
					lhsTyp := lhsTyps[i]

					// matcher := fmt.Sprintf("[Params, %d] %s %s",
					// 	i,
					// 	fn.FullName(),
					// 	fn.Type(),
					// )
					matcher := fmt.Sprintf("[Params, %d] %s",
						i,
						fn.FullName(),
						// fn.Type(),
					)
					// fmt.Printf(">>> [debug passes.go:454] matcher: %+v\n", matcher)
					if err := analyzer.CheckMatcher(matcher, lhsTyp, rhsTyps[i]); err != nil {
						analyzer.logError(fset, node.Lparen, err)
					}
				}
			}
		}
	}
}

func (ExtSendStmt) Pass(analyzer *Analyzer, typesInfo *types.Info, fset *token.FileSet, node ast.Node, f *ast.File) {
	switch node := node.(type) {
	case *ast.SendStmt:
		lhsTyp := typesInfo.Types[node.Chan].Type.(*types.Chan).Elem()
		rhsTyp := typesInfo.Types[node.Value].Type

		// matcher := fmt.Sprintf("[] %s %s",
		// 	lhsTyp,
		// 	lhsTyp.Underlying(),
		// )
		matcher := fmt.Sprintf("[] %s",
			lhsTyp,
			// lhsTyp.Underlying(),
		)
		if err := analyzer.CheckMatcher(matcher, lhsTyp, rhsTyp); err != nil {
			analyzer.logError(fset, node.Pos(), err)
		}

	}
}

func valueTypeForStructLit(typesInfo *types.Info, field *types.Var, fieldIdx int, elts []ast.Expr) types.Type {
	if len(elts) == 0 {
		return nil
	}

	_, isKeyValue := elts[0].(*ast.KeyValueExpr)

	if isKeyValue {
		for ii := range elts {
			xxElt := elts[ii]
			keyVal := xxElt.(*ast.KeyValueExpr)
			if field.Name() == keyVal.Key.(*ast.Ident).Name {
				return typesInfo.TypeOf(keyVal.Value)
			}
		}
	} else {
		xxElt := elts[fieldIdx]
		return typesInfo.TypeOf(xxElt)
	}

	return nil
}

func TypeSwitchExpr(sw *ast.TypeSwitchStmt) ast.Expr {
	switch yy := sw.Assign.(type) {
	case *ast.AssignStmt:
		return yy.Rhs[0].(*ast.TypeAssertExpr).X
	case *ast.ExprStmt:
		return yy.X.(*ast.TypeAssertExpr).X
	}
	return nil
}

func TypesAssertedInSwitch(typesInfo *types.Info, sw *ast.TypeSwitchStmt) ([]types.Type, bool) {
	var typs []types.Type
	hasDefaultCase := false

	if sw.Body == nil {
		return []types.Type{}, false
	}

	for i := range sw.Body.List {
		list := sw.Body.List[i].(*ast.CaseClause).List
		if list == nil {
			hasDefaultCase = true
		}
		for j := range list {
			expr := list[j]
			typ := typesInfo.TypeOf(expr)
			typs = append(typs, typ)
		}
	}
	return typs, hasDefaultCase
}

func Path(p token.Position) string {
	if p == (token.Position{}) {
		return "builtin"
	}
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	p.Filename, err = filepath.Rel(pwd, p.Filename)
	if err != nil {
		panic(err)
	}

	return p.String()
}

func expandAssignStmt(typesInfo *types.Info, n *ast.AssignStmt) (lhsTypes, valueTypes []types.Type) {
	var rhsTypes []types.Type

	for i := range n.Lhs {
		lhsTypes = append(lhsTypes, typesInfo.TypeOf(n.Lhs[i]))
	}

	for i := range n.Rhs {
		rhsTypes = append(rhsTypes, typesInfo.TypeOf(n.Rhs[i]))
	}

	valueTypes = make([]types.Type, len(n.Lhs))

	if len(lhsTypes) > 1 && len(rhsTypes) == 1 {
		tuple := rhsTypes[0].(*types.Tuple)
		for i := 0; i < tuple.Len(); i++ {
			valueTypes[i] = tuple.At(i).Type()
		}
	} else if len(lhsTypes) == len(rhsTypes) {
		for i := range rhsTypes {
			valueTypes[i] = rhsTypes[i]
		}
	}

	return
}

func getTypeAndExpr(typParent types.Type, exprParent ast.Expr, address []string) (typTyp types.Type, valType ast.Expr) {
	if len(address) == 0 {
		return typParent, exprParent
	}

	switch address[0] {
	case "Params":
		idx, err := strconv.Atoi(address[1])
		if err != nil {
			panic(fmt.Sprintf("want int index, got %q: %v", address[1], err))
		}

		t := typParent.(*types.Signature).Params().At(idx).Type()
		e := exprParent.(*ast.CallExpr).Args[idx]

		return getTypeAndExpr(t, e, address[2:])
	// case "MapElem":
	// return getTypeAndExpr(thing.(*types.Map).Elem(), address[1:])
	default:
		panic(fmt.Sprintf("unsupported address %q", address[0]))
	}
}
