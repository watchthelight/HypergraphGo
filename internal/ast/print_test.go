package ast

import (
	"strings"
	"testing"
)

// ============================================================================
// Sprint Basic Tests
// ============================================================================

// TestSprint_Sort tests Sort printing
func TestSprint_Sort(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{0, "Type0"},
		{1, "Type1"},
		{10, "Type10"},
	}

	for _, tt := range tests {
		result := Sprint(Sort{U: tt.level})
		if result != tt.expected {
			t.Errorf("Sprint(Sort{%d}) = %q, want %q", tt.level, result, tt.expected)
		}
	}
}

// TestSprint_Var tests Var printing
func TestSprint_Var(t *testing.T) {
	tests := []struct {
		ix       int
		expected string
	}{
		{0, "{0}"},
		{5, "{5}"},
		{100, "{100}"},
	}

	for _, tt := range tests {
		result := Sprint(Var{Ix: tt.ix})
		if result != tt.expected {
			t.Errorf("Sprint(Var{%d}) = %q, want %q", tt.ix, result, tt.expected)
		}
	}
}

// TestSprint_Global tests Global printing
func TestSprint_Global(t *testing.T) {
	tests := []string{"Nat", "Bool", "zero", "succ", "myFunction"}

	for _, name := range tests {
		result := Sprint(Global{Name: name})
		if result != name {
			t.Errorf("Sprint(Global{%q}) = %q, want %q", name, result, name)
		}
	}
}

// TestSprint_Pi tests Pi printing
func TestSprint_Pi(t *testing.T) {
	// (Pi x: A . B)
	pi := Pi{
		Binder: "x",
		A:      Global{Name: "Nat"},
		B:      Global{Name: "Nat"},
	}

	result := Sprint(pi)
	if !strings.Contains(result, "Pi") {
		t.Error("Pi output should contain 'Pi'")
	}
	if !strings.Contains(result, "x") {
		t.Error("Pi output should contain binder 'x'")
	}
	if !strings.Contains(result, "Nat") {
		t.Error("Pi output should contain 'Nat'")
	}
}

// TestSprint_Pi_EmptyBinder tests Pi with empty binder
func TestSprint_Pi_EmptyBinder(t *testing.T) {
	pi := Pi{
		Binder: "",
		A:      Global{Name: "A"},
		B:      Global{Name: "B"},
	}

	result := Sprint(pi)
	if !strings.Contains(result, "_") {
		t.Error("Pi with empty binder should use '_'")
	}
}

// TestSprint_Lam tests Lambda printing
func TestSprint_Lam(t *testing.T) {
	// Annotated lambda
	lam := Lam{
		Binder: "x",
		Ann:    Global{Name: "Nat"},
		Body:   Var{Ix: 0},
	}

	result := Sprint(lam)
	if !strings.Contains(result, "\\") {
		t.Error("Lam output should contain '\\'")
	}
	if !strings.Contains(result, "x") {
		t.Error("Lam output should contain binder 'x'")
	}
	if !strings.Contains(result, "Nat") {
		t.Error("Annotated Lam should contain annotation")
	}
}

// TestSprint_Lam_Unannotated tests unannotated Lambda
func TestSprint_Lam_Unannotated(t *testing.T) {
	lam := Lam{
		Binder: "y",
		Body:   Var{Ix: 0},
	}

	result := Sprint(lam)
	if !strings.Contains(result, "\\") {
		t.Error("Lam output should contain '\\'")
	}
	if !strings.Contains(result, "y") {
		t.Error("Lam output should contain binder 'y'")
	}
}

// TestSprint_App tests Application printing
func TestSprint_App(t *testing.T) {
	// (f x)
	app := App{
		T: Global{Name: "f"},
		U: Global{Name: "x"},
	}

	result := Sprint(app)
	if !strings.Contains(result, "f") || !strings.Contains(result, "x") {
		t.Errorf("App output should contain 'f' and 'x': got %q", result)
	}
}

// TestSprint_App_Nested tests nested applications
func TestSprint_App_Nested(t *testing.T) {
	// (f x y z) = App(App(App(f, x), y), z)
	app := MkApps(Global{Name: "f"}, Global{Name: "x"}, Global{Name: "y"}, Global{Name: "z"})

	result := Sprint(app)
	if !strings.Contains(result, "f") {
		t.Error("Should contain 'f'")
	}
	if !strings.Contains(result, "x") {
		t.Error("Should contain 'x'")
	}
	if !strings.Contains(result, "y") {
		t.Error("Should contain 'y'")
	}
	if !strings.Contains(result, "z") {
		t.Error("Should contain 'z'")
	}
}

