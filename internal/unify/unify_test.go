package unify

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

func TestUnifyIdentical(t *testing.T) {
	// Unifying identical terms should succeed with no solutions
	term := ast.Sort{U: 0}
	result := Unify(term, term)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if len(result.Solved) > 0 {
		t.Errorf("expected no solutions for identical terms, got %d", len(result.Solved))
	}
}

func TestUnifyVariables(t *testing.T) {
	// Unifying identical variables should succeed
	v := ast.Var{Ix: 0}
	result := Unify(v, v)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
}

func TestUnifyDifferentVariables(t *testing.T) {
	// Unifying different variables should fail
	v1 := ast.Var{Ix: 0}
	v2 := ast.Var{Ix: 1}
	result := Unify(v1, v2)

	if len(result.Errors) == 0 {
		t.Error("expected error for different variables")
	}
}

func TestUnifyMeta(t *testing.T) {
	// Unifying ?0 with Type0 should solve ?0 = Type0
	meta := ast.Meta{ID: 0, Args: nil}
	term := ast.Sort{U: 0}
	result := Unify(meta, term)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if sol, ok := result.Solved[0]; !ok {
		t.Error("expected solution for ?0")
	} else if _, isSortSort := sol.(ast.Sort); !isSortSort {
		t.Errorf("expected Sort solution, got %T", sol)
	}
}

func TestUnifyMetaSymmetric(t *testing.T) {
	// Unifying Type0 with ?0 should also solve ?0 = Type0
	meta := ast.Meta{ID: 0, Args: nil}
	term := ast.Sort{U: 0}
	result := Unify(term, meta) // Reversed order

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if sol, ok := result.Solved[0]; !ok {
		t.Error("expected solution for ?0")
	} else if _, isSort := sol.(ast.Sort); !isSort {
		t.Errorf("expected Sort solution, got %T", sol)
	}
}

func TestUnifyMetaPattern(t *testing.T) {
	// Unifying ?0 x y with (f x y) where x,y are distinct variables
	// should solve ?0 = \x.\y.f x y
	meta := ast.Meta{
		ID:   0,
		Args: []ast.Term{ast.Var{Ix: 0}, ast.Var{Ix: 1}},
	}
	// f x y where f is at index 2, x at 0, y at 1
	term := ast.MkApps(ast.Var{Ix: 2}, ast.Var{Ix: 0}, ast.Var{Ix: 1})
	result := Unify(meta, term)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if _, ok := result.Solved[0]; !ok {
		t.Error("expected solution for ?0")
	}
}

func TestUnifyMetaNonPattern(t *testing.T) {
	// Unifying ?0 x x (non-pattern, duplicate variable) should be deferred
	meta := ast.Meta{
		ID:   0,
		Args: []ast.Term{ast.Var{Ix: 0}, ast.Var{Ix: 0}},
	}
	term := ast.Sort{U: 0}
	result := Unify(meta, term)

	// Should be unsolved, not errored
	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if len(result.Unsolved) == 0 {
		t.Error("expected unsolved constraints for non-pattern")
	}
}

func TestUnifyOccursCheck(t *testing.T) {
	// Unifying ?0 with something containing ?0 should fail occurs check
	meta := ast.Meta{ID: 0, Args: nil}
	// Pi _ ?0 ?0 contains ?0
	term := ast.Pi{Binder: "_", A: meta, B: ast.Sort{U: 0}}
	result := Unify(meta, term)

	if len(result.Errors) == 0 {
		t.Error("expected occurs check error")
	}
}

func TestUnifyPi(t *testing.T) {
	// Unifying (x : ?0) -> ?1 with (x : Type0) -> Type0
	// should solve ?0 = Type0, ?1 = Type0
	lhs := ast.Pi{
		Binder: "x",
		A:      ast.Meta{ID: 0, Args: nil},
		B:      ast.Meta{ID: 1, Args: nil},
	}
	rhs := ast.Pi{
		Binder: "x",
		A:      ast.Sort{U: 0},
		B:      ast.Sort{U: 0},
	}
	result := Unify(lhs, rhs)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if len(result.Solved) != 2 {
		t.Errorf("expected 2 solutions, got %d", len(result.Solved))
	}
}

func TestZonk(t *testing.T) {
	// Zonk should substitute solutions into a term
	solutions := map[int]ast.Term{
		0: ast.Sort{U: 0},
		1: ast.Var{Ix: 0},
	}

	term := ast.Pi{
		Binder: "x",
		A:      ast.Meta{ID: 0, Args: nil},
		B:      ast.Meta{ID: 1, Args: nil},
	}

	result := Zonk(solutions, term)

	pi, ok := result.(ast.Pi)
	if !ok {
		t.Fatalf("expected Pi, got %T", result)
	}

	if _, isSort := pi.A.(ast.Sort); !isSort {
		t.Errorf("expected domain to be zonked to Sort, got %T", pi.A)
	}

	if _, isVar := pi.B.(ast.Var); !isVar {
		t.Errorf("expected codomain to be zonked to Var, got %T", pi.B)
	}
}

func TestZonkFull(t *testing.T) {
	// ZonkFull should error if metavariables remain
	solutions := map[int]ast.Term{
		0: ast.Sort{U: 0},
		// ?1 is not solved
	}

	term := ast.Pi{
		Binder: "x",
		A:      ast.Meta{ID: 0, Args: nil},
		B:      ast.Meta{ID: 1, Args: nil}, // Unsolved
	}

	_, err := ZonkFull(solutions, term)
	if err == nil {
		t.Error("expected error for unsolved metavariables")
	}
}

