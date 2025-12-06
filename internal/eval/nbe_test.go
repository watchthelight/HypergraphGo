package eval

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// Test helper constructors for convenience
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

func sort(level int) ast.Term {
	return ast.Sort{U: ast.Level(level)}
}

// nf normalizes a term using NbE and returns its string representation
func nf(t ast.Term) string {
	return NormalizeNBE(t)
}

// Test beta reduction: (\x. x) y ⇓ y
func TestNBE_BetaReduction(t *testing.T) {
	// Identity function applied to global "y"
	id := lam("x", vr(0))
	term := app(id, glob("y"))

	got := nf(term)
	want := "y"

	if got != want {
		t.Errorf("Beta reduction failed: got %q, want %q", got, want)
	}
}

// Test first projection: fst (pair a b) ⇓ a
func TestNBE_FirstProjection(t *testing.T) {
	p := pair(glob("a"), glob("b"))
	term := fst(p)

	got := nf(term)
	want := "a"

	if got != want {
		t.Errorf("First projection failed: got %q, want %q", got, want)
	}
}

// Test second projection: snd (pair a b) ⇓ b
func TestNBE_SecondProjection(t *testing.T) {
	p := pair(glob("a"), glob("b"))
	term := snd(p)

	got := nf(term)
	want := "b"

	if got != want {
		t.Errorf("Second projection failed: got %q, want %q", got, want)
	}
}

// Test neutral application: applying a neutral head to an arg remains neutral
func TestNBE_NeutralApplication(t *testing.T) {
	// f applied to variable 0 should remain neutral
	f := glob("f")
	arg := vr(0)
	term := app(f, arg)

	got := nf(term)
	want := "(f {0})"

	if got != want {
		t.Errorf("Neutral application failed: got %q, want %q", got, want)
	}
}

// Test spine application: f 0 1 should normalize consistently
func TestNBE_SpineApplication(t *testing.T) {
	f := glob("f")
	term := ast.MkApps(f, vr(0), vr(1))

	got := nf(term)
	want := "(f {0} {1})"

	if got != want {
		t.Errorf("Spine application failed: got %q, want %q", got, want)
	}
}

// Test complex beta reduction: (\x. \y. x) a b ⇓ a
func TestNBE_ComplexBeta(t *testing.T) {
	// K combinator: \x. \y. x
	k := lam("x", lam("y", vr(1)))
	term := ast.MkApps(k, glob("a"), glob("b"))

	got := nf(term)
	want := "a"

	if got != want {
		t.Errorf("Complex beta reduction failed: got %q, want %q", got, want)
	}
}

// Test nested projections: fst (fst (pair (pair a b) c)) ⇓ a
func TestNBE_NestedProjections(t *testing.T) {
	inner := pair(glob("a"), glob("b"))
	outer := pair(inner, glob("c"))
	term := fst(fst(outer))

	got := nf(term)
	want := "a"

	if got != want {
		t.Errorf("Nested projections failed: got %q, want %q", got, want)
	}
}

// Test projection of neutral pair remains neutral
func TestNBE_NeutralProjection(t *testing.T) {
	p := glob("p") // neutral pair
	term := fst(p)

	got := nf(term)
	want := "(fst p)"

	if got != want {
		t.Errorf("Neutral projection failed: got %q, want %q", got, want)
	}
}

// Test that variables remain as variables
func TestNBE_Variables(t *testing.T) {
	term := vr(0)

	got := nf(term)
	want := "{0}"

	if got != want {
		t.Errorf("Variable normalization failed: got %q, want %q", got, want)
	}
}

// Test that globals remain as globals
func TestNBE_Globals(t *testing.T) {
	term := glob("foo")

	got := nf(term)
	want := "foo"

	if got != want {
		t.Errorf("Global normalization failed: got %q, want %q", got, want)
	}
}

// Test sorts
func TestNBE_Sorts(t *testing.T) {
	term := sort(0)

	got := nf(term)
	want := "Type0"

	if got != want {
		t.Errorf("Sort normalization failed: got %q, want %q", got, want)
	}
}

// Test pairs normalize their components
func TestNBE_PairNormalization(t *testing.T) {
	// ((\x. x) a, (\y. y) b) should normalize to (a, b)
	id1 := lam("x", vr(0))
	id2 := lam("y", vr(0))
	term := pair(app(id1, glob("a")), app(id2, glob("b")))

	got := nf(term)
	want := "(a , b)"

	if got != want {
		t.Errorf("Pair normalization failed: got %q, want %q", got, want)
	}
}

