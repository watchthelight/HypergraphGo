package eval

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// TestEvalNilEnv tests that Eval handles nil environment gracefully.
func TestEvalNilEnv(t *testing.T) {
	term := ast.Global{Name: "x"}
	result := Eval(nil, term)

	// Should not panic and return a valid value
	if result == nil {
		t.Error("Eval(nil, term) should not return nil")
	}

	// Global should evaluate to VNeutral with global head
	if _, ok := result.(VNeutral); !ok {
		t.Errorf("expected VNeutral for Global, got %T", result)
	}
}

// TestEvalNilTerm tests that Eval handles nil term gracefully.
func TestEvalNilTerm(t *testing.T) {
	env := &Env{Bindings: nil}
	result := Eval(env, nil)

	// Should return VGlobal{"nil"}
	if g, ok := result.(VGlobal); !ok || g.Name != "nil" {
		t.Errorf("Eval(env, nil) should return VGlobal{nil}, got %v", result)
	}
}

// TestEvalLetBinding tests Let expression evaluation.
func TestEvalLetBinding(t *testing.T) {
	// let x = Global{a} in x
	term := ast.Let{
		Binder: "x",
		Ann:    ast.Sort{U: 0},
		Val:    ast.Global{Name: "a"},
		Body:   ast.Var{Ix: 0}, // x
	}

	env := &Env{Bindings: nil}
	result := Eval(env, term)

	// Should evaluate to VNeutral with Global{a} head
	if n, ok := result.(VNeutral); ok {
		if n.N.Head.Glob != "a" {
			t.Errorf("expected global head 'a', got %q", n.N.Head.Glob)
		}
	} else {
		t.Errorf("expected VNeutral, got %T", result)
	}
}

// TestEvalPiType tests Pi type evaluation.
func TestEvalPiType(t *testing.T) {
	// Pi (x : Sort 0) . Sort 1
	term := ast.Pi{
		Binder: "x",
		A:      ast.Sort{U: 0},
		B:      ast.Sort{U: 1},
	}

	env := &Env{Bindings: nil}
	result := Eval(env, term)

	pi, ok := result.(VPi)
	if !ok {
		t.Fatalf("expected VPi, got %T", result)
	}

	// Domain should be VSort{0}
	if s, ok := pi.A.(VSort); !ok || s.Level != 0 {
		t.Errorf("expected VSort{0} for domain, got %v", pi.A)
	}

	// Codomain should be a closure
	if pi.B == nil {
		t.Error("expected non-nil closure for codomain")
	}
}

// TestEvalSigmaType tests Sigma type evaluation.
func TestEvalSigmaType(t *testing.T) {
	// Sigma (x : Sort 0) . Sort 1
	term := ast.Sigma{
		Binder: "x",
		A:      ast.Sort{U: 0},
		B:      ast.Sort{U: 1},
	}

	env := &Env{Bindings: nil}
	result := Eval(env, term)

	sigma, ok := result.(VSigma)
	if !ok {
		t.Fatalf("expected VSigma, got %T", result)
	}

	// Domain should be VSort{0}
	if s, ok := sigma.A.(VSort); !ok || s.Level != 0 {
		t.Errorf("expected VSort{0} for domain, got %v", sigma.A)
	}
}

// TestApplyNonFunction tests Apply behavior on non-function values.
func TestApplyNonFunction(t *testing.T) {
	tests := []struct {
		name string
		val  Value
	}{
		{"VSort", VSort{Level: 0}},
		{"VGlobal", VGlobal{Name: "x"}},
		{"VPair", VPair{Fst: VSort{Level: 0}, Snd: VSort{Level: 1}}},
	}

	arg := VSort{Level: 0}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Apply(tt.val, arg)

			// Should return a neutral term (bad_app)
			if n, ok := result.(VNeutral); ok {
				if n.N.Head.Glob != "bad_app" {
					t.Errorf("expected bad_app neutral, got head %q", n.N.Head.Glob)
				}
			} else {
				t.Errorf("expected VNeutral for non-function application, got %T", result)
			}
		})
	}
}

// TestReifyAllValueTypes tests Reify produces valid AST for all Value types.
func TestReifyAllValueTypes(t *testing.T) {
	tests := []struct {
		name string
		val  Value
	}{
		{"VSort", VSort{Level: 0}},
		{"VGlobal", VGlobal{Name: "x"}},
		{"VNeutral_Var", VNeutral{N: Neutral{Head: Head{Var: 0}}}},
		{"VNeutral_Global", VNeutral{N: Neutral{Head: Head{Glob: "g"}}}},
		{"VPair", VPair{Fst: VSort{Level: 0}, Snd: VSort{Level: 1}}},
		{"VLam", VLam{Body: &Closure{Env: &Env{}, Term: ast.Var{Ix: 0}}}},
		{"VPi", VPi{A: VSort{Level: 0}, B: &Closure{Env: &Env{}, Term: ast.Sort{U: 1}}}},
		{"VSigma", VSigma{A: VSort{Level: 0}, B: &Closure{Env: &Env{}, Term: ast.Sort{U: 1}}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reify(tt.val)
			if result == nil {
				t.Errorf("Reify(%s) returned nil", tt.name)
			}
		})
	}
}

