package core

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// Test helper constructors
func lam(binder string, body ast.Term) ast.Term {
	return ast.Lam{Binder: binder, Body: body}
}

func app(f, a ast.Term) ast.Term {
	return ast.App{T: f, U: a}
}

func pair(fst, snd ast.Term) ast.Term {
	return ast.Pair{Fst: fst, Snd: snd}
}

func fst(p ast.Term) ast.Term {
	return ast.Fst{P: p}
}

func snd(p ast.Term) ast.Term {
	return ast.Snd{P: p}
}

func glob(name string) ast.Term {
	return ast.Global{Name: name}
}

func vr(ix int) ast.Term {
	return ast.Var{Ix: ix}
}

func pi(binder string, a, b ast.Term) ast.Term {
	return ast.Pi{Binder: binder, A: a, B: b}
}

func sigma(binder string, a, b ast.Term) ast.Term {
	return ast.Sigma{Binder: binder, A: a, B: b}
}

func sort(level int) ast.Term {
	return ast.Sort{U: ast.Level(level)}
}

func lamId() ast.Term {
	return lam("x", vr(0))
}

// Test beta reduction: Conv(env, App(Lam(x,x), y), y, ConvOptions{}) == true
func TestConv_Beta(t *testing.T) {
	env := NewEnv()
	id := lamId()
	term := app(id, glob("y"))
	expected := glob("y")

	if !Conv(env, term, expected, ConvOptions{}) {
		t.Fatal("beta reduction should yield convertibility")
	}
}

// Test neutral mismatch: Conv(env, Var(0), Lam(x,x), ConvOptions{}) == false
func TestConv_NeutralMismatch(t *testing.T) {
	env := NewEnv()
	term1 := vr(0)
	term2 := lamId()

	if Conv(env, term1, term2, ConvOptions{}) {
		t.Fatal("variable and lambda should not be convertible")
	}
}

// Test eta for functions: with EnableEta=false, Conv(f, Lam(x, App(f, Var(0))), {}) == false
// with EnableEta=true, == true
func TestConv_EtaFunctions(t *testing.T) {
	env := NewEnv()
	f := glob("f")
	etaExpanded := lam("x", app(f, vr(0)))

	// Without eta, should not be equal
	if Conv(env, f, etaExpanded, ConvOptions{EnableEta: false}) {
		t.Fatal("without eta, f and \\x. f x should not be convertible")
	}

	// With eta, should be equal
	if !Conv(env, f, etaExpanded, ConvOptions{EnableEta: true}) {
		t.Fatal("with eta, f and \\x. f x should be convertible")
	}
}

// Test eta for pairs: with EnableEta=false, Conv(v, Pair(Fst(v), Snd(v)), {}) == false
// with EnableEta=true, == true
func TestConv_EtaPairs(t *testing.T) {
	env := NewEnv()
	v := glob("v")
	etaExpanded := pair(fst(v), snd(v))

	// Without eta, should not be equal
	if Conv(env, v, etaExpanded, ConvOptions{EnableEta: false}) {
		t.Fatal("without eta, v and (fst v, snd v) should not be convertible")
	}

	// With eta, should be equal
	if !Conv(env, v, etaExpanded, ConvOptions{EnableEta: true}) {
		t.Fatal("with eta, v and (fst v, snd v) should be convertible")
	}
}

// Test that different terms are not convertible
func TestConv_NonEqual(t *testing.T) {
	env := NewEnv()
	a := pair(glob("a"), glob("b"))
	b := pair(glob("a"), glob("c"))

	if Conv(env, a, b, ConvOptions{}) {
		t.Fatal("different pairs must not be convertible")
	}
}

// Test application spine normalization
func TestConv_AppSpine(t *testing.T) {
	env := NewEnv()
	f := glob("f")
	l := ast.MkApps(f, vr(0), vr(1))
	r := app(app(f, vr(0)), vr(1))

	if !Conv(env, l, r, ConvOptions{}) {
		t.Fatal("application spine associativity should normalize")
	}
}

// Test complex beta reduction: (\x. \y. x) a b ⇓ a
func TestConv_ComplexBeta(t *testing.T) {
	env := NewEnv()
	// K combinator: \x. \y. x
	k := lam("x", lam("y", vr(1)))
	term := ast.MkApps(k, glob("a"), glob("b"))
	expected := glob("a")

	if !Conv(env, term, expected, ConvOptions{}) {
		t.Fatal("complex beta reduction should yield convertibility")
	}
}

