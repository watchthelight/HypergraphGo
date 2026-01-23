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
			"Pi implicit",
			"(Pi {A} Type (Var 0))",
			ast.Pi{Binder: "A", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}, Implicit: true},
		},
		{
			"Lam without annotation",
			"(Lam x (Var 0))",
			ast.Lam{Binder: "x", Ann: nil, Body: ast.Var{Ix: 0}},
		},
		{
			"Lam implicit",
			"(Lam {x} (Var 0))",
			ast.Lam{Binder: "x", Ann: nil, Body: ast.Var{Ix: 0}, Implicit: true},
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
		"(Pi {A} Type (Var 0))", // implicit Pi
		"(App succ zero)",
		"(Lam x (Var 0))",
		"(Lam {x} (Var 0))", // implicit Lam
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

// --- Deep Nesting Tests ---

func TestParseTerm_DeepNesting_10Levels(t *testing.T) {
	// Build nested App: (App (App (App ... f x) x) x)
	// Start with innermost: (App f x), then wrap with more Apps
	input := "(App f x)"
	for i := 1; i < 10; i++ {
		input = "(App " + input + " x)"
	}

	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm failed on 10-level nesting: %v", err)
	}

	// Verify structure by walking down the T (function) side
	current := result
	for i := 0; i < 10; i++ {
		app, ok := current.(ast.App)
		if !ok {
			t.Fatalf("Level %d: expected App, got %T", i, current)
		}
		if i < 9 {
			current = app.T
		}
	}
}

func TestParseTerm_DeepNesting_50Levels(t *testing.T) {
	// Build nested Lam: (Lam x (Lam x ... (Var 49) ...))
	input := ""
	for i := 0; i < 50; i++ {
		input += "(Lam x "
	}
	input += "(Var 49)"
	for i := 0; i < 50; i++ {
		input += ")"
	}

	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm failed on 50-level Lam nesting: %v", err)
	}

	// Verify structure
	current := result
	for i := 0; i < 50; i++ {
		lam, ok := current.(ast.Lam)
		if !ok {
			t.Fatalf("Level %d: expected Lam, got %T", i, current)
		}
		current = lam.Body
	}

	// Innermost should be Var 49
	v, ok := current.(ast.Var)
	if !ok || v.Ix != 49 {
		t.Errorf("Innermost term: expected Var 49, got %v", current)
	}
}

func TestParseTerm_DeepNesting_100Levels(t *testing.T) {
	// Build nested Pi: (Pi _ Type (Pi _ Type ...))
	input := ""
	for i := 0; i < 100; i++ {
		input += "(Pi _ Type "
	}
	input += "Type"
	for i := 0; i < 100; i++ {
		input += ")"
	}

	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm failed on 100-level Pi nesting: %v", err)
	}

	// Verify structure
	current := result
	for i := 0; i < 100; i++ {
		pi, ok := current.(ast.Pi)
		if !ok {
			t.Fatalf("Level %d: expected Pi, got %T", i, current)
		}
		current = pi.B
	}
}

func TestFormatTerm_DeepNesting(t *testing.T) {
	// Build 5-level nested term and verify round-trip
	input := "(App (App (App (App (App f x) x) x) x) x)"
	term, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm error: %v", err)
	}

	formatted := FormatTerm(term)
	reparsed, err := ParseTerm(formatted)
	if err != nil {
		t.Fatalf("Reparse error: %v", err)
	}

	if !termEqual(term, reparsed) {
		t.Errorf("Deep nesting round-trip failed")
	}
}

// --- Atom Edge Cases ---

func TestParseTerm_LongAtom_100Chars(t *testing.T) {
	// 100-character identifier
	longName := "abcdefghij"
	for len(longName) < 100 {
		longName += "abcdefghij"
	}
	longName = longName[:100]

	result, err := ParseTerm(longName)
	if err != nil {
		t.Fatalf("ParseTerm(%d chars) error: %v", len(longName), err)
	}

	g, ok := result.(ast.Global)
	if !ok {
		t.Fatalf("Expected Global, got %T", result)
	}
	if g.Name != longName {
		t.Errorf("Name mismatch: got %d chars, want %d chars", len(g.Name), len(longName))
	}
}