func TestZonkFullComplete(t *testing.T) {
	// ZonkFull should succeed if all metavariables are solved
	solutions := map[int]ast.Term{
		0: ast.Sort{U: 0},
		1: ast.Sort{U: 1},
	}

	term := ast.Pi{
		Binder: "x",
		A:      ast.Meta{ID: 0, Args: nil},
		B:      ast.Meta{ID: 1, Args: nil},
	}

	result, err := ZonkFull(solutions, term)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if hasMeta(result) {
		t.Error("result still contains metavariables")
	}
}

func TestUnifyAll(t *testing.T) {
	// Test UnifyAll with multiple constraints
	constraints := []Constraint{
		{LHS: ast.Meta{ID: 0}, RHS: ast.Sort{U: 0}},
		{LHS: ast.Meta{ID: 1}, RHS: ast.Sort{U: 1}},
	}

	result := UnifyAll(constraints)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if len(result.Solved) != 2 {
		t.Errorf("expected 2 solutions, got %d", len(result.Solved))
	}
}

func TestUnifyApp(t *testing.T) {
	// Unify (f ?0) with (f Type0)
	lhs := ast.App{T: ast.Var{Ix: 0}, U: ast.Meta{ID: 0}}
	rhs := ast.App{T: ast.Var{Ix: 0}, U: ast.Sort{U: 0}}

	result := Unify(lhs, rhs)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if sol, ok := result.Solved[0]; !ok {
		t.Error("expected solution for ?0")
	} else if _, isSort := sol.(ast.Sort); !isSort {
		t.Errorf("expected Sort solution, got %T", sol)
	}
}

func TestUnifyId(t *testing.T) {
	// Unify Id ?0 x y with Id A x y
	lhs := ast.Id{A: ast.Meta{ID: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 2}}
	rhs := ast.Id{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 2}}

	result := Unify(lhs, rhs)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if sol, ok := result.Solved[0]; !ok {
		t.Error("expected solution for ?0")
	} else if v, isVar := sol.(ast.Var); !isVar || v.Ix != 0 {
		t.Errorf("expected Var{Ix: 0} solution, got %v", sol)
	}
}

// --- Extended Unify Tests ---

func TestUnifyLam(t *testing.T) {
	// Unify (λx. ?0) with (λx. x)
	lhs := ast.Lam{Binder: "x", Body: ast.Meta{ID: 0}}
	rhs := ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}}

	result := Unify(lhs, rhs)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
}

func TestUnifySigma(t *testing.T) {
	// Unify (Σx:?0. ?1) with (Σx:Type. Type)
	lhs := ast.Sigma{Binder: "x", A: ast.Meta{ID: 0}, B: ast.Meta{ID: 1}}
	rhs := ast.Sigma{Binder: "x", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}

	result := Unify(lhs, rhs)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
	if len(result.Solved) != 2 {
		t.Errorf("expected 2 solutions, got %d", len(result.Solved))
	}
}

func TestUnifyPair(t *testing.T) {
	// Unify (?0, ?1) with (Type, Type)
	lhs := ast.Pair{Fst: ast.Meta{ID: 0}, Snd: ast.Meta{ID: 1}}
	rhs := ast.Pair{Fst: ast.Sort{U: 0}, Snd: ast.Sort{U: 0}}

	result := Unify(lhs, rhs)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
}

func TestUnifyFst(t *testing.T) {
	// Unify (fst ?0) with (fst p)
	lhs := ast.Fst{P: ast.Meta{ID: 0}}
	rhs := ast.Fst{P: ast.Var{Ix: 0}}

	result := Unify(lhs, rhs)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
}

func TestUnifySnd(t *testing.T) {
	// Unify (snd ?0) with (snd p)
	lhs := ast.Snd{P: ast.Meta{ID: 0}}
	rhs := ast.Snd{P: ast.Var{Ix: 0}}

	result := Unify(lhs, rhs)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
}

func TestUnifyRefl(t *testing.T) {
	// Unify (refl ?0 ?1) with (refl Type x)
	lhs := ast.Refl{A: ast.Meta{ID: 0}, X: ast.Meta{ID: 1}}
	rhs := ast.Refl{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}}

	result := Unify(lhs, rhs)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
}

func TestUnifyPath(t *testing.T) {
	// Unify (Path ?0 x y) with (Path ?0 x y) - same path types should unify
	lhs := ast.Path{A: ast.Meta{ID: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}}
	rhs := ast.Path{A: ast.Meta{ID: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}}

	result := Unify(lhs, rhs)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
}

func TestUnifyPathP(t *testing.T) {
	// Unify (PathP ?0 x y) with (PathP ?0 x y) - same pathP types should unify
	lhs := ast.PathP{A: ast.Meta{ID: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}}
	rhs := ast.PathP{A: ast.Meta{ID: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 1}}

	result := Unify(lhs, rhs)

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
}

func TestUnifyGlobal(t *testing.T) {
	// Same global should unify
	lhs := ast.Global{Name: "foo"}
	rhs := ast.Global{Name: "foo"}

	result := Unify(lhs, rhs)
	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}

	// Different globals should fail
	lhs2 := ast.Global{Name: "foo"}
	rhs2 := ast.Global{Name: "bar"}

	result2 := Unify(lhs2, rhs2)
	if len(result2.Errors) == 0 {
		t.Error("expected error for different globals")
	}
}

func TestUnifyDifferentSorts(t *testing.T) {
	// Type0 and Type1 should not unify
	result := Unify(ast.Sort{U: 0}, ast.Sort{U: 1})
	if len(result.Errors) == 0 {
		t.Error("expected error for different sorts")
	}
}

