package parser

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/elab"
)

func TestParseSurfaceVar(t *testing.T) {
	term, err := ParseSurfaceTerm("x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := term.(*elab.SVar)
	if !ok {
		t.Fatalf("expected SVar, got %T", term)
	}
	if v.Name != "x" {
		t.Errorf("expected name x, got %s", v.Name)
	}
}

func TestParseSurfaceType(t *testing.T) {
	tests := []struct {
		input string
		level uint
	}{
		{"Type", 0},
		{"Type0", 0},
		{"Type1", 1},
		{"Type2", 2},
		{"(Type 3)", 3},
	}

	for _, tt := range tests {
		term, err := ParseSurfaceTerm(tt.input)
		if err != nil {
			t.Errorf("ParseSurfaceTerm(%q): %v", tt.input, err)
			continue
		}
		s, ok := term.(*elab.SType)
		if !ok {
			t.Errorf("ParseSurfaceTerm(%q): expected SType, got %T", tt.input, term)
			continue
		}
		if s.Level != tt.level {
			t.Errorf("ParseSurfaceTerm(%q): expected level %d, got %d", tt.input, tt.level, s.Level)
		}
	}
}

func TestParseSurfaceHole(t *testing.T) {
	tests := []struct {
		input string
		name  string
	}{
		{"_", ""},
		{"?foo", "foo"},
		{"?bar", "bar"},
	}

	for _, tt := range tests {
		term, err := ParseSurfaceTerm(tt.input)
		if err != nil {
			t.Errorf("ParseSurfaceTerm(%q): %v", tt.input, err)
			continue
		}
		h, ok := term.(*elab.SHole)
		if !ok {
			t.Errorf("ParseSurfaceTerm(%q): expected SHole, got %T", tt.input, term)
			continue
		}
		if h.Name != tt.name {
			t.Errorf("ParseSurfaceTerm(%q): expected name %q, got %q", tt.input, tt.name, h.Name)
		}
	}
}

func TestParseSurfaceLambda(t *testing.T) {
	tests := []struct {
		input  string
		binder string
		icity  elab.Icity
	}{
		{`\x. x`, "x", elab.Explicit},
		{`\{x}. x`, "x", elab.Implicit},
		{`\y. Type`, "y", elab.Explicit},
	}

	for _, tt := range tests {
		term, err := ParseSurfaceTerm(tt.input)
		if err != nil {
			t.Errorf("ParseSurfaceTerm(%q): %v", tt.input, err)
			continue
		}
		lam, ok := term.(*elab.SLam)
		if !ok {
			t.Errorf("ParseSurfaceTerm(%q): expected SLam, got %T", tt.input, term)
			continue
		}
		if lam.Binder != tt.binder {
			t.Errorf("ParseSurfaceTerm(%q): expected binder %q, got %q", tt.input, tt.binder, lam.Binder)
		}
		if lam.Icity != tt.icity {
			t.Errorf("ParseSurfaceTerm(%q): expected icity %v, got %v", tt.input, tt.icity, lam.Icity)
		}
	}
}

func TestParseSurfaceArrow(t *testing.T) {
	term, err := ParseSurfaceTerm("Type -> Type")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pi, ok := term.(*elab.SPi)
	if !ok {
		t.Fatalf("expected SPi, got %T", term)
	}
	if pi.Binder != "_" {
		t.Errorf("expected binder _, got %s", pi.Binder)
	}
	if pi.Icity != elab.Explicit {
		t.Error("expected explicit")
	}
}

func TestParseSurfaceDependentPi(t *testing.T) {
	term, err := ParseSurfaceTerm("(A : Type) -> A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pi, ok := term.(*elab.SPi)
	if !ok {
		t.Fatalf("expected SPi, got %T", term)
	}
	if pi.Binder != "A" {
		t.Errorf("expected binder A, got %s", pi.Binder)
	}
	if pi.Icity != elab.Explicit {
		t.Error("expected explicit")
	}
}

func TestParseSurfaceImplicitPi(t *testing.T) {
	term, err := ParseSurfaceTerm("{A : Type} -> A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pi, ok := term.(*elab.SPi)
	if !ok {
		t.Fatalf("expected SPi, got %T", term)
	}
	if pi.Binder != "A" {
		t.Errorf("expected binder A, got %s", pi.Binder)
	}
	if pi.Icity != elab.Implicit {
		t.Error("expected implicit")
	}
}

func TestParseSurfaceApplication(t *testing.T) {
	term, err := ParseSurfaceTerm("f x")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	app, ok := term.(*elab.SApp)
	if !ok {
		t.Fatalf("expected SApp, got %T", term)
	}
	if app.Icity != elab.Explicit {
		t.Error("expected explicit application")
	}
}

func TestParseSurfaceImplicitApplication(t *testing.T) {
	term, err := ParseSurfaceTerm("f {A}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	app, ok := term.(*elab.SApp)
	if !ok {
		t.Fatalf("expected SApp, got %T", term)
	}
	if app.Icity != elab.Implicit {
		t.Error("expected implicit application")
	}
}

func TestParseSurfaceSExpr(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"(Pi A Type A)"},
		{"(Lam x x)"},
		{"(Sigma A Type A)"},
		{"(Pair Type Type)"},
		{"(Id Type x y)"},
		{"(Refl Type x)"},
	}

	for _, tt := range tests {
		_, err := ParseSurfaceTerm(tt.input)
		if err != nil {
			t.Errorf("ParseSurfaceTerm(%q): %v", tt.input, err)
		}
	}
}

func TestParseSurfaceComments(t *testing.T) {
	term, err := ParseSurfaceTerm(`
		-- This is a comment
		Type -- end of line comment
	`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := term.(*elab.SType); !ok {
		t.Errorf("expected SType, got %T", term)
	}
}

func TestParseSurfaceComplex(t *testing.T) {
	// Identity function type
	term, err := ParseSurfaceTerm("{A : Type} -> A -> A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pi1, ok := term.(*elab.SPi)
	if !ok {
		t.Fatalf("expected SPi, got %T", term)
	}
	if pi1.Icity != elab.Implicit {
		t.Error("expected implicit outer Pi")
	}
	pi2, ok := pi1.Cod.(*elab.SPi)
	if !ok {
		t.Fatalf("expected inner SPi, got %T", pi1.Cod)
	}
	if pi2.Icity != elab.Explicit {
		t.Error("expected explicit inner Pi")
	}
}

func TestFormatSurfaceTerm(t *testing.T) {
	tests := []struct {
		term     elab.STerm
		expected string
	}{
		{&elab.SType{Level: 0}, "Type"},
		{&elab.SType{Level: 1}, "Type1"},
		{&elab.SVar{Name: "x"}, "x"},
		{&elab.SHole{Name: ""}, "_"},
		{&elab.SHole{Name: "foo"}, "?foo"},
	}

	for _, tt := range tests {
		result := FormatSurfaceTerm(tt.term)
		if result != tt.expected {
			t.Errorf("FormatSurfaceTerm(%v): expected %q, got %q", tt.term, tt.expected, result)
		}
	}
}

// --- Additional coverage tests ---

// Note: Some S-expression keywords (fst, snd, let, J, PathP, etc.) are not
// recognized in the surface parser's syntax. They are handled in the
// S-expression parser (sexpr.go). Testing those via the surface parser
// requires the S-expression syntax with capitalized keywords.

func TestParseSurfaceFstSExpr(t *testing.T) {
	// Using S-expression syntax
	term, err := ParseSurfaceTerm("(Fst p)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fst, ok := term.(*elab.SFst)
	if !ok {
		t.Fatalf("expected SFst, got %T", term)
	}
	_ = fst
}

func TestParseSurfaceSndSExpr(t *testing.T) {
	term, err := ParseSurfaceTerm("(Snd p)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	snd, ok := term.(*elab.SSnd)
	if !ok {
		t.Fatalf("expected SSnd, got %T", term)
	}
	_ = snd
}

func TestParseSurfaceLetSExpr(t *testing.T) {
	term, err := ParseSurfaceTerm("(Let x Type Type x)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lt, ok := term.(*elab.SLet)
	if !ok {
		t.Fatalf("expected SLet, got %T", term)
	}
	if lt.Binder != "x" {
		t.Errorf("expected binder x, got %s", lt.Binder)
	}
}

func TestParseSurfaceJSExpr(t *testing.T) {
	term, err := ParseSurfaceTerm("(J A C d x y p)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	j, ok := term.(*elab.SJ)
	if !ok {
		t.Fatalf("expected SJ, got %T", term)
	}
	_ = j
}

func TestParseSurfacePath(t *testing.T) {
	term, err := ParseSurfaceTerm("(Path A x y)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	path, ok := term.(*elab.SPath)
	if !ok {
		t.Fatalf("expected SPath, got %T", term)
	}
	_ = path
}

// Note: PathP, I0, I1, Transport, PathLam, PathApp, Global are handled by
// the S-expression parser (sexpr.go) not the surface parser. We test those
// through the sexpr_test.go tests.

func TestParseSurfaceParenExpr(t *testing.T) {
	term, err := ParseSurfaceTerm("(Type)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, ok := term.(*elab.SType)
	if !ok {
		t.Fatalf("expected SType, got %T", term)
	}
}

func TestFormatSurfaceTermMore(t *testing.T) {
	tests := []elab.STerm{
		&elab.SPi{Binder: "x", Icity: elab.Explicit, Dom: &elab.SType{}, Cod: &elab.SVar{Name: "x"}},
		&elab.SPi{Binder: "_", Icity: elab.Explicit, Dom: &elab.SType{}, Cod: &elab.SType{}},
		&elab.SLam{Binder: "x", Icity: elab.Explicit, Body: &elab.SVar{Name: "x"}},
		&elab.SLam{Binder: "x", Icity: elab.Implicit, Body: &elab.SVar{Name: "x"}},
		&elab.SApp{Fn: &elab.SVar{Name: "f"}, Arg: &elab.SVar{Name: "x"}, Icity: elab.Explicit},
		&elab.SApp{Fn: &elab.SVar{Name: "f"}, Arg: &elab.SVar{Name: "x"}, Icity: elab.Implicit},
		&elab.SSigma{Binder: "x", Fst: &elab.SType{}, Snd: &elab.SVar{Name: "x"}},
		&elab.SPair{Fst: &elab.SType{}, Snd: &elab.SType{}},
		&elab.SFst{Pair: &elab.SVar{Name: "p"}},
		&elab.SSnd{Pair: &elab.SVar{Name: "p"}},
		&elab.SLet{Binder: "x", Ann: &elab.SType{}, Val: &elab.SType{}, Body: &elab.SVar{Name: "x"}},
		&elab.SId{A: &elab.SType{}, X: &elab.SVar{Name: "x"}, Y: &elab.SVar{Name: "y"}},
		&elab.SRefl{A: &elab.SType{}, X: &elab.SVar{Name: "x"}},
		&elab.SJ{A: &elab.SVar{Name: "A"}, C: &elab.SVar{Name: "C"}, D: &elab.SVar{Name: "d"}, X: &elab.SVar{Name: "x"}, Y: &elab.SVar{Name: "y"}, P: &elab.SVar{Name: "p"}},
		&elab.SPath{A: &elab.SType{}, X: &elab.SVar{Name: "x"}, Y: &elab.SVar{Name: "y"}},
		&elab.SPathP{A: &elab.SVar{Name: "A"}, X: &elab.SVar{Name: "x"}, Y: &elab.SVar{Name: "y"}},
		&elab.SPathLam{Binder: "i", Body: &elab.SVar{Name: "x"}},
		&elab.SPathApp{Path: &elab.SVar{Name: "p"}, Arg: &elab.SI0{}},
		&elab.SI0{},
		&elab.SI1{},
		&elab.STransport{A: &elab.SVar{Name: "A"}, E: &elab.SVar{Name: "e"}},
		&elab.SGlobal{Name: "foo"},
		&elab.SArrow{Dom: &elab.SType{}, Cod: &elab.SType{}},
		&elab.SProd{Fst: &elab.SType{}, Snd: &elab.SType{}},
	}

	for _, term := range tests {
		result := FormatSurfaceTerm(term)
		if result == "" {
			t.Errorf("FormatSurfaceTerm(%T) returned empty string", term)
		}
	}
}

func TestParseSurfaceErrors(t *testing.T) {
	// Various error cases
	tests := []string{
		"(",       // Unclosed paren
		")",       // Unexpected paren
		"(Pi)",    // Not enough args
		"(Lam)",   // Not enough args
		"(Sigma)", // Not enough args
	}

	for _, tt := range tests {
		_, err := ParseSurfaceTerm(tt)
		if err == nil {
			t.Errorf("ParseSurfaceTerm(%q): expected error", tt)
		}
	}
}

func TestParseSurfaceMultipleApplications(t *testing.T) {
	term, err := ParseSurfaceTerm("f x y z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be ((f x) y) z
	app, ok := term.(*elab.SApp)
	if !ok {
		t.Fatalf("expected SApp, got %T", term)
	}
	_ = app
}

func TestParseSurfaceNestedImplicitPi(t *testing.T) {
	term, err := ParseSurfaceTerm("{A : Type} -> {B : Type} -> A")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pi1, ok := term.(*elab.SPi)
	if !ok {
		t.Fatalf("expected SPi, got %T", term)
	}
	if pi1.Icity != elab.Implicit {
		t.Error("expected implicit")
	}
	pi2, ok := pi1.Cod.(*elab.SPi)
	if !ok {
		t.Fatalf("expected inner SPi, got %T", pi1.Cod)
	}
	if pi2.Icity != elab.Implicit {
		t.Error("expected implicit inner Pi")
	}
}

func TestParseSurfaceArrowRightAssoc(t *testing.T) {
	term, err := ParseSurfaceTerm("Type -> Type -> Type")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pi1, ok := term.(*elab.SPi)
	if !ok {
		t.Fatalf("expected SPi, got %T", term)
	}
	// The codomain should be another Pi
	_, ok = pi1.Cod.(*elab.SPi)
	if !ok {
		t.Fatalf("expected inner SPi (right associative), got %T", pi1.Cod)
	}
}
