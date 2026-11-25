package outcome

// Option is a value of type T that may be absent. Absence is not a failure.
//
// Its fields are unexported, so an Option is built only by Some, None and the
// To...Option functions and read only by FoldOption. It is a struct, so no
// type declared elsewhere can stand in for one, and its zero value, such as
// an unset struct field, is None.
type Option[T any] struct {
	value   T
	present bool
}

// Some is an Option holding value.
func Some[T any](value T) Option[T] {
	return Option[T]{value: value, present: true}
}

// None is an Option holding nothing.
func None[T any]() Option[T] {
	return Option[T]{}
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
// called for None, the zero Option included.
func FoldOption[T, R any](option Option[T], onSome func(T) R, onNone func() R) R {
	if option.present {
		return onSome(option.value)
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