func TestUnifyMismatchedTypes(t *testing.T) {
	// Pi and Lam should not unify
	lhs := ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	rhs := ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for Pi vs Lam")
	}

	// App and Sigma should not unify
	lhs2 := ast.App{T: ast.Var{Ix: 0}, U: ast.Var{Ix: 1}}
	rhs2 := ast.Sigma{Binder: "x", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}

	result2 := Unify(lhs2, rhs2)
	if len(result2.Errors) == 0 {
		t.Error("expected error for App vs Sigma")
	}
}

func TestUnifyErrorMessage(t *testing.T) {
	err := &UnifyError{Message: "test error"}
	if err.Error() != "test error" {
		t.Errorf("expected 'test error', got %q", err.Error())
	}
}

func TestNewUnifier(t *testing.T) {
	u := NewUnifier()
	if u.solutions == nil {
		t.Error("solutions map should not be nil")
	}
}

func TestUnifierAddConstraint(t *testing.T) {
	u := NewUnifier()
	u.AddConstraint(ast.Sort{U: 0}, ast.Sort{U: 0})

	if len(u.worklist) != 1 {
		t.Errorf("expected 1 constraint, got %d", len(u.worklist))
	}
}

func TestUnifierGetSolution(t *testing.T) {
	u := NewUnifier()
	u.AddConstraint(ast.Meta{ID: 0}, ast.Sort{U: 0})
	u.Solve()

	sol, ok := u.GetSolution(0)
	if !ok {
		t.Error("expected solution for ?0")
	}
	if _, isSort := sol.(ast.Sort); !isSort {
		t.Errorf("expected Sort, got %T", sol)
	}

	_, ok = u.GetSolution(999)
	if ok {
		t.Error("expected no solution for ?999")
	}
}

func TestUnifyConsistency(t *testing.T) {
	// Solving ?0 twice with consistent values should work
	u := NewUnifier()
	u.AddConstraint(ast.Meta{ID: 0}, ast.Sort{U: 0})
	u.AddConstraint(ast.Meta{ID: 0}, ast.Sort{U: 0})
	result := u.Solve()

	if len(result.Errors) > 0 {
		t.Errorf("unexpected errors: %v", result.Errors)
	}
}

func TestUnifyInconsistency(t *testing.T) {
	// Solving ?0 twice with inconsistent values should fail
	u := NewUnifier()
	u.AddConstraint(ast.Meta{ID: 0}, ast.Sort{U: 0})
	u.AddConstraint(ast.Meta{ID: 0}, ast.Sort{U: 1})
	result := u.Solve()

	if len(result.Errors) == 0 {
		t.Error("expected error for inconsistent solutions")
	}
}

func TestZonkNestedMeta(t *testing.T) {
	// ?0 = Type, ?1 = ?0 (should resolve to Type)
	solutions := map[int]ast.Term{
		0: ast.Sort{U: 0},
	}

	term := ast.Pi{
		Binder: "x",
		A:      ast.Meta{ID: 0},
		B:      ast.Meta{ID: 0},
	}

	result := Zonk(solutions, term)
	pi, ok := result.(ast.Pi)
	if !ok {
		t.Fatalf("expected Pi, got %T", result)
	}

	if _, isSort := pi.A.(ast.Sort); !isSort {
		t.Error("expected zonked A to be Sort")
	}
}

func TestZonkMetaWithArgs(t *testing.T) {
	// ?0 = λx. x, then ?0 Type should become Type
	solutions := map[int]ast.Term{
		0: ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}},
	}

	term := ast.Meta{ID: 0, Args: []ast.Term{ast.Sort{U: 0}}}
	result := Zonk(solutions, term)

	// Should be App(Lam, Type)
	if _, isApp := result.(ast.App); !isApp {
		t.Errorf("expected App, got %T", result)
	}
}

func TestZonkLet(t *testing.T) {
	solutions := map[int]ast.Term{
		0: ast.Sort{U: 0},
	}

	term := ast.Let{
		Binder: "x",
		Ann:    ast.Meta{ID: 0},
		Val:    ast.Meta{ID: 0},
		Body:   ast.Var{Ix: 0},
	}

	result := Zonk(solutions, term)
	lt, ok := result.(ast.Let)
	if !ok {
		t.Fatalf("expected Let, got %T", result)
	}

	if _, isSort := lt.Ann.(ast.Sort); !isSort {
		t.Error("expected Ann to be zonked")
	}
	if _, isSort := lt.Val.(ast.Sort); !isSort {
		t.Error("expected Val to be zonked")
	}
}

func TestZonkJ(t *testing.T) {
	solutions := map[int]ast.Term{
		0: ast.Sort{U: 0},
	}

	term := ast.J{
		A: ast.Meta{ID: 0},
		C: ast.Var{Ix: 0},
		D: ast.Var{Ix: 0},
		X: ast.Var{Ix: 0},
		Y: ast.Var{Ix: 0},
		P: ast.Var{Ix: 0},
	}

	result := Zonk(solutions, term)
	j, ok := result.(ast.J)
	if !ok {
		t.Fatalf("expected J, got %T", result)
	}

	if _, isSort := j.A.(ast.Sort); !isSort {
		t.Error("expected A to be zonked")
	}
}

func TestZonkPathLamAndApp(t *testing.T) {
	solutions := map[int]ast.Term{
		0: ast.Var{Ix: 0},
	}

	term1 := ast.PathLam{Binder: "i", Body: ast.Meta{ID: 0}}
	result1 := Zonk(solutions, term1)
	if _, ok := result1.(ast.PathLam); !ok {
		t.Errorf("expected PathLam, got %T", result1)
	}

	term2 := ast.PathApp{P: ast.Var{Ix: 0}, R: ast.Meta{ID: 0}}
	result2 := Zonk(solutions, term2)
	if _, ok := result2.(ast.PathApp); !ok {
		t.Errorf("expected PathApp, got %T", result2)
	}
}

