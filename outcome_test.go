package outcome_test

import (
	"errors"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/CheckPickerUpper/outcome"
)

type lookupFailure struct {
	Key string
}

func (failure lookupFailure) Error() string {
	return "no entry for " + failure.Key
}

// describeResult names the arm Fold called and what that arm received.
func describeResult(result outcome.Result[int]) string {
	return outcome.Fold(result,
		func(value int) string {
			return "ok " + strconv.Itoa(value)
		},
		describeError)
}

func describeError(err error) string {
	if lookup, isLookup := errors.AsType[lookupFailure](err); isLookup {
		return "lookup failure " + lookup.Key
	}
	if _, isMissing := errors.AsType[outcome.MissingError](err); isMissing {
		return "missing error"
	}
	if _, isUnset := errors.AsType[outcome.UnsetResult](err); isUnset {
		return "unset result"
	}
	return "another error"
}

// describeOption names the arm FoldOption called and what that arm received.
func describeOption(option outcome.Option[int]) string {
	return outcome.FoldOption(option,
		func(value int) string {
			return "some " + strconv.Itoa(value)
		},
		func() string {
			return "none"
		})
}

func TestFoldCallsTheArmOfEachResult(t *testing.T) {
	rows := []struct {
		situation string
		result    outcome.Result[int]
		want      string
	}{
		{situation: "a success passes its value to onOk", result: outcome.Ok(7), want: "ok 7"},
		{situation: "a failure passes its error to onErr", result: outcome.Err[int](lookupFailure{Key: "a"}), want: "lookup failure a"},
		{situation: "a failure built with a nil error passes MissingError to onErr", result: outcome.Err[int](nil), want: "missing error"},
		// An unset slice element holds a nil Result, as an unset struct field does.
		{situation: "a nil Result passes UnsetResult to onErr", result: make([]outcome.Result[int], 1)[0], want: "unset result"},
		{situation: "a library pair with a nil error is a success", result: outcome.TupleToResult(7, nil), want: "ok 7"},
		{situation: "a library pair with an error is that failure", result: outcome.TupleToResult(7, lookupFailure{Key: "b"}), want: "lookup failure b"},
	}
	for _, row := range rows {
		t.Run(row.situation, func(t *testing.T) {
			if got := describeResult(row.result); got != row.want {
				t.Fatalf("Fold gave %q, want %q", got, row.want)
			}
		})
	}
}

func TestFoldOptionCallsTheArmOfEachOption(t *testing.T) {
	seven := 7
	zero := 0
	rows := []struct {
		situation string
		option    outcome.Option[int]
		want      string
	}{
		{situation: "Some passes its value to onSome", option: outcome.Some(7), want: "some 7"},
		{situation: "None calls onNone", option: outcome.None[int](), want: "none"},
		// An unset slice element holds the zero Option, as an unset struct field does.
		{situation: "the zero Option calls onNone", option: make([]outcome.Option[int], 1)[0], want: "none"},
		{situation: "a present library pair is Some", option: outcome.TupleToOption(7, true), want: "some 7"},
		{situation: "an absent library pair is None", option: outcome.TupleToOption(7, false), want: "none"},
		{situation: "the zero value is None", option: outcome.EmptyableToOption(0), want: "none"},
		{situation: "a non-zero value is Some", option: outcome.EmptyableToOption(7), want: "some 7"},
		{situation: "a nil pointer is None", option: outcome.PointerToOption(make([]*int, 1)[0]), want: "none"},
		{situation: "a pointer is Some of what it points to", option: outcome.PointerToOption(&seven), want: "some 7"},
		{situation: "a pointer to a zero value is Some of zero", option: outcome.PointerToOption(&zero), want: "some 0"},
	}
	for _, row := range rows {
		t.Run(row.situation, func(t *testing.T) {
			if got := describeOption(row.option); got != row.want {
				t.Fatalf("FoldOption gave %q, want %q", got, row.want)
			}
		})
	}
}

// succeeded is the value of a successful result, and fails the test with what
// it was doing otherwise.
func succeeded[T any](t *testing.T, doing string, result outcome.Result[T]) T {
	t.Helper()
	return outcome.Fold(result,
		func(value T) T {
			return value
		},
		func(err error) T {
			t.Fatalf("%s failed: %v", doing, err)
			// The one element of an empty array literal is the zero value of T.
			return [1]T{}[0]
		})
}

// checkedPackage type-checks this package's own source, so its method sets are
// read from the declarations a caller links.
func checkedPackage(t *testing.T) *types.Package {
	t.Helper()
	files := token.NewFileSet()
	parsed := []*ast.File{}
	for _, path := range succeeded(t, "listing the package's files", outcome.TupleToResult(filepath.Glob("*.go"))) {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		parsed = append(parsed, succeeded(t, "parsing "+path, outcome.TupleToResult(parser.ParseFile(files, path, nil, parser.SkipObjectResolution))))
	}
	checker := types.Config{
		Context:                  types.NewContext(),
		GoVersion:                "",
		IgnoreFuncBodies:         true,
		FakeImportC:              false,
		Error:                    nil,
		Importer:                 importer.ForCompiler(files, "source", nil),
		Sizes:                    nil,
		DisableUnusedImportCheck: false,
	}
	return succeeded(t, "type-checking the package", outcome.TupleToResult(checker.Check("github.com/CheckPickerUpper/outcome", files, parsed, nil)))
}

