# outcome

Result and Option for Go, read with `Fold` and `FoldOption`. Result and Option have no methods, so reflection keeps none of them in your binary.

```sh
go get github.com/CheckPickerUpper/outcome
```

## Result

```go
import (
	"os"

	"github.com/CheckPickerUpper/outcome"
)

func describe(path string) string {
	read := outcome.TupleToResult(os.ReadFile(path))
	return outcome.Fold(read,
		func(content []byte) string {
			return string(content)
		},
		func(err error) string {
			return "unreadable: " + err.Error()
		})
}
```

`outcome.Ok(value)` and `outcome.Err[T](err)` build a Result directly.

## Option

```go
func greeting(names map[string]string, id string) string {
	name, found := names[id]
	return outcome.FoldOption(outcome.TupleToOption(name, found),
		func(name string) string {
			return "hello " + name
		},
		func() string {
			return "hello stranger"
		})
}
```

`outcome.Some(value)`, `outcome.None[T]()`, `outcome.EmptyableToOption(value)` and `outcome.PointerToOption(pointer)` build an Option. An Option is a struct, so no other type can stand in for one.

## JSON

A JSON struct member that may be absent is a `jsonoption.Field[T]`, which encodes absent as `null`. Convert it with `jsonoption.Option(field)` after decoding.

```go
import "github.com/CheckPickerUpper/outcome/jsonoption"

type request struct {
	Name jsonoption.Field[string] `json:"name,omitzero"`
}
```

## What each value folds to

| Value | `Fold` or `FoldOption` calls |
| --- | --- |
| an unset Result (a zero struct field, a `var` with no value) | `onErr(outcome.UnsetResult{})` |
| `outcome.Err[T](nil)` | `onErr(outcome.MissingError{})` |
| `outcome.TupleToResult(value, err)` where `err` is a nil pointer of an error type, such as `(*os.PathError)(nil)` | `onErr` with that nil pointer: an interface holding a nil pointer is not a nil error |
| an unset Option | `onNone()` |

An unset Result never reads as a success. Neither type has a `Get`, `MustGet` or unwrap method: a value comes out only through `Fold`, `FoldOption`, `ForEach` or `ForEachOption`, or as JSON through a `jsonoption.Field`'s `MarshalJSON`, so every caller handles both arms.

## Compared with samber/mo and IBM/fp-go

[samber/mo](https://github.com/samber/mo) (v1.17.0) gives `Result` and `Option` methods such as `Get`, `MustGet`, `OrElse` and `Map`. Its zero `Result` reports `IsOk()` as true with a zero value, and `MustGet` panics on a failure. Result and Option have no methods, so reflection over a value keeps none alive in the linked binary, and an unset Result folds to the error arm.

[IBM/fp-go](https://github.com/IBM/fp-go) ports fp-ts to Go: curried functions composed with `Pipe`, and `Either`, `IO`, `Reader` and their combinations, with optics. outcome holds two types and the functions that build and fold them.

## Go version and releases

outcome needs Go 1.26.5 or later, as `go.mod` declares. Releases follow semantic versioning. A breaking change is one that stops a caller compiling or changes which arm `Fold` or `FoldOption` calls for a value: removing or renaming an exported name, or changing a signature. Before v1.0.0 a minor release (v0.2.0) may break; a patch release (v0.1.1) never does.

## Licence

MIT. See [LICENSE](LICENSE).
