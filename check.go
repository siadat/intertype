package intertype

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/token"
	"go/types"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type Address []string

type Constraints struct {
	OneOf        []string          `yaml:"OneOf,omitempty" json:"OneOf,omitempty"`
	NoneOf       []string          `yaml:"NoneOf,omitempty" json:"NoneOf,omitempty"`
	SameTypes    []Address         `yaml:"SameTypes,omitempty" json:"SameTypes,omitempty"`
	Fields       map[string]string `yaml:"Fields,omitempty" json:"Fields,omitempty"`
	Tags         []string          `yaml:"Tags,omitempty" json:"Tags,omitempty"`
	FieldsRegex  map[string]string `yaml:"FieldsRegex,omitempty" json:"FieldsRegex,omitempty"`
	IsInterface  bool              `yaml:"IsInterface,omitempty" json:"IsInterface,omitempty"`
	IsChan       bool              `yaml:"IsChan,omitempty" json:"IsChan,omitempty"`
	IsFunc       bool              `yaml:"IsFunc,omitempty" json:"IsFunc,omitempty"`
	IsSlice      bool              `yaml:"IsSlice,omitempty" json:"IsSlice,omitempty"`
	IsStruct     bool              `yaml:"IsStruct,omitempty" json:"IsStruct,omitempty"`
	IsMap        bool              `yaml:"IsMap,omitempty" json:"IsMap,omitempty"`
	IsPointer    bool              `yaml:"IsPointer,omitempty" json:"IsPointer,omitempty"`
	IsReference  bool              `yaml:"Reference,omitempty" json:"Reference,omitempty"`
	IsNotPointer bool              `yaml:"IsNotPointer,omitempty" json:"IsNotPointer,omitempty"`
}