func TestZonkTransport(t *testing.T) {
	solutions := map[int]ast.Term{
		0: ast.Var{Ix: 0},
	}

	term := ast.Transport{A: ast.Meta{ID: 0}, E: ast.Meta{ID: 0}}
	result := Zonk(solutions, term)
	if _, ok := result.(ast.Transport); !ok {
		t.Errorf("expected Transport, got %T", result)
	}
}

func TestZonkNilAndUnknown(t *testing.T) {
	solutions := map[int]ast.Term{}

	// Nil should stay nil
	if Zonk(solutions, nil) != nil {
		t.Error("expected nil for nil input")
	}

	// Unsolved meta should stay as meta
	term := ast.Meta{ID: 0}
	result := Zonk(solutions, term)
	if _, ok := result.(ast.Meta); !ok {
		t.Errorf("expected unsolved Meta, got %T", result)
	}
}

func TestHasMetaComprehensive(t *testing.T) {
	// Test hasMeta on various term types
	tests := []struct {
		term    ast.Term
		hasMeta bool
	}{
		{nil, false},
		{ast.Var{Ix: 0}, false},
		{ast.Global{Name: "x"}, false},
		{ast.Sort{U: 0}, false},
		{ast.Meta{ID: 0}, true},
		{ast.Pi{A: ast.Meta{ID: 0}, B: ast.Var{Ix: 0}}, true},
		{ast.Lam{Body: ast.Meta{ID: 0}}, true},
		{ast.App{T: ast.Meta{ID: 0}, U: ast.Var{Ix: 0}}, true},
		{ast.Sigma{A: ast.Var{Ix: 0}, B: ast.Meta{ID: 0}}, true},
		{ast.Pair{Fst: ast.Var{Ix: 0}, Snd: ast.Meta{ID: 0}}, true},
		{ast.Fst{P: ast.Meta{ID: 0}}, true},
		{ast.Snd{P: ast.Meta{ID: 0}}, true},
		{ast.Let{Val: ast.Meta{ID: 0}}, true},
		{ast.Id{A: ast.Meta{ID: 0}}, true},
		{ast.Refl{A: ast.Meta{ID: 0}}, true},
		{ast.J{P: ast.Meta{ID: 0}}, true},
		{ast.I0{}, false},
		{ast.I1{}, false},
	}

	for _, tt := range tests {
		if hasMeta(tt.term) != tt.hasMeta {
			t.Errorf("hasMeta(%T) = %v, want %v", tt.term, hasMeta(tt.term), tt.hasMeta)
		}
	}
}

func TestOccursCheckComprehensive(t *testing.T) {
	u := NewUnifier()

	// Direct occurrence
	if !u.occurs(0, ast.Meta{ID: 0}) {
		t.Error("expected ?0 to occur in ?0")
	}

	// Nested in args
	if !u.occurs(0, ast.Meta{ID: 1, Args: []ast.Term{ast.Meta{ID: 0}}}) {
		t.Error("expected ?0 to occur in ?1[?0]")
	}

	// Various term types
	tests := []struct {
		term   ast.Term
		occurs bool
	}{
		{nil, false},
		{ast.Var{Ix: 0}, false},
		{ast.Global{Name: "x"}, false},
		{ast.Sort{U: 0}, false},
		{ast.Pi{A: ast.Meta{ID: 0}}, true},
		{ast.Lam{Ann: ast.Meta{ID: 0}}, true},
		{ast.App{U: ast.Meta{ID: 0}}, true},
		{ast.Sigma{B: ast.Meta{ID: 0}}, true},
		{ast.Pair{Snd: ast.Meta{ID: 0}}, true},
		{ast.Fst{P: ast.Meta{ID: 0}}, true},
		{ast.Snd{P: ast.Meta{ID: 0}}, true},
		{ast.Let{Body: ast.Meta{ID: 0}}, true},
		{ast.Id{Y: ast.Meta{ID: 0}}, true},
		{ast.Refl{X: ast.Meta{ID: 0}}, true},
		{ast.J{C: ast.Meta{ID: 0}}, true},
		{ast.Path{X: ast.Meta{ID: 0}}, true},
		{ast.PathP{Y: ast.Meta{ID: 0}}, true},
		{ast.PathLam{Body: ast.Meta{ID: 0}}, true},
		{ast.PathApp{R: ast.Meta{ID: 0}}, true},
		{ast.Transport{E: ast.Meta{ID: 0}}, true},
		{ast.I0{}, false},
	}

	for _, tt := range tests {
		if u.occurs(0, tt.term) != tt.occurs {
			t.Errorf("occurs(0, %T) = %v, want %v", tt.term, u.occurs(0, tt.term), tt.occurs)
		}
	}
}

func TestIsPattern(t *testing.T) {
	u := NewUnifier()

	// Valid pattern - distinct variables
	if !u.isPattern(ast.Meta{ID: 0, Args: []ast.Term{ast.Var{Ix: 0}, ast.Var{Ix: 1}}}) {
		t.Error("expected valid pattern")
	}

	// Invalid - duplicate variables
	if u.isPattern(ast.Meta{ID: 0, Args: []ast.Term{ast.Var{Ix: 0}, ast.Var{Ix: 0}}}) {
		t.Error("expected invalid pattern for duplicate vars")
	}

	// Invalid - non-variable argument
	if u.isPattern(ast.Meta{ID: 0, Args: []ast.Term{ast.Sort{U: 0}}}) {
		t.Error("expected invalid pattern for non-variable arg")
	}
}

