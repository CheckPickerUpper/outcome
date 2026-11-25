# outcome

Result and Option for Go, read with `Fold` and `FoldOption`. No type here has an exported method, so reflection keeps none of them in your binary.

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

`outcome.Some(value)`, `outcome.None[T]()`, `outcome.EmptyableToOption(value)` and `outcome.PointerToOption(pointer)` build an Option.

## JSON

A JSON struct member that may be absent is a `jsonoption.Field[T]`, which encodes absent as `null`. Convert it with `jsonoption.Option(field)` after decoding.

```go
import "github.com/CheckPickerUpper/outcome/jsonoption"

type request struct {
	Name jsonoption.Field[string] `json:"name,omitzero"`
}
```

## Licence

MIT. See [LICENSE](LICENSE).