// Table test for various normalization cases
func TestNBE_TableTests(t *testing.T) {
	tests := []struct {
		name string
		term ast.Term
		want string
	}{
		{
			name: "identity",
			term: app(lam("x", vr(0)), glob("y")),
			want: "y",
		},
		{
			name: "fst_pair",
			term: fst(pair(glob("a"), glob("b"))),
			want: "a",
		},
		{
			name: "snd_pair",
			term: snd(pair(glob("a"), glob("b"))),
			want: "b",
		},
		{
			name: "neutral_app",
			term: app(glob("f"), vr(0)),
			want: "(f {0})",
		},
		{
			name: "neutral_fst",
			term: fst(glob("p")),
			want: "(fst p)",
		},
		{
			name: "neutral_snd",
			term: snd(glob("p")),
			want: "(snd p)",
		},
		{
			name: "const_combinator",
			term: ast.MkApps(lam("x", lam("y", vr(1))), glob("a"), glob("b")),
			want: "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nf(tt.term)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

// Test Value equality functions
func TestNBE_ValueEquality(t *testing.T) {
	env := &Env{Bindings: nil}

	// Test that same terms evaluate to equal values
	term1 := glob("a")
	term2 := glob("a")

	val1 := Eval(env, term1)
	val2 := Eval(env, term2)

	if !ValueEqual(val1, val2) {
		t.Errorf("Equal terms should evaluate to equal values")
	}

	// Test that different terms evaluate to different values
	term3 := glob("b")
	val3 := Eval(env, term3)

	if ValueEqual(val1, val3) {
		t.Errorf("Different terms should evaluate to different values")
	}
}

// Test reify/reflect round-trip
func TestNBE_ReifyReflectRoundTrip(t *testing.T) {
	env := &Env{Bindings: nil}

	tests := []ast.Term{
		glob("a"),
		vr(0),
		sort(0),
		pair(glob("a"), glob("b")),
	}

	for i, term := range tests {
		val := Eval(env, term)
		reified := Reify(val)

		// The reified term should normalize to the same result
		original := nf(term)
		roundtrip := nf(reified)

		if original != roundtrip {
			t.Errorf("Test %d: reify/reflect round-trip failed: original %q, roundtrip %q",
				i, original, roundtrip)
		}
	}
}

// Benchmark basic operations
func BenchmarkNBE_BetaReduction(b *testing.B) {
	id := lam("x", vr(0))
	term := app(id, glob("y"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EvalNBE(term)
	}
}

func BenchmarkNBE_Projection(b *testing.B) {
	p := pair(glob("a"), glob("b"))
	term := fst(p)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EvalNBE(term)
	}
}

func BenchmarkNBE_ComplexTerm(b *testing.B) {
	// (\x. \y. fst (pair x y)) a b
	inner := lam("y", fst(pair(vr(1), vr(0))))
	outer := lam("x", inner)
	term := ast.MkApps(outer, glob("a"), glob("b"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EvalNBE(term)
	}
}

// Test stuck J reification - regression test for J reification bug
func TestNBE_StuckJReification(t *testing.T) {
	// Create a stuck J: J A C d x y p where p is a neutral variable
	// This should reify to ast.J, not nested App nodes
	jTerm := ast.J{
		A: ast.Sort{U: 0},
		C: ast.Global{Name: "C"},
		D: ast.Global{Name: "d"},
		X: ast.Global{Name: "x"},
		Y: ast.Global{Name: "y"},
		P: ast.Var{Ix: 0}, // neutral proof variable (not refl)
	}

	// The proof variable is neutral (not refl), so J is stuck
	env := &Env{Bindings: []Value{vVar(0)}}
	val := Eval(env, jTerm)
	reified := Reify(val)

	// Should be ast.J, not nested App nodes
	if _, ok := reified.(ast.J); !ok {
		t.Errorf("Stuck J should reify to ast.J, got %T: %v", reified, ast.Sprint(reified))
	}

	// Also verify the structure is correct
	j := reified.(ast.J)
	if _, ok := j.A.(ast.Sort); !ok {
		t.Error("J.A should be Sort")
	}
	if g, ok := j.C.(ast.Global); !ok || g.Name != "C" {
		t.Error("J.C should be Global{C}")
	}
}

// Test J computation rule: J A C d x x (refl A x) --> d
func TestNBE_JComputation(t *testing.T) {
	// J A C d x x (refl A x) should reduce to d
	jTerm := ast.J{
		A: ast.Sort{U: 0},
		C: ast.Global{Name: "C"},
		D: ast.Global{Name: "d"},
		X: ast.Global{Name: "x"},
		Y: ast.Global{Name: "x"}, // same as X
		P: ast.Refl{A: ast.Sort{U: 0}, X: ast.Global{Name: "x"}},
	}

	env := &Env{Bindings: nil}
	val := Eval(env, jTerm)
	reified := Reify(val)

	// Should reduce to just "d"
	if g, ok := reified.(ast.Global); !ok || g.Name != "d" {
		t.Errorf("J with refl should reduce to d, got %T: %v", reified, ast.Sprint(reified))
	}
}

// Test error handling - ensure no panics
func TestNBE_ErrorHandling(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("NbE should not panic, but got: %v", r)
		}
	}()

	env := &Env{Bindings: nil}

	// Test with nil term
	val := Eval(env, nil)
	_ = Reify(val)

	// Test with out-of-bounds variable
	val2 := Eval(env, vr(10))
	_ = Reify(val2)

	// Test applying non-function
	val3 := Apply(VSort{Level: 0}, VGlobal{Name: "arg"})
	_ = Reify(val3)

	// Test projecting non-pair
	val4 := Fst(VSort{Level: 0})
	_ = Reify(val4)
}