func TestCheckScope(t *testing.T) {
	u := NewUnifier()
	scope := map[int]int{0: 0, 1: 1}

	// Variable in scope
	if !u.checkScope(ast.Var{Ix: 0}, scope) {
		t.Error("expected var 0 in scope")
	}

	// Variable not in scope
	if u.checkScope(ast.Var{Ix: 5}, scope) {
		t.Error("expected var 5 not in scope")
	}

	// Nil
	if !u.checkScope(nil, scope) {
		t.Error("expected nil to be in scope")
	}

	// Global and Sort always in scope
	if !u.checkScope(ast.Global{Name: "x"}, scope) {
		t.Error("expected global in scope")
	}
	if !u.checkScope(ast.Sort{U: 0}, scope) {
		t.Error("expected sort in scope")
	}

	// Meta with args
	if !u.checkScope(ast.Meta{ID: 0, Args: []ast.Term{ast.Var{Ix: 0}}}, scope) {
		t.Error("expected meta in scope")
	}

	// Various term types with binders
	if !u.checkScope(ast.Pi{A: ast.Var{Ix: 0}, B: ast.Var{Ix: 0}}, scope) {
		t.Error("expected Pi in scope")
	}

	if !u.checkScope(ast.Lam{Body: ast.Var{Ix: 0}}, scope) {
		t.Error("expected Lam in scope")
	}

	if !u.checkScope(ast.Sigma{A: ast.Var{Ix: 0}, B: ast.Var{Ix: 0}}, scope) {
		t.Error("expected Sigma in scope")
	}

	if !u.checkScope(ast.Let{Val: ast.Var{Ix: 0}, Body: ast.Var{Ix: 0}}, scope) {
		t.Error("expected Let in scope")
	}
}

func TestShiftScope(t *testing.T) {
	u := NewUnifier()
	scope := map[int]int{0: 0, 1: 1}

	shifted := u.shiftScope(scope)

	if shifted[1] != 1 {
		t.Errorf("expected shifted[1] = 1, got %d", shifted[1])
	}
	if shifted[2] != 2 {
		t.Errorf("expected shifted[2] = 2, got %d", shifted[2])
	}
}

func TestShiftRenaming(t *testing.T) {
	u := NewUnifier()
	renaming := map[int]int{0: 5, 1: 6}

	shifted := u.shiftRenaming(renaming)

	if shifted[0] != 0 {
		t.Errorf("expected shifted[0] = 0, got %d", shifted[0])
	}
	if shifted[1] != 6 {
		t.Errorf("expected shifted[1] = 6, got %d", shifted[1])
	}
	if shifted[2] != 7 {
		t.Errorf("expected shifted[2] = 7, got %d", shifted[2])
	}
}

func TestRenameVars(t *testing.T) {
	u := NewUnifier()
	renaming := map[int]int{0: 5}

	// Simple variable
	result := u.renameVars(ast.Var{Ix: 0}, renaming)
	if v, ok := result.(ast.Var); !ok || v.Ix != 5 {
		t.Errorf("expected Var{Ix: 5}, got %v", result)
	}

	// Variable not in renaming
	result2 := u.renameVars(ast.Var{Ix: 1}, renaming)
	if v, ok := result2.(ast.Var); !ok || v.Ix != 1 {
		t.Errorf("expected Var{Ix: 1}, got %v", result2)
	}

	// Nil
	if u.renameVars(nil, renaming) != nil {
		t.Error("expected nil for nil input")
	}

	// Global and Sort unchanged
	if g, ok := u.renameVars(ast.Global{Name: "x"}, renaming).(ast.Global); !ok || g.Name != "x" {
		t.Error("expected unchanged global")
	}
	if s, ok := u.renameVars(ast.Sort{U: 0}, renaming).(ast.Sort); !ok || s.U != 0 {
		t.Error("expected unchanged sort")
	}

	// Various term types
	_ = u.renameVars(ast.Pi{A: ast.Var{Ix: 0}, B: ast.Var{Ix: 1}}, renaming)
	_ = u.renameVars(ast.Lam{Ann: ast.Var{Ix: 0}, Body: ast.Var{Ix: 1}}, renaming)
	_ = u.renameVars(ast.App{T: ast.Var{Ix: 0}, U: ast.Var{Ix: 1}}, renaming)
	_ = u.renameVars(ast.Sigma{A: ast.Var{Ix: 0}, B: ast.Var{Ix: 1}}, renaming)
	_ = u.renameVars(ast.Pair{Fst: ast.Var{Ix: 0}, Snd: ast.Var{Ix: 1}}, renaming)
	_ = u.renameVars(ast.Fst{P: ast.Var{Ix: 0}}, renaming)
	_ = u.renameVars(ast.Snd{P: ast.Var{Ix: 0}}, renaming)
	_ = u.renameVars(ast.Let{Ann: ast.Var{Ix: 0}, Val: ast.Var{Ix: 0}, Body: ast.Var{Ix: 1}}, renaming)
	_ = u.renameVars(ast.Id{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}, renaming)
	_ = u.renameVars(ast.Refl{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 0}}, renaming)
	_ = u.renameVars(ast.Meta{ID: 0, Args: []ast.Term{ast.Var{Ix: 0}}}, renaming)
}