// TestSprint_Sigma tests Sigma printing
func TestSprint_Sigma(t *testing.T) {
	sigma := Sigma{
		Binder: "x",
		A:      Global{Name: "Nat"},
		B:      Global{Name: "Bool"},
	}

	result := Sprint(sigma)
	if !strings.Contains(result, "Sigma") {
		t.Error("Sigma output should contain 'Sigma'")
	}
}

// TestSprint_Pair tests Pair printing
func TestSprint_Pair(t *testing.T) {
	pair := Pair{
		Fst: Global{Name: "x"},
		Snd: Global{Name: "y"},
	}

	result := Sprint(pair)
	if !strings.Contains(result, ",") {
		t.Error("Pair output should contain ','")
	}
}

// TestSprint_Fst tests Fst printing
func TestSprint_Fst(t *testing.T) {
	fst := Fst{P: Global{Name: "p"}}

	result := Sprint(fst)
	if !strings.Contains(result, "fst") {
		t.Error("Fst output should contain 'fst'")
	}
}

// TestSprint_Snd tests Snd printing
func TestSprint_Snd(t *testing.T) {
	snd := Snd{P: Global{Name: "p"}}

	result := Sprint(snd)
	if !strings.Contains(result, "snd") {
		t.Error("Snd output should contain 'snd'")
	}
}

// TestSprint_Let tests Let printing
func TestSprint_Let(t *testing.T) {
	let := Let{
		Binder: "x",
		Ann:    Global{Name: "Nat"},
		Val:    Global{Name: "zero"},
		Body:   Var{Ix: 0},
	}

	result := Sprint(let)
	if !strings.Contains(result, "let") {
		t.Error("Let output should contain 'let'")
	}
	if !strings.Contains(result, "in") {
		t.Error("Let output should contain 'in'")
	}
}

// TestSprint_Let_Unannotated tests Let without annotation
func TestSprint_Let_Unannotated(t *testing.T) {
	let := Let{
		Binder: "x",
		Val:    Global{Name: "zero"},
		Body:   Var{Ix: 0},
	}

	result := Sprint(let)
	if !strings.Contains(result, "let") {
		t.Error("Let output should contain 'let'")
	}
}

// TestSprint_Id tests Id printing
func TestSprint_Id(t *testing.T) {
	id := Id{
		A: Global{Name: "Nat"},
		X: Global{Name: "zero"},
		Y: Global{Name: "zero"},
	}

	result := Sprint(id)
	if !strings.Contains(result, "Id") {
		t.Error("Id output should contain 'Id'")
	}
}

// TestSprint_Refl tests Refl printing
func TestSprint_Refl(t *testing.T) {
	refl := Refl{
		A: Global{Name: "Nat"},
		X: Global{Name: "zero"},
	}

	result := Sprint(refl)
	if !strings.Contains(result, "refl") {
		t.Error("Refl output should contain 'refl'")
	}
}

// TestSprint_J tests J printing
func TestSprint_J(t *testing.T) {
	j := J{
		A: Global{Name: "Nat"},
		C: Global{Name: "C"},
		D: Global{Name: "d"},
		X: Global{Name: "x"},
		Y: Global{Name: "y"},
		P: Global{Name: "p"},
	}

	result := Sprint(j)
	if !strings.Contains(result, "J") {
		t.Error("J output should contain 'J'")
	}
}

// ============================================================================
// Complex Term Tests
// ============================================================================

// TestSprint_IdentityFunction tests identity function
func TestSprint_IdentityFunction(t *testing.T) {
	// λA. λx. x
	id := Lam{
		Binder: "A",
		Ann:    Sort{U: 0},
		Body: Lam{
			Binder: "x",
			Ann:    Var{Ix: 0},
			Body:   Var{Ix: 0},
		},
	}

	result := Sprint(id)
	if result == "" {
		t.Error("Sprint returned empty for identity function")
	}
	// Just verify it doesn't panic and produces output
}

// TestSprint_DependentPair tests dependent pair type
func TestSprint_DependentPair(t *testing.T) {
	// Σ(n : Nat). Vec n
	sigma := Sigma{
		Binder: "n",
		A:      Global{Name: "Nat"},
		B:      App{T: Global{Name: "Vec"}, U: Var{Ix: 0}},
	}

	result := Sprint(sigma)
	if result == "" {
		t.Error("Sprint returned empty for dependent pair")
	}
}

// TestSprint_NestedPi tests deeply nested Pi types
func TestSprint_NestedPi(t *testing.T) {
	// Π(A:Type). Π(B:Type). Π(f: A->B). A -> B
	nested := Pi{
		Binder: "A", A: Sort{U: 0},
		B: Pi{
			Binder: "B", A: Sort{U: 0},
			B: Pi{
				Binder: "f",
				A:      Pi{Binder: "_", A: Var{Ix: 1}, B: Var{Ix: 1}},
				B:      Pi{Binder: "_", A: Var{Ix: 2}, B: Var{Ix: 2}},
			},
		},
	}

	result := Sprint(nested)
	if result == "" {
		t.Error("Sprint returned empty for nested Pi")
	}
}