func TestParseTerm_LongAtom_1000Chars(t *testing.T) {
	// 1000-character identifier
	longName := ""
	for len(longName) < 1000 {
		longName += "identifier_"
	}
	longName = longName[:1000]

	result, err := ParseTerm(longName)
	if err != nil {
		t.Fatalf("ParseTerm(%d chars) error: %v", len(longName), err)
	}

	g, ok := result.(ast.Global)
	if !ok {
		t.Fatalf("Expected Global, got %T", result)
	}
	if len(g.Name) != 1000 {
		t.Errorf("Name length: got %d, want 1000", len(g.Name))
	}
}

func TestParseTerm_UnicodeIdentifier(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Greek alpha", "α", "α"},
		{"Greek beta", "β", "β"},
		{"Greek gamma", "γ", "γ"},
		{"Greek delta", "δ", "δ"},
		{"Greek omega", "ω", "ω"},
		{"Greek Sigma prefix", "Σ_type", "Σ_type"}, // Σ alone triggers Sigma form
		{"Mixed greek", "αβγ", "αβγ"},
		{"Hebrew aleph", "ℵ", "ℵ"},
		{"Subscript", "x₁", "x₁"},
		{"Superscript", "x²", "x²"},
		{"Japanese hiragana", "あ", "あ"},
		{"Emoji prefix", "✓ok", "✓ok"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTerm(tt.input)
			if err != nil {
				t.Fatalf("ParseTerm(%q) error: %v", tt.input, err)
			}
			g, ok := result.(ast.Global)
			if !ok {
				t.Fatalf("Expected Global, got %T", result)
			}
			if g.Name != tt.expected {
				t.Errorf("Name = %q, want %q", g.Name, tt.expected)
			}
		})
	}
}

func TestParseTerm_UnicodeGlobal(t *testing.T) {
	// Unicode in (Global ...) form
	tests := []string{"αβ", "τype", "ℕat", "ℤero"}
	for _, name := range tests {
		input := "(Global " + name + ")"
		result, err := ParseTerm(input)
		if err != nil {
			t.Fatalf("ParseTerm(%q) error: %v", input, err)
		}
		g, ok := result.(ast.Global)
		if !ok {
			t.Fatalf("Expected Global, got %T", result)
		}
		if g.Name != name {
			t.Errorf("Name = %q, want %q", g.Name, name)
		}
	}
}

