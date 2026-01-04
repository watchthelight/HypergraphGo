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

// Test AlphaEq lambda annotations - regression test for type soundness bug
func TestAlphaEq_LamAnnotations(t *testing.T) {
	// λ(x:Nat).x should NOT equal λ(x:Bool).x
	lam1 := ast.Lam{Binder: "x", Ann: ast.Global{Name: "Nat"}, Body: ast.Var{Ix: 0}}
	lam2 := ast.Lam{Binder: "x", Ann: ast.Global{Name: "Bool"}, Body: ast.Var{Ix: 0}}
	if AlphaEq(lam1, lam2) {
		t.Error("Lambdas with different annotations should not be alpha-equal")
	}

	// λ(x:Nat).x should equal λ(y:Nat).y (alpha equivalent)
	lam3 := ast.Lam{Binder: "y", Ann: ast.Global{Name: "Nat"}, Body: ast.Var{Ix: 0}}
	if !AlphaEq(lam1, lam3) {
		t.Error("Alpha-equivalent lambdas with same annotations should be equal")
	}

	// Unannotated lambdas should still work
	lam4 := ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}}
	lam5 := ast.Lam{Binder: "y", Body: ast.Var{Ix: 0}}
	if !AlphaEq(lam4, lam5) {
		t.Error("Unannotated alpha-equivalent lambdas should be equal")
	}

	// Annotated vs unannotated should NOT be equal
	if AlphaEq(lam1, lam4) {
		t.Error("Annotated lambda should not equal unannotated lambda")
	}

	// Complex annotations should be compared structurally
	lam6 := ast.Lam{Binder: "x", Ann: ast.Pi{Binder: "a", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}}, Body: ast.Var{Ix: 0}}
	lam7 := ast.Lam{Binder: "y", Ann: ast.Pi{Binder: "b", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}}, Body: ast.Var{Ix: 0}}
	if !AlphaEq(lam6, lam7) {
		t.Error("Lambdas with alpha-equivalent complex annotations should be equal")
	}
}

// Test Env.Extend method
func TestEnv_Extend(t *testing.T) {
	t.Parallel()
	env := NewEnv()

	// Extend with a simple term
	extEnv := env.Extend(glob("x"))
	if extEnv == nil {
		t.Fatal("Extend should return a new environment")
	}
	if extEnv == env {
		t.Fatal("Extend should return a new environment, not mutate existing")
	}

	// Extend twice
	extEnv2 := extEnv.Extend(glob("y"))
	if extEnv2 == nil {
		t.Fatal("Second Extend should return a new environment")
	}

	// Extend with more complex terms
	complexEnv := env.Extend(lam("x", app(vr(0), vr(0))))
	if complexEnv == nil {
		t.Fatal("Extend with complex term should work")
	}

	// Extend with pair
	pairEnv := env.Extend(pair(glob("a"), glob("b")))
	if pairEnv == nil {
		t.Fatal("Extend with pair should work")
	}
}

// Test shiftTerm for various term types
func TestShiftTerm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		term     ast.Term
		d        int
		cutoff   int
		checkVar int // expected Var.Ix for Var terms, -1 to skip check
	}{
		// Var above cutoff should be shifted
		{"Var above cutoff", vr(2), 1, 0, 3},
		{"Var at cutoff", vr(0), 1, 0, 1},
		// Var below cutoff should not be shifted
		{"Var below cutoff", vr(0), 1, 1, 0},
		{"Var below cutoff 2", vr(1), 1, 3, 1},
		// Global is unchanged
		{"Global", glob("x"), 1, 0, -1},
		// Sort is unchanged
		{"Sort", sort(0), 1, 0, -1},
		// Negative shift
		{"Var negative shift", vr(5), -2, 0, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shiftTerm(tt.term, tt.d, tt.cutoff)
			if result == nil {
				t.Fatal("shiftTerm returned nil")
			}
			if tt.checkVar >= 0 {
				if v, ok := result.(ast.Var); ok {
					if v.Ix != tt.checkVar {
						t.Errorf("shiftTerm(%v, %d, %d) = Var{%d}, want Var{%d}",
							tt.term, tt.d, tt.cutoff, v.Ix, tt.checkVar)
					}
				} else {
					t.Errorf("expected Var, got %T", result)
				}
			}
		})
	}
}