// TestDeepNesting tests evaluation of deeply nested terms (stack safety).
func TestDeepNesting(t *testing.T) {
	// Create a deeply nested application: f x x x ... x (100 applications)
	var term ast.Term = ast.Global{Name: "f"}
	arg := ast.Global{Name: "x"}

	for i := 0; i < 100; i++ {
		term = ast.App{T: term, U: arg}
	}

	env := &Env{Bindings: nil}

	// Should not stack overflow
	result := Eval(env, term)
	if result == nil {
		t.Error("deeply nested eval returned nil")
	}

	// Should be a neutral term with long spine
	if n, ok := result.(VNeutral); ok {
		if len(n.N.Sp) != 100 {
			t.Errorf("expected spine length 100, got %d", len(n.N.Sp))
		}
	} else {
		t.Errorf("expected VNeutral, got %T", result)
	}
}

// TestEvalBothNilEnvAndTerm tests both nil env and nil term.
func TestEvalBothNilEnvAndTerm(t *testing.T) {
	result := Eval(nil, nil)

	// Should return VGlobal{"nil"}
	if g, ok := result.(VGlobal); !ok || g.Name != "nil" {
		t.Errorf("Eval(nil, nil) should return VGlobal{nil}, got %v", result)
	}
}

// TestEnvLookupOutOfBounds tests environment lookup with out-of-bounds index.
func TestEnvLookupOutOfBounds(t *testing.T) {
	env := &Env{Bindings: []Value{VSort{Level: 0}}}

	// Positive out of bounds
	result := env.Lookup(5)
	if n, ok := result.(VNeutral); !ok || n.N.Head.Var != 5 {
		t.Errorf("expected VNeutral{Var: 5}, got %v", result)
	}

	// Negative index
	result = env.Lookup(-1)
	if n, ok := result.(VNeutral); !ok || n.N.Head.Var != -1 {
		t.Errorf("expected VNeutral{Var: -1}, got %v", result)
	}
}

// TestEnvExtend tests environment extension.
func TestEnvExtend(t *testing.T) {
	env := &Env{Bindings: []Value{VSort{Level: 0}}}
	newEnv := env.Extend(VSort{Level: 1})

	// New binding should be at index 0
	result := newEnv.Lookup(0)
	if s, ok := result.(VSort); !ok || s.Level != 1 {
		t.Errorf("expected VSort{1} at index 0, got %v", result)
	}

	// Old binding should be at index 1
	result = newEnv.Lookup(1)
	if s, ok := result.(VSort); !ok || s.Level != 0 {
		t.Errorf("expected VSort{0} at index 1, got %v", result)
	}
}

// TestFstSndOnNonPair tests Fst and Snd on non-pair values.
func TestFstSndOnNonPair(t *testing.T) {
	// Fst on a neutral
	neutral := VNeutral{N: Neutral{Head: Head{Glob: "p"}}}
	result := Fst(neutral)
	if _, ok := result.(VNeutral); !ok {
		t.Errorf("Fst(neutral) should return VNeutral, got %T", result)
	}

	// Snd on a non-pair (VSort)
	sort := VSort{Level: 0}
	result = Snd(sort)
	if _, ok := result.(VNeutral); !ok {
		t.Errorf("Snd(VSort) should return VNeutral, got %T", result)
	}
}

// TestReifyNeutralWithSpine tests reification of neutral terms with argument spine.
func TestReifyNeutralWithSpine(t *testing.T) {
	// f x y (neutral with spine [x, y])
	neutral := VNeutral{
		N: Neutral{
			Head: Head{Glob: "f"},
			Sp:   []Value{VGlobal{Name: "x"}, VGlobal{Name: "y"}},
		},
	}

	result := Reify(neutral)

	// Should be App{App{Global{f}, Global{x}}, Global{y}}
	app1, ok := result.(ast.App)
	if !ok {
		t.Fatalf("expected outer App, got %T", result)
	}

	app2, ok := app1.T.(ast.App)
	if !ok {
		t.Fatalf("expected inner App, got %T", app1.T)
	}

	if g, ok := app2.T.(ast.Global); !ok || g.Name != "f" {
		t.Errorf("expected Global{f} as head, got %v", app2.T)
	}
}

// BenchmarkEvalDeep benchmarks evaluation of deeply nested terms.
func BenchmarkEvalDeep(b *testing.B) {
	// Build a deep term once
	var term ast.Term = ast.Global{Name: "f"}
	arg := ast.Global{Name: "x"}
	for i := 0; i < 50; i++ {
		term = ast.App{T: term, U: arg}
	}
	env := &Env{Bindings: nil}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Eval(env, term)
	}
}

// BenchmarkReify benchmarks reification performance.
func BenchmarkReify(b *testing.B) {
	// Create a value with moderate complexity
	val := VNeutral{
		N: Neutral{
			Head: Head{Glob: "f"},
			Sp: []Value{
				VSort{Level: 0},
				VGlobal{Name: "x"},
				VPair{Fst: VSort{Level: 0}, Snd: VSort{Level: 1}},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Reify(val)
	}
}