func TestRenameAndShiftVarsDepth(t *testing.T) {
	u := NewUnifier()
	patternVars := map[int]int{0: 0, 1: 1}
	shift := 2

	// Variable bound within solution (ix < depth)
	result := u.renameAndShiftVarsDepth(ast.Var{Ix: 0}, patternVars, shift, 1)
	if v, ok := result.(ast.Var); !ok || v.Ix != 0 {
		t.Errorf("expected Var{Ix: 0} (local var), got %v", result)
	}

	// Pattern variable
	result2 := u.renameAndShiftVarsDepth(ast.Var{Ix: 0}, patternVars, shift, 0)
	if v, ok := result2.(ast.Var); !ok || v.Ix != 0 {
		t.Errorf("expected renamed pattern var, got %v", result2)
	}

	// Non-pattern variable - should be shifted
	result3 := u.renameAndShiftVarsDepth(ast.Var{Ix: 5}, patternVars, shift, 0)
	if v, ok := result3.(ast.Var); !ok || v.Ix != 7 {
		t.Errorf("expected Var{Ix: 7} (shifted), got %v", result3)
	}

	// Nil
	if u.renameAndShiftVarsDepth(nil, patternVars, shift, 0) != nil {
		t.Error("expected nil for nil input")
	}

	// Various term types
	_ = u.renameAndShiftVarsDepth(ast.Global{Name: "x"}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.Sort{U: 0}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.Pi{A: ast.Var{Ix: 0}, B: ast.Var{Ix: 1}}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.Lam{Body: ast.Var{Ix: 0}}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.App{T: ast.Var{Ix: 0}, U: ast.Var{Ix: 1}}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.Sigma{A: ast.Var{Ix: 0}, B: ast.Var{Ix: 1}}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.Pair{Fst: ast.Var{Ix: 0}, Snd: ast.Var{Ix: 1}}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.Fst{P: ast.Var{Ix: 0}}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.Snd{P: ast.Var{Ix: 0}}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.Let{Val: ast.Var{Ix: 0}, Body: ast.Var{Ix: 1}}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.Id{A: ast.Var{Ix: 0}}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.Refl{A: ast.Var{Ix: 0}}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.Meta{ID: 0, Args: []ast.Term{ast.Var{Ix: 0}}}, patternVars, shift, 0)
}

func TestConstraintStruct(t *testing.T) {
	c := Constraint{LHS: ast.Sort{U: 0}, RHS: ast.Sort{U: 1}}
	if _, ok := c.LHS.(ast.Sort); !ok {
		t.Error("expected Sort LHS")
	}
	if _, ok := c.RHS.(ast.Sort); !ok {
		t.Error("expected Sort RHS")
	}
}

func TestUnifyResultStruct(t *testing.T) {
	result := UnifyResult{
		Solved:   map[int]ast.Term{0: ast.Sort{U: 0}},
		Unsolved: []Constraint{{LHS: ast.Var{Ix: 0}, RHS: ast.Var{Ix: 1}}},
		Errors:   []UnifyError{{Message: "test"}},
	}

	if len(result.Solved) != 1 {
		t.Error("expected 1 solved")
	}
	if len(result.Unsolved) != 1 {
		t.Error("expected 1 unsolved")
	}
	if len(result.Errors) != 1 {
		t.Error("expected 1 error")
	}
}

func TestUnifyUnknownTerm(t *testing.T) {
	// Unknown term types should be deferred
	result := Unify(ast.I0{}, ast.I1{})
	// Should add to unsolved, not error
	if len(result.Unsolved) == 0 && len(result.Errors) == 0 {
		t.Error("expected unsolved or error for I0 vs I1")
	}
}

// --- Additional unify tests for coverage ---

func TestUnifyPiMismatch(t *testing.T) {
	// Pi vs non-Pi should fail
	lhs := ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	rhs := ast.Sort{U: 0}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for Pi vs Sort")
	}
}

func TestUnifyLamMismatch(t *testing.T) {
	// Lam vs non-Lam should fail
	lhs := ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}}
	rhs := ast.Sort{U: 0}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for Lam vs Sort")
	}
}

func TestUnifyAppMismatch(t *testing.T) {
	// App vs non-App should fail
	lhs := ast.App{T: ast.Var{Ix: 0}, U: ast.Var{Ix: 1}}
	rhs := ast.Sort{U: 0}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for App vs Sort")
	}
}

func TestUnifySigmaMismatch(t *testing.T) {
	// Sigma vs non-Sigma should fail
	lhs := ast.Sigma{Binder: "x", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	rhs := ast.Sort{U: 0}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for Sigma vs Sort")
	}
}

func TestUnifyPairMismatch(t *testing.T) {
	// Pair vs non-Pair should fail
	lhs := ast.Pair{Fst: ast.Sort{U: 0}, Snd: ast.Sort{U: 0}}
	rhs := ast.Sort{U: 0}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for Pair vs Sort")
	}
}

func TestUnifyFstMismatch(t *testing.T) {
	// Fst vs non-Fst should fail
	lhs := ast.Fst{P: ast.Var{Ix: 0}}
	rhs := ast.Sort{U: 0}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for Fst vs Sort")
	}
}

func TestUnifySndMismatch(t *testing.T) {
	// Snd vs non-Snd should fail
	lhs := ast.Snd{P: ast.Var{Ix: 0}}
	rhs := ast.Sort{U: 0}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for Snd vs Sort")
	}
}

func TestUnifyIdMismatch(t *testing.T) {
	// Id vs non-Id should fail
	lhs := ast.Id{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}
	rhs := ast.Sort{U: 0}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for Id vs Sort")
	}
}

func TestUnifyReflMismatch(t *testing.T) {
	// Refl vs non-Refl should fail
	lhs := ast.Refl{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}}
	rhs := ast.Sort{U: 0}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for Refl vs Sort")
	}
}

func TestUnifyPathMismatch(t *testing.T) {
	// Path vs non-Path should fail
	lhs := ast.Path{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}
	rhs := ast.Sort{U: 0}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for Path vs Sort")
	}
}

func TestUnifyPathPMismatch(t *testing.T) {
	// PathP vs non-PathP should fail
	lhs := ast.PathP{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}
	rhs := ast.Sort{U: 0}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for PathP vs Sort")
	}
}