func TestParseTerm_NumericEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ast.Term
	}{
		{"Var 0", "(Var 0)", ast.Var{Ix: 0}},
		{"Var large", "(Var 999)", ast.Var{Ix: 999}},
		{"Var very large", "(Var 2147483647)", ast.Var{Ix: 2147483647}},
		{"Sort 0", "(Sort 0)", ast.Sort{U: 0}},
		{"Sort large", "(Sort 100)", ast.Sort{U: 100}},
		{"Shorthand 0", "0", ast.Var{Ix: 0}},
		{"Shorthand 42", "42", ast.Var{Ix: 42}},
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

// --- Alternative Syntax Tests ---

func TestParseTerm_LambdaVariants(t *testing.T) {
	// All these should parse to the same Lam structure
	expected := ast.Lam{Binder: "x", Ann: nil, Body: ast.Var{Ix: 0}}
	variants := []string{
		`(Lam x (Var 0))`,
		`(λ x (Var 0))`,
		`(\ x (Var 0))`,
		`(lambda x (Var 0))`,
	}

	for _, input := range variants {
		t.Run(input, func(t *testing.T) {
			result, err := ParseTerm(input)
			if err != nil {
				t.Fatalf("ParseTerm(%q) error: %v", input, err)
			}
			if !termEqual(result, expected) {
				t.Errorf("ParseTerm(%q) = %v, want %v", input, result, expected)
			}
		})
	}
}

func TestParseTerm_ArrowSyntax(t *testing.T) {
	// -> is an alternative for Pi
	expected := ast.Pi{Binder: "x", A: ast.Global{Name: "Nat"}, B: ast.Global{Name: "Bool"}}
	variants := []string{
		`(Pi x Nat Bool)`,
		`(-> x Nat Bool)`,
	}

	for _, input := range variants {
		t.Run(input, func(t *testing.T) {
			result, err := ParseTerm(input)
			if err != nil {
				t.Fatalf("ParseTerm(%q) error: %v", input, err)
			}
			if !termEqual(result, expected) {
				t.Errorf("ParseTerm(%q) = %v, want %v", input, result, expected)
			}
		})
	}
}

func TestParseTerm_SigmaVariant(t *testing.T) {
	// Σ is an alternative for Sigma
	expected := ast.Sigma{Binder: "x", A: ast.Global{Name: "Nat"}, B: ast.Global{Name: "Bool"}}
	variants := []string{
		`(Sigma x Nat Bool)`,
		`(Σ x Nat Bool)`,
	}

	for _, input := range variants {
		t.Run(input, func(t *testing.T) {
			result, err := ParseTerm(input)
			if err != nil {
				t.Fatalf("ParseTerm(%q) error: %v", input, err)
			}
			if !termEqual(result, expected) {
				t.Errorf("ParseTerm(%q) = %v, want %v", input, result, expected)
			}
		})
	}
}

func TestParseTerm_MixedSyntax(t *testing.T) {
	// Mix of alternative syntax forms in one expression
	tests := []struct {
		name     string
		input    string
		expected ast.Term
	}{
		{
			"Lambda with arrow",
			`(λ f (-> x Nat Nat))`,
			ast.Lam{
				Binder: "f",
				Ann:    nil,
				Body:   ast.Pi{Binder: "x", A: ast.Global{Name: "Nat"}, B: ast.Global{Name: "Nat"}},
			},
		},
		{
			"Arrow with Sigma body",
			`(-> x Nat (Σ y Nat Bool))`,
			ast.Pi{
				Binder: "x",
				A:      ast.Global{Name: "Nat"},
				B:      ast.Sigma{Binder: "y", A: ast.Global{Name: "Nat"}, B: ast.Global{Name: "Bool"}},
			},
		},
		{
			"Nested lambdas",
			`(\ x (λ y (Var 0)))`,
			ast.Lam{
				Binder: "x",
				Ann:    nil,
				Body:   ast.Lam{Binder: "y", Ann: nil, Body: ast.Var{Ix: 0}},
			},
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

// --- Whitespace and Comment Tests ---

func TestParseTerm_WhitespaceOnly(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Spaces only", "   "},
		{"Tabs only", "\t\t\t"},
		{"Newlines only", "\n\n\n"},
		{"Mixed whitespace", "  \t\n  \t\n  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTerm(tt.input)
			if err == nil {
				t.Errorf("ParseTerm(%q) expected error for whitespace-only input", tt.input)
			}
		})
	}
}

func TestParseTerm_MultipleComments(t *testing.T) {
	input := `
; First comment
; Second comment
; Third comment
Type
`
	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm error: %v", err)
	}
	if !termEqual(result, ast.Sort{U: 0}) {
		t.Errorf("Expected Type, got %v", result)
	}
}

func TestParseTerm_CommentAtEnd(t *testing.T) {
	input := "Nat ; trailing comment"
	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm error: %v", err)
	}
	if !termEqual(result, ast.Global{Name: "Nat"}) {
		t.Errorf("Expected Nat, got %v", result)
	}
}

func TestParseTerm_CommentOnly(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Single comment", "; just a comment"},
		{"Comment with newline", "; comment\n"},
		{"Multiple comments only", "; comment 1\n; comment 2\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTerm(tt.input)
			if err == nil {
				t.Errorf("ParseTerm(%q) expected error for comment-only input", tt.input)
			}
		})
	}
}

