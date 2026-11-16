package jsonoption_test

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/CheckPickerUpper/outcome"
	"github.com/CheckPickerUpper/outcome/jsonoption"
)

type wire struct {
	Count jsonoption.Field[int]    `json:"count"`
	Name  jsonoption.Field[string] `json:"name,omitzero"`
}

// describeField names whether a decoded Field is present and what it holds.
func describeField[T any](field jsonoption.Field[T], show func(T) string) string {
	return outcome.FoldOption(jsonoption.Option(field),
		func(value T) string {
			return "present " + show(value)
		},
		func() string {
			return "absent"
		})
}

func describeWire(decoded wire) string {
	return "count " + describeField(decoded.Count, strconv.Itoa) + ", name " + describeField(decoded.Name, strconv.Quote)
}

func TestFieldsEncodeAsSamberMoOptionsDo(t *testing.T) {
	rows := []struct {
		situation string
		value     wire
		want      string
	}{
		{
			situation: "a present field is its value and an absent omitzero field is left out",
			value:     wire{Count: jsonoption.Present(3), Name: jsonoption.Absent[string]()},
			want:      `{"count":3}`,
		},
		{
			situation: "an absent field is null and a present zero value is kept under omitzero",
			value:     wire{Count: jsonoption.Absent[int](), Name: jsonoption.Present("")},
			want:      `{"count":null,"name":""}`,
		},
		{
			situation: "an unset field is absent",
			value:     wire{Count: jsonoption.Field[int]{}, Name: jsonoption.Field[string]{}},
			want:      `{"count":null}`,
		},
		{
			situation: "a field from Some is present and a field from None is absent",
			value:     wire{Count: jsonoption.FromOption(outcome.Some(4)), Name: jsonoption.FromOption(outcome.None[string]())},
			want:      `{"count":4}`,
		},
	}
	for _, row := range rows {
		t.Run(row.situation, func(t *testing.T) {
			got := outcome.Fold(outcome.TupleToResult(json.Marshal(row.value)),
				func(encoded []byte) string {
					return string(encoded)
				},
				func(err error) string {
					return "Marshal failed: " + err.Error()
				})
			if got != row.want {
				t.Fatalf("Marshal gave %s, want %s", got, row.want)
			}
		})
	}
}

func TestFieldsDecodeAsSamberMoOptionsDo(t *testing.T) {
	unset := wire{Count: jsonoption.Field[int]{}, Name: jsonoption.Field[string]{}}
	rows := []struct {
		situation string
		before    wire
		encoded   string
		want      string
	}{
		{
			situation: "null replaces a present value with absent, and a value is present",
			before:    wire{Count: jsonoption.Present(9), Name: jsonoption.Absent[string]()},
			encoded:   `{"count":null,"name":"x"}`,
			want:      `count absent, name present "x"`,
		},
		{situation: "a present zero value stays present", before: unset, encoded: `{"count":0,"name":""}`, want: `count present 0, name present ""`},
		{situation: "an omitted member of a new struct is absent", before: unset, encoded: `{}`, want: "count absent, name absent"},
	}
	for _, row := range rows {
		t.Run(row.situation, func(t *testing.T) {
			decoded := row.before
			got := outcome.FoldOption(outcome.EmptyableToOption(json.Unmarshal([]byte(row.encoded), &decoded)),
				func(err error) string {
					return "Unmarshal failed: " + err.Error()
				},
				func() string {
					return describeWire(decoded)
				})
			if got != row.want {
				t.Fatalf("Unmarshal gave %s, want %s", got, row.want)
			}
		})
	}
}

func TestAValueOfTheWrongTypeFailsToDecode(t *testing.T) {
	decoded := wire{Count: jsonoption.Field[int]{}, Name: jsonoption.Field[string]{}}
	err := json.Unmarshal([]byte(`{"count":"three"}`), &decoded)
	if _, isTypeError := errors.AsType[*json.UnmarshalTypeError](err); !isTypeError {
		t.Fatalf("Unmarshal gave %v, want a *json.UnmarshalTypeError", err)
	}
}
