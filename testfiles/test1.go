package main

import (
	"context"
	"encoding/json"
	"go/ast"
	"sort"
	"strings"
)

// XXX is XXX lol
type XXX int

// XX is blah
type XX interface {
	// #type int, float64, string
	// #intertype {"OneOf": ["int", "float64", "string"]}
}

type YY interface {
	// #intertype {"OneOf": ["float64"]}
}

type MyStruct struct {
	MyXX XX
	Yo   int
}

var xx XX = 555

func tuple() (XX, error) {
	myStruct := MyStruct{
		MyXX: 3.14,
		Yo:   3,
	}

	_ = struct {
		Bool   bool
		YourXX XX
	}{
		YourXX: 6.1,
	}

	_ = struct {
		Bool   bool
		YourXX XX
	}{
		Bool: true,
	}

	_ = struct {
		Int    bool
		YourXX XX
	}{
		true,
		6.1,
	}

	myMap := map[XX]XX{
		3.14: 3,
	}
	_ = myMap[true]
	myMap[3] = struct{}{}
	myMap[struct{}{}] = 3

	myStruct.MyXX = true
	type ZZ interface {
		// #intertype {"OneOf": ["float64"]}
		// #intertype {"OneOf": ["blah", "bloo"]}
	}
	return 1, nil
}

func generateTuple() (XX, bool) {
	return nil, false
}

func makeChan() chan XX {
	return make(chan XX)
}

func _() {
	ch := make(chan XX)
	ch <- true
	makeChan() <- true
}

func _() {
	var t XX
	var err error
	t, err = tuple()
	_ = t
	_ = err

	t2, err2 := tuple()
	_ = t2
	_ = err2

	x1 := 3.14
	var x2 = 3.14
	var x3 float64 = 3.14
	var x4 interface{} = 3.14
	var x5 XX = 3.14
	var x6 XX
	_ = x1
	_ = x2
	_ = x3
	_ = x4
	_ = x5
	_ = x6
}

func f_switch1() {
	var x XX
	switch x.(type) {
	case string:
	case int:
	case nil:
	}
}

func f_switch2() {
	var x XX
	switch x.(type) {
	case int:
	case struct{}:
	case nil:
	default:
	}
}

func f1() {
	var m0 interface{} = nil
	var m1 XX = m0
	_ = m1
}
func f2() {
	var m0 interface{} = nil
	var m1 interface{} = m0
	_ = m1
}

func main() {
	var m0 interface{} = nil
	var m1 XX = "hi"
	var m2, m3 XX = m0, 42
	ret := haha(3.14, true, m1, m2, m3)
	_ = ret

	haha(generateTuple())

	var m4 XX
	m4 = struct{}{}
	_ = m4

	ch := make(chan XX)
	ch <- true
	m4 = <-ch

}

func haha(xs ...XX) XX {
	return nil
}

func haha2(chan XX) {
	return
}

func _() XX {
	return true
}

func retval() (struct{}, int) {
	return struct{}{}, 5
}

func _() (XX, int) {
	return retval()
}

type SpecialFields interface {
	// #intertype {"Fields": {"Name": "string", "Cache": "map[string]string"}}
}

func _() {
	type S struct {
		Name  string
		Cache map[string]string
	}

	var x SpecialFields = struct {
		Name  string
		Cache map[string]string
	}{}

	x = S{}
	x = &S{}

	x = struct {
		Name_ string
		Cache map[string]string
	}{}

	switch x.(type) {
	case S:
	case struct{}:
	}

	_ = x
}

func _() {
	type Func interface {
		// #intertype {"IsFunc": true}
	}

	type FuncType func() float64

	myFunc := func() {}
	var f Func
	f = func() {}
	f = func() bool { return false }
	f = func(string) {}
	f = func(string) bool { return false }
	f = myFunc
	f = generateTuple
	f = 1
	f = struct{}{}
	f = nil
	f = strings.Builder{}

	switch f.(type) {
	case func():
	case func() bool:
	case FuncType:
	case int:
	}

	_ = f
}