func TestParseTerm_CommentsInsideExpr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ast.Term
	}{
		{
			"Comment between parens",
			"( ; comment\n Pi x Nat Nat)",
			ast.Pi{Binder: "x", A: ast.Global{Name: "Nat"}, B: ast.Global{Name: "Nat"}},
		},
		{
			"Comment between args",
			"(App ; comment here\n f x)",
			ast.App{T: ast.Global{Name: "f"}, U: ast.Global{Name: "x"}},
		},
		{
			"Comment after each arg",
			"(Sigma ; 1\n x ; 2\n Nat ; 3\n Bool ; 4\n)",
			ast.Sigma{Binder: "x", A: ast.Global{Name: "Nat"}, B: ast.Global{Name: "Bool"}},
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

func TestParseTerm_ExtremeWhitespace(t *testing.T) {
	// Lots of whitespace around a simple term
	input := "\n\n\n   \t\t\t   \n\n\n   Type   \n\n\n   \t\t\t   \n\n\n"
	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm error: %v", err)
	}
	if !termEqual(result, ast.Sort{U: 0}) {
		t.Errorf("Expected Type, got %v", result)
	}
}

// --- Complex Form Tests ---

func TestParseTerm_J_Eliminator(t *testing.T) {
	// J eliminator has 6 arguments: A C d x y p
	input := "(J Nat C d zero zero (Refl Nat zero))"
	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm error: %v", err)
	}

	j, ok := result.(ast.J)
	if !ok {
		t.Fatalf("Expected J, got %T", result)
	}

	// Verify all 6 fields
	if !termEqual(j.A, ast.Global{Name: "Nat"}) {
		t.Errorf("J.A = %v, want Nat", j.A)
	}
	if !termEqual(j.C, ast.Global{Name: "C"}) {
		t.Errorf("J.C = %v, want C", j.C)
	}
	if !termEqual(j.D, ast.Global{Name: "d"}) {
		t.Errorf("J.D = %v, want d", j.D)
	}
	if !termEqual(j.X, ast.Global{Name: "zero"}) {
		t.Errorf("J.X = %v, want zero", j.X)
	}
	if !termEqual(j.Y, ast.Global{Name: "zero"}) {
		t.Errorf("J.Y = %v, want zero", j.Y)
	}
	expectedRefl := ast.Refl{A: ast.Global{Name: "Nat"}, X: ast.Global{Name: "zero"}}
	if !termEqual(j.P, expectedRefl) {
		t.Errorf("J.P = %v, want Refl Nat zero", j.P)
	}
}

func TestParseTerm_Let_Shadowing(t *testing.T) {
	// Nested lets with same binder name (valid, uses shadowing)
	input := "(Let x Nat zero (Let x Nat (Var 0) (Var 0)))"
	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm error: %v", err)
	}

	outer, ok := result.(ast.Let)
	if !ok {
		t.Fatalf("Expected Let, got %T", result)
	}
	if outer.Binder != "x" {
		t.Errorf("Outer.Binder = %q, want x", outer.Binder)
	}

	inner, ok := outer.Body.(ast.Let)
	if !ok {
		t.Fatalf("Expected inner Let, got %T", outer.Body)
	}
	if inner.Binder != "x" {
		t.Errorf("Inner.Binder = %q, want x", inner.Binder)
	}
}

func TestParseTerm_EmptyBinder(t *testing.T) {
	// _ is a valid binder for unused variables
	tests := []struct {
		name  string
		input string
	}{
		{"Pi with underscore", "(Pi _ Nat Nat)"},
		{"Lam with underscore", "(Lam _ (Var 0))"},
		{"Sigma with underscore", "(Sigma _ Nat Nat)"},
		{"Let with underscore", "(Let _ Nat zero (Var 0))"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTerm(tt.input)
			if err != nil {
				t.Fatalf("ParseTerm(%q) error: %v", tt.input, err)
			}
			// Just check that it parses; the binder should be "_"
			switch v := result.(type) {
			case ast.Pi:
				if v.Binder != "_" {
					t.Errorf("Expected binder '_', got %q", v.Binder)
				}
			case ast.Lam:
				if v.Binder != "_" {
					t.Errorf("Expected binder '_', got %q", v.Binder)
				}
			case ast.Sigma:
				if v.Binder != "_" {
					t.Errorf("Expected binder '_', got %q", v.Binder)
				}
			case ast.Let:
				if v.Binder != "_" {
					t.Errorf("Expected binder '_', got %q", v.Binder)
				}
			}
		})
	}
}

