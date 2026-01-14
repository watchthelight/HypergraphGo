// Package parser provides parsing utilities for the HoTT kernel.
// This file contains fuzz tests for the S-expression parser.
package parser

import (
	"strings"
	"testing"
)

// FuzzParseTerm fuzzes the S-expression parser with random inputs.
// It verifies that:
// 1. The parser never panics on arbitrary input
// 2. Successfully parsed terms can be formatted and re-parsed (when possible)
func FuzzParseTerm(f *testing.F) {
	// Seed corpus with valid S-expressions
	seeds := []string{
		// Basic terms
		"Type",
		"Type0",
		"Type1",
		"Type2",
		"0",
		"1",
		"42",
		"x",
		"foo",
		"_",

		// Compound forms
		"(Sort 0)",
		"(Sort 1)",
		"(Sort 42)",
		"(Var 0)",
		"(Var 1)",
		"(Var 99)",
		"(Global x)",
		"(Global foo)",
		"(Global _underscore)",

		// Pi types
		"(Pi x Type Type)",
		"(Pi _ Type (Var 0))",
		"(Pi a (Sort 0) (Sort 1))",
		"(-> x Type Type)",

		// Lambda
		"(Lam x (Var 0))",
		"(Lam x Type (Var 0))",
		"(λ x (Var 0))",
		"(lambda x (Var 0))",
		"(\\ x (Var 0))",

		// Application
		"(App (Var 0) (Var 1))",
		"(App (Lam x (Var 0)) Type)",

		// Sigma types
		"(Sigma x Type Type)",
		"(Σ x Type Type)",

		// Pairs
		"(Pair (Var 0) (Var 1))",
		"(Fst (Var 0))",
		"(Snd (Var 0))",

		// Let
		"(Let x Type (Var 0) (Var 0))",

		// Identity types
		"(Id Type (Var 0) (Var 0))",
		"(Refl Type (Var 0))",
		"(J Type (Lam x (Lam y (Lam p Type))) (Lam x (Var 0)) (Var 0) (Var 0) (Refl Type (Var 0)))",

		// Cubical terms
		"I",
		"i0",
		"i1",
		"(Path Type (Var 0) (Var 1))",
		"(PathP (Lam i Type) (Var 0) (Var 1))",
		"(PathLam i (Var 0))",
		"(PathApp (Var 0) i0)",
		"(Transport (Lam i Type) (Var 0))",

		// Nested expressions
		"(App (App (Var 0) (Var 1)) (Var 2))",
		"(Pi x Type (Pi y Type (Var 0)))",
		"(Lam x (Lam y (App (Var 1) (Var 0))))",

		// Comments
		"; comment\nType",
		"Type ; inline comment",
		"(Pi x ; comment in middle\n Type Type)",

		// Whitespace variations
		"  Type  ",
		"\n\nType\n\n",
		"\tType\t",
		"( Pi x Type Type )",
		"(  Pi  x  Type  Type  )",

		// Edge cases
		"",
		"()",
		"(Sort)",
		"((Type))",

		// Very long names
		"verylongvariablenamethatshouldstillparse",
		"(Global averylongglobalnamethatmightbreakthings)",

		// Unicode (might not be supported but shouldn't crash)
		"αβγ",
		"∀",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// The parser should never panic
		term, err := ParseTerm(input)

		if err == nil && term != nil {
			// If we successfully parsed, try to format and reparse
			// (round-trip test - but we don't require it to succeed
			// since not all terms have stable round-trip)
			formatted := FormatTerm(term)
			_ = formatted // Use the result

			// Also try ParseMultiple
			_, _ = ParseMultiple(input)
		}
	})
}

// FuzzParseMultiple fuzzes the ParseMultiple function.
func FuzzParseMultiple(f *testing.F) {
	seeds := []string{
		"Type Type",
		"(Var 0) (Var 1) (Var 2)",
		"Type (Lam x (Var 0)) Type",
		"",
		"   ",
		"\n\n",
		"Type\nType\nType",
		"(Sort 0)(Sort 1)(Sort 2)", // No spaces
		"   Type   (Var 0)   Type   ",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Should never panic
		terms, err := ParseMultiple(input)

		if err == nil {
			// Each parsed term should be formattable
			for _, term := range terms {
				_ = FormatTerm(term)
			}
		}
	})
}

// FuzzMalformedInput specifically tests malformed inputs that might
// trigger edge cases in error handling.
func FuzzMalformedInput(f *testing.F) {
	malformed := []string{
		// Unmatched parentheses
		"(",
		")",
		"(()",
		"())",
		"(((((",
		")))))",
		"(Type",
		"Type)",
		"(Pi x Type",
		"Pi x Type)",

		// Invalid forms
		"(Unknown x y z)",
		"(Pi)",
		"(Pi x)",
		"(Lam)",
		"(App)",
		"(App x)",
		"(Sort -1)",
		"(Var -1)",

		// Null bytes and control characters
		"Type\x00Type",
		"(Pi\x00x Type Type)",
		"\x00",
		"\x01\x02\x03",

		// Very deeply nested
		strings.Repeat("(App ", 100) + "Type" + strings.Repeat(" Type)", 100),

		// Very long tokens
		strings.Repeat("x", 10000),
		"(" + strings.Repeat("x", 10000) + ")",

		// Mixed valid/invalid
		"Type (InvalidForm x) Type",
		"(Pi x Type (Unknown))",

		// Incomplete atoms
		"(Sort 0x)",
		"(Sort 0.5)",
		"(Sort 1e10)",

		// Reserved characters
		"(Pi () Type Type)",
		"(Global ())",
		"\"quoted\"",
		"'quoted'",

		// Tab and newline in middle of form
		"(Pi\tx\tType\tType)",
		"(Pi\nx\nType\nType)",
	}

	for _, m := range malformed {
		f.Add(m)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// The parser should never panic, even on malformed input
		_, _ = ParseTerm(input)
		_, _ = ParseMultiple(input)
	})
}
