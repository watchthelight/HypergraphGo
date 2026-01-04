package check

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// ============================================================================
// Error Constructor Tests
// ============================================================================

// TestErrUnboundVar tests unbound variable error construction
func TestErrUnboundVar(t *testing.T) {
	span := NewSpan("test.hott", 10, 5, 10, 6)
	err := errUnboundVar(span, 5)

	if err.Kind != ErrUnboundVariable {
		t.Errorf("Expected ErrUnboundVariable, got %v", err.Kind)
	}
	if err.Span != span {
		t.Errorf("Span not preserved")
	}

	details, ok := err.Details.(UnboundVariableDetails)
	if !ok {
		t.Fatalf("Expected UnboundVariableDetails, got %T", err.Details)
	}
	if details.Index != 5 {
		t.Errorf("Expected index 5, got %d", details.Index)
	}

	msg := err.Error()
	if msg == "" {
		t.Error("Error message is empty")
	}
}

// TestErrTypeMismatch tests type mismatch error construction
func TestErrTypeMismatch(t *testing.T) {
	span := NewSpan("test.hott", 20, 1, 20, 10)
	expected := ast.Global{Name: "Nat"}
	actual := ast.Global{Name: "Bool"}

	err := errTypeMismatch(span, expected, actual)

	if err.Kind != ErrTypeMismatch {
		t.Errorf("Expected ErrTypeMismatch, got %v", err.Kind)
	}

	details, ok := err.Details.(TypeMismatchDetails)
	if !ok {
		t.Fatalf("Expected TypeMismatchDetails, got %T", err.Details)
	}
	if g, ok := details.Expected.(ast.Global); !ok || g.Name != "Nat" {
		t.Errorf("Expected Nat in details.Expected")
	}
	if g, ok := details.Actual.(ast.Global); !ok || g.Name != "Bool" {
		t.Errorf("Expected Bool in details.Actual")
	}

	msg := err.Error()
	if msg == "" {
		t.Error("Error message is empty")
	}
}

// TestErrNotAFunction tests not-a-function error construction
func TestErrNotAFunction(t *testing.T) {
	span := NewSpan("test.hott", 1, 1, 1, 5)
	actual := ast.Global{Name: "zero"}

	err := errNotAFunction(span, actual)

	if err.Kind != ErrNotAFunction {
		t.Errorf("Expected ErrNotAFunction, got %v", err.Kind)
	}

	details, ok := err.Details.(NotAFunctionDetails)
	if !ok {
		t.Fatalf("Expected NotAFunctionDetails, got %T", err.Details)
	}
	if g, ok := details.Actual.(ast.Global); !ok || g.Name != "zero" {
		t.Errorf("Expected zero in details.Actual")
	}
}

// TestErrNotAPair tests not-a-pair error construction
func TestErrNotAPair(t *testing.T) {
	span := NoSpan()
	actual := ast.Global{Name: "zero"}

	err := errNotAPair(span, actual)

	if err.Kind != ErrNotAPair {
		t.Errorf("Expected ErrNotAPair, got %v", err.Kind)
	}

	details, ok := err.Details.(NotAPairDetails)
	if !ok {
		t.Fatalf("Expected NotAPairDetails, got %T", err.Details)
	}
	if g, ok := details.Actual.(ast.Global); !ok || g.Name != "zero" {
		t.Errorf("Expected zero in details.Actual")
	}
}

// TestErrNotAType tests not-a-type error construction
func TestErrNotAType(t *testing.T) {
	span := NoSpan()
	actual := ast.Global{Name: "zero"}

	err := errNotAType(span, actual)

	if err.Kind != ErrNotAType {
		t.Errorf("Expected ErrNotAType, got %v", err.Kind)
	}

	details, ok := err.Details.(NotATypeDetails)
	if !ok {
		t.Fatalf("Expected NotATypeDetails, got %T", err.Details)
	}
	if g, ok := details.Actual.(ast.Global); !ok || g.Name != "zero" {
		t.Errorf("Expected zero in details.Actual")
	}
}

// TestErrUnknownGlobal tests unknown global error construction
func TestErrUnknownGlobal(t *testing.T) {
	span := NewSpan("test.hott", 5, 10, 5, 20)

	err := errUnknownGlobal(span, "undefined")

	if err.Kind != ErrUnknownGlobal {
		t.Errorf("Expected ErrUnknownGlobal, got %v", err.Kind)
	}

	details, ok := err.Details.(UnknownGlobalDetails)
	if !ok {
		t.Fatalf("Expected UnknownGlobalDetails, got %T", err.Details)
	}
	if details.Name != "undefined" {
		t.Errorf("Expected 'undefined', got %s", details.Name)
	}
}

