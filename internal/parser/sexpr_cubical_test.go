package parser

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// ============================================================================
// Cubical Atom Parsing Tests
// ============================================================================

func TestParseCubicalAtom_Interval(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ast.Term
	}{
		{"I", "I", ast.Interval{}},
		{"Interval", "Interval", ast.Interval{}},
		{"i0", "i0", ast.I0{}},
		{"i1", "i1", ast.I1{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTerm(tt.input)
			if err != nil {
				t.Fatalf("ParseTerm(%q) error: %v", tt.input, err)
			}
			if !cubicalTermEqual(result, tt.expected) {
				t.Errorf("ParseTerm(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// IVar Parsing Tests
// ============================================================================

func TestParseIVar(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ast.Term
	}{
		{"IVar 0", "(IVar 0)", ast.IVar{Ix: 0}},
		{"IVar 1", "(IVar 1)", ast.IVar{Ix: 1}},
		{"IVar 10", "(IVar 10)", ast.IVar{Ix: 10}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTerm(tt.input)
			if err != nil {
				t.Fatalf("ParseTerm(%q) error: %v", tt.input, err)
			}
			if !cubicalTermEqual(result, tt.expected) {
				t.Errorf("ParseTerm(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseIVar_Error(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"non-numeric index", "(IVar abc)"},
		{"missing index", "(IVar )"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTerm(tt.input)
			if err == nil {
				t.Errorf("ParseTerm(%q) expected error, got nil", tt.input)
			}
		})
	}
}

// ============================================================================
// Path Type Parsing Tests
// ============================================================================

func TestParsePath(t *testing.T) {
	input := "(Path Nat zero one)"
	expected := ast.Path{
		A: ast.Global{Name: "Nat"},
		X: ast.Global{Name: "zero"},
		Y: ast.Global{Name: "one"},
	}

	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm(%q) error: %v", input, err)
	}
	if !cubicalTermEqual(result, expected) {
		t.Errorf("ParseTerm(%q) = %v, want %v", input, result, expected)
	}
}

func TestParsePath_Nested(t *testing.T) {
	// Path with nested types
	input := "(Path (Pi x Nat Nat) f g)"
	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm error: %v", err)
	}

	path, ok := result.(ast.Path)
	if !ok {
		t.Fatalf("Expected Path, got %T", result)
	}
	if _, ok := path.A.(ast.Pi); !ok {
		t.Errorf("Path.A should be Pi")
	}
}

// ============================================================================
// PathP Parsing Tests
// ============================================================================

func TestParsePathP(t *testing.T) {
	input := "(PathP TypeFamily x0 x1)"
	expected := ast.PathP{
		A: ast.Global{Name: "TypeFamily"},
		X: ast.Global{Name: "x0"},
		Y: ast.Global{Name: "x1"},
	}

	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm(%q) error: %v", input, err)
	}
	if !cubicalTermEqual(result, expected) {
		t.Errorf("ParseTerm(%q) = %v, want %v", input, result, expected)
	}
}

// ============================================================================
// PathLam Parsing Tests
// ============================================================================

func TestParsePathLam(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ast.PathLam
	}{
		{
			"with binder",
			"(PathLam i x)",
			ast.PathLam{Binder: "i", Body: ast.Global{Name: "x"}},
		},
		{
			"alternative syntax <>",
			"(<> j y)",
			ast.PathLam{Binder: "j", Body: ast.Global{Name: "y"}},
		},
		{
			"with nested body",
			"(PathLam i (App f (IVar 0)))",
			ast.PathLam{
				Binder: "i",
				Body:   ast.App{T: ast.Global{Name: "f"}, U: ast.IVar{Ix: 0}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTerm(tt.input)
			if err != nil {
				t.Fatalf("ParseTerm(%q) error: %v", tt.input, err)
			}
			pl, ok := result.(ast.PathLam)
			if !ok {
				t.Fatalf("Expected PathLam, got %T", result)
			}
			if pl.Binder != tt.expected.Binder {
				t.Errorf("Binder = %q, want %q", pl.Binder, tt.expected.Binder)
			}
			if !cubicalTermEqual(pl.Body, tt.expected.Body) {
				t.Errorf("Body = %v, want %v", pl.Body, tt.expected.Body)
			}
		})
	}
}

// ============================================================================
// PathApp Parsing Tests
// ============================================================================

func TestParsePathApp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ast.PathApp
	}{
		{
			"with i0",
			"(PathApp p i0)",
			ast.PathApp{P: ast.Global{Name: "p"}, R: ast.I0{}},
		},
		{
			"with i1",
			"(PathApp p i1)",
			ast.PathApp{P: ast.Global{Name: "p"}, R: ast.I1{}},
		},
		{
			"with IVar",
			"(PathApp q (IVar 0))",
			ast.PathApp{P: ast.Global{Name: "q"}, R: ast.IVar{Ix: 0}},
		},
		{
			"alternative syntax @",
			"(@ p i1)",
			ast.PathApp{P: ast.Global{Name: "p"}, R: ast.I1{}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTerm(tt.input)
			if err != nil {
				t.Fatalf("ParseTerm(%q) error: %v", tt.input, err)
			}
			pa, ok := result.(ast.PathApp)
			if !ok {
				t.Fatalf("Expected PathApp, got %T", result)
			}
			if !cubicalTermEqual(pa.P, tt.expected.P) {
				t.Errorf("P = %v, want %v", pa.P, tt.expected.P)
			}
			if !cubicalTermEqual(pa.R, tt.expected.R) {
				t.Errorf("R = %v, want %v", pa.R, tt.expected.R)
			}
		})
	}
}

// ============================================================================
// Transport Parsing Tests
// ============================================================================

func TestParseTransport(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ast.Transport
	}{
		{
			"simple",
			"(Transport TypeFamily elem)",
			ast.Transport{A: ast.Global{Name: "TypeFamily"}, E: ast.Global{Name: "elem"}},
		},
		{
			"with nested types",
			"(Transport (PathLam i A) x)",
			ast.Transport{
				A: ast.PathLam{Binder: "i", Body: ast.Global{Name: "A"}},
				E: ast.Global{Name: "x"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTerm(tt.input)
			if err != nil {
				t.Fatalf("ParseTerm(%q) error: %v", tt.input, err)
			}
			tr, ok := result.(ast.Transport)
			if !ok {
				t.Fatalf("Expected Transport, got %T", result)
			}
			if !cubicalTermEqual(tr.A, tt.expected.A) {
				t.Errorf("A = %v, want %v", tr.A, tt.expected.A)
			}
			if !cubicalTermEqual(tr.E, tt.expected.E) {
				t.Errorf("E = %v, want %v", tr.E, tt.expected.E)
			}
		})
	}
}

// ============================================================================
// HITApp Parsing Tests
// ============================================================================

func TestParseHITApp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ast.HITApp
	}{
		{
			"no args",
			"(HITApp S1 loop () ())",
			ast.HITApp{HITName: "S1", Ctor: "loop", Args: nil, IArgs: nil},
		},
		{
			"with term args",
			"(HITApp Susp merid (a) ())",
			ast.HITApp{
				HITName: "Susp",
				Ctor:    "merid",
				Args:    []ast.Term{ast.Global{Name: "a"}},
				IArgs:   nil,
			},
		},
		{
			"with interval args",
			"(HITApp S1 loop () (i0))",
			ast.HITApp{
				HITName: "S1",
				Ctor:    "loop",
				Args:    nil,
				IArgs:   []ast.Term{ast.I0{}},
			},
		},
		{
			"with both args",
			"(HITApp Quot eq (x y) ((IVar 0)))",
			ast.HITApp{
				HITName: "Quot",
				Ctor:    "eq",
				Args:    []ast.Term{ast.Global{Name: "x"}, ast.Global{Name: "y"}},
				IArgs:   []ast.Term{ast.IVar{Ix: 0}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTerm(tt.input)
			if err != nil {
				t.Fatalf("ParseTerm(%q) error: %v", tt.input, err)
			}
			ha, ok := result.(ast.HITApp)
			if !ok {
				t.Fatalf("Expected HITApp, got %T", result)
			}
			if ha.HITName != tt.expected.HITName {
				t.Errorf("HITName = %q, want %q", ha.HITName, tt.expected.HITName)
			}
			if ha.Ctor != tt.expected.Ctor {
				t.Errorf("Ctor = %q, want %q", ha.Ctor, tt.expected.Ctor)
			}
			if len(ha.Args) != len(tt.expected.Args) {
				t.Errorf("len(Args) = %d, want %d", len(ha.Args), len(tt.expected.Args))
			}
			if len(ha.IArgs) != len(tt.expected.IArgs) {
				t.Errorf("len(IArgs) = %d, want %d", len(ha.IArgs), len(tt.expected.IArgs))
			}
		})
	}
}