// ============================================================================
// Edge Cases
// ============================================================================

// TestSprint_EmptyBinders tests terms with empty binders
func TestSprint_EmptyBinders(t *testing.T) {
	terms := []Term{
		Pi{Binder: "", A: Sort{U: 0}, B: Sort{U: 0}},
		Lam{Binder: "", Body: Var{Ix: 0}},
		Sigma{Binder: "", A: Sort{U: 0}, B: Sort{U: 0}},
		Let{Binder: "", Val: Global{Name: "x"}, Body: Var{Ix: 0}},
	}

	for _, term := range terms {
		result := Sprint(term)
		if result == "" {
			t.Errorf("Sprint returned empty for %T with empty binder", term)
		}
		// Empty binder should be replaced with "_"
		if !strings.Contains(result, "_") {
			t.Errorf("Empty binder not replaced with '_' in %T: %s", term, result)
		}
	}
}

// TestSprint_Nil tests that Sprint handles nil gracefully
// (if it does - otherwise this tests the expected behavior)
func TestSprint_OutputNonEmpty(t *testing.T) {
	terms := []Term{
		Sort{U: 0},
		Var{Ix: 0},
		Global{Name: "x"},
		Pi{Binder: "x", A: Sort{U: 0}, B: Sort{U: 0}},
		Lam{Binder: "x", Body: Var{Ix: 0}},
		App{T: Global{Name: "f"}, U: Global{Name: "x"}},
		Sigma{Binder: "x", A: Sort{U: 0}, B: Sort{U: 0}},
		Pair{Fst: Global{Name: "a"}, Snd: Global{Name: "b"}},
		Fst{P: Global{Name: "p"}},
		Snd{P: Global{Name: "p"}},
		Let{Binder: "x", Val: Global{Name: "v"}, Body: Var{Ix: 0}},
		Id{A: Sort{U: 0}, X: Global{Name: "x"}, Y: Global{Name: "y"}},
		Refl{A: Sort{U: 0}, X: Global{Name: "x"}},
		J{A: Sort{U: 0}, C: Global{Name: "C"}, D: Global{Name: "d"},
			X: Global{Name: "x"}, Y: Global{Name: "y"}, P: Global{Name: "p"}},
	}

	for _, term := range terms {
		result := Sprint(term)
		if result == "" {
			t.Errorf("Sprint returned empty for %T", term)
		}
	}
}

// ============================================================================
// collectSpine Tests
// ============================================================================

// TestCollectSpine tests the collectSpine helper
func TestCollectSpine(t *testing.T) {
	// Single application: (f x)
	app1 := App{T: Global{Name: "f"}, U: Global{Name: "x"}}
	fun, args := collectSpine(app1)

	if g, ok := fun.(Global); !ok || g.Name != "f" {
		t.Errorf("Expected Global{f}, got %T", fun)
	}
	if len(args) != 1 {
		t.Errorf("Expected 1 arg, got %d", len(args))
	}

	// Nested: (f x y) = App(App(f, x), y)
	app2 := App{T: App{T: Global{Name: "g"}, U: Global{Name: "a"}}, U: Global{Name: "b"}}
	fun2, args2 := collectSpine(app2)

	if g, ok := fun2.(Global); !ok || g.Name != "g" {
		t.Errorf("Expected Global{g}, got %T", fun2)
	}
	if len(args2) != 2 {
		t.Errorf("Expected 2 args, got %d", len(args2))
	}

	// Non-application
	nonApp := Global{Name: "h"}
	fun3, args3 := collectSpine(nonApp)

	if g, ok := fun3.(Global); !ok || g.Name != "h" {
		t.Errorf("Expected Global{h}, got %T", fun3)
	}
	if len(args3) != 0 {
		t.Errorf("Expected 0 args for non-app, got %d", len(args3))
	}
}

// ============================================================================
// Round-Trip Style Tests
// ============================================================================

// TestSprint_ContainsExpectedParts verifies Sprint output structure
func TestSprint_ContainsExpectedParts(t *testing.T) {
	// Identity type representation
	result := Sprint(Pi{
		Binder: "A",
		A:      Sort{U: 0},
		B: Pi{
			Binder: "x",
			A:      Var{Ix: 0},
			B: Pi{
				Binder: "_",
				A:      Var{Ix: 1},
				B:      Var{Ix: 2},
			},
		},
	})

	// Should have basic structure elements
	if !strings.HasPrefix(result, "(Pi") {
		t.Error("Pi should start with '(Pi'")
	}
	if !strings.HasSuffix(result, ")") {
		t.Error("Term should end with ')'")
	}
}