// TestErrCannotInfer tests cannot-infer error construction
func TestErrCannotInfer(t *testing.T) {
	span := NoSpan()
	term := ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}}

	err := errCannotInfer(span, term)

	if err.Kind != ErrCannotInfer {
		t.Errorf("Expected ErrCannotInfer, got %v", err.Kind)
	}

	details, ok := err.Details.(CannotInferDetails)
	if !ok {
		t.Fatalf("Expected CannotInferDetails, got %T", err.Details)
	}
	if _, ok := details.Term.(ast.Lam); !ok {
		t.Errorf("Expected Lam in details.Term")
	}
}

// ============================================================================
// TypeError.Error() Tests
// ============================================================================

// TestTypeError_ErrorWithSpan tests error string with location
func TestTypeError_ErrorWithSpan(t *testing.T) {
	span := NewSpan("foo.hott", 10, 5, 10, 15)
	err := &TypeError{
		Span:    span,
		Kind:    ErrTypeMismatch,
		Message: "test message",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("Error message is empty")
	}
	// Should contain file and line info
	if len(msg) < len("test message") {
		t.Error("Message should include span info")
	}
}

// TestTypeError_ErrorWithoutSpan tests error string without location
func TestTypeError_ErrorWithoutSpan(t *testing.T) {
	err := &TypeError{
		Span:    NoSpan(),
		Kind:    ErrTypeMismatch,
		Message: "test message",
	}

	msg := err.Error()
	if msg != "test message" {
		t.Errorf("Expected 'test message', got '%s'", msg)
	}
}

// ============================================================================
// sprintTerm Tests
// ============================================================================

// TestSprintTerm_Nil tests sprintTerm with nil
func TestSprintTerm_Nil(t *testing.T) {
	result := sprintTerm(nil)
	if result != "<nil>" {
		t.Errorf("Expected '<nil>', got '%s'", result)
	}
}

// TestSprintTerm_Term tests sprintTerm with actual term
func TestSprintTerm_Term(t *testing.T) {
	result := sprintTerm(ast.Global{Name: "Nat"})
	if result == "" {
		t.Error("sprintTerm returned empty string")
	}
}

// ============================================================================
// ErrorDetails Interface Tests
// ============================================================================

// TestErrorDetailsInterface tests that all detail types implement the interface
func TestErrorDetailsInterface(t *testing.T) {
	// These should compile and not panic
	var _ ErrorDetails = TypeMismatchDetails{}
	var _ ErrorDetails = NotAFunctionDetails{}
	var _ ErrorDetails = NotAPairDetails{}
	var _ ErrorDetails = NotATypeDetails{}
	var _ ErrorDetails = UnboundVariableDetails{}
	var _ ErrorDetails = UnknownGlobalDetails{}
	var _ ErrorDetails = CannotInferDetails{}

	// Call the marker methods to ensure they exist
	TypeMismatchDetails{}.isErrorDetails()
	NotAFunctionDetails{}.isErrorDetails()
	NotAPairDetails{}.isErrorDetails()
	NotATypeDetails{}.isErrorDetails()
	UnboundVariableDetails{}.isErrorDetails()
	UnknownGlobalDetails{}.isErrorDetails()
	CannotInferDetails{}.isErrorDetails()
}

// ============================================================================
// Complex Error Scenario Tests
// ============================================================================

// TestTypeMismatch_ComplexTypes tests type mismatch with complex types
func TestTypeMismatch_ComplexTypes(t *testing.T) {
	expected := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "x",
			A:      ast.Var{Ix: 0},
			B:      ast.Var{Ix: 1},
		},
	}
	actual := ast.Sigma{
		Binder: "x",
		A:      ast.Global{Name: "Nat"},
		B:      ast.Global{Name: "Bool"},
	}

	err := errTypeMismatch(NoSpan(), expected, actual)

	details, ok := err.Details.(TypeMismatchDetails)
	if !ok {
		t.Fatal("Expected TypeMismatchDetails")
	}

	// Verify complex types are preserved
	if _, ok := details.Expected.(ast.Pi); !ok {
		t.Error("Expected Pi in details.Expected")
	}
	if _, ok := details.Actual.(ast.Sigma); !ok {
		t.Error("Expected Sigma in details.Actual")
	}

	// Error message should be non-empty and include type info
	msg := err.Error()
	if msg == "" {
		t.Error("Error message is empty")
	}
}