// Test projection normalization
func TestConv_Projections(t *testing.T) {
	env := NewEnv()

	// fst (pair a b) ⇓ a
	p := pair(glob("a"), glob("b"))
	term1 := fst(p)
	expected1 := glob("a")

	if !Conv(env, term1, expected1, ConvOptions{}) {
		t.Fatal("first projection should normalize")
	}

	// snd (pair a b) ⇓ b
	term2 := snd(p)
	expected2 := glob("b")

	if !Conv(env, term2, expected2, ConvOptions{}) {
		t.Fatal("second projection should normalize")
	}
}

// Test reflexivity
func TestConv_Reflexive(t *testing.T) {
	env := NewEnv()
	terms := []ast.Term{
		glob("a"),
		vr(0),
		sort(0),
		lamId(),
		pair(glob("a"), glob("b")),
	}

	for i, term := range terms {
		if !Conv(env, term, term, ConvOptions{}) {
			t.Fatalf("Test %d: term should be convertible with itself", i)
		}
	}
}

// Test symmetry
func TestConv_Symmetric(t *testing.T) {
	env := NewEnv()

	// Beta reduction is symmetric
	id := lamId()
	term1 := app(id, glob("y"))
	term2 := glob("y")

	if Conv(env, term1, term2, ConvOptions{}) != Conv(env, term2, term1, ConvOptions{}) {
		t.Fatal("conversion should be symmetric")
	}
}

// Test with environment bindings
func TestConv_WithEnvironment(t *testing.T) {
	env := NewEnv()
	// For now, we don't have a way to extend the environment with terms
	// This test ensures the environment parameter works

	term1 := vr(0)
	term2 := vr(0)

	if !Conv(env, term1, term2, ConvOptions{}) {
		t.Fatal("same variables should be convertible")
	}
}

// Table-driven tests
func TestConv_TableTests(t *testing.T) {
	tests := []struct {
		name     string
		term1    ast.Term
		term2    ast.Term
		opts     ConvOptions
		expected bool
	}{
		{
			name:     "beta_identity",
			term1:    app(lamId(), glob("y")),
			term2:    glob("y"),
			opts:     ConvOptions{},
			expected: true,
		},
		{
			name:     "fst_pair",
			term1:    fst(pair(glob("a"), glob("b"))),
			term2:    glob("a"),
			opts:     ConvOptions{},
			expected: true,
		},
		{
			name:     "snd_pair",
			term1:    snd(pair(glob("a"), glob("b"))),
			term2:    glob("b"),
			opts:     ConvOptions{},
			expected: true,
		},
		{
			name:     "different_globals",
			term1:    glob("a"),
			term2:    glob("b"),
			opts:     ConvOptions{},
			expected: false,
		},
		{
			name:     "eta_function_disabled",
			term1:    glob("f"),
			term2:    lam("x", app(glob("f"), vr(0))),
			opts:     ConvOptions{EnableEta: false},
			expected: false,
		},
		{
			name:     "eta_function_enabled",
			term1:    glob("f"),
			term2:    lam("x", app(glob("f"), vr(0))),
			opts:     ConvOptions{EnableEta: true},
			expected: true,
		},
		{
			name:     "eta_pair_disabled",
			term1:    glob("p"),
			term2:    pair(fst(glob("p")), snd(glob("p"))),
			opts:     ConvOptions{EnableEta: false},
			expected: false,
		},
		{
			name:     "eta_pair_enabled",
			term1:    glob("p"),
			term2:    pair(fst(glob("p")), snd(glob("p"))),
			opts:     ConvOptions{EnableEta: true},
			expected: true,
		},
	}

	env := NewEnv()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Conv(env, tt.term1, tt.term2, tt.opts)
			if got != tt.expected {
				t.Errorf("got %v, want %v", got, tt.expected)
			}
		})
	}
}

// Legacy API compatibility tests
func TestConv_LegacyAPI(t *testing.T) {
	// Test that the legacy API still works
	id := lamId()
	app := ast.App{T: id, U: glob("y")}

	if !ConvLegacy(app, glob("y"), EtaFlags{}) {
		t.Fatal("legacy API should work for beta reduction")
	}
}

// Test error handling - ensure no panics
func TestConv_ErrorHandling(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Conv should not panic, but got: %v", r)
		}
	}()

	env := NewEnv()

	// Test with nil terms (should be handled gracefully)
	_ = Conv(env, nil, glob("a"), ConvOptions{})
	_ = Conv(env, glob("a"), nil, ConvOptions{})
	_ = Conv(env, nil, nil, ConvOptions{})

	// Test with nil environment
	_ = Conv(nil, glob("a"), glob("a"), ConvOptions{})
}