func TestParseTerm_LamWithAnnotation(t *testing.T) {
	// Lam can have an optional type annotation when the type is in parens
	// and starts with a recognized form like (Pi ...), (Sort ...), etc.
	tests := []struct {
		name     string
		input    string
		binder   string
		hasAnn   bool
		annTerm  ast.Term
		bodyTerm ast.Term
	}{
		{
			"Lam without annotation",
			"(Lam x (Var 0))",
			"x",
			false,
			nil,
			ast.Var{Ix: 0},
		},
		{
			"Lam with Sort annotation",
			"(Lam x (Sort 0) (Var 0))",
			"x",
			true,
			ast.Sort{U: 0},
			ast.Var{Ix: 0},
		},
		{
			"Lam with Pi annotation",
			"(Lam f (Pi x Nat Nat) (App (Var 0) zero))",
			"f",
			true,
			ast.Pi{Binder: "x", A: ast.Global{Name: "Nat"}, B: ast.Global{Name: "Nat"}},
			ast.App{T: ast.Var{Ix: 0}, U: ast.Global{Name: "zero"}},
		},
		{
			"Lam with Sigma annotation",
			"(Lam p (Sigma x Nat Nat) (Fst (Var 0)))",
			"p",
			true,
			ast.Sigma{Binder: "x", A: ast.Global{Name: "Nat"}, B: ast.Global{Name: "Nat"}},
			ast.Fst{P: ast.Var{Ix: 0}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTerm(tt.input)
			if err != nil {
				t.Fatalf("ParseTerm error: %v", err)
			}

			lam, ok := result.(ast.Lam)
			if !ok {
				t.Fatalf("Expected Lam, got %T", result)
			}
			if lam.Binder != tt.binder {
				t.Errorf("Binder = %q, want %q", lam.Binder, tt.binder)
			}
			if tt.hasAnn {
				if !termEqual(lam.Ann, tt.annTerm) {
					t.Errorf("Ann = %v, want %v", lam.Ann, tt.annTerm)
				}
			} else {
				if lam.Ann != nil {
					t.Errorf("Expected no annotation, got %v", lam.Ann)
				}
			}
			if !termEqual(lam.Body, tt.bodyTerm) {
				t.Errorf("Body = %v, want %v", lam.Body, tt.bodyTerm)
			}
		})
	}
}

func TestParseTerm_NestedApplications_50(t *testing.T) {
	// Build a chain of 50 applications
	input := "f"
	for i := 0; i < 50; i++ {
		input = "(App " + input + " x)"
	}

	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm failed on 50 nested applications: %v", err)
	}

	// Verify it's a chain of Apps
	current := result
	for i := 0; i < 50; i++ {
		app, ok := current.(ast.App)
		if !ok {
			t.Fatalf("Level %d: expected App, got %T", i, current)
		}
		if !termEqual(app.U, ast.Global{Name: "x"}) {
			t.Errorf("Level %d: arg should be x", i)
		}
		current = app.T
	}

	// Innermost should be f
	if !termEqual(current, ast.Global{Name: "f"}) {
		t.Errorf("Innermost: expected f, got %v", current)
	}
}

func TestParseTerm_ComplexNesting(t *testing.T) {
	// A realistically complex term
	input := `(Pi A Type
               (Pi x (Var 0)
                 (Sigma y (Var 1)
                   (Id (Var 2) (Var 1) (Var 0)))))`
	result, err := ParseTerm(input)
	if err != nil {
		t.Fatalf("ParseTerm error: %v", err)
	}

	// Verify outer structure
	pi1, ok := result.(ast.Pi)
	if !ok {
		t.Fatalf("Expected Pi, got %T", result)
	}
	if pi1.Binder != "A" {
		t.Errorf("First Pi binder = %q, want A", pi1.Binder)
	}

	pi2, ok := pi1.B.(ast.Pi)
	if !ok {
		t.Fatalf("Expected nested Pi, got %T", pi1.B)
	}
	if pi2.Binder != "x" {
		t.Errorf("Second Pi binder = %q, want x", pi2.Binder)
	}

	sigma, ok := pi2.B.(ast.Sigma)
	if !ok {
		t.Fatalf("Expected Sigma, got %T", pi2.B)
	}
	if sigma.Binder != "y" {
		t.Errorf("Sigma binder = %q, want y", sigma.Binder)
	}

	id, ok := sigma.B.(ast.Id)
	if !ok {
		t.Fatalf("Expected Id, got %T", sigma.B)
	}
	if !termEqual(id.A, ast.Var{Ix: 2}) {
		t.Errorf("Id.A = %v, want Var 2", id.A)
	}
}

// --- Malformed Input Tests ---