func TestUnifyVarVsNonVar(t *testing.T) {
	// Var vs non-Var should fail
	lhs := ast.Var{Ix: 0}
	rhs := ast.Sort{U: 0}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for Var vs Sort")
	}
}

func TestUnifyGlobalVsNonGlobal(t *testing.T) {
	// Global vs non-Global should fail
	lhs := ast.Global{Name: "foo"}
	rhs := ast.Sort{U: 0}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for Global vs Sort")
	}
}

func TestUnifySortVsNonSort(t *testing.T) {
	// Sort vs non-Sort should fail
	lhs := ast.Sort{U: 0}
	rhs := ast.Var{Ix: 0}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for Sort vs Var")
	}
}

func TestZonkPath(t *testing.T) {
	solutions := map[int]ast.Term{
		0: ast.Sort{U: 0},
	}

	term := ast.Path{A: ast.Meta{ID: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}
	result := Zonk(solutions, term)
	if _, ok := result.(ast.Path); !ok {
		t.Errorf("expected Path, got %T", result)
	}
}

func TestZonkPathP(t *testing.T) {
	solutions := map[int]ast.Term{
		0: ast.Var{Ix: 0},
	}

	term := ast.PathP{A: ast.Meta{ID: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}
	result := Zonk(solutions, term)
	if _, ok := result.(ast.PathP); !ok {
		t.Errorf("expected PathP, got %T", result)
	}
}

func TestZonkId(t *testing.T) {
	solutions := map[int]ast.Term{
		0: ast.Sort{U: 0},
	}

	term := ast.Id{A: ast.Meta{ID: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}
	result := Zonk(solutions, term)
	if _, ok := result.(ast.Id); !ok {
		t.Errorf("expected Id, got %T", result)
	}
}

func TestZonkRefl(t *testing.T) {
	solutions := map[int]ast.Term{
		0: ast.Sort{U: 0},
	}

	term := ast.Refl{A: ast.Meta{ID: 0}, X: ast.Var{Ix: 0}}
	result := Zonk(solutions, term)
	if _, ok := result.(ast.Refl); !ok {
		t.Errorf("expected Refl, got %T", result)
	}
}

func TestZonkSigma(t *testing.T) {
	solutions := map[int]ast.Term{
		0: ast.Sort{U: 0},
	}

	term := ast.Sigma{Binder: "x", A: ast.Meta{ID: 0}, B: ast.Var{Ix: 0}}
	result := Zonk(solutions, term)
	if _, ok := result.(ast.Sigma); !ok {
		t.Errorf("expected Sigma, got %T", result)
	}
}

func TestZonkPair(t *testing.T) {
	solutions := map[int]ast.Term{
		0: ast.Sort{U: 0},
	}

	term := ast.Pair{Fst: ast.Meta{ID: 0}, Snd: ast.Var{Ix: 0}}
	result := Zonk(solutions, term)
	if _, ok := result.(ast.Pair); !ok {
		t.Errorf("expected Pair, got %T", result)
	}
}

func TestZonkFstSnd(t *testing.T) {
	solutions := map[int]ast.Term{
		0: ast.Var{Ix: 0},
	}

	term1 := ast.Fst{P: ast.Meta{ID: 0}}
	result1 := Zonk(solutions, term1)
	if _, ok := result1.(ast.Fst); !ok {
		t.Errorf("expected Fst, got %T", result1)
	}

	term2 := ast.Snd{P: ast.Meta{ID: 0}}
	result2 := Zonk(solutions, term2)
	if _, ok := result2.(ast.Snd); !ok {
		t.Errorf("expected Snd, got %T", result2)
	}
}

func TestZonkApp(t *testing.T) {
	solutions := map[int]ast.Term{
		0: ast.Var{Ix: 0},
	}

	term := ast.App{T: ast.Meta{ID: 0}, U: ast.Meta{ID: 0}}
	result := Zonk(solutions, term)
	if _, ok := result.(ast.App); !ok {
		t.Errorf("expected App, got %T", result)
	}
}

func TestZonkLam(t *testing.T) {
	solutions := map[int]ast.Term{
		0: ast.Var{Ix: 0},
	}

	term := ast.Lam{Binder: "x", Ann: ast.Meta{ID: 0}, Body: ast.Meta{ID: 0}}
	result := Zonk(solutions, term)
	if _, ok := result.(ast.Lam); !ok {
		t.Errorf("expected Lam, got %T", result)
	}
}

func TestCheckScopeComprehensive(t *testing.T) {
	u := NewUnifier()
	scope := map[int]int{0: 0, 1: 1}

	// Test more term types
	tests := []struct {
		term    ast.Term
		inScope bool
	}{
		{ast.I0{}, true},
		{ast.I1{}, true},
		{ast.Interval{}, true},
		{ast.App{T: ast.Var{Ix: 0}, U: ast.Var{Ix: 1}}, true},
		{ast.App{T: ast.Var{Ix: 5}, U: ast.Var{Ix: 0}}, false},
		{ast.Pair{Fst: ast.Var{Ix: 0}, Snd: ast.Var{Ix: 1}}, true},
		{ast.Pair{Fst: ast.Var{Ix: 0}, Snd: ast.Var{Ix: 5}}, false},
		{ast.Fst{P: ast.Var{Ix: 5}}, false},
		{ast.Snd{P: ast.Var{Ix: 5}}, false},
		{ast.Id{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 0}}, true},
		{ast.Id{A: ast.Var{Ix: 5}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}, false},
		{ast.Refl{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 1}}, true},
		{ast.Refl{A: ast.Var{Ix: 5}, X: ast.Var{Ix: 0}}, false},
		{ast.J{A: ast.Var{Ix: 0}, C: ast.Var{Ix: 1}, D: ast.Var{Ix: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 0}, P: ast.Var{Ix: 1}}, true},
		{ast.Path{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 0}}, true},
		{ast.PathP{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 0}}, true},
		{ast.PathLam{Binder: "i", Body: ast.Var{Ix: 0}}, true},
		{ast.PathApp{P: ast.Var{Ix: 0}, R: ast.Var{Ix: 1}}, true},
		{ast.Transport{A: ast.Var{Ix: 0}, E: ast.Var{Ix: 1}}, true},
	}

	for _, tt := range tests {
		if u.checkScope(tt.term, scope) != tt.inScope {
			t.Errorf("checkScope(%T) = %v, want %v", tt.term, u.checkScope(tt.term, scope), tt.inScope)
		}
	}
}

