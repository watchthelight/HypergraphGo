package ast_test

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

func TestSprint_Id(t *testing.T) {
	// id : Π(A:Type0). A → A  represented as λ . Var0
	id := ast.Lam{
		Binder: "x",
		Ann:    nil,
		Body:   ast.Var{Ix: 0},
	}
	got := ast.Sprint(id)
	want := "(\\x => {0})"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSprint_Pi(t *testing.T) {
	// (Pi x:Type0 . x)
	pi := ast.Pi{
		Binder: "x",
		A:      ast.Sort{U: 0},
		B:      ast.Var{Ix: 0},
	}
	got := ast.Sprint(pi)
	want := "(Pi x: Type0 . {0})"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSprint_AppChain(t *testing.T) {
	tm := ast.MkApps(ast.Global{Name: "f"}, ast.Var{Ix: 0}, ast.Var{Ix: 1})
	got := ast.Sprint(tm)
	want := "(f {0} {1})"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

// ============================================================================
// HIT Term Structure Tests
// ============================================================================

func TestHITSpec_IsHIT(t *testing.T) {
	tests := []struct {
		name     string
		spec     ast.HITSpec
		expected bool
	}{
		{
			name: "with path constructors is HIT",
			spec: ast.HITSpec{
				Name: "S1",
				PathCtors: []ast.PathConstructor{
					{Name: "loop", Level: 1},
				},
			},
			expected: true,
		},
		{
			name: "without path constructors is not HIT",
			spec: ast.HITSpec{
				Name:      "Nat",
				PathCtors: nil,
			},
			expected: false,
		},
		{
			name: "empty path constructors is not HIT",
			spec: ast.HITSpec{
				Name:      "Bool",
				PathCtors: []ast.PathConstructor{},
			},
			expected: false,
		},
		{
			name: "multiple path constructors is HIT",
			spec: ast.HITSpec{
				Name: "Quot",
				PathCtors: []ast.PathConstructor{
					{Name: "eq", Level: 1},
					{Name: "trunc", Level: 2},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.IsHIT()
			if got != tt.expected {
				t.Errorf("HITSpec{%s}.IsHIT() = %v, want %v", tt.spec.Name, got, tt.expected)
			}
		})
	}
}

func TestHITSpec_MaxLevel(t *testing.T) {
	tests := []struct {
		name     string
		spec     ast.HITSpec
		expected int
	}{
		{
			name: "no path constructors returns 0",
			spec: ast.HITSpec{
				Name:      "Nat",
				PathCtors: nil,
			},
			expected: 0,
		},
		{
			name: "single level-1 path constructor",
			spec: ast.HITSpec{
				Name: "S1",
				PathCtors: []ast.PathConstructor{
					{Name: "loop", Level: 1},
				},
			},
			expected: 1,
		},
		{
			name: "single level-2 path constructor",
			spec: ast.HITSpec{
				Name: "Trunc",
				PathCtors: []ast.PathConstructor{
					{Name: "squash", Level: 2},
				},
			},
			expected: 2,
		},
		{
			name: "mixed levels returns max",
			spec: ast.HITSpec{
				Name: "Complex",
				PathCtors: []ast.PathConstructor{
					{Name: "path1", Level: 1},
					{Name: "path2", Level: 3},
					{Name: "path3", Level: 2},
				},
			},
			expected: 3,
		},
		{
			name: "all same level",
			spec: ast.HITSpec{
				Name: "Multi",
				PathCtors: []ast.PathConstructor{
					{Name: "a", Level: 2},
					{Name: "b", Level: 2},
					{Name: "c", Level: 2},
				},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.MaxLevel()
			if got != tt.expected {
				t.Errorf("HITSpec{%s}.MaxLevel() = %d, want %d", tt.spec.Name, got, tt.expected)
			}
		})
	}
}

func TestPathConstructor_Structure(t *testing.T) {
	// Test PathConstructor field access and structure
	pc := ast.PathConstructor{
		Name:  "loop",
		Level: 1,
		Type:  ast.Path{A: ast.Global{Name: "S1"}, X: ast.Global{Name: "base"}, Y: ast.Global{Name: "base"}},
		Boundaries: []ast.Boundary{
			{AtZero: ast.Global{Name: "base"}, AtOne: ast.Global{Name: "base"}},
		},
	}

	if pc.Name != "loop" {
		t.Errorf("PathConstructor.Name = %q, want %q", pc.Name, "loop")
	}
	if pc.Level != 1 {
		t.Errorf("PathConstructor.Level = %d, want %d", pc.Level, 1)
	}
	if len(pc.Boundaries) != 1 {
		t.Errorf("len(PathConstructor.Boundaries) = %d, want %d", len(pc.Boundaries), 1)
	}
	if _, ok := pc.Type.(ast.Path); !ok {
		t.Errorf("PathConstructor.Type should be ast.Path")
	}
}

func TestBoundary_Structure(t *testing.T) {
	// Test Boundary field access
	b := ast.Boundary{
		AtZero: ast.Global{Name: "left"},
		AtOne:  ast.Global{Name: "right"},
	}

	if g, ok := b.AtZero.(ast.Global); !ok || g.Name != "left" {
		t.Errorf("Boundary.AtZero = %v, want Global{left}", b.AtZero)
	}
	if g, ok := b.AtOne.(ast.Global); !ok || g.Name != "right" {
		t.Errorf("Boundary.AtOne = %v, want Global{right}", b.AtOne)
	}
}

func TestHITApp_Structure(t *testing.T) {
	// Test HITApp field access and isCoreTerm interface
	hitApp := ast.HITApp{
		HITName: "S1",
		Ctor:    "loop",
		Args:    []ast.Term{ast.Global{Name: "A"}},
		IArgs:   []ast.Term{ast.I0{}, ast.I1{}},
	}

	if hitApp.HITName != "S1" {
		t.Errorf("HITApp.HITName = %q, want %q", hitApp.HITName, "S1")
	}
	if hitApp.Ctor != "loop" {
		t.Errorf("HITApp.Ctor = %q, want %q", hitApp.Ctor, "loop")
	}
	if len(hitApp.Args) != 1 {
		t.Errorf("len(HITApp.Args) = %d, want %d", len(hitApp.Args), 1)
	}
	if len(hitApp.IArgs) != 2 {
		t.Errorf("len(HITApp.IArgs) = %d, want %d", len(hitApp.IArgs), 2)
	}

	// Verify HITApp implements Term interface (via isCoreTerm)
	var _ ast.Term = hitApp
}

func TestConstructor_Structure(t *testing.T) {
	// Test Constructor field access
	ctor := ast.Constructor{
		Name: "zero",
		Type: ast.Global{Name: "Nat"},
	}

	if ctor.Name != "zero" {
		t.Errorf("Constructor.Name = %q, want %q", ctor.Name, "zero")
	}
	if g, ok := ctor.Type.(ast.Global); !ok || g.Name != "Nat" {
		t.Errorf("Constructor.Type = %v, want Global{Nat}", ctor.Type)
	}
}

func TestHITSpec_FullStructure(t *testing.T) {
	// Test a complete HITSpec (like Circle)
	s1Spec := ast.HITSpec{
		Name:       "S1",
		Type:       ast.Sort{U: 0},
		NumParams:  0,
		ParamTypes: nil,
		PointCtors: []ast.Constructor{
			{Name: "base", Type: ast.Global{Name: "S1"}},
		},
		PathCtors: []ast.PathConstructor{
			{
				Name:  "loop",
				Level: 1,
				Type: ast.Path{
					A: ast.Global{Name: "S1"},
					X: ast.Global{Name: "base"},
					Y: ast.Global{Name: "base"},
				},
				Boundaries: []ast.Boundary{
					{AtZero: ast.Global{Name: "base"}, AtOne: ast.Global{Name: "base"}},
				},
			},
		},
		Eliminator: "S1-elim",
	}

	if s1Spec.Name != "S1" {
		t.Errorf("HITSpec.Name = %q, want %q", s1Spec.Name, "S1")
	}
	if !s1Spec.IsHIT() {
		t.Error("S1 should be a HIT")
	}
	if s1Spec.MaxLevel() != 1 {
		t.Errorf("S1.MaxLevel() = %d, want %d", s1Spec.MaxLevel(), 1)
	}
	if len(s1Spec.PointCtors) != 1 {
		t.Errorf("len(S1.PointCtors) = %d, want %d", len(s1Spec.PointCtors), 1)
	}
	if len(s1Spec.PathCtors) != 1 {
		t.Errorf("len(S1.PathCtors) = %d, want %d", len(s1Spec.PathCtors), 1)
	}
	if s1Spec.Eliminator != "S1-elim" {
		t.Errorf("S1.Eliminator = %q, want %q", s1Spec.Eliminator, "S1-elim")
	}
}

// ============================================================================
// MkApps Tests
// ============================================================================

func TestMkApps_Empty(t *testing.T) {
	// MkApps with just a function, no args
	result := ast.MkApps(ast.Global{Name: "f"})
	if g, ok := result.(ast.Global); !ok || g.Name != "f" {
		t.Errorf("MkApps(f) = %v, want Global{f}", result)
	}
}

func TestMkApps_SingleArg(t *testing.T) {
	result := ast.MkApps(ast.Global{Name: "f"}, ast.Global{Name: "x"})
	app, ok := result.(ast.App)
	if !ok {
		t.Fatalf("MkApps(f, x) should be App, got %T", result)
	}
	if g, ok := app.T.(ast.Global); !ok || g.Name != "f" {
		t.Errorf("App.T = %v, want Global{f}", app.T)
	}
	if g, ok := app.U.(ast.Global); !ok || g.Name != "x" {
		t.Errorf("App.U = %v, want Global{x}", app.U)
	}
}

func TestMkApps_MultipleArgs(t *testing.T) {
	// (f x y z) = App(App(App(f, x), y), z)
	result := ast.MkApps(ast.Global{Name: "f"}, ast.Global{Name: "x"}, ast.Global{Name: "y"}, ast.Global{Name: "z"})

	// Outer is App with z
	app1, ok := result.(ast.App)
	if !ok {
		t.Fatalf("result should be App")
	}
	if g, ok := app1.U.(ast.Global); !ok || g.Name != "z" {
		t.Errorf("outermost arg should be z")
	}

	// Inner is App with y
	app2, ok := app1.T.(ast.App)
	if !ok {
		t.Fatalf("inner should be App")
	}
	if g, ok := app2.U.(ast.Global); !ok || g.Name != "y" {
		t.Errorf("second arg should be y")
	}
}

// ============================================================================
// Sort.IsZeroLevel Tests
// ============================================================================

func TestSort_IsZeroLevel(t *testing.T) {
	tests := []struct {
		level    ast.Level
		expected bool
	}{
		{0, true},
		{1, false},
		{10, false},
	}

	for _, tt := range tests {
		s := ast.Sort{U: tt.level}
		got := s.IsZeroLevel()
		if got != tt.expected {
			t.Errorf("Sort{%d}.IsZeroLevel() = %v, want %v", tt.level, got, tt.expected)
		}
	}
}