func _() {
	type IsSlice interface {
		// #intertype {"IsSlice": true}
	}

	type SliceType []float64

	mySlice := []int{}
	var s IsSlice
	s = []int{}
	s = []bool{}
	s = []struct {
		bool
		int
	}{}
	s = []interface{}{}
	s = mySlice
	s = 1
	s = struct{}{}
	s = nil
	s = strings.Builder{}

	switch s.(type) {
	case func():
	case func() bool:
	case SliceType:
	case int:
	}

	_ = s
}

func _() {
	type IsChan interface {
		// #intertype {"IsChan": true}
	}

	type ChanType chan string

	myChan := make(chan int)
	var s IsChan
	s = make(chan int)
	s = make(chan bool)
	s = make(chan<- string)
	s = make(<-chan struct{})
	s = myChan
	s = 1
	s = struct{}{}
	s = nil
	s = strings.Builder{}

	switch s.(type) {
	case chan int:
	case chan struct{}:
	case ChanType:
	case int:
	}

	_ = s
}

func _() {
	//
	//
	//
}

type StructWithMethod struct{}

func (s *StructWithMethod) Do(x XX) {}

func _() {
	s := StructWithMethod{}
	s.Do(true)
}

type Doer interface {
	Do(x XX)
}

func _() {
	var s Doer = &StructWithMethod{}
	s.Do(true)
	DoCopy := s.Do
	DoCopy(true) // known limitation: this cannot be checked atm
	_ = XX(true)
	_ = XX(3.14)

	ctx := context.Background()
	context.WithValue(ctx, 3, 4)
}

func _() {
	s := 5
	s2 := []string{}
	sort.Slice(s, func(i, j int) bool { return true })
	sort.Slice(3.14, func(i, j int) bool { return true })
	sort.Slice(s2, func(i, j int) bool { return true })
	json.Unmarshal([]byte("{}"), 3)
	type S struct{}
	var st S
	json.Unmarshal([]byte("{}"), st)
	json.Unmarshal([]byte("{}"), &st)

	ctx := context.Background()
	ctx.Value(true)
	ctx.Value("key")

	json.NewDecoder(nil).Decode(5)
	obj := ast.Object{}
	obj.Data = []string{}
	obj.Data = 3.14

	_ = ast.Object{Data: 3.14}
	_ = ast.Object{1, "name", nil, 3.14, nil}

	type ExtMapWithInterfaceKey map[interface{}]string
	m := ExtMapWithInterfaceKey{1: "hi"}
	m[2] = "hi"
	_ = m[3]
	m[[]string{}] = "hi"
	_ = m[[]string{}]

	type ExtXX interface{}
	var xx ExtXX
	xx = true
	xx = []string{}
	_ = xx

	_ = ExtMapWithInterfaceKey{[]int{}: "hi"}
	_ = ExtXX(true)
}

type ExtYY interface{}

func _() {
	ch := make(chan ExtYY)
	ch <- "chanvalue"
	makeChan() <- true
}

func ReturnIntOrFloat() interface{} {
	if true {
		return 100
	} else {
		return "bad"
	}
}

func Sum(a interface{}, b interface{}) interface{} {
	return 0
}

func _() {
	Sum(1, 1)
	Sum(1.1, 1.1)
	Sum("good", "good")
	Sum(2.2, "bad") // just this one is bad
	Sum(true, false)
}

type myOutput struct {
	Field1 string `yaml:"Field1"`
	Field2 int    `json:"Field2"`
	field  bool
}

func _() {
	json.Marshal(myOutput{})
}

type TemplateFunction interface {
	// #intertype {OneOf: ["func(x string) string", "func(x string) (string, error)"]}
}

func _() {
	var f TemplateFunction
	f = func(x string) string {
		return "TODO"
	}
	f = func(x string) (string, error) {
		return "TODO", nil
	}
	f = func() (string, error) {
		return "TODO", nil
	}
	_ = f
}

func _() {
	var xs []XX
	_ = append(xs, true)
}

type Deprecated interface {
	// #intertype {"NoneOf": ["int", "float64"]}
}

func _() {
	var nn Deprecated = 3
	_ = nn
}