func TestRenameVarsComprehensive(t *testing.T) {
	u := NewUnifier()
	renaming := map[int]int{0: 5, 1: 6}

	// Test more term types
	_ = u.renameVars(ast.J{A: ast.Var{Ix: 0}, C: ast.Var{Ix: 1}, D: ast.Var{Ix: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 0}, P: ast.Var{Ix: 1}}, renaming)
	_ = u.renameVars(ast.Path{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 0}}, renaming)
	_ = u.renameVars(ast.PathP{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 0}}, renaming)
	_ = u.renameVars(ast.PathLam{Binder: "i", Body: ast.Var{Ix: 0}}, renaming)
	_ = u.renameVars(ast.PathApp{P: ast.Var{Ix: 0}, R: ast.Var{Ix: 1}}, renaming)
	_ = u.renameVars(ast.Transport{A: ast.Var{Ix: 0}, E: ast.Var{Ix: 1}}, renaming)
	_ = u.renameVars(ast.I0{}, renaming)
	_ = u.renameVars(ast.I1{}, renaming)
	_ = u.renameVars(ast.Interval{}, renaming)
}

func TestRenameAndShiftVarsDepthComprehensive(t *testing.T) {
	u := NewUnifier()
	patternVars := map[int]int{0: 0, 1: 1}
	shift := 2

	// Test more term types
	_ = u.renameAndShiftVarsDepth(ast.J{A: ast.Var{Ix: 0}}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.Path{A: ast.Var{Ix: 0}}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.PathP{A: ast.Var{Ix: 0}}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.PathLam{Body: ast.Var{Ix: 0}}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.PathApp{P: ast.Var{Ix: 0}}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.Transport{A: ast.Var{Ix: 0}}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.I0{}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.I1{}, patternVars, shift, 0)
	_ = u.renameAndShiftVarsDepth(ast.Interval{}, patternVars, shift, 0)
}

func TestUnifyPiCodDomainFails(t *testing.T) {
	// Domain unifies, but codomain fails
	lhs := ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	rhs := ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Sort{U: 1}} // Different codomain

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for different codomains")
	}
}

func TestUnifyAppFnFails(t *testing.T) {
	// Function parts don't unify
	lhs := ast.App{T: ast.Var{Ix: 0}, U: ast.Sort{U: 0}}
	rhs := ast.App{T: ast.Var{Ix: 1}, U: ast.Sort{U: 0}}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for different function heads")
	}
}

func TestUnifySigmaAFails(t *testing.T) {
	// First component doesn't unify
	lhs := ast.Sigma{Binder: "x", A: ast.Sort{U: 0}, B: ast.Sort{U: 0}}
	rhs := ast.Sigma{Binder: "x", A: ast.Sort{U: 1}, B: ast.Sort{U: 0}}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for different first components")
	}
}

func TestUnifyPairFstFails(t *testing.T) {
	// First element doesn't unify
	lhs := ast.Pair{Fst: ast.Sort{U: 0}, Snd: ast.Sort{U: 0}}
	rhs := ast.Pair{Fst: ast.Sort{U: 1}, Snd: ast.Sort{U: 0}}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for different first elements")
	}
}

func TestUnifyIdAFails(t *testing.T) {
	// Type component doesn't unify
	lhs := ast.Id{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}
	rhs := ast.Id{A: ast.Sort{U: 1}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for different type component")
	}
}

func TestUnifyIdXFails(t *testing.T) {
	// X component doesn't unify
	lhs := ast.Id{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}
	rhs := ast.Id{A: ast.Sort{U: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 0}}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for different X component")
	}
}

func TestUnifyReflAFails(t *testing.T) {
	// Type component doesn't unify
	lhs := ast.Refl{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}}
	rhs := ast.Refl{A: ast.Sort{U: 1}, X: ast.Var{Ix: 0}}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for different type component")
	}
}

func TestUnifyPathAFails(t *testing.T) {
	// Type component doesn't unify
	lhs := ast.Path{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}
	rhs := ast.Path{A: ast.Sort{U: 1}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for different type component")
	}
}

func TestUnifyPathXFails(t *testing.T) {
	// X component doesn't unify
	lhs := ast.Path{A: ast.Sort{U: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}
	rhs := ast.Path{A: ast.Sort{U: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 0}}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for different X component")
	}
}

func TestUnifyPathPAFails(t *testing.T) {
	// Line type doesn't unify
	lhs := ast.PathP{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}
	rhs := ast.PathP{A: ast.Var{Ix: 1}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for different line type")
	}
}

func TestUnifyPathPXFails(t *testing.T) {
	// X component doesn't unify
	lhs := ast.PathP{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 0}, Y: ast.Var{Ix: 0}}
	rhs := ast.PathP{A: ast.Var{Ix: 0}, X: ast.Var{Ix: 1}, Y: ast.Var{Ix: 0}}

	result := Unify(lhs, rhs)
	if len(result.Errors) == 0 {
		t.Error("expected error for different X component")
	}
}
