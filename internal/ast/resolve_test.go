package ast_test

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

type fakeGlobals map[string]struct{}

func (g fakeGlobals) Has(name string) bool {
	_, ok := g[name]
	return ok
}

func TestResolve_PiLamApp(t *testing.T) {
	// Raw: (\x => x) y  where y is global
	raw := ast.RApp{
		T: ast.RLam{Binder: "x", Body: ast.RVar{Name: "x"}},
		U: ast.RGlobal{Name: "y"},
	}
	core, err := ast.Resolve(fakeGlobals{"y": {}}, nil, raw)
	if err != nil {
		t.Fatal(err)
	}
	got := ast.Sprint(core)
	want := "((\\x => {0}) y)"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolve_ScopeShadow(t *testing.T) {
	// let x = {global x} in (\x => {0})  should resolve inner x to Var0 and let body Var0 to bound x
	raw := ast.RLet{
		Binder: "x",
		Val:    ast.RGlobal{Name: "x"},
		Body:   ast.RLam{Binder: "x", Body: ast.RVar{Name: "x"}},
	}
	core, err := ast.Resolve(fakeGlobals{"x": {}}, nil, raw)
	if err != nil {
		t.Fatal(err)
	}
	got := ast.Sprint(core)
	want := "(let x = x in (\\x => {0}))"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestResolve_Unbound(t *testing.T) {
	_, err := ast.Resolve(nil, nil, ast.RVar{Name: "ghost"})
	if err == nil {
		t.Fatal("expected error for unbound variable")
	}
}

func TestResolve_Sort(t *testing.T) {
	raw := ast.RSort{U: 0}
	core, err := ast.Resolve(nil, nil, raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := core.(ast.Sort); !ok {
		t.Fatalf("expected Sort, got %T", core)
	}
}

func TestResolve_Pi(t *testing.T) {
	// Pi (x : Type) -> x
	raw := ast.RPi{
		Binder: "x",
		A:      ast.RSort{U: 0},
		B:      ast.RVar{Name: "x"},
	}
	core, err := ast.Resolve(nil, nil, raw)
	if err != nil {
		t.Fatal(err)
	}
	pi, ok := core.(ast.Pi)
	if !ok {
		t.Fatalf("expected Pi, got %T", core)
	}
	if pi.Binder != "x" {
		t.Fatalf("expected binder 'x', got %q", pi.Binder)
	}
	// Body should reference bound variable
	if v, ok := pi.B.(ast.Var); !ok || v.Ix != 0 {
		t.Fatalf("expected Var{0} in body, got %v", pi.B)
	}
}

func TestResolve_Pi_ErrorInA(t *testing.T) {
	raw := ast.RPi{
		Binder: "x",
		A:      ast.RVar{Name: "unbound"},
		B:      ast.RSort{U: 0},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Pi.A")
	}
}

func TestResolve_Pi_ErrorInB(t *testing.T) {
	raw := ast.RPi{
		Binder: "x",
		A:      ast.RSort{U: 0},
		B:      ast.RVar{Name: "unbound"},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Pi.B")
	}
}

func TestResolve_LamWithAnnotation(t *testing.T) {
	// \(x : Type) => x
	raw := ast.RLam{
		Binder: "x",
		Ann:    ast.RSort{U: 0},
		Body:   ast.RVar{Name: "x"},
	}
	core, err := ast.Resolve(nil, nil, raw)
	if err != nil {
		t.Fatal(err)
	}
	lam, ok := core.(ast.Lam)
	if !ok {
		t.Fatalf("expected Lam, got %T", core)
	}
	if lam.Ann == nil {
		t.Fatal("expected annotation, got nil")
	}
}

func TestResolve_Lam_ErrorInAnn(t *testing.T) {
	raw := ast.RLam{
		Binder: "x",
		Ann:    ast.RVar{Name: "unbound"},
		Body:   ast.RSort{U: 0},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Lam.Ann")
	}
}

func TestResolve_Lam_ErrorInBody(t *testing.T) {
	raw := ast.RLam{
		Binder: "x",
		Body:   ast.RVar{Name: "unbound"},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Lam.Body")
	}
}

func TestResolve_App_ErrorInT(t *testing.T) {
	raw := ast.RApp{
		T: ast.RVar{Name: "unbound"},
		U: ast.RSort{U: 0},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in App.T")
	}
}

func TestResolve_App_ErrorInU(t *testing.T) {
	raw := ast.RApp{
		T: ast.RSort{U: 0},
		U: ast.RVar{Name: "unbound"},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in App.U")
	}
}

func TestResolve_Sigma(t *testing.T) {
	// Sigma (x : Type) x
	raw := ast.RSigma{
		Binder: "x",
		A:      ast.RSort{U: 0},
		B:      ast.RVar{Name: "x"},
	}
	core, err := ast.Resolve(nil, nil, raw)
	if err != nil {
		t.Fatal(err)
	}
	sigma, ok := core.(ast.Sigma)
	if !ok {
		t.Fatalf("expected Sigma, got %T", core)
	}
	if sigma.Binder != "x" {
		t.Fatalf("expected binder 'x', got %q", sigma.Binder)
	}
}

func TestResolve_Sigma_ErrorInA(t *testing.T) {
	raw := ast.RSigma{
		Binder: "x",
		A:      ast.RVar{Name: "unbound"},
		B:      ast.RSort{U: 0},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Sigma.A")
	}
}

func TestResolve_Sigma_ErrorInB(t *testing.T) {
	raw := ast.RSigma{
		Binder: "x",
		A:      ast.RSort{U: 0},
		B:      ast.RVar{Name: "unbound"},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Sigma.B")
	}
}

func TestResolve_Pair(t *testing.T) {
	raw := ast.RPair{
		Fst: ast.RSort{U: 0},
		Snd: ast.RSort{U: 1},
	}
	core, err := ast.Resolve(nil, nil, raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := core.(ast.Pair); !ok {
		t.Fatalf("expected Pair, got %T", core)
	}
}

func TestResolve_Pair_ErrorInFst(t *testing.T) {
	raw := ast.RPair{
		Fst: ast.RVar{Name: "unbound"},
		Snd: ast.RSort{U: 0},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Pair.Fst")
	}
}

func TestResolve_Pair_ErrorInSnd(t *testing.T) {
	raw := ast.RPair{
		Fst: ast.RSort{U: 0},
		Snd: ast.RVar{Name: "unbound"},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Pair.Snd")
	}
}

func TestResolve_Fst(t *testing.T) {
	raw := ast.RFst{P: ast.RGlobal{Name: "p"}}
	core, err := ast.Resolve(fakeGlobals{"p": {}}, nil, raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := core.(ast.Fst); !ok {
		t.Fatalf("expected Fst, got %T", core)
	}
}

func TestResolve_Fst_Error(t *testing.T) {
	raw := ast.RFst{P: ast.RVar{Name: "unbound"}}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Fst.P")
	}
}

func TestResolve_Snd(t *testing.T) {
	raw := ast.RSnd{P: ast.RGlobal{Name: "p"}}
	core, err := ast.Resolve(fakeGlobals{"p": {}}, nil, raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := core.(ast.Snd); !ok {
		t.Fatalf("expected Snd, got %T", core)
	}
}

func TestResolve_Snd_Error(t *testing.T) {
	raw := ast.RSnd{P: ast.RVar{Name: "unbound"}}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Snd.P")
	}
}

func TestResolve_LetWithAnnotation(t *testing.T) {
	raw := ast.RLet{
		Binder: "x",
		Ann:    ast.RSort{U: 0},
		Val:    ast.RSort{U: 0},
		Body:   ast.RVar{Name: "x"},
	}
	core, err := ast.Resolve(nil, nil, raw)
	if err != nil {
		t.Fatal(err)
	}
	let, ok := core.(ast.Let)
	if !ok {
		t.Fatalf("expected Let, got %T", core)
	}
	if let.Ann == nil {
		t.Fatal("expected annotation, got nil")
	}
}

func TestResolve_Let_ErrorInAnn(t *testing.T) {
	raw := ast.RLet{
		Binder: "x",
		Ann:    ast.RVar{Name: "unbound"},
		Val:    ast.RSort{U: 0},
		Body:   ast.RSort{U: 0},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Let.Ann")
	}
}

func TestResolve_Let_ErrorInVal(t *testing.T) {
	raw := ast.RLet{
		Binder: "x",
		Val:    ast.RVar{Name: "unbound"},
		Body:   ast.RSort{U: 0},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Let.Val")
	}
}

func TestResolve_Let_ErrorInBody(t *testing.T) {
	raw := ast.RLet{
		Binder: "x",
		Val:    ast.RSort{U: 0},
		Body:   ast.RVar{Name: "unbound"},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Let.Body")
	}
}

func TestResolve_Id(t *testing.T) {
	raw := ast.RId{
		A: ast.RSort{U: 0},
		X: ast.RGlobal{Name: "a"},
		Y: ast.RGlobal{Name: "b"},
	}
	core, err := ast.Resolve(fakeGlobals{"a": {}, "b": {}}, nil, raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := core.(ast.Id); !ok {
		t.Fatalf("expected Id, got %T", core)
	}
}

func TestResolve_Id_ErrorInA(t *testing.T) {
	raw := ast.RId{
		A: ast.RVar{Name: "unbound"},
		X: ast.RSort{U: 0},
		Y: ast.RSort{U: 0},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Id.A")
	}
}

func TestResolve_Id_ErrorInX(t *testing.T) {
	raw := ast.RId{
		A: ast.RSort{U: 0},
		X: ast.RVar{Name: "unbound"},
		Y: ast.RSort{U: 0},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Id.X")
	}
}

func TestResolve_Id_ErrorInY(t *testing.T) {
	raw := ast.RId{
		A: ast.RSort{U: 0},
		X: ast.RSort{U: 0},
		Y: ast.RVar{Name: "unbound"},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Id.Y")
	}
}

func TestResolve_Refl(t *testing.T) {
	raw := ast.RRefl{
		A: ast.RSort{U: 0},
		X: ast.RGlobal{Name: "a"},
	}
	core, err := ast.Resolve(fakeGlobals{"a": {}}, nil, raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := core.(ast.Refl); !ok {
		t.Fatalf("expected Refl, got %T", core)
	}
}

func TestResolve_Refl_ErrorInA(t *testing.T) {
	raw := ast.RRefl{
		A: ast.RVar{Name: "unbound"},
		X: ast.RSort{U: 0},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Refl.A")
	}
}

func TestResolve_Refl_ErrorInX(t *testing.T) {
	raw := ast.RRefl{
		A: ast.RSort{U: 0},
		X: ast.RVar{Name: "unbound"},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in Refl.X")
	}
}

func TestResolve_J(t *testing.T) {
	globals := fakeGlobals{"a": {}, "b": {}, "C": {}, "d": {}, "p": {}}
	raw := ast.RJ{
		A: ast.RSort{U: 0},
		C: ast.RGlobal{Name: "C"},
		D: ast.RGlobal{Name: "d"},
		X: ast.RGlobal{Name: "a"},
		Y: ast.RGlobal{Name: "b"},
		P: ast.RGlobal{Name: "p"},
	}
	core, err := ast.Resolve(globals, nil, raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := core.(ast.J); !ok {
		t.Fatalf("expected J, got %T", core)
	}
}

func TestResolve_J_ErrorInA(t *testing.T) {
	raw := ast.RJ{
		A: ast.RVar{Name: "unbound"},
		C: ast.RSort{U: 0},
		D: ast.RSort{U: 0},
		X: ast.RSort{U: 0},
		Y: ast.RSort{U: 0},
		P: ast.RSort{U: 0},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in J.A")
	}
}

func TestResolve_J_ErrorInC(t *testing.T) {
	raw := ast.RJ{
		A: ast.RSort{U: 0},
		C: ast.RVar{Name: "unbound"},
		D: ast.RSort{U: 0},
		X: ast.RSort{U: 0},
		Y: ast.RSort{U: 0},
		P: ast.RSort{U: 0},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in J.C")
	}
}

func TestResolve_J_ErrorInD(t *testing.T) {
	raw := ast.RJ{
		A: ast.RSort{U: 0},
		C: ast.RSort{U: 0},
		D: ast.RVar{Name: "unbound"},
		X: ast.RSort{U: 0},
		Y: ast.RSort{U: 0},
		P: ast.RSort{U: 0},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in J.D")
	}
}

func TestResolve_J_ErrorInX(t *testing.T) {
	raw := ast.RJ{
		A: ast.RSort{U: 0},
		C: ast.RSort{U: 0},
		D: ast.RSort{U: 0},
		X: ast.RVar{Name: "unbound"},
		Y: ast.RSort{U: 0},
		P: ast.RSort{U: 0},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in J.X")
	}
}

func TestResolve_J_ErrorInY(t *testing.T) {
	raw := ast.RJ{
		A: ast.RSort{U: 0},
		C: ast.RSort{U: 0},
		D: ast.RSort{U: 0},
		X: ast.RSort{U: 0},
		Y: ast.RVar{Name: "unbound"},
		P: ast.RSort{U: 0},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in J.Y")
	}
}

func TestResolve_J_ErrorInP(t *testing.T) {
	raw := ast.RJ{
		A: ast.RSort{U: 0},
		C: ast.RSort{U: 0},
		D: ast.RSort{U: 0},
		X: ast.RSort{U: 0},
		Y: ast.RSort{U: 0},
		P: ast.RVar{Name: "unbound"},
	}
	_, err := ast.Resolve(nil, nil, raw)
	if err == nil {
		t.Fatal("expected error for unbound in J.P")
	}
}

func TestResolve_VarBoundToGlobal(t *testing.T) {
	// RVar that resolves to global when not locally bound
	raw := ast.RVar{Name: "x"}
	core, err := ast.Resolve(fakeGlobals{"x": {}}, nil, raw)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := core.(ast.Global); !ok {
		t.Fatalf("expected Global, got %T", core)
	}
}