// Test shiftTerm for Lam
func TestShiftTerm_Lam(t *testing.T) {
	t.Parallel()

	// Shift Lam: body cutoff increases by 1
	term := ast.Lam{Binder: "x", Body: vr(1)}
	result := shiftTerm(term, 1, 0)

	lam, ok := result.(ast.Lam)
	if !ok {
		t.Fatalf("expected Lam, got %T", result)
	}

	// Body should have Var{Ix: 2} because original Var{1} is above cutoff 0+1=1
	body, ok := lam.Body.(ast.Var)
	if !ok || body.Ix != 2 {
		t.Errorf("Lam body should be Var{2}, got %v", lam.Body)
	}

	// Var bound by lambda (Var{0} in body) should not shift
	term2 := ast.Lam{Binder: "x", Body: vr(0)}
	result2 := shiftTerm(term2, 1, 0)
	lam2 := result2.(ast.Lam)
	body2 := lam2.Body.(ast.Var)
	if body2.Ix != 0 {
		t.Errorf("Bound variable should not shift, got Var{%d}", body2.Ix)
	}
}

// Test shiftTerm for Lam with annotation
func TestShiftTerm_LamWithAnn(t *testing.T) {
	t.Parallel()

	term := ast.Lam{Binder: "x", Ann: vr(0), Body: vr(1)}
	result := shiftTerm(term, 1, 0)

	lam := result.(ast.Lam)
	// Ann is not under the binder, so Var{0} shifts to Var{1}
	ann, ok := lam.Ann.(ast.Var)
	if !ok || ann.Ix != 1 {
		t.Errorf("Lam.Ann should be Var{1}, got %v", lam.Ann)
	}
}

// Test shiftTerm for Pi
func TestShiftTerm_Pi(t *testing.T) {
	t.Parallel()

	term := ast.Pi{Binder: "x", A: vr(0), B: vr(1)}
	result := shiftTerm(term, 1, 0)

	pi := result.(ast.Pi)
	// A shifts: Var{0} -> Var{1}
	a := pi.A.(ast.Var)
	if a.Ix != 1 {
		t.Errorf("Pi.A should be Var{1}, got Var{%d}", a.Ix)
	}
	// B: cutoff increases, so Var{1} in body (referencing outside the binder) shifts to Var{2}
	b := pi.B.(ast.Var)
	if b.Ix != 2 {
		t.Errorf("Pi.B should be Var{2}, got Var{%d}", b.Ix)
	}
}

// Test shiftTerm for Sigma
func TestShiftTerm_Sigma(t *testing.T) {
	t.Parallel()

	term := ast.Sigma{Binder: "x", A: vr(0), B: vr(1)}
	result := shiftTerm(term, 1, 0)

	sigma := result.(ast.Sigma)
	a := sigma.A.(ast.Var)
	if a.Ix != 1 {
		t.Errorf("Sigma.A should be Var{1}, got Var{%d}", a.Ix)
	}
	b := sigma.B.(ast.Var)
	if b.Ix != 2 {
		t.Errorf("Sigma.B should be Var{2}, got Var{%d}", b.Ix)
	}
}

// Test shiftTerm for Pair
func TestShiftTerm_Pair(t *testing.T) {
	t.Parallel()

	term := ast.Pair{Fst: vr(0), Snd: vr(1)}
	result := shiftTerm(term, 2, 0)

	p := result.(ast.Pair)
	fst := p.Fst.(ast.Var)
	snd := p.Snd.(ast.Var)
	if fst.Ix != 2 {
		t.Errorf("Pair.Fst should be Var{2}, got Var{%d}", fst.Ix)
	}
	if snd.Ix != 3 {
		t.Errorf("Pair.Snd should be Var{3}, got Var{%d}", snd.Ix)
	}
}

