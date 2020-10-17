# Intertype
## Type analysis for annotated empty interfaces

See Also: [DefinitelyIntertyped](https://github.com/siadat/DefinitelyIntertyped), which
is the place to share intertype annotations for Go packages.

This is an experiment.

The type of the dynamic value of empty interfaces
is not type checked by the Go type checker.

Intertype allows you to perform some analysis on the type of the dynamic values
inside an empty interface.

## Usage:

There are 2 ways of annotating a type:

1. Annotate inside a comment

For example:

```go
type Number interface {
  // #intertype {OneOf: [int, float64]}
}
```

2. Annotate inside intertype.yaml

For example:

```yaml
"[Params, 0] sort.Slice":
  - check: {"IsSlice": true}
```
Then you could run intertype:

```bash
intertype .
intertype ./...
intertype file.go
```
