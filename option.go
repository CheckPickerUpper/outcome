package outcome

// Option is a value of type T that may be absent. Absence is not a failure.
//
// Its members are unexported, so an Option is built only by Some, None and the
// To...Option functions and read only by FoldOption. Its one method mentions
// T, so an Option[string] cannot be passed as an Option[int].
//
//sumtype:decl
type Option[T any] interface {
	isOption(T)
}

type some[T any] struct {
	value T
}

type none[T any] struct{}

func (some[T]) isOption(T) {}
func (none[T]) isOption(T) {}

// Some is an Option holding value.
func Some[T any](value T) Option[T] {
	return some[T]{value: value}
}

// None is an Option holding nothing.
func None[T any]() Option[T] {
	return none[T]{}
}

// TupleToOption turns a library's (value, present) pair into an Option:
// present gives Some(value), absent gives None.
func TupleToOption[T any](value T, present bool) Option[T] {
	if present {
		return Some(value)
	}
	return None[T]()
}

// EmptyableToOption is None for the zero value of T and Some(value) for any
// other value.
func EmptyableToOption[T comparable](value T) Option[T] {
	// The one element of an empty array literal is the zero value of T.
	if value == [1]T{}[0] {
		return None[T]()
	}
	return Some(value)
}

// PointerToOption is None for a nil pointer and Some of the pointed-to value
// otherwise.
func PointerToOption[T any](value *T) Option[T] {
	return FoldOption(EmptyableToOption(value),
		func(pointer *T) Option[T] {
			return Some(*pointer)
		},
		None[T])
}

// FoldOption reads an Option: onSome receives the value of Some, and onNone is
// called for None. A nil Option, such as an unset struct field, folds to
// onNone(), the way a zero samber/mo Option is None. A type outside this
// package that embeds Option also folds to onNone().
func FoldOption[T, R any](option Option[T], onSome func(T) R, onNone func() R) R {
	switch member := option.(type) {
	case some[T]:
		return onSome(member.value)
	case none[T]:
		return onNone()
	case nil:
		return onNone()
	}
	return onNone()
}

// ForEachOption calls action with the value of Some, and does nothing for
// None.
func ForEachOption[T any](option Option[T], action func(T)) {
	FoldOption(option,
		func(value T) struct{} {
			action(value)
			return struct{}{}
		},
		func() struct{} {
			return struct{}{}
		})
}

// EqualOption is true when both Options are None, or both are Some holding
// equal values.
func EqualOption[T comparable](option, other Option[T]) bool {
	return FoldOption(option,
		func(value T) bool {
			return FoldOption(other,
				func(otherValue T) bool {
					return value == otherValue
				},
				func() bool {
					return false
				})
		},
		func() bool {
			return FoldOption(other,
				func(T) bool {
					return false
				},
				func() bool {
					return true
				})
		})
}