// Test shiftTerm for Fst and Snd
func TestShiftTerm_FstSnd(t *testing.T) {
	t.Parallel()

	fstTerm := ast.Fst{P: vr(0)}
	result := shiftTerm(fstTerm, 1, 0)
	fst := result.(ast.Fst)
	if v := fst.P.(ast.Var); v.Ix != 1 {
		t.Errorf("Fst.P should be Var{1}, got Var{%d}", v.Ix)
	}

	sndTerm := ast.Snd{P: vr(0)}
	result2 := shiftTerm(sndTerm, 1, 0)
	snd := result2.(ast.Snd)
	if v := snd.P.(ast.Var); v.Ix != 1 {
		t.Errorf("Snd.P should be Var{1}, got Var{%d}", v.Ix)
	}
}

// Test shiftTerm for Let
func TestShiftTerm_Let(t *testing.T) {
	t.Parallel()

	term := ast.Let{
		Binder: "x",
		Ann:    vr(0),
		Val:    vr(1),
		Body:   vr(2),
	}
	result := shiftTerm(term, 1, 0)

	lt := result.(ast.Let)
	// Ann and Val shift normally (not under binder)
	if v := lt.Ann.(ast.Var); v.Ix != 1 {
		t.Errorf("Let.Ann should be Var{1}, got Var{%d}", v.Ix)
	}
	if v := lt.Val.(ast.Var); v.Ix != 2 {
		t.Errorf("Let.Val should be Var{2}, got Var{%d}", v.Ix)
	}
	// Body is under binder, so cutoff+1: Var{2} -> Var{3}
	if v := lt.Body.(ast.Var); v.Ix != 3 {
		t.Errorf("Let.Body should be Var{3}, got Var{%d}", v.Ix)
	}
}

// Test shiftTerm for Id
func TestShiftTerm_Id(t *testing.T) {
	t.Parallel()

	term := ast.Id{A: vr(0), X: vr(1), Y: vr(2)}
	result := shiftTerm(term, 1, 0)

	id := result.(ast.Id)
	if v := id.A.(ast.Var); v.Ix != 1 {
		t.Errorf("Id.A should be Var{1}, got Var{%d}", v.Ix)
	}
	if v := id.X.(ast.Var); v.Ix != 2 {
		t.Errorf("Id.X should be Var{2}, got Var{%d}", v.Ix)
	}
	if v := id.Y.(ast.Var); v.Ix != 3 {
		t.Errorf("Id.Y should be Var{3}, got Var{%d}", v.Ix)
	}
}

// Test shiftTerm for Refl
func TestShiftTerm_Refl(t *testing.T) {
	t.Parallel()

	term := ast.Refl{A: vr(0), X: vr(1)}
	result := shiftTerm(term, 1, 0)

	refl := result.(ast.Refl)
	if v := refl.A.(ast.Var); v.Ix != 1 {
		t.Errorf("Refl.A should be Var{1}, got Var{%d}", v.Ix)
	}
	if v := refl.X.(ast.Var); v.Ix != 2 {
		t.Errorf("Refl.X should be Var{2}, got Var{%d}", v.Ix)
	}
}

// Test shiftTerm for J
func TestShiftTerm_J(t *testing.T) {
	t.Parallel()

	term := ast.J{
		A: vr(0),
		C: vr(1),
		D: vr(2),
		X: vr(3),
		Y: vr(4),
		P: vr(5),
	}
	result := shiftTerm(term, 1, 0)

	j := result.(ast.J)
	if v := j.A.(ast.Var); v.Ix != 1 {
		t.Errorf("J.A should be Var{1}, got Var{%d}", v.Ix)
	}
	if v := j.C.(ast.Var); v.Ix != 2 {
		t.Errorf("J.C should be Var{2}, got Var{%d}", v.Ix)
	}
	if v := j.D.(ast.Var); v.Ix != 3 {
		t.Errorf("J.D should be Var{3}, got Var{%d}", v.Ix)
	}
	if v := j.X.(ast.Var); v.Ix != 4 {
		t.Errorf("J.X should be Var{4}, got Var{%d}", v.Ix)
	}
	if v := j.Y.(ast.Var); v.Ix != 5 {
		t.Errorf("J.Y should be Var{5}, got Var{%d}", v.Ix)
	}
	if v := j.P.(ast.Var); v.Ix != 6 {
		t.Errorf("J.P should be Var{6}, got Var{%d}", v.Ix)
	}
}

