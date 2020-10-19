# Intertype
### Type analysis for annotated empty interfaces

### What is Intertype?

- This is an experiment.
- The type of the dynamic value of empty interfaces
  is not type checked by the Go type checker.
- Intertype is a cli tool that performs some checks on empty interface{}s.
- Intertype parses your code and your type annotations
  (in special comments or in intertype.yaml)
  and checks if you have violated your type annotations.
- Type annotations are how we impose some form of constraint
  on the type of the dynamic values inside empty interface{}s.

A few examples should make it more clear:

### Demo 1

Without Intertype:

```go
/*    */  package main
/*    */
/*    */  type Numeric interface{}
/*    */  
/*    */  func main() {
/*    */  	var n Numeric
/* OK */  	n = 3
/* OK */  	n = 3.14
/* OK */  	n = "abcd"
/*    */  	_ = n
/*    */  }
```

With Intertype:

```go
/*     */  package main
/*     */
/*     */  type Numeric interface{
/* --> */  	// #intertype {OneOf: [int, float64]}
/*     */  }
/*     */  
/*     */  func main() {
/*     */  	var n Numeric
/* OK  */  	n = 3
/* OK  */  	n = 3.14
/* ERR */  	n = "abcd"
/*     */  	_ = n
/*     */  }
```

### Demo 2

Without Intertype:

```go
/*    */  package main
/*    */
/*    */  import "context"
/*    */  
/*    */  func main() {
/* OK */    _ = context.WithValue(context.TODO(), "key1", "value")
/* OK */    _ = context.WithValue(context.TODO(), "key2", "value")
/* OK */    _ = context.WithValue(context.TODO(), 1234,   "value")
/* OK */    _ = context.WithValue(context.TODO(), 3.14,   "value")
/* OK */    _ = context.WithValue(context.TODO(), true,   "value")
/*    */  }
```

With Intertype:

```shell
$ cat intertype.yaml
```

```yaml
"[Params, 1] context.WithValue":
  - check: {OneOf: [string]}
```

```go
/*     */  package main
/*     */
/*     */  import "context"
/*     */  
/*     */  func main() {
/* OK  */    _ = context.WithValue(context.TODO(), "key1", "value")
/* OK  */    _ = context.WithValue(context.TODO(), "key2", "value")
/* ERR */    _ = context.WithValue(context.TODO(), 1234,   "value")
/* ERR */    _ = context.WithValue(context.TODO(), 3.14,   "value")
/* ERR */    _ = context.WithValue(context.TODO(), true,   "value")
/*     */  }
```

### Example (context.WithValue)

Here is the signature of the context.WithValue function in the standard library:

```go
func WithValue(parent Context, key, val interface{}) Context
```

key and val are both empty interface{}s.
Because the library wants to allow us to pass any value we want.

To make our system simpler,
let's say you want to make sure that all keys we pass to context.WithValue
are strings, and nothing else.
We could annotate context.WithValue inside a intertype.yaml file in our package:

```shell
$ cat intertype.yaml
```

```yaml
"[Params, 1] context.WithValue":
  - check: {OneOf: [string]}
```

This yaml thing is what we mean by a "type annotation".

All you need to do now is to run Intertype:

```bash
$ intertype program.go
```

Done! 

Now, if you pass a non-string key to WithValue, Intertype will warn you:

```go
context.WithValue("key", "value") // OK
context.WithValue(123,   "value") // [intertype] [W] cannot contain dynamic type int, allowed types: string
```

Note that the 2nd line is compiled by the Go compiler with no errors.
It is valid Go code.
But it violates the constraint we specified in intertype.yaml.

### Example (json.Marshal)

Let's say you want all fields in the structs you give json.Marshal to define a "json" tag.
Easy:

```yaml
"[Params, 0] encoding/json.Marshal":
  - check: {Tags: [json, yaml]}
```

### Example (yaml.Unmarshal)

json.Unmarshal (encoding/json package) has an analyzer in Gopls that ensures that we only pass pointers to it.
But how about yaml.Unmarshal (gopkg.in/yaml.v2 package)? Easy:

```yaml
"[Params, 1] gopkg.in/yaml.v2.Unmarshal":
  - check: {IsPointer: true}
```

### Example (template.FuncMap)

```yaml
"[Elem] text/template.FuncMap":
  - check: {"IsFunc": true}
```

### Example (sort package)

sort.Slice has an analyzer in Gopls that ensures that we only pass pointers to it.
But how about sort.SliceIsSorted? How about sort.SliceStable?
Easy (note that I included sort.Slice as well):

```yaml
"[Params, 0] sort.Slice":
  - check: {IsSlice: true}

"[Params, 0] sort.SliceIsSorted":
  - check: {IsSlice: true}

"[Params, 0] sort.SliceStable":
  - check: {IsSlice: true}
```

### Example (embedded comment)

You could declare an empty interface type using a special comment
that starts with `// #intertype ` followed by the type annotation:

```go
type Number interface {
  // #intertype {OneOf: [int, float64]}
}
```

### DefinitelyIntertyped (a shared collection of type annotations)

Because some of these annotations could also be used by others, I created a repository
that will contain them: [DefinitelyIntertyped](https://github.com/siadat/DefinitelyIntertyped).

Feel free to add/edit annotations for your own library types/functions!

### Contributing

- All contributions are very much appreciated! :)
- The syntax for the annotations is just YAML at the moment.
  I think we need to see more usecases for Intertype and then redesign this syntax.
  Meanwhile, I'd love to hear your opinion about it!