func TestParseHITApp_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"missing HIT name", "(HITApp)"},
		{"missing ctor name", "(HITApp S1)"},
		{"missing args list", "(HITApp S1 loop)"},
		{"malformed args", "(HITApp S1 loop [x] ())"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTerm(tt.input)
			if err == nil {
				t.Errorf("ParseTerm(%q) expected error, got nil", tt.input)
			}
		})
	}
}

// ============================================================================
// parseTermList Tests
// ============================================================================

func TestParseTermList(t *testing.T) {
	// Tested indirectly through HITApp, but let's verify edge cases
	tests := []struct {
		name        string
		input       string
		expectedLen int
	}{
		{"empty list in HITApp", "(HITApp S1 loop () ())", 0},
		{"single element", "(HITApp S1 loop (x) ())", 1},
		{"multiple elements", "(HITApp S1 loop (a b c) ())", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTerm(tt.input)
			if err != nil {
				t.Fatalf("ParseTerm error: %v", err)
			}
			ha, ok := result.(ast.HITApp)
			if !ok {
				t.Fatalf("Expected HITApp, got %T", result)
			}
			if len(ha.Args) != tt.expectedLen {
				t.Errorf("len(Args) = %d, want %d", len(ha.Args), tt.expectedLen)
			}
		})
	}
}