func MustMarshalYaml(whatever interface{}) string {
	buf := bytes.NewBuffer(nil)
	enc := yaml.NewEncoder(buf)
	// enc.SetIndent("", "    ")
	err := enc.Encode(whatever)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func MustMarshalJson(whatever interface{}) string {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "    ")
	err := enc.Encode(whatever)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

func (c Constraints) String() string {
	return MustMarshalYaml(c)
}

type SameTypes struct{}

func (*SameTypes) MultiCheckAssign(spec *Constraints, lhsTyps, rhsTyps []types.Type) error {
	if len(spec.SameTypes) == 0 {
		return nil
	}

	var indexes []int

	for i := range spec.SameTypes {
		if spec.SameTypes[i][0] != "Params" {
			panic("only Params is supported")
		}
		idx, err := strconv.Atoi(spec.SameTypes[i][1])
		if err != nil {
			return err
		}
		indexes = append(indexes, idx)
	}

	if len(rhsTyps) == 0 {
		return nil
	}

	for i := range indexes {
		idx := indexes[i]
		if !types.Identical(rhsTyps[indexes[0]], rhsTyps[idx]) {
			return fmt.Errorf("expected same types, got %s != %s", rhsTyps[0], rhsTyps[idx])
		}
	}
	return nil
}

// ---
type IsPointer struct{}

func (ch *IsPointer) CheckSwitchTypes(spec *Constraints, lhs types.Type, switchTypes []types.Type, hasDefaultCase bool) error {
	if !spec.IsPointer {
		return nil
	}
	for i := range switchTypes {
		err := ch.CheckAssign(spec, lhs, switchTypes[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (ch *IsPointer) CheckAssign(spec *Constraints, lhs, rhs types.Type) error {
	if !spec.IsPointer {
		return nil
	}

	if _, ok := rhs.Underlying().(*types.Pointer); ok {
		return nil
	}

	return fmt.Errorf("expected a pointer, got %s", rhs)
}

// ---

type IsInterface struct{}

func (ch *IsInterface) CheckSwitchTypes(spec *Constraints, lhs types.Type, switchTypes []types.Type, hasDefaultCase bool) error {
	if !spec.IsInterface {
		return nil
	}
	for i := range switchTypes {
		err := ch.CheckAssign(spec, lhs, switchTypes[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (ch *IsInterface) CheckAssign(spec *Constraints, lhs, rhs types.Type) error {
	if !spec.IsInterface {
		return nil
	}

	if _, ok := rhs.Underlying().(*types.Interface); ok {
		return nil
	}

	return fmt.Errorf("expected an interface, got %s", rhs)
}

// --

type IsChan struct{}

func (ch *IsChan) CheckSwitchTypes(spec *Constraints, lhs types.Type, switchTypes []types.Type, hasDefaultCase bool) error {
	if !spec.IsChan {
		return nil
	}
	for i := range switchTypes {
		err := ch.CheckAssign(spec, lhs, switchTypes[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (ch *IsChan) CheckAssign(spec *Constraints, lhs, rhs types.Type) error {
	if !spec.IsChan {
		return nil
	}

	if _, ok := rhs.Underlying().(*types.Chan); ok {
		return nil
	}

	return fmt.Errorf("expected a channel, got %s", rhs)
}

//
type IsStruct struct{}

func (ch *IsStruct) CheckSwitchTypes(spec *Constraints, lhs types.Type, switchTypes []types.Type, hasDefaultCase bool) error {
	if !spec.IsStruct {
		return nil
	}
	for i := range switchTypes {
		err := ch.CheckAssign(spec, lhs, switchTypes[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (ch *IsStruct) CheckAssign(spec *Constraints, lhs, rhs types.Type) error {
	if !spec.IsStruct {
		return nil
	}

	if _, ok := rhs.Underlying().(*types.Struct); ok {
		return nil
	}

	return fmt.Errorf("expected a struct, got %s", rhs)
}

//

type IsMap struct{}

func (ch *IsMap) CheckSwitchTypes(spec *Constraints, lhs types.Type, switchTypes []types.Type, hasDefaultCase bool) error {
	if !spec.IsMap {
		return nil
	}
	for i := range switchTypes {
		err := ch.CheckAssign(spec, lhs, switchTypes[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (ch *IsMap) CheckAssign(spec *Constraints, lhs, rhs types.Type) error {
	if !spec.IsMap {
		return nil
	}

	if _, ok := rhs.Underlying().(*types.Map); ok {
		return nil
	}

	return fmt.Errorf("expected a map, got %s", rhs)
}

///

type IsSlice struct{}

func (ch *IsSlice) CheckSwitchTypes(spec *Constraints, lhs types.Type, switchTypes []types.Type, hasDefaultCase bool) error {
	if !spec.IsSlice {
		return nil
	}
	for i := range switchTypes {
		err := ch.CheckAssign(spec, lhs, switchTypes[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (ch *IsSlice) CheckAssign(spec *Constraints, lhs, rhs types.Type) error {
	if !spec.IsSlice {
		return nil
	}

	if _, ok := rhs.Underlying().(*types.Slice); ok {
		return nil
	}

	return fmt.Errorf("expected a slice, got %s", rhs)
}

// --

type IsFunc struct{}

func (ch *IsFunc) CheckSwitchTypes(spec *Constraints, lhs types.Type, switchTypes []types.Type, hasDefaultCase bool) error {
	if !spec.IsFunc {
		return nil
	}
	for i := range switchTypes {
		err := ch.CheckAssign(spec, lhs, switchTypes[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func (ch *IsFunc) CheckAssign(spec *Constraints, lhs, rhs types.Type) error {
	if !spec.IsFunc {
		return nil
	}

	if _, ok := rhs.Underlying().(*types.Signature); ok {
		return nil
	}

	return fmt.Errorf("expected a function, got %s", rhs)
}

type TagsChecker struct{}

func (ch *TagsChecker) CheckSwitchTypes(spec *Constraints, lhs types.Type, switchTypes []types.Type, hasDefaultCase bool) error {
	if len(spec.Tags) == 0 {
		return nil
	}
	for i := range switchTypes {
		err := ch.CheckAssign(spec, lhs, switchTypes[i])
		if err != nil {
			return err
		}
	}
	return nil

}

func (ch *TagsChecker) CheckAssign(spec *Constraints, lhs, rhs types.Type) error {
	if len(spec.Tags) == 0 {
		return nil
	}

	var structTyp *types.Struct
	var isStruct bool

	if ptr, ok := rhs.(*types.Pointer); ok {
		structTyp, isStruct = ptr.Elem().Underlying().(*types.Struct)
	} else {
		structTyp, isStruct = rhs.Underlying().(*types.Struct)
	}

	if !isStruct {
		return nil
	}

	missingFieldsTags := make(map[string][]string)

	for i := 0; i < structTyp.NumFields(); i++ {
		f := structTyp.Field(i)
		if !token.IsExported(f.Name()) {
			continue
		}

		for _, requiredTag := range spec.Tags {

			if got := reflect.StructTag(structTyp.Tag(i)).Get(requiredTag); got == "" {
				m := missingFieldsTags[f.Name()]
				m = append(m, requiredTag)
				missingFieldsTags[f.Name()] = m
			}
		}
	}

	if len(missingFieldsTags) == 0 {
		return nil
	}

	var missingFieldsSlice []string

	for name, missingRequiredTags := range missingFieldsTags {
		missingFieldsSlice = append(missingFieldsSlice, fmt.Sprintf("%q for field %v", missingRequiredTags, name))
	}

	// deterministic output:
	sort.Slice(missingFieldsSlice, func(i, j int) bool {
		return missingFieldsSlice[i] < missingFieldsSlice[j]
	})

	return fmt.Errorf("missing tags %s of %s", strings.Join(missingFieldsSlice, ", "), rhs)
}

type FieldsChecker struct{}

func (ch *FieldsChecker) CheckSwitchTypes(spec *Constraints, lhs types.Type, switchTypes []types.Type, hasDefaultCase bool) error {
	if len(spec.Fields) == 0 {
		return nil
	}
	for i := range switchTypes {
		err := ch.CheckAssign(spec, lhs, switchTypes[i])
		if err != nil {
			return err
		}
	}
	return nil

}

func (ch *FieldsChecker) CheckAssign(spec *Constraints, lhs, rhs types.Type) error {
	if len(spec.Fields) == 0 {
		return nil
	}

	var structTyp *types.Struct
	var isStruct bool

	if ptr, ok := rhs.(*types.Pointer); ok {
		structTyp, isStruct = ptr.Elem().Underlying().(*types.Struct)
	} else {
		structTyp, isStruct = rhs.Underlying().(*types.Struct)
	}

	var missingFieldNames []string
	missingFields := make(map[string]string)
	for name, typ := range spec.Fields {
		missingFields[name] = typ
		missingFieldNames = append(missingFieldNames, fmt.Sprintf(".%s", name))
	}

	if !isStruct {
		return fmt.Errorf("want a struct with fields %+v", strings.Join(missingFieldNames, ", "))
	}

	for i := 0; i < structTyp.NumFields(); i++ {
		f := structTyp.Field(i)
		f.Name()
		f.Type().String()

		for name, typ := range spec.Fields {
			if name == f.Name() && typ == f.Type().String() {
				delete(missingFields, name)
			}
		}
	}

	if len(missingFields) == 0 {
		return nil
	}

	var missingFieldsSlice []string

	for name, typ := range missingFields {
		missingFieldsSlice = append(missingFieldsSlice, fmt.Sprintf(".%s %s", name, typ))
	}

	// deterministic output:
	sort.Slice(missingFieldsSlice, func(i, j int) bool {
		return missingFieldsSlice[i] < missingFieldsSlice[j]
	})

	return fmt.Errorf("missing fields [%s] in %s", strings.Join(missingFieldsSlice, ", "), rhs)
}

type OneOfChecker struct{}
type NoneOfChecker struct{}

func (ch *OneOfChecker) CheckSwitchTypes(spec *Constraints, lhs types.Type, switchTypes []types.Type, hasDefaultCase bool) error {
	if len(spec.OneOf) == 0 {
		return nil
	}

	missingTyps, impossibleTyps := checkPossibleTypes(spec.OneOf, switchTypes)

	var errParts []string

	if len(impossibleTyps) > 0 {
		errParts = append(errParts, fmt.Sprintf("impossible types %v", impossibleTyps))
	}

	if !hasDefaultCase && len(missingTyps) > 0 {
		errParts = append(errParts, fmt.Sprintf("missing types %v", missingTyps))
	}

	if len(errParts) == 0 {
		return nil
	}

	return fmt.Errorf(strings.Join(errParts, ", "))
}

func (ch *OneOfChecker) CheckAssign(spec *Constraints, lhs, rhs types.Type) error {
	if len(spec.OneOf) == 0 {
		return nil
	}
	if types.Identical(lhs, rhs) {
		return nil
	}

	_, impossibleTypes := checkPossibleTypes(spec.OneOf, []types.Type{rhs})

	if len(impossibleTypes) > 0 {
		annTypeStr := types.TypeString(lhs, func(*types.Package) string { return "" })
		dynTypeStr := types.TypeString(rhs, func(*types.Package) string { return "" })
		return fmt.Errorf("%s cannot contain dynamic type %s, allowed types: %+v",
			annTypeStr, dynTypeStr, strings.Join(spec.OneOf, ", "))
	}

	return nil
}

func (ch *NoneOfChecker) CheckSwitchTypes(spec *Constraints, lhs types.Type, switchTypes []types.Type, hasDefaultCase bool) error {
	if len(spec.NoneOf) == 0 {
		return nil
	}

	impossibleTyps := checkImpossibleTypes(spec.NoneOf, switchTypes)

	var errParts []string

	if len(impossibleTyps) > 0 {
		errParts = append(errParts, fmt.Sprintf("impossible types %v", impossibleTyps))
	}

	if hasDefaultCase {
		errParts = append(errParts, fmt.Sprintf("default case not allowed"))
	}

	if len(errParts) == 0 {
		return nil
	}

	return fmt.Errorf(strings.Join(errParts, ", "))
}

func (ch *NoneOfChecker) CheckAssign(spec *Constraints, lhs, rhs types.Type) error {
	if len(spec.NoneOf) == 0 {
		return nil
	}
	if types.Identical(lhs, rhs) {
		return nil
	}

	impossibleTypes := checkImpossibleTypes(spec.NoneOf, []types.Type{rhs})

	if len(impossibleTypes) > 0 {
		annTypeStr := types.TypeString(lhs, func(*types.Package) string { return "" })
		dynTypeStr := types.TypeString(rhs, func(*types.Package) string { return "" })
		return fmt.Errorf("%s cannot contain dynamic type %s, forbidden types: %+v",
			annTypeStr, dynTypeStr, strings.Join(spec.NoneOf, ", "))
	}

	return nil
}

func parseIntertypeCommentLines(anns []string) (*Constraints, error) {
	lines := make([]string, 0, len(anns))
	for i := range anns {
		line := anns[i]
		if !strings.HasPrefix(line, "// #intertype ") {
			continue
		}
		line = strings.TrimPrefix(line, "// #intertype ")
		lines = append(lines, line)
	}

	if len(lines) == 0 {
		return nil, nil
	}

	ann := strings.Join(lines, "\n")
	var spec Constraints
	err := yaml.Unmarshal([]byte(ann), &spec)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %v %q", err, ann)
	}

	return &spec, nil
}