// instantiated is the named type declared in checked under name, given the
// type argument argument.
func instantiated(t *testing.T, checked *types.Package, name string, argument types.Type) types.Type {
	t.Helper()
	declared, isTypeName := checked.Scope().Lookup(name).(*types.TypeName)
	if !isTypeName {
		t.Fatalf("package outcome declares no type %s", name)
	}
	return succeeded(t, "instantiating "+name, outcome.TupleToResult(types.Instantiate(nil, declared.Type(), []types.Type{argument}, true)))
}

// methodSet is the interface an instantiated Result or Option declares.
func methodSet(t *testing.T, set types.Type) *types.Interface {
	t.Helper()
	declared, isInterface := set.Underlying().(*types.Interface)
	if !isInterface {
		t.Fatalf("%v is not an interface", set)
	}
	return declared
}

// exportedMethods names the exported methods of declared and of a pointer to
// it; the pointer's method set holds both.
func exportedMethods(declared types.Type) []string {
	names := []string{}
	methods := types.NewMethodSet(types.NewPointer(declared))
	for index := range methods.Len() {
		if method := methods.At(index).Obj(); method.Exported() {
			names = append(names, method.Name())
		}
	}
	return names
}

func TestNoMemberOfResultOrOptionHasAnExportedMethod(t *testing.T) {
	checked := checkedPackage(t)
	integer := types.Typ[types.Int]
	result := methodSet(t, instantiated(t, checked, "Result", integer))
	option := instantiated(t, checked, "Option", integer)
	members := []string{}
	for _, name := range checked.Scope().Names() {
		declared, isTypeName := checked.Scope().Lookup(name).(*types.TypeName)
		if !isTypeName || types.IsInterface(declared.Type()) {
			continue
		}
		named, isNamed := declared.Type().(*types.Named)
		if !isNamed || named.TypeParams().Len() != 1 {
			continue
		}
		member := instantiated(t, checked, name, integer)
		if !types.Implements(member, result) && !types.Identical(member, option) {
			continue
		}
		members = append(members, name)
		if methods := exportedMethods(member); len(methods) != 0 {
			t.Errorf("%s has exported methods %v, want none", name, methods)
		}
	}
	if !slices.Equal(members, []string{"Option", "failed", "ok"}) {
		t.Fatalf("found members %v, want [Option failed ok]", members)
	}
}

func TestAFailureForOneTypeIsNotAResultOfAnother(t *testing.T) {
	checked := checkedPackage(t)
	resultOfInt := methodSet(t, instantiated(t, checked, "Result", types.Typ[types.Int]))
	if !types.Implements(instantiated(t, checked, "failed", types.Typ[types.Int]), resultOfInt) {
		t.Fatalf("failed[int] does not implement Result[int]")
	}
	if types.Implements(instantiated(t, checked, "failed", types.Typ[types.String]), resultOfInt) {
		t.Fatalf("failed[string] implements Result[int]")
	}
}

func TestForEachCallsActionOnlyForASuccess(t *testing.T) {
	received := []int{}
	collect := func(value int) {
		received = append(received, value)
	}
	outcome.ForEach(outcome.Ok(7), collect)
	outcome.ForEach(outcome.Err[int](lookupFailure{Key: "a"}), collect)
	if !slices.Equal(received, []int{7}) {
		t.Fatalf("ForEach passed %v, want [7]", received)
	}
}

func TestForEachOptionCallsActionOnlyForSome(t *testing.T) {
	received := []int{}
	collect := func(value int) {
		received = append(received, value)
	}
	outcome.ForEachOption(outcome.Some(7), collect)
	outcome.ForEachOption(outcome.None[int](), collect)
	if !slices.Equal(received, []int{7}) {
		t.Fatalf("ForEachOption passed %v, want [7]", received)
	}
}

func TestEqualOptionComparesPresenceAndValue(t *testing.T) {
	rows := []struct {
		situation string
		option    outcome.Option[int]
		other     outcome.Option[int]
		want      bool
	}{
		{situation: "Some of equal values", option: outcome.Some(1), other: outcome.Some(1), want: true},
		{situation: "Some of different values", option: outcome.Some(1), other: outcome.Some(2), want: false},
		{situation: "Some and None", option: outcome.Some(0), other: outcome.None[int](), want: false},
		{situation: "None and Some", option: outcome.None[int](), other: outcome.Some(0), want: false},
		{situation: "None and None", option: outcome.None[int](), other: outcome.None[int](), want: true},
	}
	for _, row := range rows {
		t.Run(row.situation, func(t *testing.T) {
			if got := outcome.EqualOption(row.option, row.other); got != row.want {
				t.Fatalf("EqualOption gave %v, want %v", got, row.want)
			}
		})
	}
}
