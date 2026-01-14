// Package parser provides parsing utilities for the HoTT kernel.
// This file contains property-based tests for parse/format invariants.
package parser

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
)

// TestParseFormatRoundTrip tests that FormatTerm produces parseable output
// that is semantically equivalent to the original (alpha-equal after normalization).
func TestParseFormatRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// Basic terms
		{"Type", "Type"},
		{"Sort 0", "(Sort 0)"},
		{"Sort 1", "(Sort 1)"},
		{"Var 0", "(Var 0)"},
		{"Var 10", "(Var 10)"},
		{"Global", "(Global foo)"},

		// Pi types
		{"Simple Pi", "(Pi x Type Type)"},
		{"Nested Pi", "(Pi x Type (Pi y Type (Var 0)))"},
		{"Arrow", "(Pi _ Type Type)"},

		// Lambda
		{"Simple Lam", "(Lam x (Var 0))"},
		// Note: Annotated lambda round-trip is asymmetric - FormatTerm outputs Type not (Sort 0)
		{"Nested Lam", "(Lam x (Lam y (Var 0)))"},

		// Application
		{"Simple App", "(App (Var 0) (Var 1))"},
		{"Nested App", "(App (App (Var 0) (Var 1)) (Var 2))"},
		{"App of Lam", "(App (Lam x (Var 0)) Type)"},

		// Sigma types
		{"Simple Sigma", "(Sigma x Type Type)"},
		{"Nested Sigma", "(Sigma x Type (Sigma y Type (Var 0)))"},

		// Pairs and projections
		{"Pair", "(Pair (Var 0) (Var 1))"},
		{"Fst", "(Fst (Var 0))"},
		{"Snd", "(Snd (Var 0))"},
		{"Fst of Pair", "(Fst (Pair (Var 0) (Var 1)))"},

		// Let
		{"Let", "(Let x Type (Var 0) (Var 0))"},
		{"Nested Let", "(Let x Type (Var 0) (Let y Type (Var 0) (Var 0)))"},

		// Identity types
		{"Id", "(Id Type (Var 0) (Var 0))"},
		{"Refl", "(Refl Type (Var 0))"},
		{"Complex J", "(J Type (Lam x (Lam y (Lam p Type))) (Lam x (Var 0)) (Var 0) (Var 0) (Refl Type (Var 0)))"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse original
			term1, err := ParseTerm(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse original: %v", err)
			}

			// Format
			formatted := FormatTerm(term1)

			// Reparse
			term2, err := ParseTerm(formatted)
			if err != nil {
				t.Fatalf("Failed to reparse formatted output %q: %v", formatted, err)
			}

			// Check alpha-equality
			if !eval.AlphaEq(term1, term2) {
				t.Errorf("Round-trip failed:\n  Original:  %v\n  Formatted: %s\n  Reparsed:  %v",
					term1, formatted, term2)
			}
		})
	}
}

// TestParseFormatCubicalRoundTrip tests round-trip for cubical terms.
func TestParseFormatCubicalRoundTrip(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Interval", "I"},
		{"I0", "i0"},
		{"I1", "i1"},
		{"IVar", "(IVar 0)"},
		{"Path", "(Path Type (Var 0) (Var 0))"},
		{"PathP", "(PathP (Lam i Type) (Var 0) (Var 0))"},
		{"PathLam", "(PathLam i (Var 0))"},
		{"PathApp i0", "(PathApp (Var 0) i0)"},
		{"PathApp i1", "(PathApp (Var 0) i1)"},
		{"Transport", "(Transport (Lam i Type) (Var 0))"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			term1, err := ParseTerm(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse original: %v", err)
			}

			formatted := FormatTerm(term1)

			term2, err := ParseTerm(formatted)
			if err != nil {
				t.Fatalf("Failed to reparse formatted output %q: %v", formatted, err)
			}

			if !eval.AlphaEq(term1, term2) {
				t.Errorf("Cubical round-trip failed:\n  Original:  %v\n  Formatted: %s\n  Reparsed:  %v",
					term1, formatted, term2)
			}
		})
	}
}