// ============================================================================
// formatCubicalTerm Tests
// ============================================================================

func TestFormatCubicalTerm(t *testing.T) {
	tests := []struct {
		term     ast.Term
		expected string
	}{
		{ast.Interval{}, "I"},
		{ast.I0{}, "i0"},
		{ast.I1{}, "i1"},
		{ast.IVar{Ix: 0}, "(IVar 0)"},
		{ast.IVar{Ix: 5}, "(IVar 5)"},
		{
			ast.Path{A: ast.Global{Name: "A"}, X: ast.Global{Name: "x"}, Y: ast.Global{Name: "y"}},
			"(Path A x y)",
		},
		{
			ast.PathP{A: ast.Global{Name: "F"}, X: ast.Global{Name: "a"}, Y: ast.Global{Name: "b"}},
			"(PathP F a b)",
		},
		{
			ast.PathLam{Binder: "i", Body: ast.Global{Name: "t"}},
			"(PathLam i t)",
		},
		{
			ast.PathApp{P: ast.Global{Name: "p"}, R: ast.I0{}},
			"(PathApp p i0)",
		},
		{
			ast.Transport{A: ast.Global{Name: "F"}, E: ast.Global{Name: "e"}},
			"(Transport F e)",
		},
		{
			ast.HITApp{HITName: "S1", Ctor: "loop", Args: nil, IArgs: nil},
			"(HITApp S1 loop () ())",
		},
		{
			ast.HITApp{
				HITName: "Susp",
				Ctor:    "merid",
				Args:    []ast.Term{ast.Global{Name: "a"}},
				IArgs:   []ast.Term{ast.I0{}},
			},
			"(HITApp Susp merid (a) (i0))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatTerm(tt.term)
			if result != tt.expected {
				t.Errorf("FormatTerm(%T) = %q, want %q", tt.term, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Round-Trip Tests for Cubical Terms
// ============================================================================

func TestCubicalRoundTrip(t *testing.T) {
	inputs := []string{
		"I",
		"i0",
		"i1",
		"(IVar 0)",
		"(IVar 5)",
		"(Path Nat zero zero)",
		"(PathP F x y)",
		"(PathLam i x)",
		"(PathApp p i0)",
		"(Transport F e)",
		"(HITApp S1 loop () ())",
		"(HITApp Susp merid (a) (i0))",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			term1, err := ParseTerm(input)
			if err != nil {
				t.Fatalf("First parse error: %v", err)
			}

			formatted := FormatTerm(term1)
			term2, err := ParseTerm(formatted)
			if err != nil {
				t.Fatalf("Second parse error: %v", err)
			}

			if !cubicalTermEqual(term1, term2) {
				t.Errorf("Round trip failed: %v != %v", term1, term2)
			}
		})
	}
}

// ============================================================================
// Complex Cubical Expressions
// ============================================================================

func TestParseCubicalComplex(t *testing.T) {
	// PathApp on PathLam
	input := "(PathApp (PathLam i x) i0)"
	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm error: %v", err)
	}

	pa, ok := result.(ast.PathApp)
	if !ok {
		t.Fatalf("Expected PathApp, got %T", result)
	}
	if _, ok := pa.P.(ast.PathLam); !ok {
		t.Errorf("PathApp.P should be PathLam")
	}
	if _, ok := pa.R.(ast.I0); !ok {
		t.Errorf("PathApp.R should be I0")
	}
}

func TestParseCubicalWithMixedTerms(t *testing.T) {
	// Path with Pi type
	input := "(Path (Pi x Nat Nat) f g)"
	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm error: %v", err)
	}

	path, ok := result.(ast.Path)
	if !ok {
		t.Fatalf("Expected Path, got %T", result)
	}
	if _, ok := path.A.(ast.Pi); !ok {
		t.Errorf("Path.A should be Pi")
	}
}

// ============================================================================
// Helper: cubicalTermEqual
// ============================================================================

// cubicalTermEqual extends termEqual with cubical types.
func cubicalTermEqual(a, b ast.Term) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch av := a.(type) {
	// Cubical interval types
	case ast.Interval:
		_, ok := b.(ast.Interval)
		return ok
	case ast.I0:
		_, ok := b.(ast.I0)
		return ok
	case ast.I1:
		_, ok := b.(ast.I1)
		return ok
	case ast.IVar:
		if bv, ok := b.(ast.IVar); ok {
			return av.Ix == bv.Ix
		}

	// Cubical path types
	case ast.Path:
		if bv, ok := b.(ast.Path); ok {
			return cubicalTermEqual(av.A, bv.A) && cubicalTermEqual(av.X, bv.X) && cubicalTermEqual(av.Y, bv.Y)
		}
	case ast.PathP:
		if bv, ok := b.(ast.PathP); ok {
			return cubicalTermEqual(av.A, bv.A) && cubicalTermEqual(av.X, bv.X) && cubicalTermEqual(av.Y, bv.Y)
		}
	case ast.PathLam:
		if bv, ok := b.(ast.PathLam); ok {
			return av.Binder == bv.Binder && cubicalTermEqual(av.Body, bv.Body)
		}
	case ast.PathApp:
		if bv, ok := b.(ast.PathApp); ok {
			return cubicalTermEqual(av.P, bv.P) && cubicalTermEqual(av.R, bv.R)
		}
	case ast.Transport:
		if bv, ok := b.(ast.Transport); ok {
			return cubicalTermEqual(av.A, bv.A) && cubicalTermEqual(av.E, bv.E)
		}

	// HIT
	case ast.HITApp:
		if bv, ok := b.(ast.HITApp); ok {
			if av.HITName != bv.HITName || av.Ctor != bv.Ctor {
				return false
			}
			if len(av.Args) != len(bv.Args) || len(av.IArgs) != len(bv.IArgs) {
				return false
			}
			for i := range av.Args {
				if !cubicalTermEqual(av.Args[i], bv.Args[i]) {
					return false
				}
			}
			for i := range av.IArgs {
				if !cubicalTermEqual(av.IArgs[i], bv.IArgs[i]) {
					return false
				}
			}
			return true
		}

	// Fall back to non-cubical comparison
	case ast.Sort:
		if bv, ok := b.(ast.Sort); ok {
			return av.U == bv.U
		}
	case ast.Var:
		if bv, ok := b.(ast.Var); ok {
			return av.Ix == bv.Ix
		}
	case ast.Global:
		if bv, ok := b.(ast.Global); ok {
			return av.Name == bv.Name
		}
	case ast.Pi:
		if bv, ok := b.(ast.Pi); ok {
			return av.Binder == bv.Binder && cubicalTermEqual(av.A, bv.A) && cubicalTermEqual(av.B, bv.B)
		}
	case ast.Lam:
		if bv, ok := b.(ast.Lam); ok {
			return av.Binder == bv.Binder && cubicalTermEqual(av.Ann, bv.Ann) && cubicalTermEqual(av.Body, bv.Body)
		}
	case ast.App:
		if bv, ok := b.(ast.App); ok {
			return cubicalTermEqual(av.T, bv.T) && cubicalTermEqual(av.U, bv.U)
		}
	case ast.Sigma:
		if bv, ok := b.(ast.Sigma); ok {
			return av.Binder == bv.Binder && cubicalTermEqual(av.A, bv.A) && cubicalTermEqual(av.B, bv.B)
		}
	case ast.Pair:
		if bv, ok := b.(ast.Pair); ok {
			return cubicalTermEqual(av.Fst, bv.Fst) && cubicalTermEqual(av.Snd, bv.Snd)
		}
	case ast.Fst:
		if bv, ok := b.(ast.Fst); ok {
			return cubicalTermEqual(av.P, bv.P)
		}
	case ast.Snd:
		if bv, ok := b.(ast.Snd); ok {
			return cubicalTermEqual(av.P, bv.P)
		}
	case ast.Let:
		if bv, ok := b.(ast.Let); ok {
			return av.Binder == bv.Binder && cubicalTermEqual(av.Ann, bv.Ann) && cubicalTermEqual(av.Val, bv.Val) && cubicalTermEqual(av.Body, bv.Body)
		}
	case ast.Id:
		if bv, ok := b.(ast.Id); ok {
			return cubicalTermEqual(av.A, bv.A) && cubicalTermEqual(av.X, bv.X) && cubicalTermEqual(av.Y, bv.Y)
		}
	case ast.Refl:
		if bv, ok := b.(ast.Refl); ok {
			return cubicalTermEqual(av.A, bv.A) && cubicalTermEqual(av.X, bv.X)
		}
	}
	return false
}