// TestErrorKind_AllValues tests all ErrorKind values have valid strings
func TestErrorKind_AllValues(t *testing.T) {
	kinds := []ErrorKind{
		ErrUnboundVariable,
		ErrTypeMismatch,
		ErrNotAFunction,
		ErrNotAPair,
		ErrNotAType,
		ErrUnknownGlobal,
		ErrCannotInfer,
		ErrOccursCheck,
	}

	for _, k := range kinds {
		s := k.String()
		if s == "" {
			t.Errorf("ErrorKind(%d).String() returned empty", k)
		}
		if s == "unknown error" && k != ErrorKind(99) {
			t.Errorf("ErrorKind(%d).String() returned 'unknown error' unexpectedly", k)
		}
	}
}

// TestErrorKind_Unknown tests unknown error kind
func TestErrorKind_Unknown(t *testing.T) {
	unknown := ErrorKind(999)
	s := unknown.String()
	if s != "unknown error" {
		t.Errorf("Expected 'unknown error' for unknown kind, got '%s'", s)
	}
}

// ============================================================================
// Span Edge Cases
// ============================================================================

// TestSpan_IsEmpty tests Span.IsEmpty
func TestSpan_IsEmpty(t *testing.T) {
	empty := NoSpan()
	if !empty.IsEmpty() {
		t.Error("NoSpan should be empty")
	}

	notEmpty := NewSpan("test.hott", 1, 1, 1, 5)
	if notEmpty.IsEmpty() {
		t.Error("NewSpan should not be empty")
	}
}

// TestSpan_SameLine tests span on same line
func TestSpan_SameLine(t *testing.T) {
	span := NewSpan("test.hott", 5, 1, 5, 10)
	s := span.String()
	// Should be "test.hott:5:1-10"
	if s == "" {
		t.Error("Span string is empty")
	}
}

// TestSpan_MultiLine tests span across multiple lines
func TestSpan_MultiLine(t *testing.T) {
	span := NewSpan("test.hott", 5, 1, 10, 5)
	s := span.String()
	// Should be "test.hott:5:1-10:5"
	if s == "" {
		t.Error("Span string is empty")
	}
}

// TestSpan_SingleChar tests span for single character
func TestSpan_SingleChar(t *testing.T) {
	span := NewSpan("", 5, 10, 5, 10)
	s := span.String()
	// Should be "5:10"
	if s == "" {
		t.Error("Span string is empty")
	}
}

// ============================================================================
// Cubical Error Constructor Tests
// ============================================================================

// TestErrNotAPath tests errNotAPath construction
func TestErrNotAPath(t *testing.T) {
	span := NewSpan("test.hott", 5, 1, 5, 10)
	err := errNotAPath(span, "expected path type")

	if err.Kind != ErrNotAFunction {
		t.Errorf("Expected ErrNotAFunction (reused), got %v", err.Kind)
	}
	if err.Span != span {
		t.Error("Span not preserved")
	}
	if err.Message != "expected path type" {
		t.Errorf("Expected message 'expected path type', got '%s'", err.Message)
	}
}

// TestErrPathEndpointMismatch tests errPathEndpointMismatch construction
func TestErrPathEndpointMismatch(t *testing.T) {
	span := NewSpan("test.hott", 10, 5, 10, 20)
	err := errPathEndpointMismatch(span, "left endpoint mismatch")

	if err.Kind != ErrTypeMismatch {
		t.Errorf("Expected ErrTypeMismatch (reused), got %v", err.Kind)
	}
	if err.Span != span {
		t.Error("Span not preserved")
	}
	if err.Message != "left endpoint mismatch" {
		t.Errorf("Expected message 'left endpoint mismatch', got '%s'", err.Message)
	}
}

// TestErrUnboundIVar tests errUnboundIVar construction
func TestErrUnboundIVar(t *testing.T) {
	span := NewSpan("test.hott", 15, 3, 15, 5)
	err := errUnboundIVar(span, 7)

	if err.Kind != ErrUnboundVariable {
		t.Errorf("Expected ErrUnboundVariable (reused), got %v", err.Kind)
	}
	if err.Span != span {
		t.Error("Span not preserved")
	}
	if err.Message != "unbound interval variable 7" {
		t.Errorf("Expected 'unbound interval variable 7', got '%s'", err.Message)
	}
}