func TestParseTerm_UnmatchedCloseParen(t *testing.T) {
	tests := []string{
		")",
		"Type)",
		"(Pi x Nat Nat))",
		"(App f x))",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseTerm(input)
			if err == nil {
				t.Errorf("ParseTerm(%q) expected error for unmatched close paren", input)
			}
		})
	}
}

func TestParseTerm_UnmatchedOpenParen(t *testing.T) {
	tests := []string{
		"(",
		"((",
		"(Pi x Nat",
		"(App (App f x)",
		"(Lam x (Var 0",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseTerm(input)
			if err == nil {
				t.Errorf("ParseTerm(%q) expected error for unmatched open paren", input)
			}
		})
	}
}

func TestParseTerm_MissingArgs(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Pi missing 2 args", "(Pi x)"},
		{"Pi missing 1 arg", "(Pi x Nat)"},
		{"App missing arg", "(App f)"},
		{"App empty", "(App)"},
		{"Sigma missing args", "(Sigma x)"},
		{"Pair missing arg", "(Pair x)"},
		{"Fst empty", "(Fst)"},
		{"Snd empty", "(Snd)"},
		{"Let missing body", "(Let x Nat zero)"},
		{"Id missing args", "(Id Nat x)"},
		{"Refl missing arg", "(Refl Nat)"},
		{"J missing args", "(J Nat C d x y)"},
		{"Var empty", "(Var)"},
		// Note: (Sort) parses as Sort{U:0}, (Global) as Global{Name:""} - not errors
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTerm(tt.input)
			if err == nil {
				t.Errorf("ParseTerm(%q) expected error for missing args", tt.input)
			}
		})
	}
}

func TestParseTerm_BadVarIndex(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		// Note: (Var -1) parses -1 as an atom, not a negative number
		{"Non-numeric", "(Var abc)"},
		{"Float", "(Var 1.5)"},
		{"Hex", "(Var 0x10)"},
		{"Empty atom after Var", "(Var )"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTerm(tt.input)
			if err == nil {
				t.Errorf("ParseTerm(%q) expected error for bad Var index", tt.input)
			}
		})
	}
}

func TestParseTerm_UnknownForm(t *testing.T) {
	tests := []string{
		"(Unknown x y)",
		"(Foo)",
		"(lambda2 x body)",
		"(Π x Nat Nat)", // Π instead of Pi - not a recognized form
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseTerm(input)
			if err == nil {
				t.Errorf("ParseTerm(%q) expected error for unknown form", input)
			}
		})
	}
}

func TestParseTerm_ExtraTrailingContent(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"Extra atom", "Type extra"},
		{"Extra paren", "Nat (extra)"},
		{"Two terms", "(Pi x Nat Nat) (App f x)"},
		{"Term then garbage", "Type xyz123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseTerm(tt.input)
			if err == nil {
				t.Errorf("ParseTerm(%q) expected error for trailing content", tt.input)
			}
		})
	}
}

func TestParseTerm_EmptyParens(t *testing.T) {
	_, err := ParseTerm("()")
	if err == nil {
		t.Error("ParseTerm(\"()\") expected error for empty parens")
	}
}

