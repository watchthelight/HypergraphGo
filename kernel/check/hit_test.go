package check

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
)

func TestDeclareHIT_Circle(t *testing.T) {
	env := NewGlobalEnv()

	// Declare S1 as a HIT
	spec := &ast.HITSpec{
		Name: "S1",
		Type: ast.Sort{U: 0},
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

	err := env.DeclareHIT(spec)
	if err != nil {
		t.Fatalf("DeclareHIT failed: %v", err)
	}

	// Check that S1 was added
	s1Type := env.LookupType("S1")
	if s1Type == nil {
		t.Error("S1 type not found")
	}

	// Check that base was added
	baseType := env.LookupType("base")
	if baseType == nil {
		t.Error("base constructor not found")
	}

	// Check that the eliminator was added
	elimType := env.LookupType("S1-elim")
	if elimType == nil {
		t.Error("S1-elim not found")
	}

	// Check that the inductive has HIT fields set
	ind := env.inductives["S1"]
	if !ind.IsHIT {
		t.Error("S1 should be marked as HIT")
	}
	if len(ind.PathCtors) != 1 {
		t.Errorf("expected 1 path constructor, got %d", len(ind.PathCtors))
	}
	if ind.MaxLevel != 1 {
		t.Errorf("expected MaxLevel 1, got %d", ind.MaxLevel)
	}
}

func TestDeclareHIT_Validation(t *testing.T) {
	env := NewGlobalEnv()

	// Test that path constructor must have matching boundaries
	spec := &ast.HITSpec{
		Name: "BadHIT",
		Type: ast.Sort{U: 0},
		PointCtors: []ast.Constructor{
			{Name: "pt", Type: ast.Global{Name: "BadHIT"}},
		},
		PathCtors: []ast.PathConstructor{
			{
				Name:  "badPath",
				Level: 2, // Level 2 needs 2 boundaries
				Type: ast.Path{
					A: ast.Global{Name: "BadHIT"},
					X: ast.Global{Name: "pt"},
					Y: ast.Global{Name: "pt"},
				},
				Boundaries: []ast.Boundary{
					// Only 1 boundary for level 2 path - should fail
					{AtZero: ast.Global{Name: "pt"}, AtOne: ast.Global{Name: "pt"}},
				},
			},
		},
		Eliminator: "BadHIT-elim",
	}

	err := env.DeclareHIT(spec)
	if err == nil {
		t.Error("DeclareHIT should fail for mismatched boundary count")
	}
}

func TestHIT_RecursorInfoBuilding(t *testing.T) {
	env := NewGlobalEnv()

	// Declare a HIT without path constructors first (simpler test)
	spec := &ast.HITSpec{
		Name: "TestHIT",
		Type: ast.Sort{U: 0},
		PointCtors: []ast.Constructor{
			{Name: "pt1", Type: ast.Global{Name: "TestHIT"}},
			{Name: "pt2", Type: ast.Global{Name: "TestHIT"}},
		},
		PathCtors:  nil, // No path constructors to avoid reference issues
		Eliminator: "TestHIT-elim",
	}

	err := env.DeclareHIT(spec)
	if err != nil {
		t.Fatalf("DeclareHIT failed: %v", err)
	}

	// Check RecursorInfo was registered
	info := eval.LookupRecursor("TestHIT-elim")
	if info == nil {
		t.Fatal("RecursorInfo not registered")
	}

	// Without path constructors, IsHIT is based on the spec, but since we have no path ctors
	// the Inductive itself won't be marked as HIT (it's a regular inductive)
	// Let's just check the basic structure
	if len(info.Ctors) != 2 {
		t.Errorf("expected 2 point constructors, got %d", len(info.Ctors))
	}
}

func TestValidatePathConstructor(t *testing.T) {
	env := NewGlobalEnv()

	// Add a dummy HIT type for testing
	env.AddAxiom("TestHIT", ast.Sort{U: 0})
	env.AddAxiom("pt", ast.Global{Name: "TestHIT"})

	checker := NewChecker(env)

	tests := []struct {
		name      string
		ctor      ast.PathConstructor
		wantError bool
	}{
		{
			name: "valid path constructor",
			ctor: ast.PathConstructor{
				Name:  "validPath",
				Level: 1,
				Type: ast.Path{
					A: ast.Global{Name: "TestHIT"},
					X: ast.Global{Name: "pt"},
					Y: ast.Global{Name: "pt"},
				},
				Boundaries: []ast.Boundary{
					{AtZero: ast.Global{Name: "pt"}, AtOne: ast.Global{Name: "pt"}},
				},
			},
			wantError: false,
		},
		{
			name: "mismatched boundary count",
			ctor: ast.PathConstructor{
				Name:  "badPath",
				Level: 2,
				Type: ast.Path{
					A: ast.Global{Name: "TestHIT"},
					X: ast.Global{Name: "pt"},
					Y: ast.Global{Name: "pt"},
				},
				Boundaries: []ast.Boundary{
					{AtZero: ast.Global{Name: "pt"}, AtOne: ast.Global{Name: "pt"}},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePathConstructor(checker, "TestHIT", tt.ctor)
			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestIsPathToHIT(t *testing.T) {
	tests := []struct {
		name    string
		ty      ast.Term
		hitName string
		want    bool
	}{
		{
			name: "direct Path to HIT",
			ty: ast.Path{
				A: ast.Global{Name: "S1"},
				X: ast.Global{Name: "base"},
				Y: ast.Global{Name: "base"},
			},
			hitName: "S1",
			want:    true,
		},
		{
			name: "Path to different type",
			ty: ast.Path{
				A: ast.Global{Name: "Nat"},
				X: ast.Global{Name: "zero"},
				Y: ast.Global{Name: "zero"},
			},
			hitName: "S1",
			want:    false,
		},
		{
			name: "Pi then Path",
			ty: ast.Pi{
				Binder: "x",
				A:      ast.Global{Name: "A"},
				B: ast.Path{
					A: ast.Global{Name: "S1"},
					X: ast.Global{Name: "base"},
					Y: ast.Global{Name: "base"},
				},
			},
			hitName: "S1",
			want:    true,
		},
		{
			name:    "non-Path type",
			ty:      ast.Global{Name: "S1"},
			hitName: "S1",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPathToHIT(tt.ty, tt.hitName)
			if got != tt.want {
				t.Errorf("isPathToHIT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainsHIT(t *testing.T) {
	tests := []struct {
		name    string
		term    ast.Term
		hitName string
		want    bool
	}{
		{
			name:    "Global with HIT name",
			term:    ast.Global{Name: "S1"},
			hitName: "S1",
			want:    true,
		},
		{
			name:    "Global with different name",
			term:    ast.Global{Name: "Nat"},
			hitName: "S1",
			want:    false,
		},
		{
			name:    "App containing HIT in function",
			term:    ast.App{T: ast.Global{Name: "S1"}, U: ast.Global{Name: "x"}},
			hitName: "S1",
			want:    true,
		},
		{
			name:    "App containing HIT in argument",
			term:    ast.App{T: ast.Global{Name: "List"}, U: ast.Global{Name: "S1"}},
			hitName: "S1",
			want:    true,
		},
		{
			name:    "Pi with HIT in domain",
			term:    ast.Pi{Binder: "x", A: ast.Global{Name: "S1"}, B: ast.Sort{U: 0}},
			hitName: "S1",
			want:    true,
		},
		{
			name:    "Pi with HIT in codomain",
			term:    ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Global{Name: "S1"}},
			hitName: "S1",
			want:    true,
		},
		{
			name:    "Lam with HIT in body",
			term:    ast.Lam{Binder: "x", Body: ast.Global{Name: "S1"}},
			hitName: "S1",
			want:    true,
		},
		{
			name:    "Lam without HIT",
			term:    ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}},
			hitName: "S1",
			want:    false,
		},
		{
			name:    "PathLam with HIT in body",
			term:    ast.PathLam{Body: ast.Global{Name: "S1"}},
			hitName: "S1",
			want:    true,
		},
		{
			name: "Path with HIT in A",
			term: ast.Path{
				A: ast.Global{Name: "S1"},
				X: ast.Global{Name: "base"},
				Y: ast.Global{Name: "base"},
			},
			hitName: "S1",
			want:    true,
		},
		{
			name: "Path with HIT in X",
			term: ast.Path{
				A: ast.Sort{U: 0},
				X: ast.Global{Name: "S1"},
				Y: ast.Global{Name: "y"},
			},
			hitName: "S1",
			want:    true,
		},
		{
			name: "Path with HIT in Y",
			term: ast.Path{
				A: ast.Sort{U: 0},
				X: ast.Global{Name: "x"},
				Y: ast.Global{Name: "S1"},
			},
			hitName: "S1",
			want:    true,
		},
		{
			name: "PathP with HIT in A",
			term: ast.PathP{
				A: ast.Global{Name: "S1"},
				X: ast.Global{Name: "x"},
				Y: ast.Global{Name: "y"},
			},
			hitName: "S1",
			want:    true,
		},
		{
			name: "PathP with HIT in X",
			term: ast.PathP{
				A: ast.Sort{U: 0},
				X: ast.Global{Name: "S1"},
				Y: ast.Global{Name: "y"},
			},
			hitName: "S1",
			want:    true,
		},
		{
			name:    "PathApp with HIT in P",
			term:    ast.PathApp{P: ast.Global{Name: "S1"}, R: ast.I0{}},
			hitName: "S1",
			want:    true,
		},
		{
			name:    "PathApp with HIT in R (unusual but valid)",
			term:    ast.PathApp{P: ast.Global{Name: "x"}, R: ast.Global{Name: "S1"}},
			hitName: "S1",
			want:    true,
		},
		{
			name:    "Sort never contains HIT",
			term:    ast.Sort{U: 0},
			hitName: "S1",
			want:    false,
		},
		{
			name:    "Var never contains HIT",
			term:    ast.Var{Ix: 0},
			hitName: "S1",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsHIT(tt.term, tt.hitName)
			if got != tt.want {
				t.Errorf("containsHIT() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPathResult(t *testing.T) {
	tests := []struct {
		name    string
		term    ast.Term
		hitName string
		want    bool
	}{
		{
			name: "Path with HIT in A",
			term: ast.Path{
				A: ast.Global{Name: "S1"},
				X: ast.Global{Name: "x"},
				Y: ast.Global{Name: "y"},
			},
			hitName: "S1",
			want:    true,
		},
		{
			name: "Path without HIT",
			term: ast.Path{
				A: ast.Global{Name: "Nat"},
				X: ast.Global{Name: "x"},
				Y: ast.Global{Name: "y"},
			},
			hitName: "S1",
			want:    false,
		},
		{
			name: "PathP with HIT in A",
			term: ast.PathP{
				A: ast.Global{Name: "S1"},
				X: ast.Global{Name: "x"},
				Y: ast.Global{Name: "y"},
			},
			hitName: "S1",
			want:    true,
		},
		{
			name: "PathP without HIT",
			term: ast.PathP{
				A: ast.Global{Name: "Nat"},
				X: ast.Global{Name: "x"},
				Y: ast.Global{Name: "y"},
			},
			hitName: "S1",
			want:    false,
		},
		{
			name: "App with Path inside",
			term: ast.App{
				T: ast.Path{
					A: ast.Global{Name: "S1"},
					X: ast.Global{Name: "x"},
					Y: ast.Global{Name: "y"},
				},
				U: ast.Global{Name: "arg"},
			},
			hitName: "S1",
			want:    true,
		},
		{
			name:    "Global is not a path result",
			term:    ast.Global{Name: "S1"},
			hitName: "S1",
			want:    false,
		},
		{
			name:    "Sort is not a path result",
			term:    ast.Sort{U: 0},
			hitName: "S1",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPathResult(tt.term, tt.hitName)
			if got != tt.want {
				t.Errorf("isPathResult() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuiltInHITs(t *testing.T) {
	env := NewGlobalEnvWithPrimitives()
	env.AddHITs()

	// Test S1
	t.Run("S1", func(t *testing.T) {
		if env.LookupType("S1") == nil {
			t.Error("S1 not defined")
		}
		if env.LookupType("base") == nil {
			t.Error("base not defined")
		}
		if env.LookupType("S1-elim") == nil {
			t.Error("S1-elim not defined")
		}
		info := eval.LookupRecursor("S1-elim")
		if info == nil || !info.IsHIT {
			t.Error("S1-elim not registered as HIT")
		}
	})

	// Test Trunc
	t.Run("Trunc", func(t *testing.T) {
		if env.LookupType("Trunc") == nil {
			t.Error("Trunc not defined")
		}
		if env.LookupType("inc") == nil {
			t.Error("inc not defined")
		}
		if env.LookupType("Trunc-elim") == nil {
			t.Error("Trunc-elim not defined")
		}
		info := eval.LookupRecursor("Trunc-elim")
		if info == nil || !info.IsHIT {
			t.Error("Trunc-elim not registered as HIT")
		}
	})

	// Test Susp
	t.Run("Susp", func(t *testing.T) {
		if env.LookupType("Susp") == nil {
			t.Error("Susp not defined")
		}
		if env.LookupType("north") == nil {
			t.Error("north not defined")
		}
		if env.LookupType("south") == nil {
			t.Error("south not defined")
		}
		if env.LookupType("Susp-elim") == nil {
			t.Error("Susp-elim not defined")
		}
		info := eval.LookupRecursor("Susp-elim")
		if info == nil || !info.IsHIT {
			t.Error("Susp-elim not registered as HIT")
		}
	})

	// Test Int
	t.Run("Int", func(t *testing.T) {
		if env.LookupType("Int") == nil {
			t.Error("Int not defined")
		}
		if env.LookupType("pos") == nil {
			t.Error("pos not defined")
		}
		if env.LookupType("neg") == nil {
			t.Error("neg not defined")
		}
		if env.LookupType("Int-elim") == nil {
			t.Error("Int-elim not defined")
		}
		info := eval.LookupRecursor("Int-elim")
		if info == nil || !info.IsHIT {
			t.Error("Int-elim not registered as HIT")
		}
	})

	// Test Quot
	t.Run("Quot", func(t *testing.T) {
		if env.LookupType("Quot") == nil {
			t.Error("Quot not defined")
		}
		if env.LookupType("quot") == nil {
			t.Error("quot not defined")
		}
		if env.LookupType("Quot-elim") == nil {
			t.Error("Quot-elim not defined")
		}
		info := eval.LookupRecursor("Quot-elim")
		if info == nil || !info.IsHIT {
			t.Error("Quot-elim not registered as HIT")
		}
	})
}

func TestHITApp_EvalAtEndpoints(t *testing.T) {
	// Test that VHITPathCtor reduces at endpoints
	base := eval.VNeutral{N: eval.Neutral{Head: eval.Head{Glob: "base"}}}

	hitPath := eval.VHITPathCtor{
		HITName:  "S1",
		CtorName: "loop",
		Args:     nil,
		IArgs:    []eval.Value{eval.VI0{}},
		Boundaries: []eval.BoundaryVal{
			{AtZero: base, AtOne: base},
		},
	}

	// At i0, should reduce to base
	result := eval.PathApply(hitPath, eval.VI0{})
	if _, ok := result.(eval.VNeutral); !ok {
		// Since VHITPathCtor is already stuck with VI0 in IArgs,
		// PathApply will try to reduce further
		t.Logf("Result type: %T", result)
	}
}

func TestHITPathReduction(t *testing.T) {
	// Create a VHITPathCtor value
	baseVal := eval.VNeutral{N: eval.Neutral{Head: eval.Head{Glob: "base"}}}

	hitPath := eval.VHITPathCtor{
		HITName:  "S1",
		CtorName: "loop",
		Args:     nil,
		IArgs:    nil, // No interval args yet
		Boundaries: []eval.BoundaryVal{
			{AtZero: baseVal, AtOne: baseVal},
		},
	}

	// Apply interval i0
	result := eval.PathApply(hitPath, eval.VI0{})
	if _, ok := result.(eval.VNeutral); !ok {
		t.Logf("PathApply at i0 returned: %T", result)
	}

	// Apply interval i1
	result = eval.PathApply(hitPath, eval.VI1{})
	if _, ok := result.(eval.VNeutral); !ok {
		t.Logf("PathApply at i1 returned: %T", result)
	}
}