// Test shiftTerm for App
func TestShiftTerm_App(t *testing.T) {
	t.Parallel()

	term := ast.App{T: vr(0), U: vr(1)}
	result := shiftTerm(term, 1, 0)

	app := result.(ast.App)
	if v := app.T.(ast.Var); v.Ix != 1 {
		t.Errorf("App.T should be Var{1}, got Var{%d}", v.Ix)
	}
	if v := app.U.(ast.Var); v.Ix != 2 {
		t.Errorf("App.U should be Var{2}, got Var{%d}", v.Ix)
	}
}

// Test shiftTerm for nil
func TestShiftTerm_Nil(t *testing.T) {
	t.Parallel()
	result := shiftTerm(nil, 1, 0)
	if result != nil {
		t.Error("shiftTerm(nil) should return nil")
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

// Test AlphaEq for edge cases
func TestAlphaEq_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b ast.Term
		want bool
	}{
		// nil handling
		{"nil == nil", nil, nil, true},
		{"nil != Global", nil, glob("x"), false},
		{"Global != nil", glob("x"), nil, false},

		// Sort with different universe levels
		{"Sort same level", sort(0), sort(0), true},
		{"Sort different level", sort(0), sort(1), false},
		{"Sort high level", sort(100), sort(100), true},

		// Var
		{"Var same", vr(0), vr(0), true},
		{"Var different", vr(0), vr(1), false},

		// Global
		{"Global same", glob("x"), glob("x"), true},
		{"Global different", glob("x"), glob("y"), false},

		// Pi
		{"Pi same", pi("x", sort(0), vr(0)), pi("y", sort(0), vr(0)), true},
		{"Pi different domain", pi("x", sort(0), vr(0)), pi("x", sort(1), vr(0)), false},
		{"Pi different codomain", pi("x", sort(0), vr(0)), pi("x", sort(0), vr(1)), false},

		// Sigma
		{"Sigma same", ast.Sigma{Binder: "x", A: sort(0), B: vr(0)}, ast.Sigma{Binder: "y", A: sort(0), B: vr(0)}, true},
		{"Sigma different A", ast.Sigma{Binder: "x", A: sort(0), B: vr(0)}, ast.Sigma{Binder: "x", A: sort(1), B: vr(0)}, false},
		{"Sigma different B", ast.Sigma{Binder: "x", A: sort(0), B: vr(0)}, ast.Sigma{Binder: "x", A: sort(0), B: vr(1)}, false},

		// Pair
		{"Pair same", pair(glob("a"), glob("b")), pair(glob("a"), glob("b")), true},
		{"Pair different fst", pair(glob("a"), glob("b")), pair(glob("c"), glob("b")), false},
		{"Pair different snd", pair(glob("a"), glob("b")), pair(glob("a"), glob("c")), false},

		// Fst/Snd
		{"Fst same", fst(glob("p")), fst(glob("p")), true},
		{"Fst different", fst(glob("p")), fst(glob("q")), false},
		{"Snd same", snd(glob("p")), snd(glob("p")), true},
		{"Snd different", snd(glob("p")), snd(glob("q")), false},

		// Let
		{
			"Let same",
			ast.Let{Binder: "x", Val: glob("a"), Body: vr(0)},
			ast.Let{Binder: "y", Val: glob("a"), Body: vr(0)},
			true,
		},
		{
			"Let different val",
			ast.Let{Binder: "x", Val: glob("a"), Body: vr(0)},
			ast.Let{Binder: "x", Val: glob("b"), Body: vr(0)},
			false,
		},
		{
			"Let different body",
			ast.Let{Binder: "x", Val: glob("a"), Body: vr(0)},
			ast.Let{Binder: "x", Val: glob("a"), Body: vr(1)},
			false,
		},
		{
			"Let with ann same",
			ast.Let{Binder: "x", Ann: sort(0), Val: glob("a"), Body: vr(0)},
			ast.Let{Binder: "y", Ann: sort(0), Val: glob("a"), Body: vr(0)},
			true,
		},
		{
			"Let with ann different",
			ast.Let{Binder: "x", Ann: sort(0), Val: glob("a"), Body: vr(0)},
			ast.Let{Binder: "x", Ann: sort(1), Val: glob("a"), Body: vr(0)},
			false,
		},
		{
			"Let ann vs no ann",
			ast.Let{Binder: "x", Ann: sort(0), Val: glob("a"), Body: vr(0)},
			ast.Let{Binder: "x", Val: glob("a"), Body: vr(0)},
			false,
		},

		// Id
		{
			"Id same",
			ast.Id{A: sort(0), X: glob("a"), Y: glob("b")},
			ast.Id{A: sort(0), X: glob("a"), Y: glob("b")},
			true,
		},
		{
			"Id different A",
			ast.Id{A: sort(0), X: glob("a"), Y: glob("b")},
			ast.Id{A: sort(1), X: glob("a"), Y: glob("b")},
			false,
		},
		{
			"Id different X",
			ast.Id{A: sort(0), X: glob("a"), Y: glob("b")},
			ast.Id{A: sort(0), X: glob("c"), Y: glob("b")},
			false,
		},
		{
			"Id different Y",
			ast.Id{A: sort(0), X: glob("a"), Y: glob("b")},
			ast.Id{A: sort(0), X: glob("a"), Y: glob("c")},
			false,
		},

		// Refl
		{
			"Refl same",
			ast.Refl{A: sort(0), X: glob("a")},
			ast.Refl{A: sort(0), X: glob("a")},
			true,
		},
		{
			"Refl different A",
			ast.Refl{A: sort(0), X: glob("a")},
			ast.Refl{A: sort(1), X: glob("a")},
			false,
		},
		{
			"Refl different X",
			ast.Refl{A: sort(0), X: glob("a")},
			ast.Refl{A: sort(0), X: glob("b")},
			false,
		},

		// J
		{
			"J same",
			ast.J{A: glob("A"), C: glob("C"), D: glob("D"), X: glob("x"), Y: glob("y"), P: glob("p")},
			ast.J{A: glob("A"), C: glob("C"), D: glob("D"), X: glob("x"), Y: glob("y"), P: glob("p")},
			true,
		},
		{
			"J different A",
			ast.J{A: glob("A"), C: glob("C"), D: glob("D"), X: glob("x"), Y: glob("y"), P: glob("p")},
			ast.J{A: glob("B"), C: glob("C"), D: glob("D"), X: glob("x"), Y: glob("y"), P: glob("p")},
			false,
		},

		// App
		{"App same", app(glob("f"), glob("x")), app(glob("f"), glob("x")), true},
		{"App different T", app(glob("f"), glob("x")), app(glob("g"), glob("x")), false},
		{"App different U", app(glob("f"), glob("x")), app(glob("f"), glob("y")), false},

		// Cross-type comparisons
		{"Sort vs Var", sort(0), vr(0), false},
		{"Lam vs Pi", lam("x", vr(0)), pi("x", sort(0), vr(0)), false},
		{"Pair vs App", pair(glob("a"), glob("b")), app(glob("a"), glob("b")), false},
		{"Fst vs Snd", fst(glob("p")), snd(glob("p")), false},
		{"Id vs Refl", ast.Id{A: sort(0), X: glob("a"), Y: glob("a")}, ast.Refl{A: sort(0), X: glob("a")}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AlphaEq(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("AlphaEq(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// Test etaEqual edge cases
func TestEtaEqual_EdgeCases(t *testing.T) {
	t.Parallel()
	env := NewEnv()

	tests := []struct {
		name string
		a, b ast.Term
		eta  bool
		want bool
	}{
		// Same term should always be equal regardless of eta
		{"Same global", glob("f"), glob("f"), false, true},
		{"Same global eta", glob("f"), glob("f"), true, true},
		{"Same lambda", lam("x", vr(0)), lam("x", vr(0)), false, true},

		// Function eta: f = \x. f x
		{"Eta function", glob("f"), lam("x", app(glob("f"), vr(0))), true, true},
		{"Eta function reversed", lam("x", app(glob("f"), vr(0))), glob("f"), true, true},

		// Pair eta: p = (fst p, snd p)
		{"Eta pair", glob("p"), pair(fst(glob("p")), snd(glob("p"))), true, true},
		{"Eta pair reversed", pair(fst(glob("p")), snd(glob("p"))), glob("p"), true, true},

		// Not eta equal - wrong argument
		{"Not eta - wrong arg", glob("f"), lam("x", app(glob("f"), vr(1))), true, false},

		// Not eta equal - wrong function
		{"Not eta - wrong func", glob("f"), lam("x", app(glob("g"), vr(0))), true, false},

		// Not eta equal - wrong fst
		{"Not eta pair - wrong fst", glob("p"), pair(fst(glob("q")), snd(glob("p"))), true, false},

		// Not eta equal - wrong snd
		{"Not eta pair - wrong snd", glob("p"), pair(fst(glob("p")), snd(glob("q"))), true, false},

		// Nested lambda (not simple eta expansion)
		{"Nested lambda", glob("f"), lam("x", lam("y", app(app(glob("f"), vr(1)), vr(0)))), true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Conv(env, tt.a, tt.b, ConvOptions{EnableEta: tt.eta})
			if got != tt.want {
				t.Errorf("Conv(%v, %v, {Eta:%v}) = %v, want %v", tt.a, tt.b, tt.eta, got, tt.want)
			}
		})
	}
}

// Test etaEqualPair edge cases directly
func TestEtaEqualPair_DirectCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		neutral ast.Term
		pair    ast.Term
		want    bool
	}{
		// Valid eta pair
		{"Valid eta", glob("p"), pair(fst(glob("p")), snd(glob("p"))), true},

		// Not a pair
		{"Not a pair - global", glob("p"), glob("q"), false},
		{"Not a pair - app", glob("p"), app(glob("f"), glob("x")), false},

		// Fst not on neutral
		{"Wrong fst", glob("p"), pair(fst(glob("q")), snd(glob("p"))), false},

		// Snd not on neutral
		{"Wrong snd", glob("p"), pair(fst(glob("p")), snd(glob("q"))), false},

		// Fst component is not Fst
		{"Fst not Fst", glob("p"), pair(glob("a"), snd(glob("p"))), false},

		// Snd component is not Snd
		{"Snd not Snd", glob("p"), pair(fst(glob("p")), glob("b")), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := etaEqualPair(tt.neutral, tt.pair)
			if got != tt.want {
				t.Errorf("etaEqualPair(%v, %v) = %v, want %v", tt.neutral, tt.pair, got, tt.want)
			}
		})
	}
}

// Test etaEqualFunction edge cases directly
func TestEtaEqualFunction_DirectCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		neutral ast.Term
		lambda  ast.Term
		want    bool
	}{
		// Valid eta expansion
		{"Valid eta", glob("f"), lam("x", app(glob("f"), vr(0))), true},

		// Not a lambda
		{"Not a lambda - global", glob("f"), glob("g"), false},
		{"Not a lambda - app", glob("f"), app(glob("g"), glob("x")), false},

		// Body not application
		{"Body not app", glob("f"), lam("x", vr(0)), false},
		{"Body not app - global", glob("f"), lam("x", glob("y")), false},

		// Argument not Var(0)
		{"Wrong arg - Var(1)", glob("f"), lam("x", app(glob("f"), vr(1))), false},
		{"Wrong arg - Global", glob("f"), lam("x", app(glob("f"), glob("x"))), false},

		// Function not shifted neutral
		{"Wrong func", glob("f"), lam("x", app(glob("g"), vr(0))), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := etaEqualFunction(tt.neutral, tt.lambda)
			if got != tt.want {
				t.Errorf("etaEqualFunction(%v, %v) = %v, want %v", tt.neutral, tt.lambda, got, tt.want)
			}
		})
	}
}

// Test Conv with nil environment
func TestConv_NilEnvironment(t *testing.T) {
	t.Parallel()

	// nil environment should be handled gracefully
	result := Conv(nil, glob("a"), glob("a"), ConvOptions{})
	if !result {
		t.Error("Conv with nil env should work for equal terms")
	}

	result = Conv(nil, glob("a"), glob("b"), ConvOptions{})
	if result {
		t.Error("Conv with nil env should return false for different terms")
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