func TestParseTerm_NestedEmptyParens(t *testing.T) {
	tests := []string{
		"(App () x)",
		"(Pi x () Nat)",
		"(Lam x ())",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, err := ParseTerm(input)
			if err == nil {
				t.Errorf("ParseTerm(%q) expected error", input)
			}
		})
	}
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
			return av.Binder == bv.Binder && termEqual(av.A, bv.A) && termEqual(av.B, bv.B) && av.Implicit == bv.Implicit
		}
	case ast.Lam:
		if bv, ok := b.(ast.Lam); ok {
			return av.Binder == bv.Binder && termEqual(av.Ann, bv.Ann) && termEqual(av.Body, bv.Body) && av.Implicit == bv.Implicit
		}
	case ast.App:
		if bv, ok := b.(ast.App); ok {
			return termEqual(av.T, bv.T) && termEqual(av.U, bv.U) && av.Implicit == bv.Implicit
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

// ============================================================================
// FormatTerm Tests
// ============================================================================

func TestFormatTerm_Nil(t *testing.T) {
	result := FormatTerm(nil)
	if result != "nil" {
		t.Errorf("FormatTerm(nil) = %q, want %q", result, "nil")
	}
}

func TestFormatTerm_LamWithoutAnn(t *testing.T) {
	// Test Lam without annotation
	lam := ast.Lam{Binder: "x", Ann: nil, Body: ast.Var{Ix: 0}}
	result := FormatTerm(lam)
	expected := "(Lam x (Var 0))"
	if result != expected {
		t.Errorf("FormatTerm(Lam without ann) = %q, want %q", result, expected)
	}
}

func TestFormatTerm_SortHigherLevel(t *testing.T) {
	// Test Sort with non-zero level
	sort := ast.Sort{U: 5}
	result := FormatTerm(sort)
	expected := "(Sort 5)"
	if result != expected {
		t.Errorf("FormatTerm(Sort 5) = %q, want %q", result, expected)
	}
}

// ============================================================================
// Normalize Tests
// ============================================================================

func TestNormalize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"trim leading space", "  Type", "Type"},
		{"trim trailing space", "Type  ", "Type"},
		{"trim both", "  Type  ", "Type"},
		{"already clean", "Type", "Type"},
		{"empty input", "", ""},
		{"only whitespace", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Normalize(tt.input)
			if result != tt.expected {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Additional Error Path Tests
// ============================================================================

func TestParseJ_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"error in A", "(J ("},
		{"error in C", "(J Nat ("},
		{"error in D", "(J Nat C ("},
		{"error in X", "(J Nat C D ("},
		{"error in Y", "(J Nat C D X ("},
		{"error in P", "(J Nat C D X Y ("},
		{"missing args", "(J)"},
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

func TestParseSigma_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"error in A", "(Sigma x ("},
		{"error in B", "(Sigma x Nat ("},
		{"missing args", "(Sigma)"},
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

func TestParseLet_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"error in ann", "(Let x ("},
		{"error in val", "(Let x Nat ("},
		{"error in body", "(Let x Nat zero ("},
		{"missing args", "(Let)"},
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

func TestParseId_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"error in A", "(Id ("},
		{"error in X", "(Id Nat ("},
		{"error in Y", "(Id Nat zero ("},
		{"missing args", "(Id)"},
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

func TestParseRefl_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"error in A", "(Refl ("},
		{"error in X", "(Refl Nat ("},
		{"missing args", "(Refl)"},
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

func TestParsePair_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"error in fst", "(Pair ("},
		{"error in snd", "(Pair zero ("},
		{"missing args", "(Pair)"},
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

func TestParseGlobal_WithParen(t *testing.T) {
	// (Global foo) should parse to Global{Name: "foo"}
	result, err := ParseTerm("(Global foo)")
	if err != nil {
		t.Fatalf("ParseTerm error: %v", err)
	}
	g, ok := result.(ast.Global)
	if !ok {
		t.Fatalf("Expected Global, got %T", result)
	}
	if g.Name != "foo" {
		t.Errorf("Global.Name = %q, want %q", g.Name, "foo")
	}
}

func TestParseSort_InvalidLevel(t *testing.T) {
	// (Sort abc) should fail - non-numeric level
	_, err := ParseTerm("(Sort abc)")
	if err == nil {
		t.Error("ParseTerm((Sort abc)) expected error, got nil")
	}
}

// ============================================================================
// Context-Aware Formatting Tests
// ============================================================================

func TestFormatTermWithContext_Var(t *testing.T) {
	// Var 0 with context ["x"] should print as "x"
	term := ast.Var{Ix: 0}
	result := FormatTermWithContext(term, []string{"x"})
	if result != "x" {
		t.Errorf("FormatTermWithContext(Var 0, [x]) = %q, want %q", result, "x")
	}
}

func TestFormatTermWithContext_VarMultiple(t *testing.T) {
	// Var 0 with context ["A", "x"] should print as "x" (innermost)
	// Var 1 with context ["A", "x"] should print as "A"
	ctx := []string{"A", "x"}

	term0 := ast.Var{Ix: 0}
	result0 := FormatTermWithContext(term0, ctx)
	if result0 != "x" {
		t.Errorf("FormatTermWithContext(Var 0, [A, x]) = %q, want %q", result0, "x")
	}

	term1 := ast.Var{Ix: 1}
	result1 := FormatTermWithContext(term1, ctx)
	if result1 != "A" {
		t.Errorf("FormatTermWithContext(Var 1, [A, x]) = %q, want %q", result1, "A")
	}
}