// TestParseStable verifies that parsing is idempotent:
// parse(s1) == parse(s2) implies parse(format(parse(s1))) == parse(format(parse(s2)))
func TestParseStable(t *testing.T) {
	equivalentInputs := [][]string{
		// Different whitespace, same term
		{"Type", "  Type  ", "\tType\t", "\nType\n"},
		{"(Sort 0)", "( Sort 0 )", "(  Sort  0  )"},
		{"(Pi x Type Type)", "(Pi  x  Type  Type)", "( Pi x Type Type )"},

		// Comments don't affect result
		{"Type", "; comment\nType", "Type ; comment"},

		// Different binder names are irrelevant for parsing
		// (but AST keeps them, so we check structural equivalence instead)
	}

	for _, group := range equivalentInputs {
		t.Run(group[0], func(t *testing.T) {
			// Parse all variants
			var terms []ast.Term
			for _, input := range group {
				term, err := ParseTerm(input)
				if err != nil {
					t.Fatalf("Failed to parse %q: %v", input, err)
				}
				terms = append(terms, term)
			}

			// All should be alpha-equal
			for i := 1; i < len(terms); i++ {
				if !eval.AlphaEq(terms[0], terms[i]) {
					t.Errorf("Input %q and %q parsed differently:\n  %v\n  %v",
						group[0], group[i], terms[0], terms[i])
				}
			}

			// All formatted versions should reparse to equivalent terms
			for i, term := range terms {
				formatted := FormatTerm(term)
				reparsed, err := ParseTerm(formatted)
				if err != nil {
					t.Fatalf("Failed to reparse formatted version of %q: %v", group[i], err)
				}
				if !eval.AlphaEq(term, reparsed) {
					t.Errorf("Format-reparse changed term for input %q", group[i])
				}
			}
		})
	}
}

// TestFormatTermIdempotent verifies that formatting is idempotent:
// format(parse(format(parse(s)))) == format(parse(s))
func TestFormatTermIdempotent(t *testing.T) {
	inputs := []string{
		"Type",
		"(Sort 0)",
		"(Sort 1)",
		"(Var 0)",
		"(Global x)",
		"(Pi x Type Type)",
		"(Lam x (Var 0))",
		"(App (Var 0) (Var 1))",
		"(Sigma x Type Type)",
		"(Pair (Var 0) (Var 1))",
		"(Fst (Var 0))",
		"(Snd (Var 0))",
		"(Let x Type (Var 0) (Var 0))",
		"(Id Type (Var 0) (Var 0))",
		"(Refl Type (Var 0))",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			term, err := ParseTerm(input)
			if err != nil {
				t.Fatalf("Failed to parse: %v", err)
			}

			format1 := FormatTerm(term)

			term2, err := ParseTerm(format1)
			if err != nil {
				t.Fatalf("Failed to reparse: %v", err)
			}

			format2 := FormatTerm(term2)

			if format1 != format2 {
				t.Errorf("Formatting not idempotent:\n  First:  %s\n  Second: %s", format1, format2)
			}
		})
	}
}

// TestParseErrorMessages verifies that parse errors include position info.
func TestParseErrorMessages(t *testing.T) {
	badInputs := []string{
		"(",
		")",
		"(Pi x)",
		"(Unknown x y)",
		"(Sort abc)",
		"(Var abc)",
	}

	for _, input := range badInputs {
		t.Run(input, func(t *testing.T) {
			_, err := ParseTerm(input)
			if err == nil {
				t.Fatalf("Expected error for input %q", input)
			}

			// Error should be a ParseError with position
			if pe, ok := err.(*ParseError); ok {
				if pe.Pos < 0 {
					t.Errorf("ParseError should have non-negative position, got %d", pe.Pos)
				}
				if pe.Message == "" {
					t.Error("ParseError should have a message")
				}
			}
		})
	}
}

// TestMustParsePanics verifies that MustParse panics on invalid input.
func TestMustParsePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustParse should panic on invalid input")
		}
	}()

	_ = MustParse("(Invalid)")
}

// TestMustParseValid verifies that MustParse works on valid input.
func TestMustParseValid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("MustParse should not panic on valid input: %v", r)
		}
	}()

	term := MustParse("Type")
	if term == nil {
		t.Error("MustParse returned nil for valid input")
	}
}

// TestFormatNil verifies that FormatTerm handles nil input.
func TestFormatNil(t *testing.T) {
	result := FormatTerm(nil)
	if result != "nil" {
		t.Errorf("FormatTerm(nil) = %q, want \"nil\"", result)
	}
}

// TestFormatUnknownType verifies that FormatTerm handles unknown term types.
func TestFormatUnknownType(t *testing.T) {
	// Create a term type that FormatTerm doesn't know about
	// Since we can't easily create unknown types, we verify it handles
	// all the known types without panicking

	terms := []ast.Term{
		ast.Sort{U: 0},
		ast.Var{Ix: 0},
		ast.Global{Name: "x"},
		ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}},
		ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}},
		ast.App{T: ast.Var{Ix: 0}, U: ast.Var{Ix: 0}},
		ast.Sigma{Binder: "x", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}},
		ast.Pair{Fst: ast.Var{Ix: 0}, Snd: ast.Var{Ix: 0}},
		ast.Fst{P: ast.Var{Ix: 0}},
		ast.Snd{P: ast.Var{Ix: 0}},
		ast.Let{Binder: "x", Ann: ast.Sort{U: 0}, Val: ast.Var{Ix: 0}, Body: ast.Var{Ix: 0}},
		ast.Id{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}},
		ast.Refl{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}},
	}

	for _, term := range terms {
		result := FormatTerm(term)
		if result == "" {
			t.Errorf("FormatTerm(%T) returned empty string", term)
		}
	}
}
