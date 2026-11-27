// Package outcome holds Result, Either and Option, read with Fold, FoldEither
// and FoldOption.
//
// No type in this package has an exported method, so reflection over a value
// keeps no method alive in the linked binary, and each generic function is
// instantiated only where a caller calls it.
package outcome

import "fmt"

// Result is a value of type T, or the error that stopped it being made.
//
// Its members are unexported, so a Result is built only by Ok, Err and
// TupleToResult and read only by Fold. Its one method mentions T, so a
// Result[string] failure cannot be passed as a Result[int].
//
//sumtype:decl
type Result[T any] interface {
	isResult(T)
}

type ok[T any] struct {
	value T
}

type failed[T any] struct {
	err error
}

func (ok[T]) isResult(T)     {}
func (failed[T]) isResult(T) {}

// MissingError is the failure Err holds when it was given a nil error, so a
// nil never reaches a caller as success.
type MissingError struct{}

// Error says a failure was built with no error.
func (MissingError) Error() string {
	return "a Result failure was built with a nil error"
}

// UnsetResult is the error Fold passes to onErr for a nil Result.
type UnsetResult struct{}

// Error says an unset Result was folded.
func (UnsetResult) Error() string {
	return "an unset Result was folded"
}

// ForeignResult is the error Fold passes to onErr for a type declared outside
// this package that embeds Result, with that type's name.
type ForeignResult struct {
	Type string
}

// Error names the foreign type that was folded.
func (foreign ForeignResult) Error() string {
	return fmt.Sprintf("a Result of type %s, declared outside package outcome, was folded", foreign.Type)
}

// Ok is a Result holding value.
func Ok[T any](value T) Result[T] {
	return ok[T]{value: value}
}

// Err is a Result that failed with err. A nil err gives a failure holding
// MissingError.
func Err[T any](err error) Result[T] {
	return FoldOption(EmptyableToOption(err),
		func(present error) Result[T] {
			return failed[T]{err: present}
		},
		func() Result[T] {
			return failed[T]{err: MissingError{}}
		})
}

// TupleToResult turns a library's (value, err) pair into a Result: a nil err
// gives Ok(value), any other err gives Err(err).
func TupleToResult[T any](value T, err error) Result[T] {
	return FoldOption(EmptyableToOption(err),
		Err[T],
		func() Result[T] {
			return Ok(value)
		})
}

// Fold reads a Result: onOk receives the value of a success, onErr the error
// of a failure. A nil Result gives onErr(UnsetResult{}), and a type outside
// this package that embeds Result gives onErr(ForeignResult{}) naming it.
func Fold[T, R any](result Result[T], onOk func(T) R, onErr func(error) R) R {
	switch member := result.(type) {
	case ok[T]:
		return onOk(member.value)
	case failed[T]:
		return onErr(member.err)
	case nil:
		return onErr(UnsetResult{})
	}
	return onErr(ForeignResult{Type: fmt.Sprintf("%T", result)})
}

// ForEach calls action with the value of a successful Result, and does
// nothing for a failure.
func ForEach[T any](result Result[T], action func(T)) {
	Fold(result,
		func(value T) struct{} {
			action(value)
			return struct{}{}
		},
		func(error) struct{} {
			return struct{}{}
		})
}
