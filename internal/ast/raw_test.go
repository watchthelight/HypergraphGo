package ast

import (
	"testing"
)

// fakeGlobals implements Globals for testing.
type fakeGlobals map[string]struct{}

func (g fakeGlobals) Has(name string) bool {
	_, ok := g[name]
	return ok
}

func TestResolve_RId(t *testing.T) {
	// Id Nat x y where Nat is a global
	raw := RId{
		A: RGlobal{Name: "Nat"},
		X: RVar{Name: "x"},
		Y: RVar{Name: "y"},
	}
	globals := fakeGlobals{"Nat": {}, "x": {}, "y": {}}

	term, err := Resolve(globals, nil, raw)
	if err != nil {
		t.Fatal(err)
	}

	got := Sprint(term)
	want := "(Id Nat x y)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolve_RRefl(t *testing.T) {
	// refl Nat z where z is a global
	raw := RRefl{
		A: RGlobal{Name: "Nat"},
		X: RVar{Name: "z"},
	}
	globals := fakeGlobals{"Nat": {}, "z": {}}

	term, err := Resolve(globals, nil, raw)
	if err != nil {
		t.Fatal(err)
	}

	got := Sprint(term)
	want := "(refl Nat z)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolve_RJ(t *testing.T) {
	// J A C d x y p - all globals
	raw := RJ{
		A: RGlobal{Name: "A"},
		C: RGlobal{Name: "C"},
		D: RGlobal{Name: "d"},
		X: RGlobal{Name: "x"},
		Y: RGlobal{Name: "y"},
		P: RGlobal{Name: "p"},
	}
	globals := fakeGlobals{"A": {}, "C": {}, "d": {}, "x": {}, "y": {}, "p": {}}

	term, err := Resolve(globals, nil, raw)
	if err != nil {
		t.Fatal(err)
	}

	got := Sprint(term)
	want := "(J A C d x y p)"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolve_RId_Nested(t *testing.T) {
	// Id (Id A x x) (refl A x) (refl A x)
	// Tests nested identity types
	innerRefl := RRefl{
		A: RGlobal{Name: "A"},
		X: RGlobal{Name: "x"},
	}
	innerIdType := RId{
		A: RGlobal{Name: "A"},
		X: RGlobal{Name: "x"},
		Y: RGlobal{Name: "x"},
	}
	raw := RId{
		A: innerIdType,
		X: innerRefl,
		Y: innerRefl,
	}
	globals := fakeGlobals{"A": {}, "x": {}}

	term, err := Resolve(globals, nil, raw)
	if err != nil {
		t.Fatal(err)
	}

	got := Sprint(term)
	want := "(Id (Id A x x) (refl A x) (refl A x))"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolve_RId_WithBoundVars(t *testing.T) {
	// \x => Id Nat x x
	// Tests identity type with bound variables
	raw := RLam{
		Binder: "x",
		Ann:    RGlobal{Name: "Nat"},
		Body: RId{
			A: RGlobal{Name: "Nat"},
			X: RVar{Name: "x"},
			Y: RVar{Name: "x"},
		},
	}
	globals := fakeGlobals{"Nat": {}}

	term, err := Resolve(globals, nil, raw)
	if err != nil {
		t.Fatal(err)
	}

	got := Sprint(term)
	want := "(\\x : Nat => (Id Nat {0} {0}))"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolve_RRefl_WithBoundVar(t *testing.T) {
	// \x => refl Nat x
	raw := RLam{
		Binder: "x",
		Ann:    RGlobal{Name: "Nat"},
		Body: RRefl{
			A: RGlobal{Name: "Nat"},
			X: RVar{Name: "x"},
		},
	}
	globals := fakeGlobals{"Nat": {}}

	term, err := Resolve(globals, nil, raw)
	if err != nil {
		t.Fatal(err)
	}

	got := Sprint(term)
	want := "(\\x : Nat => (refl Nat {0}))"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolve_RJ_WithBoundVars(t *testing.T) {
	// \x => \y => \p => J A C d x y p
	// Tests J with bound variables
	raw := RLam{
		Binder: "x",
		Ann:    RGlobal{Name: "A"},
		Body: RLam{
			Binder: "y",
			Ann:    RGlobal{Name: "A"},
			Body: RLam{
				Binder: "p",
				Ann: RId{
					A: RGlobal{Name: "A"},
					X: RVar{Name: "x"},
					Y: RVar{Name: "y"},
				},
				Body: RJ{
					A: RGlobal{Name: "A"},
					C: RGlobal{Name: "C"},
					D: RGlobal{Name: "d"},
					X: RVar{Name: "x"},
					Y: RVar{Name: "y"},
					P: RVar{Name: "p"},
				},
			},
		},
	}
	globals := fakeGlobals{"A": {}, "C": {}, "d": {}}

	term, err := Resolve(globals, nil, raw)
	if err != nil {
		t.Fatal(err)
	}

	got := Sprint(term)
	want := "(\\x : A => (\\y : A => (\\p : (Id A {1} {0}) => (J A C d {2} {1} {0}))))"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestResolve_RId_UnboundError(t *testing.T) {
	// Id A x y where x is unbound
	raw := RId{
		A: RGlobal{Name: "A"},
		X: RVar{Name: "x"}, // unbound
		Y: RGlobal{Name: "y"},
	}
	globals := fakeGlobals{"A": {}, "y": {}}

	_, err := Resolve(globals, nil, raw)
	if err == nil {
		t.Error("expected error for unbound variable, got nil")
	}
}

func TestResolve_RRefl_UnboundError(t *testing.T) {
	// refl A x where x is unbound
	raw := RRefl{
		A: RGlobal{Name: "A"},
		X: RVar{Name: "x"}, // unbound
	}
	globals := fakeGlobals{"A": {}}

	_, err := Resolve(globals, nil, raw)
	if err == nil {
		t.Error("expected error for unbound variable, got nil")
	}
}

func TestResolve_RJ_UnboundError(t *testing.T) {
	// J A C d x y p where p is unbound
	raw := RJ{
		A: RGlobal{Name: "A"},
		C: RGlobal{Name: "C"},
		D: RGlobal{Name: "d"},
		X: RGlobal{Name: "x"},
		Y: RGlobal{Name: "y"},
		P: RVar{Name: "p"}, // unbound
	}
	globals := fakeGlobals{"A": {}, "C": {}, "d": {}, "x": {}, "y": {}}

	_, err := Resolve(globals, nil, raw)
	if err == nil {
		t.Error("expected error for unbound variable, got nil")
	}
}
