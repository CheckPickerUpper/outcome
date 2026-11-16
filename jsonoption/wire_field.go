// Package jsonoption holds Field, an optional member of a JSON wire struct.
//
// Field encodes exactly as a samber/mo Option does: absent as null, present as
// its value, and null decodes to absent. Its only methods are the ones
// encoding needs (MarshalJSON, UnmarshalJSON and IsZero, which omitzero
// reads). Only wire structs hold a Field: domain code keeps outcome.Option,
// and a wire struct converts at its decode boundary with Option(field).
package jsonoption

import (
	"bytes"
	"encoding/json"

	"github.com/CheckPickerUpper/outcome"
)

// Field is an optional member of a JSON wire struct. Its zero value is
// absent, so a member missing from the decoded JSON is absent.
type Field[T any] struct {
	option outcome.Option[T]
}

// Present is a Field holding value.
func Present[T any](value T) Field[T] {
	return Field[T]{option: outcome.Some(value)}
}

// Absent is a Field holding nothing.
func Absent[T any]() Field[T] {
	return Field[T]{option: outcome.None[T]()}
}

// Option is the Field as an outcome.Option: Some for a present Field, None for
// an absent one, the zero Field included.
func Option[T any](field Field[T]) outcome.Option[T] {
	return outcome.FoldOption(field.option, outcome.Some[T], outcome.None[T])
}

// FromOption is a Field present when option is Some and absent when it is
// None.
func FromOption[T any](option outcome.Option[T]) Field[T] {
	return Field[T]{option: option}
}

// MarshalJSON encodes a present Field as its value and an absent one as null.
func (field Field[T]) MarshalJSON() ([]byte, error) {
	return outcome.FoldOption(field.option, encodePresent[T], encodeAbsent)()
}

// UnmarshalJSON decodes null, in any letter case, to absent and any other
// JSON to a present Field holding the decoded value. A value that does not
// decode leaves the Field unchanged and returns the decoding error.
func (field *Field[T]) UnmarshalJSON(encoded []byte) error {
	if bytes.Equal([]byte("null"), bytes.ToLower(encoded)) {
		*field = Absent[T]()
		return nil
	}
	// The one element of an empty array literal is the zero value of T.
	decoded := [1]T{}
	return outcome.FoldOption(outcome.EmptyableToOption(json.Unmarshal(encoded, &decoded[0])),
		func(decodeError error) error {
			return decodeError
		},
		func() error {
			*field = Present(decoded[0])
			return nil
		})
}

// IsZero is true only for an absent Field, so omitzero leaves out absent
// members and keeps every present one.
func (field Field[T]) IsZero() bool {
	return outcome.FoldOption(field.option,
		func(T) bool {
			return false
		},
		func() bool {
			return true
		})
}

func encodePresent[T any](value T) func() ([]byte, error) {
	return func() ([]byte, error) {
		return json.Marshal(value)
	}
}

func encodeAbsent() func() ([]byte, error) {
	return func() ([]byte, error) {
		return []byte("null"), nil
	}
}
