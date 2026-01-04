package parser

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

func TestParseTerm_Simple(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ast.Term
	}{
		{"Type", "Type", ast.Sort{U: 0}},
		{"Type0", "Type0", ast.Sort{U: 0}},
		{"Type1", "Type1", ast.Sort{U: 1}},
		{"Global", "Nat", ast.Global{Name: "Nat"}},
		{"Var shorthand", "0", ast.Var{Ix: 0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTerm(tt.input)
			if err != nil {
				t.Fatalf("ParseTerm(%q) error: %v", tt.input, err)
			}
			if !termEqual(result, tt.expected) {
				t.Errorf("ParseTerm(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseTerm_Compound(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ast.Term
	}{
		{
			"Sort",
			"(Sort 2)",
			ast.Sort{U: 2},
		},
		{
			"Var",
			"(Var 3)",
			ast.Var{Ix: 3},
		},
		{
			"Global",
			"(Global foo)",
			ast.Global{Name: "foo"},
		},
		{
			"Pi",
			"(Pi x Nat Nat)",
			ast.Pi{Binder: "x", A: ast.Global{Name: "Nat"}, B: ast.Global{Name: "Nat"}},
		},
		{
			"Lam without annotation",
			"(Lam x (Var 0))",
			ast.Lam{Binder: "x", Ann: nil, Body: ast.Var{Ix: 0}},
		},
		{
			"App",
			"(App succ zero)",
			ast.App{T: ast.Global{Name: "succ"}, U: ast.Global{Name: "zero"}},
		},
		{
			"Sigma",
			"(Sigma x Type Type)",
			ast.Sigma{Binder: "x", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}},
		},
		{
			"Pair",
			"(Pair zero zero)",
			ast.Pair{Fst: ast.Global{Name: "zero"}, Snd: ast.Global{Name: "zero"}},
		},
		{
			"Fst",
			"(Fst (Var 0))",
			ast.Fst{P: ast.Var{Ix: 0}},
		},
		{
			"Snd",
			"(Snd (Var 0))",
			ast.Snd{P: ast.Var{Ix: 0}},
		},
		{
			"Let",
			"(Let x Nat zero (Var 0))",
			ast.Let{Binder: "x", Ann: ast.Global{Name: "Nat"}, Val: ast.Global{Name: "zero"}, Body: ast.Var{Ix: 0}},
		},
		{
			"Id",
			"(Id Nat zero zero)",
			ast.Id{A: ast.Global{Name: "Nat"}, X: ast.Global{Name: "zero"}, Y: ast.Global{Name: "zero"}},
		},
		{
			"Refl",
			"(Refl Nat zero)",
			ast.Refl{A: ast.Global{Name: "Nat"}, X: ast.Global{Name: "zero"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTerm(tt.input)
			if err != nil {
				t.Fatalf("ParseTerm(%q) error: %v", tt.input, err)
			}
			if !termEqual(result, tt.expected) {
				t.Errorf("ParseTerm(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseTerm_Nested(t *testing.T) {
	// (App (App succ zero) zero)
	input := "(App (App succ zero) zero)"
	expected := ast.App{
		T: ast.App{
			T: ast.Global{Name: "succ"},
			U: ast.Global{Name: "zero"},
		},
		U: ast.Global{Name: "zero"},
	}

	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm(%q) error: %v", input, err)
	}
	if !termEqual(result, expected) {
		t.Errorf("ParseTerm(%q) = %v, want %v", input, result, expected)
	}
}

func TestParseTerm_WithComments(t *testing.T) {
	input := `
	; This is a comment
	(Pi x Nat ; another comment
	    Nat)
	`
	expected := ast.Pi{Binder: "x", A: ast.Global{Name: "Nat"}, B: ast.Global{Name: "Nat"}}

	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm error: %v", err)
	}
	if !termEqual(result, expected) {
		t.Errorf("ParseTerm = %v, want %v", result, expected)
	}
}

func TestParseTerm_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Empty", ""},
		{"Unmatched paren", "(Pi x Nat"},
		{"Unknown form", "(Unknown x)"},
		{"Bad Var index", "(Var abc)"},
		{"Extra chars", "Type extra"},
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

func TestFormatTerm(t *testing.T) {
	tests := []struct {
		term     ast.Term
		expected string
	}{
		{ast.Sort{U: 0}, "Type"},
		{ast.Sort{U: 1}, "(Sort 1)"},
		{ast.Var{Ix: 0}, "(Var 0)"},
		{ast.Global{Name: "Nat"}, "Nat"},
		{ast.Pi{Binder: "x", A: ast.Global{Name: "Nat"}, B: ast.Global{Name: "Nat"}}, "(Pi x Nat Nat)"},
		{ast.App{T: ast.Global{Name: "succ"}, U: ast.Global{Name: "zero"}}, "(App succ zero)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatTerm(tt.term)
			if result != tt.expected {
				t.Errorf("FormatTerm(%v) = %q, want %q", tt.term, result, tt.expected)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Parse, format, parse again should give same result
	inputs := []string{
		"Type",
		"(Sort 2)",
		"(Pi x Nat Nat)",
		"(App succ zero)",
		"(Lam x (Var 0))",
		"(Id Nat zero zero)",
		"(Refl Nat zero)",
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

			if !termEqual(term1, term2) {
				t.Errorf("Round trip failed: %v != %v", term1, term2)
			}
		})
	}
}

func TestParseMultiple(t *testing.T) {
	input := "Type Nat (Pi x Nat Nat)"
	terms, err := ParseMultiple(input)
	if err != nil {
		t.Fatalf("ParseMultiple error: %v", err)
	}

	if len(terms) != 3 {
		t.Fatalf("Expected 3 terms, got %d", len(terms))
	}

	if !termEqual(terms[0], ast.Sort{U: 0}) {
		t.Errorf("terms[0] = %v, want Type", terms[0])
	}
	if !termEqual(terms[1], ast.Global{Name: "Nat"}) {
		t.Errorf("terms[1] = %v, want Nat", terms[1])
	}
}

func TestMustParse(t *testing.T) {
	// Should not panic
	term := MustParse("(Pi x Nat Nat)")
	if term == nil {
		t.Error("MustParse returned nil")
	}

	// Should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParse should panic on invalid input")
		}
	}()
	MustParse("(Invalid")
}

func TestParseJ(t *testing.T) {
	// Test J eliminator parsing: (J A C d x y p)
	input := "(J Nat C d x y p)"
	term, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm(%q) error: %v", input, err)
	}

	j, ok := term.(ast.J)
	if !ok {
		t.Fatalf("expected ast.J, got %T", term)
	}

	// Check each field was parsed correctly
	if g, ok := j.A.(ast.Global); !ok || g.Name != "Nat" {
		t.Errorf("J.A = %v, want Nat", j.A)
	}
	if g, ok := j.C.(ast.Global); !ok || g.Name != "C" {
		t.Errorf("J.C = %v, want C", j.C)
	}
	if g, ok := j.D.(ast.Global); !ok || g.Name != "d" {
		t.Errorf("J.D = %v, want d", j.D)
	}
	if g, ok := j.X.(ast.Global); !ok || g.Name != "x" {
		t.Errorf("J.X = %v, want x", j.X)
	}
	if g, ok := j.Y.(ast.Global); !ok || g.Name != "y" {
		t.Errorf("J.Y = %v, want y", j.Y)
	}
	if g, ok := j.P.(ast.Global); !ok || g.Name != "p" {
		t.Errorf("J.P = %v, want p", j.P)
	}
}

func TestParseJ_Nested(t *testing.T) {
	// Test J with nested terms
	input := "(J (Pi x Nat Nat) (Lam c (Var 0)) (Refl Nat zero) zero zero (Refl Nat zero))"
	term, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm(%q) error: %v", input, err)
	}

	j, ok := term.(ast.J)
	if !ok {
		t.Fatalf("expected ast.J, got %T", term)
	}

	// A should be a Pi
	if _, ok := j.A.(ast.Pi); !ok {
		t.Errorf("J.A should be Pi, got %T", j.A)
	}

	// C should be a Lam
	if _, ok := j.C.(ast.Lam); !ok {
		t.Errorf("J.C should be Lam, got %T", j.C)
	}

	// D should be a Refl
	if _, ok := j.D.(ast.Refl); !ok {
		t.Errorf("J.D should be Refl, got %T", j.D)
	}
}

func TestNormalize(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello  ", "hello"},
		{"no spaces", "no spaces"},
		{"\t\ttabs\t\t", "tabs"},
		{"\n\nnewlines\n\n", "newlines"},
		{"  mixed \t spaces  ", "mixed \t spaces"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := Normalize(tt.input)
			if got != tt.expected {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestFormatTerm_Extended(t *testing.T) {
	tests := []struct {
		name     string
		term     ast.Term
		contains string
	}{
		{"nil", nil, "nil"},
		{"Lam with Ann", ast.Lam{Binder: "x", Ann: ast.Global{Name: "Nat"}, Body: ast.Var{Ix: 0}}, "Lam"},
		{"Sigma", ast.Sigma{Binder: "x", A: ast.Global{Name: "A"}, B: ast.Var{Ix: 0}}, "Sigma"},
		{"Pair", ast.Pair{Fst: ast.Global{Name: "a"}, Snd: ast.Global{Name: "b"}}, "Pair"},
		{"Fst", ast.Fst{P: ast.Global{Name: "p"}}, "Fst"},
		{"Snd", ast.Snd{P: ast.Global{Name: "p"}}, "Snd"},
		{"Let", ast.Let{Binder: "x", Val: ast.Global{Name: "v"}, Body: ast.Var{Ix: 0}}, "Let"},
		{"Id", ast.Id{A: ast.Global{Name: "Nat"}, X: ast.Global{Name: "x"}, Y: ast.Global{Name: "y"}}, "Id"},
		{"Refl", ast.Refl{A: ast.Global{Name: "Nat"}, X: ast.Global{Name: "x"}}, "Refl"},
		{"J", ast.J{A: ast.Global{Name: "Nat"}, C: ast.Global{Name: "C"}, D: ast.Global{Name: "d"}, X: ast.Global{Name: "x"}, Y: ast.Global{Name: "y"}, P: ast.Global{Name: "p"}}, "J"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatTerm(tt.term)
			if !containsString(result, tt.contains) {
				t.Errorf("FormatTerm(%v) = %q, should contain %q", tt.term, result, tt.contains)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// termEqual compares two terms for structural equality.
func termEqual(a, b ast.Term) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch av := a.(type) {
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
			return av.Binder == bv.Binder && termEqual(av.A, bv.A) && termEqual(av.B, bv.B)
		}
	case ast.Lam:
		if bv, ok := b.(ast.Lam); ok {
			return av.Binder == bv.Binder && termEqual(av.Ann, bv.Ann) && termEqual(av.Body, bv.Body)
		}
	case ast.App:
		if bv, ok := b.(ast.App); ok {
			return termEqual(av.T, bv.T) && termEqual(av.U, bv.U)
		}
	case ast.Sigma:
		if bv, ok := b.(ast.Sigma); ok {
			return av.Binder == bv.Binder && termEqual(av.A, bv.A) && termEqual(av.B, bv.B)
		}
	case ast.Pair:
		if bv, ok := b.(ast.Pair); ok {
			return termEqual(av.Fst, bv.Fst) && termEqual(av.Snd, bv.Snd)
		}
	case ast.Fst:
		if bv, ok := b.(ast.Fst); ok {
			return termEqual(av.P, bv.P)
		}
	case ast.Snd:
		if bv, ok := b.(ast.Snd); ok {
			return termEqual(av.P, bv.P)
		}
	case ast.Let:
		if bv, ok := b.(ast.Let); ok {
			return av.Binder == bv.Binder && termEqual(av.Ann, bv.Ann) && termEqual(av.Val, bv.Val) && termEqual(av.Body, bv.Body)
		}
	case ast.Id:
		if bv, ok := b.(ast.Id); ok {
			return termEqual(av.A, bv.A) && termEqual(av.X, bv.X) && termEqual(av.Y, bv.Y)
		}
	case ast.Refl:
		if bv, ok := b.(ast.Refl); ok {
			return termEqual(av.A, bv.A) && termEqual(av.X, bv.X)
		}
	case ast.J:
		if bv, ok := b.(ast.J); ok {
			return termEqual(av.A, bv.A) && termEqual(av.C, bv.C) && termEqual(av.D, bv.D) &&
				termEqual(av.X, bv.X) && termEqual(av.Y, bv.Y) && termEqual(av.P, bv.P)
		}
	}
	return false
}
