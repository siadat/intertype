package intertype

import (
	"fmt"
	"go/types"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/tools/go/analysis"
	"gopkg.in/yaml.v2"
)

type MultiChecker interface {
	MultiCheckAssign(spec *Constraints, lhsTyps, rhsTyps []types.Type) error
}

type Checker interface {
	CheckAssign(spec *Constraints, lhs, rhs types.Type) error
	CheckSwitchTypes(spec *Constraints, lhs types.Type, switchTypes []types.Type, hasDefaultCase bool) error
}

type YamlAnnotItem struct {
	Address []string    `yaml:"address"`
	Check   Constraints `yaml:"check"`
}

func ParseTypes(filename string) map[string][]YamlAnnotItem {
	result := make(map[string][]YamlAnnotItem)

	f, err := os.Open(filename)
	if err != nil {
		return result
		// panic(err)
	}
	defer f.Close()

	byts, err := ioutil.ReadAll(f)
	if err != nil {
		return result
		// panic(err)
	}

	err = yaml.Unmarshal(byts, &result)
	if err != nil {
		panic(err)
	}
	return result
}

func NewAnalyzer(analysisPass *analysis.Pass) *Analyzer {
	return &Analyzer{
		AnalysisPass: analysisPass,
		Passes:       DefaultPasses,
		Annots:       ParseTypes("./intertype.yaml"),
		MultiCheckers: []MultiChecker{
			&SameTypes{},
		},
		Checkers: []Checker{
			&IsPointer{},
			&IsInterface{},
			&IsChan{},
			&IsStruct{},
			&IsMap{},
			&IsSlice{},
			&IsFunc{},
			&FieldsChecker{},
			&OneOfChecker{},
			&NoneOfChecker{},
			&TagsChecker{},
		},
	}
}

type Analyzer struct {
	AnalysisPass  *analysis.Pass
	Passes        []Passer
	Annots        map[string][]YamlAnnotItem
	Checkers      []Checker
	MultiCheckers []MultiChecker
}

func (an *Analyzer) String() string {
	var builder strings.Builder

	for match, itemSlice := range an.Annots {
		fmt.Fprintf(&builder, "Match %q\n", match)
		for i := range itemSlice {
			fmt.Fprintf(&builder, "    Address: %q\n", itemSlice[i].Address)
			fmt.Fprintf(&builder, "    Check: %s\n", itemSlice[i].Check)
		}
	}
	return builder.String()
}

func (an *Analyzer) AddAsExt(t types.Type, annotation string) {
	// matcher := fmt.Sprintf("[] %s %s", t, t.Underlying())
	matcher := fmt.Sprintf("[] %s", t)
	constraint, err := parseIntertypeCommentLines([]string{annotation})
	if err != nil {
		panic(err)
	}
	if constraint == nil {
		return
	}

	checks := an.Annots[matcher]
	checks = append(checks, YamlAnnotItem{
		Address: []string{},
		Check:   *constraint,
	})
	an.Annots[matcher] = checks
}

func (an *Analyzer) CheckSwitchStmt(matcher string, lhsType types.Type, rhsType []types.Type, hasDefaultCase bool) error {
	annotItems, found := an.Annots[matcher]
	if !found {
		if *debugMode {
			fmt.Fprintf(os.Stderr, "NOTFOUND %q\n", matcher)
		}
		return nil
	}

	if *debugMode {
		fmt.Fprintf(os.Stderr, "FOUND %q\n", matcher)
	}

	for ii := range annotItems {
		err := an.CheckSwitchTypesSpec(lhsType, rhsType, hasDefaultCase, annotItems[ii].Check)
		if err != nil {
			return err
		}
	}
	return nil
}

func (an *Analyzer) CheckMatcher(matcher string, lhsType, rhsType types.Type) error {
	annotItems, found := an.Annots[matcher]
	if !found {
		if *debugMode {
			fmt.Fprintf(os.Stderr, "NOTFOUND %q\n", matcher)
		}
		return nil
	}
	if *debugMode {
		fmt.Fprintf(os.Stderr, "FOUND %q\n", matcher)
	}

	for ii := range annotItems {
		err := an.checkAssignWithSpec(lhsType, rhsType, annotItems[ii].Check)
		if err != nil {
			return err
		}
	}
	return nil
}

func (an *Analyzer) CheckMatcherMultiple(matcher string, lhsTypes, rhsTypes []types.Type) error {
	annotItems, found := an.Annots[matcher]
	if !found {
		if *debugMode {
			fmt.Fprintf(os.Stderr, "NOTFOUND %q\n", matcher)
		}
		return nil
	}

	if *debugMode {
		fmt.Fprintf(os.Stderr, "FOUND %q\n", matcher)
	}

	for ii := range annotItems {
		err := an.checkAssignWithSpecMultiple(lhsTypes, rhsTypes, annotItems[ii].Check)
		if err != nil {
			return err
		}
	}
	return nil
}

func (an *Analyzer) checkAssignWithSpecMultiple(annotatedTypes, dynTypes []types.Type, spec Constraints) error {
	for _, ch := range an.MultiCheckers {
		if err := ch.MultiCheckAssign(&spec, annotatedTypes, dynTypes); err != nil {
			return fmt.Errorf("%v", err)
		}
	}
	return nil
}

func (an *Analyzer) checkAssignWithSpec(lhs, rhs types.Type, spec Constraints) error {
	for _, ch := range an.Checkers {
		if err := ch.CheckAssign(&spec, lhs, rhs); err != nil {
			return fmt.Errorf("%v", err)
		}
	}
	return nil
}

func (an *Analyzer) CheckSwitchTypesSpec(lhs types.Type, switchTypes []types.Type, hasDefaultCase bool, spec Constraints) error {
	for _, ch := range an.Checkers {
		if err := ch.CheckSwitchTypes(&spec, lhs, switchTypes, hasDefaultCase); err != nil {
			return fmt.Errorf("%v", err)
		}
	}
	return nil
}
