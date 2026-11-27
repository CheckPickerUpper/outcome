package outcome

// Either is a value of type R, or a failure of type L: Right holds the value
// and Left the failure. Use it where callers branch on the failure's own type,
// which Result's error interface hides.
//
// Its members are unexported, so an Either is built only by Left and Right and
// read only by FoldEither. Its one method mentions L and R, so an
// Either[A, int] cannot be passed as an Either[B, int].
//
// Go cannot forbid an unset Either: a nil interface, as a zero struct field or
// a missing map entry gives. No failure of an arbitrary L can stand for it, so
// FoldEither takes a third arm for it rather than reading it as either side.
//
//sumtype:decl
type Either[L, R any] interface {
	isEither(L, R)
}

type left[L, R any] struct {
	value L
}

type right[L, R any] struct {
	value R
}

func (left[L, R]) isEither(L, R)  {}
func (right[L, R]) isEither(L, R) {}

// Left is an Either holding the failure value.
func Left[L, R any](value L) Either[L, R] {
	return left[L, R]{value: value}
}

// Right is an Either holding value.
func Right[L, R any](value R) Either[L, R] {
	return right[L, R]{value: value}
}

// FoldEither reads an Either: onRight receives the value of a Right, onLeft
// the failure of a Left, and onUnset is called for a nil Either and for a type
// declared outside this package that embeds Either, neither of which holds a
// value or a failure.
func FoldEither[L, R, U any](either Either[L, R], onRight func(R) U, onLeft func(L) U, onUnset func() U) U {
	switch member := either.(type) {
	case right[L, R]:
		return onRight(member.value)
	case left[L, R]:
		return onLeft(member.value)
	case nil:
	}
	return onUnset()
}