func TestFormatTermWithContext_VarOutOfBounds(t *testing.T) {
	// Var 2 with context ["x"] should fall back to "(Var 2)"
	term := ast.Var{Ix: 2}
	result := FormatTermWithContext(term, []string{"x"})
	if result != "(Var 2)" {
		t.Errorf("FormatTermWithContext(Var 2, [x]) = %q, want %q", result, "(Var 2)")
	}
}

func TestFormatTermWithContext_VarUnderscore(t *testing.T) {
	// Var 0 with context ["_"] should fall back to "(Var 0)"
	term := ast.Var{Ix: 0}
	result := FormatTermWithContext(term, []string{"_"})
	if result != "(Var 0)" {
		t.Errorf("FormatTermWithContext(Var 0, [_]) = %q, want %q", result, "(Var 0)")
	}
}

func TestFormatTermWithContext_Pi(t *testing.T) {
	// (Pi A Type (Pi x (Var 0) (Var 1))) with ctx [] should show:
	// (Pi A Type (Pi x A A))
	term := ast.Pi{
		Binder: "A",
		A:      ast.Sort{U: 0},
		B: ast.Pi{
			Binder: "x",
			A:      ast.Var{Ix: 0},
			B:      ast.Var{Ix: 1},
		},
	}
	result := FormatTermWithContext(term, nil)
	expected := "(Pi A Type (Pi x A A))"
	if result != expected {
		t.Errorf("FormatTermWithContext = %q, want %q", result, expected)
	}
}

func TestFormatTermWithContext_Lam(t *testing.T) {
	// (Lam A (Lam x (Var 0))) with ctx [] should show:
	// (Lam A (Lam x x))
	term := ast.Lam{
		Binder: "A",
		Body: ast.Lam{
			Binder: "x",
			Body:   ast.Var{Ix: 0},
		},
	}
	result := FormatTermWithContext(term, nil)
	expected := "(Lam A (Lam x x))"
	if result != expected {
		t.Errorf("FormatTermWithContext = %q, want %q", result, expected)
	}
}

func TestFormatTermWithContext_Id(t *testing.T) {
	// (Id Nat (Var 0) (Var 0)) with ctx ["n"] should show:
	// (Id Nat n n)
	term := ast.Id{
		A: ast.Global{Name: "Nat"},
		X: ast.Var{Ix: 0},
		Y: ast.Var{Ix: 0},
	}
	result := FormatTermWithContext(term, []string{"n"})
	expected := "(Id Nat n n)"
	if result != expected {
		t.Errorf("FormatTermWithContext = %q, want %q", result, expected)
	}
}

func TestFormatTermWithContext_App(t *testing.T) {
	// (App f (Var 0)) with ctx ["x"] should show:
	// (App f x)
	term := ast.App{
		T: ast.Global{Name: "f"},
		U: ast.Var{Ix: 0},
	}
	result := FormatTermWithContext(term, []string{"x"})
	expected := "(App f x)"
	if result != expected {
		t.Errorf("FormatTermWithContext = %q, want %q", result, expected)
	}
}

func TestFormatTermWithContext_FullExample(t *testing.T) {
	// Full example: (Pi n Nat (Id Nat (App (App add (Var 0)) zero) (Var 0)))
	// Should format with ctx [] as: (Pi n Nat (Id Nat (App (App add n) zero) n))
	term := ast.Pi{
		Binder: "n",
		A:      ast.Global{Name: "Nat"},
		B: ast.Id{
			A: ast.Global{Name: "Nat"},
			X: ast.App{
				T: ast.App{
					T: ast.Global{Name: "add"},
					U: ast.Var{Ix: 0},
				},
				U: ast.Global{Name: "zero"},
			},
			Y: ast.Var{Ix: 0},
		},
	}
	result := FormatTermWithContext(term, nil)
	expected := "(Pi n Nat (Id Nat (App (App add n) zero) n))"
	if result != expected {
		t.Errorf("FormatTermWithContext = %q, want %q", result, expected)
	}
}
