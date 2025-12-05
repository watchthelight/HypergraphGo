package check

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// ErrorKind categorizes type errors for programmatic handling.
type ErrorKind int

const (
	ErrUnboundVariable ErrorKind = iota
	ErrTypeMismatch
	ErrNotAFunction
	ErrNotAPair
	ErrNotAType
	ErrUnknownGlobal
	ErrCannotInfer
	ErrOccursCheck
)

// String returns the error kind name.
func (k ErrorKind) String() string {
	switch k {
	case ErrUnboundVariable:
		return "unbound variable"
	case ErrTypeMismatch:
		return "type mismatch"
	case ErrNotAFunction:
		return "not a function"
	case ErrNotAPair:
		return "not a pair"
	case ErrNotAType:
		return "not a type"
	case ErrUnknownGlobal:
		return "unknown global"
	case ErrCannotInfer:
		return "cannot infer type"
	case ErrOccursCheck:
		return "occurs check"
	default:
		return "unknown error"
	}
}

// ErrorDetails provides additional context for specific error kinds.
type ErrorDetails interface {
	isErrorDetails()
}

// TypeMismatchDetails provides information about type mismatches.
type TypeMismatchDetails struct {
	Expected ast.Term
	Actual   ast.Term
}

func (TypeMismatchDetails) isErrorDetails() {}

// NotAFunctionDetails provides information when a non-function is applied.
type NotAFunctionDetails struct {
	Actual ast.Term
}

func (NotAFunctionDetails) isErrorDetails() {}

// NotAPairDetails provides information when projecting from a non-pair.
type NotAPairDetails struct {
	Actual ast.Term
}

func (NotAPairDetails) isErrorDetails() {}

// NotATypeDetails provides information when a type was expected.
type NotATypeDetails struct {
	Actual ast.Term
}

func (NotATypeDetails) isErrorDetails() {}

// UnboundVariableDetails provides information about unbound variables.
type UnboundVariableDetails struct {
	Index int
}

func (UnboundVariableDetails) isErrorDetails() {}

// UnknownGlobalDetails provides information about unknown global names.
type UnknownGlobalDetails struct {
	Name string
}

func (UnknownGlobalDetails) isErrorDetails() {}

// CannotInferDetails provides information about terms whose type cannot be inferred.
type CannotInferDetails struct {
	Term ast.Term
}

func (CannotInferDetails) isErrorDetails() {}

// TypeError represents a type checking error with location and details.
type TypeError struct {
	Span    Span
	Kind    ErrorKind
	Message string
	Details ErrorDetails
}

// Error implements the error interface.
func (e *TypeError) Error() string {
	if e.Span.IsEmpty() {
		return e.Message
	}
	return e.Span.String() + ": " + e.Message
}

// Error constructors

func errUnboundVar(span Span, ix int) *TypeError {
	return &TypeError{
		Span:    span,
		Kind:    ErrUnboundVariable,
		Message: "unbound variable with index " + itoa(ix),
		Details: UnboundVariableDetails{Index: ix},
	}
}

func errTypeMismatch(span Span, expected, actual ast.Term) *TypeError {
	return &TypeError{
		Span:    span,
		Kind:    ErrTypeMismatch,
		Message: "type mismatch: expected " + sprintTerm(expected) + ", got " + sprintTerm(actual),
		Details: TypeMismatchDetails{Expected: expected, Actual: actual},
	}
}

func errNotAFunction(span Span, actual ast.Term) *TypeError {
	return &TypeError{
		Span:    span,
		Kind:    ErrNotAFunction,
		Message: "expected function type, got " + sprintTerm(actual),
		Details: NotAFunctionDetails{Actual: actual},
	}
}

func errNotAPair(span Span, actual ast.Term) *TypeError {
	return &TypeError{
		Span:    span,
		Kind:    ErrNotAPair,
		Message: "expected pair type, got " + sprintTerm(actual),
		Details: NotAPairDetails{Actual: actual},
	}
}

func errNotAType(span Span, actual ast.Term) *TypeError {
	return &TypeError{
		Span:    span,
		Kind:    ErrNotAType,
		Message: "expected type, got " + sprintTerm(actual),
		Details: NotATypeDetails{Actual: actual},
	}
}

func errUnknownGlobal(span Span, name string) *TypeError {
	return &TypeError{
		Span:    span,
		Kind:    ErrUnknownGlobal,
		Message: "unknown global: " + name,
		Details: UnknownGlobalDetails{Name: name},
	}
}

func errCannotInfer(span Span, t ast.Term) *TypeError {
	return &TypeError{
		Span:    span,
		Kind:    ErrCannotInfer,
		Message: "cannot infer type for " + sprintTerm(t),
		Details: CannotInferDetails{Term: t},
	}
}

// sprintTerm returns a simple string representation of a term for error messages.
func sprintTerm(t ast.Term) string {
	if t == nil {
		return "<nil>"
	}
	return ast.Sprint(t)
}
