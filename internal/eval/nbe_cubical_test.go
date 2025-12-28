package eval

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// =============================================================================
// Phase 2: Interval Environment Tests
// =============================================================================

func TestEmptyIEnv(t *testing.T) {
	ienv := EmptyIEnv()
	if ienv == nil {
		t.Fatal("EmptyIEnv() returned nil")
	}
	if len(ienv.Bindings) != 0 {
		t.Errorf("EmptyIEnv() has %d bindings, want 0", len(ienv.Bindings))
	}
}

func TestIEnv_ILen(t *testing.T) {
	tests := []struct {
		name string
		ienv *IEnv
		want int
	}{
		{"nil", nil, 0},
		{"empty", EmptyIEnv(), 0},
		{"one binding", EmptyIEnv().Extend(VI0{}), 1},
		{"two bindings", EmptyIEnv().Extend(VI0{}).Extend(VI1{}), 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ienv.ILen()
			if got != tt.want {
				t.Errorf("ILen() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIEnv_Extend(t *testing.T) {
	// Extend nil environment
	t.Run("nil receiver", func(t *testing.T) {
		var ienv *IEnv = nil
		extended := ienv.Extend(VI0{})
		if extended == nil {
			t.Fatal("Extend on nil returned nil")
		}
		if extended.ILen() != 1 {
			t.Errorf("ILen() = %d, want 1", extended.ILen())
		}
	})

	// Extend empty environment
	t.Run("empty", func(t *testing.T) {
		ienv := EmptyIEnv()
		extended := ienv.Extend(VI1{})
		if extended.ILen() != 1 {
			t.Errorf("ILen() = %d, want 1", extended.ILen())
		}
		// Original should be unchanged
		if ienv.ILen() != 0 {
			t.Errorf("original ILen() = %d, want 0", ienv.ILen())
		}
	})

	// Extend with multiple values
	t.Run("multiple", func(t *testing.T) {
		ienv := EmptyIEnv().Extend(VI0{}).Extend(VI1{}).Extend(VIVar{Level: 5})
		if ienv.ILen() != 3 {
			t.Errorf("ILen() = %d, want 3", ienv.ILen())
		}
	})
}

func TestIEnv_Lookup(t *testing.T) {
	// Build environment: [VIVar{5}, VI1, VI0] (most recent first)
	ienv := EmptyIEnv().Extend(VI0{}).Extend(VI1{}).Extend(VIVar{Level: 5})

	tests := []struct {
		name    string
		ix      int
		wantTyp string
	}{
		{"index 0 (most recent)", 0, "VIVar"},
		{"index 1", 1, "VI1"},
		{"index 2 (oldest)", 2, "VI0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ienv.Lookup(tt.ix)
			switch tt.wantTyp {
			case "VI0":
				if _, ok := got.(VI0); !ok {
					t.Errorf("Lookup(%d) = %T, want VI0", tt.ix, got)
				}
			case "VI1":
				if _, ok := got.(VI1); !ok {
					t.Errorf("Lookup(%d) = %T, want VI1", tt.ix, got)
				}
			case "VIVar":
				if _, ok := got.(VIVar); !ok {
					t.Errorf("Lookup(%d) = %T, want VIVar", tt.ix, got)
				}
			}
		})
	}
}

func TestIEnv_Lookup_OutOfBounds(t *testing.T) {
	ienv := EmptyIEnv().Extend(VI0{})

	tests := []struct {
		name string
		ix   int
	}{
		{"negative index", -1},
		{"out of bounds", 5},
		{"just past end", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ienv.Lookup(tt.ix)
			// Out of bounds should return VIVar with the index as level
			v, ok := got.(VIVar)
			if !ok {
				t.Errorf("Lookup(%d) = %T, want VIVar", tt.ix, got)
				return
			}
			if v.Level != tt.ix {
				t.Errorf("VIVar.Level = %d, want %d", v.Level, tt.ix)
			}
		})
	}
}

func TestIEnv_Lookup_NilReceiver(t *testing.T) {
	var ienv *IEnv = nil
	got := ienv.Lookup(0)
	v, ok := got.(VIVar)
	if !ok {
		t.Errorf("Lookup on nil = %T, want VIVar", got)
		return
	}
	if v.Level != 0 {
		t.Errorf("VIVar.Level = %d, want 0", v.Level)
	}
}

// =============================================================================
// Phase 3: Face Formula Tests
// =============================================================================

func TestIsFaceTrue(t *testing.T) {
	tests := []struct {
		name string
		face FaceValue
		want bool
	}{
		{"VFaceTop", VFaceTop{}, true},
		{"VFaceBot", VFaceBot{}, false},
		{"VFaceEq", VFaceEq{ILevel: 0, IsOne: false}, false},
		{"VFaceAnd", VFaceAnd{Left: VFaceTop{}, Right: VFaceTop{}}, false},
		{"VFaceOr", VFaceOr{Left: VFaceBot{}, Right: VFaceTop{}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFaceTrue(tt.face)
			if got != tt.want {
				t.Errorf("IsFaceTrue(%T) = %v, want %v", tt.face, got, tt.want)
			}
		})
	}
}

func TestIsFaceFalse(t *testing.T) {
	tests := []struct {
		name string
		face FaceValue
		want bool
	}{
		{"VFaceBot", VFaceBot{}, true},
		{"VFaceTop", VFaceTop{}, false},
		{"VFaceEq", VFaceEq{ILevel: 0, IsOne: true}, false},
		{"VFaceAnd", VFaceAnd{Left: VFaceBot{}, Right: VFaceBot{}}, false},
		{"VFaceOr", VFaceOr{Left: VFaceBot{}, Right: VFaceBot{}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsFaceFalse(tt.face)
			if got != tt.want {
				t.Errorf("IsFaceFalse(%T) = %v, want %v", tt.face, got, tt.want)
			}
		})
	}
}

func TestSimplifyFaceAnd(t *testing.T) {
	eq0 := VFaceEq{ILevel: 0, IsOne: false} // (i=0)
	eq1 := VFaceEq{ILevel: 0, IsOne: true}  // (i=1)
	eqJ := VFaceEq{ILevel: 1, IsOne: false} // (j=0)

	tests := []struct {
		name    string
		left    FaceValue
		right   FaceValue
		wantTyp string
	}{
		{"Bot AND X = Bot", VFaceBot{}, eq0, "VFaceBot"},
		{"X AND Bot = Bot", eq0, VFaceBot{}, "VFaceBot"},
		{"Top AND X = X", VFaceTop{}, eq0, "VFaceEq"},
		{"X AND Top = X", eq0, VFaceTop{}, "VFaceEq"},
		{"(i=0) AND (i=1) = Bot", eq0, eq1, "VFaceBot"},
		{"(i=0) AND (j=0) = And", eq0, eqJ, "VFaceAnd"},
		{"Top AND Top = Top", VFaceTop{}, VFaceTop{}, "VFaceTop"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := simplifyFaceAnd(tt.left, tt.right)
			switch tt.wantTyp {
			case "VFaceBot":
				if _, ok := got.(VFaceBot); !ok {
					t.Errorf("got %T, want VFaceBot", got)
				}
			case "VFaceTop":
				if _, ok := got.(VFaceTop); !ok {
					t.Errorf("got %T, want VFaceTop", got)
				}
			case "VFaceEq":
				if _, ok := got.(VFaceEq); !ok {
					t.Errorf("got %T, want VFaceEq", got)
				}
			case "VFaceAnd":
				if _, ok := got.(VFaceAnd); !ok {
					t.Errorf("got %T, want VFaceAnd", got)
				}
			}
		})
	}
}

func TestSimplifyFaceOr(t *testing.T) {
	eq0 := VFaceEq{ILevel: 0, IsOne: false} // (i=0)
	eq1 := VFaceEq{ILevel: 0, IsOne: true}  // (i=1)
	eqJ := VFaceEq{ILevel: 1, IsOne: false} // (j=0)

	tests := []struct {
		name    string
		left    FaceValue
		right   FaceValue
		wantTyp string
	}{
		{"Top OR X = Top", VFaceTop{}, eq0, "VFaceTop"},
		{"X OR Top = Top", eq0, VFaceTop{}, "VFaceTop"},
		{"Bot OR X = X", VFaceBot{}, eq0, "VFaceEq"},
		{"X OR Bot = X", eq0, VFaceBot{}, "VFaceEq"},
		{"(i=0) OR (i=1) = Top", eq0, eq1, "VFaceTop"},
		{"(i=0) OR (j=0) = Or", eq0, eqJ, "VFaceOr"},
		{"Bot OR Bot = Bot", VFaceBot{}, VFaceBot{}, "VFaceBot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := simplifyFaceOr(tt.left, tt.right)
			switch tt.wantTyp {
			case "VFaceBot":
				if _, ok := got.(VFaceBot); !ok {
					t.Errorf("got %T, want VFaceBot", got)
				}
			case "VFaceTop":
				if _, ok := got.(VFaceTop); !ok {
					t.Errorf("got %T, want VFaceTop", got)
				}
			case "VFaceEq":
				if _, ok := got.(VFaceEq); !ok {
					t.Errorf("got %T, want VFaceEq", got)
				}
			case "VFaceOr":
				if _, ok := got.(VFaceOr); !ok {
					t.Errorf("got %T, want VFaceOr", got)
				}
			}
		})
	}
}

func TestEvalFaceEq(t *testing.T) {
	tests := []struct {
		name    string
		iVal    Value
		ivar    int
		isOne   bool
		ienvLen int
		wantTyp string
	}{
		// VI0 cases
		{"VI0 with isOne=false -> Top", VI0{}, 0, false, 1, "VFaceTop"},
		{"VI0 with isOne=true -> Bot", VI0{}, 0, true, 1, "VFaceBot"},
		// VI1 cases
		{"VI1 with isOne=true -> Top", VI1{}, 0, true, 1, "VFaceTop"},
		{"VI1 with isOne=false -> Bot", VI1{}, 0, false, 1, "VFaceBot"},
		// VIVar case
		{"VIVar -> VFaceEq", VIVar{Level: 2}, 0, true, 3, "VFaceEq"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evalFaceEq(tt.iVal, tt.ivar, tt.isOne, tt.ienvLen)
			switch tt.wantTyp {
			case "VFaceTop":
				if _, ok := got.(VFaceTop); !ok {
					t.Errorf("got %T, want VFaceTop", got)
				}
			case "VFaceBot":
				if _, ok := got.(VFaceBot); !ok {
					t.Errorf("got %T, want VFaceBot", got)
				}
			case "VFaceEq":
				if _, ok := got.(VFaceEq); !ok {
					t.Errorf("got %T, want VFaceEq", got)
				}
			}
		})
	}
}

func TestEvalFace(t *testing.T) {
	env := &Env{}
	ienv := EmptyIEnv().Extend(VI0{}) // index 0 -> VI0

	tests := []struct {
		name    string
		face    ast.Face
		wantTyp string
	}{
		{"nil -> Bot", nil, "VFaceBot"},
		{"FaceTop", ast.FaceTop{}, "VFaceTop"},
		{"FaceBot", ast.FaceBot{}, "VFaceBot"},
		{"FaceEq (i=0) with i=VI0 -> Top", ast.FaceEq{IVar: 0, IsOne: false}, "VFaceTop"},
		{"FaceEq (i=1) with i=VI0 -> Bot", ast.FaceEq{IVar: 0, IsOne: true}, "VFaceBot"},
		{"FaceAnd Top Top -> Top", ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceTop{}}, "VFaceTop"},
		{"FaceAnd Top Bot -> Bot", ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}}, "VFaceBot"},
		{"FaceOr Bot Top -> Top", ast.FaceOr{Left: ast.FaceBot{}, Right: ast.FaceTop{}}, "VFaceTop"},
		{"FaceOr Bot Bot -> Bot", ast.FaceOr{Left: ast.FaceBot{}, Right: ast.FaceBot{}}, "VFaceBot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evalFace(env, ienv, tt.face)
			switch tt.wantTyp {
			case "VFaceTop":
				if _, ok := got.(VFaceTop); !ok {
					t.Errorf("got %T, want VFaceTop", got)
				}
			case "VFaceBot":
				if _, ok := got.(VFaceBot); !ok {
					t.Errorf("got %T, want VFaceBot", got)
				}
			}
		})
	}
}

// =============================================================================
// Phase 4: Path Operation Tests
// =============================================================================

func TestPathApply_PathLam(t *testing.T) {
	// Create a path lambda: <i> x where x is a global
	// When applied to i0 or i1, it should return x
	env := &Env{}
	ienv := EmptyIEnv()
	body := ast.Global{Name: "x"}
	pathLamVal := VPathLam{Body: &IClosure{Env: env, IEnv: ienv, Term: body}}

	tests := []struct {
		name    string
		r       Value
		wantTyp string
	}{
		{"PathLam @ i0", VI0{}, "VNeutral"},
		{"PathLam @ i1", VI1{}, "VNeutral"},
		{"PathLam @ ivar", VIVar{Level: 0}, "VNeutral"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PathApply(pathLamVal, tt.r)
			// All should produce VNeutral with global "x"
			if n, ok := got.(VNeutral); ok {
				if n.N.Head.Glob != "x" {
					t.Errorf("got global %q, want x", n.N.Head.Glob)
				}
			} else {
				t.Errorf("got %T, want VNeutral", got)
			}
		})
	}
}

func TestPathApply_PathP_Endpoints(t *testing.T) {
	xVal := VNeutral{N: Neutral{Head: Head{Glob: "x"}, Sp: nil}}
	yVal := VNeutral{N: Neutral{Head: Head{Glob: "y"}, Sp: nil}}
	aVal := VSort{Level: 0}

	pathP := VPathP{
		A: &IClosure{Env: nil, IEnv: nil, Term: ast.Sort{U: 0}},
		X: xVal,
		Y: yVal,
	}

	t.Run("PathP @ i0 -> X", func(t *testing.T) {
		got := PathApply(pathP, VI0{})
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "x" {
				t.Errorf("got %q, want x", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want VNeutral with x", got)
		}
	})

	t.Run("PathP @ i1 -> Y", func(t *testing.T) {
		got := PathApply(pathP, VI1{})
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "y" {
				t.Errorf("got %q, want y", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want VNeutral with y", got)
		}
	})

	t.Run("PathP @ ivar -> stuck", func(t *testing.T) {
		got := PathApply(pathP, VIVar{Level: 0})
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "@" {
				t.Errorf("got head %q, want @", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want stuck VNeutral", got)
		}
	})

	_ = aVal // silence unused warning
}

func TestPathApply_Path_Endpoints(t *testing.T) {
	xVal := VNeutral{N: Neutral{Head: Head{Glob: "x"}, Sp: nil}}
	yVal := VNeutral{N: Neutral{Head: Head{Glob: "y"}, Sp: nil}}

	pathVal := VPath{
		A: VSort{Level: 0},
		X: xVal,
		Y: yVal,
	}

	t.Run("Path @ i0 -> X", func(t *testing.T) {
		got := PathApply(pathVal, VI0{})
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "x" {
				t.Errorf("got %q, want x", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want VNeutral with x", got)
		}
	})

	t.Run("Path @ i1 -> Y", func(t *testing.T) {
		got := PathApply(pathVal, VI1{})
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "y" {
				t.Errorf("got %q, want y", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want VNeutral with y", got)
		}
	})

	t.Run("Path @ ivar -> stuck", func(t *testing.T) {
		got := PathApply(pathVal, VIVar{Level: 0})
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "@" {
				t.Errorf("got head %q, want @", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want stuck VNeutral", got)
		}
	})
}

func TestPathApply_Neutral(t *testing.T) {
	// Stuck path application: neutral @ r -> stuck
	neutralPath := VNeutral{N: Neutral{Head: Head{Glob: "p"}, Sp: nil}}
	got := PathApply(neutralPath, VI0{})

	if n, ok := got.(VNeutral); ok {
		if n.N.Head.Glob != "@" {
			t.Errorf("got head %q, want @", n.N.Head.Glob)
		}
		if len(n.N.Sp) != 2 {
			t.Errorf("got spine length %d, want 2", len(n.N.Sp))
		}
	} else {
		t.Errorf("got %T, want VNeutral", got)
	}
}

func TestPathApply_NonPath(t *testing.T) {
	// Non-path value @ r -> stuck
	nonPath := VSort{Level: 0}
	got := PathApply(nonPath, VI0{})

	if n, ok := got.(VNeutral); ok {
		if n.N.Head.Glob != "@" {
			t.Errorf("got head %q, want @", n.N.Head.Glob)
		}
	} else {
		t.Errorf("got %T, want VNeutral", got)
	}
}

func TestEvalTransport_Constant(t *testing.T) {
	// transport (λi. A) e -> e when A is constant
	// Create a closure where the body doesn't use the interval variable
	env := &Env{}
	ienv := EmptyIEnv()
	body := ast.Sort{U: 0} // Type₀, constant in i

	aClosure := &IClosure{Env: env, IEnv: ienv, Term: body}
	eVal := VNeutral{N: Neutral{Head: Head{Glob: "x"}, Sp: nil}}

	got := EvalTransport(aClosure, eVal)

	// Should return e unchanged (identity transport)
	if n, ok := got.(VNeutral); ok {
		if n.N.Head.Glob != "x" {
			t.Errorf("got %q, want x", n.N.Head.Glob)
		}
	} else {
		t.Errorf("got %T, want VNeutral (identity transport)", got)
	}
}

func TestEvalTransport_NonConstant(t *testing.T) {
	// transport (λi. ...) e -> stuck VTransport when body uses interval
	env := &Env{}
	ienv := EmptyIEnv()
	// Body that uses the interval variable: IVar at index 0
	body := ast.IVar{Ix: 0}

	aClosure := &IClosure{Env: env, IEnv: ienv, Term: body}
	eVal := VNeutral{N: Neutral{Head: Head{Glob: "x"}, Sp: nil}}

	got := EvalTransport(aClosure, eVal)

	// Should return stuck VTransport
	if _, ok := got.(VTransport); !ok {
		t.Errorf("got %T, want VTransport (stuck)", got)
	}
}

func TestIsConstantFamily(t *testing.T) {
	tests := []struct {
		name string
		body ast.Term
		want bool
	}{
		{"constant Sort", ast.Sort{U: 0}, true},
		{"constant Global", ast.Global{Name: "A"}, true},
		{"uses interval IVar{0}", ast.IVar{Ix: 0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			closure := &IClosure{Env: &Env{}, IEnv: EmptyIEnv(), Term: tt.body}
			got := isConstantFamily(closure)
			if got != tt.want {
				t.Errorf("isConstantFamily = %v, want %v", got, tt.want)
			}
		})
	}
}

// =============================================================================
// Phase 5: EvalCubical Main Tests
// =============================================================================

func TestEvalCubical_Interval(t *testing.T) {
	env := &Env{}
	ienv := EmptyIEnv()

	t.Run("Interval type", func(t *testing.T) {
		got := EvalCubical(env, ienv, ast.Interval{})
		if g, ok := got.(VGlobal); ok {
			if g.Name != "I" {
				t.Errorf("got %q, want I", g.Name)
			}
		} else {
			t.Errorf("got %T, want VGlobal", got)
		}
	})

	t.Run("I0", func(t *testing.T) {
		got := EvalCubical(env, ienv, ast.I0{})
		if _, ok := got.(VI0); !ok {
			t.Errorf("got %T, want VI0", got)
		}
	})

	t.Run("I1", func(t *testing.T) {
		got := EvalCubical(env, ienv, ast.I1{})
		if _, ok := got.(VI1); !ok {
			t.Errorf("got %T, want VI1", got)
		}
	})

	t.Run("IVar lookup", func(t *testing.T) {
		ienvWithVar := EmptyIEnv().Extend(VI0{})
		got := EvalCubical(env, ienvWithVar, ast.IVar{Ix: 0})
		if _, ok := got.(VI0); !ok {
			t.Errorf("got %T, want VI0", got)
		}
	})
}

func TestEvalCubical_Path(t *testing.T) {
	env := &Env{}
	ienv := EmptyIEnv()

	t.Run("Path type", func(t *testing.T) {
		term := ast.Path{
			A: ast.Sort{U: 0},
			X: ast.Global{Name: "x"},
			Y: ast.Global{Name: "y"},
		}
		got := EvalCubical(env, ienv, term)
		if p, ok := got.(VPath); ok {
			if _, ok := p.A.(VSort); !ok {
				t.Errorf("A is %T, want VSort", p.A)
			}
		} else {
			t.Errorf("got %T, want VPath", got)
		}
	})

	t.Run("PathP type", func(t *testing.T) {
		term := ast.PathP{
			A: ast.Sort{U: 0},
			X: ast.Global{Name: "x"},
			Y: ast.Global{Name: "y"},
		}
		got := EvalCubical(env, ienv, term)
		if _, ok := got.(VPathP); !ok {
			t.Errorf("got %T, want VPathP", got)
		}
	})

	t.Run("PathLam", func(t *testing.T) {
		term := ast.PathLam{Binder: "i", Body: ast.Global{Name: "x"}}
		got := EvalCubical(env, ienv, term)
		if _, ok := got.(VPathLam); !ok {
			t.Errorf("got %T, want VPathLam", got)
		}
	})

	t.Run("PathApp", func(t *testing.T) {
		// (<i> x) @ i0 -> x
		term := ast.PathApp{
			P: ast.PathLam{Binder: "i", Body: ast.Global{Name: "x"}},
			R: ast.I0{},
		}
		got := EvalCubical(env, ienv, term)
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "x" {
				t.Errorf("got global %q, want x", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want VNeutral with x", got)
		}
	})
}

func TestEvalCubical_Transport(t *testing.T) {
	env := &Env{}
	ienv := EmptyIEnv()

	t.Run("constant family", func(t *testing.T) {
		// transport (λi. Type₀) x -> x
		term := ast.Transport{
			A: ast.Sort{U: 0},
			E: ast.Global{Name: "x"},
		}
		got := EvalCubical(env, ienv, term)
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "x" {
				t.Errorf("got global %q, want x", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want VNeutral (identity transport)", got)
		}
	})
}

func TestEvalCubical_FaceFormulas(t *testing.T) {
	env := &Env{}
	ienv := EmptyIEnv().Extend(VI0{}) // i -> VI0

	t.Run("FaceTop", func(t *testing.T) {
		got := EvalCubical(env, ienv, ast.FaceTop{})
		if _, ok := got.(VFaceTop); !ok {
			t.Errorf("got %T, want VFaceTop", got)
		}
	})

	t.Run("FaceBot", func(t *testing.T) {
		got := EvalCubical(env, ienv, ast.FaceBot{})
		if _, ok := got.(VFaceBot); !ok {
			t.Errorf("got %T, want VFaceBot", got)
		}
	})

	t.Run("FaceEq resolved", func(t *testing.T) {
		// (i=0) where i=VI0 -> Top
		got := EvalCubical(env, ienv, ast.FaceEq{IVar: 0, IsOne: false})
		if _, ok := got.(VFaceTop); !ok {
			t.Errorf("got %T, want VFaceTop", got)
		}
	})

	t.Run("FaceAnd", func(t *testing.T) {
		term := ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceTop{}}
		got := EvalCubical(env, ienv, term)
		if _, ok := got.(VFaceTop); !ok {
			t.Errorf("got %T, want VFaceTop", got)
		}
	})

	t.Run("FaceOr", func(t *testing.T) {
		term := ast.FaceOr{Left: ast.FaceBot{}, Right: ast.FaceTop{}}
		got := EvalCubical(env, ienv, term)
		if _, ok := got.(VFaceTop); !ok {
			t.Errorf("got %T, want VFaceTop", got)
		}
	})
}

func TestEvalCubical_Partial(t *testing.T) {
	env := &Env{}
	ienv := EmptyIEnv()

	term := ast.Partial{
		Phi: ast.FaceTop{},
		A:   ast.Sort{U: 0},
	}
	got := EvalCubical(env, ienv, term)
	if p, ok := got.(VPartial); ok {
		if _, ok := p.Phi.(VFaceTop); !ok {
			t.Errorf("Phi is %T, want VFaceTop", p.Phi)
		}
		if _, ok := p.A.(VSort); !ok {
			t.Errorf("A is %T, want VSort", p.A)
		}
	} else {
		t.Errorf("got %T, want VPartial", got)
	}
}

func TestEvalCubical_System(t *testing.T) {
	env := &Env{}
	ienv := EmptyIEnv()

	term := ast.System{
		Branches: []ast.SystemBranch{
			{Phi: ast.FaceTop{}, Term: ast.Global{Name: "x"}},
			{Phi: ast.FaceBot{}, Term: ast.Global{Name: "y"}},
		},
	}
	got := EvalCubical(env, ienv, term)
	if s, ok := got.(VSystem); ok {
		if len(s.Branches) != 2 {
			t.Errorf("got %d branches, want 2", len(s.Branches))
		}
	} else {
		t.Errorf("got %T, want VSystem", got)
	}
}

func TestEvalCubical_StandardTerms(t *testing.T) {
	env := &Env{}
	ienv := EmptyIEnv()

	t.Run("Var lookup", func(t *testing.T) {
		envWithVar := env.Extend(VSort{Level: 42})
		got := EvalCubical(envWithVar, ienv, ast.Var{Ix: 0})
		if s, ok := got.(VSort); ok {
			if s.Level != 42 {
				t.Errorf("got level %d, want 42", s.Level)
			}
		} else {
			t.Errorf("got %T, want VSort", got)
		}
	})

	t.Run("Global", func(t *testing.T) {
		got := EvalCubical(env, ienv, ast.Global{Name: "x"})
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "x" {
				t.Errorf("got %q, want x", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want VNeutral", got)
		}
	})

	t.Run("Sort", func(t *testing.T) {
		got := EvalCubical(env, ienv, ast.Sort{U: 5})
		if s, ok := got.(VSort); ok {
			if s.Level != 5 {
				t.Errorf("got level %d, want 5", s.Level)
			}
		} else {
			t.Errorf("got %T, want VSort", got)
		}
	})

	t.Run("Lam", func(t *testing.T) {
		term := ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}}
		got := EvalCubical(env, ienv, term)
		if _, ok := got.(VLam); !ok {
			t.Errorf("got %T, want VLam", got)
		}
	})

	t.Run("App beta", func(t *testing.T) {
		// (λx.x) y -> y
		term := ast.App{
			T: ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}},
			U: ast.Global{Name: "y"},
		}
		got := EvalCubical(env, ienv, term)
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "y" {
				t.Errorf("got %q, want y", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want VNeutral", got)
		}
	})

	t.Run("Pair", func(t *testing.T) {
		term := ast.Pair{Fst: ast.Global{Name: "a"}, Snd: ast.Global{Name: "b"}}
		got := EvalCubical(env, ienv, term)
		if _, ok := got.(VPair); !ok {
			t.Errorf("got %T, want VPair", got)
		}
	})

	t.Run("Fst", func(t *testing.T) {
		term := ast.Fst{P: ast.Pair{Fst: ast.Global{Name: "a"}, Snd: ast.Global{Name: "b"}}}
		got := EvalCubical(env, ienv, term)
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "a" {
				t.Errorf("got %q, want a", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want VNeutral with a", got)
		}
	})

	t.Run("Snd", func(t *testing.T) {
		term := ast.Snd{P: ast.Pair{Fst: ast.Global{Name: "a"}, Snd: ast.Global{Name: "b"}}}
		got := EvalCubical(env, ienv, term)
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "b" {
				t.Errorf("got %q, want b", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want VNeutral with b", got)
		}
	})

	t.Run("Pi", func(t *testing.T) {
		term := ast.Pi{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}}
		got := EvalCubical(env, ienv, term)
		if _, ok := got.(VPi); !ok {
			t.Errorf("got %T, want VPi", got)
		}
	})

	t.Run("Sigma", func(t *testing.T) {
		term := ast.Sigma{Binder: "x", A: ast.Sort{U: 0}, B: ast.Var{Ix: 0}}
		got := EvalCubical(env, ienv, term)
		if _, ok := got.(VSigma); !ok {
			t.Errorf("got %T, want VSigma", got)
		}
	})

	t.Run("Let", func(t *testing.T) {
		// let x = Type₀ in x
		term := ast.Let{Binder: "x", Val: ast.Sort{U: 0}, Body: ast.Var{Ix: 0}}
		got := EvalCubical(env, ienv, term)
		if s, ok := got.(VSort); ok {
			if s.Level != 0 {
				t.Errorf("got level %d, want 0", s.Level)
			}
		} else {
			t.Errorf("got %T, want VSort", got)
		}
	})

	t.Run("Id", func(t *testing.T) {
		term := ast.Id{A: ast.Sort{U: 0}, X: ast.Global{Name: "x"}, Y: ast.Global{Name: "y"}}
		got := EvalCubical(env, ienv, term)
		if _, ok := got.(VId); !ok {
			t.Errorf("got %T, want VId", got)
		}
	})

	t.Run("Refl", func(t *testing.T) {
		term := ast.Refl{A: ast.Sort{U: 0}, X: ast.Global{Name: "x"}}
		got := EvalCubical(env, ienv, term)
		if _, ok := got.(VRefl); !ok {
			t.Errorf("got %T, want VRefl", got)
		}
	})
}

func TestEvalCubical_NilHandling(t *testing.T) {
	t.Run("nil env", func(t *testing.T) {
		got := EvalCubical(nil, nil, ast.Sort{U: 0})
		if _, ok := got.(VSort); !ok {
			t.Errorf("got %T, want VSort", got)
		}
	})

	t.Run("nil term", func(t *testing.T) {
		got := EvalCubical(nil, nil, nil)
		// Should return error value (VGlobal with "error:" prefix)
		if g, ok := got.(VGlobal); ok {
			if g.Name != "error:nil term (cubical)" {
				t.Errorf("got %q, want error:nil term (cubical)", g.Name)
			}
		} else {
			t.Errorf("got %T, want VGlobal error", got)
		}
	})
}

// =============================================================================
// Phase 6: Composition Tests
// =============================================================================

func TestEvalComp_FaceTrue(t *testing.T) {
	env := &Env{}
	ienv := EmptyIEnv()
	aClosure := &IClosure{Env: env, IEnv: ienv, Term: ast.Sort{U: 0}}
	tubeClosure := &IClosure{Env: env, IEnv: ienv, Term: ast.Global{Name: "tube"}}
	base := VNeutral{N: Neutral{Head: Head{Glob: "base"}, Sp: nil}}

	got := EvalComp(aClosure, VFaceTop{}, tubeClosure, base)
	if n, ok := got.(VNeutral); ok {
		if n.N.Head.Glob != "tube" {
			t.Errorf("got %q, want tube", n.N.Head.Glob)
		}
	} else {
		t.Errorf("got %T, want VNeutral", got)
	}
}

func TestEvalComp_FaceFalse(t *testing.T) {
	env := &Env{}
	ienv := EmptyIEnv()
	aClosure := &IClosure{Env: env, IEnv: ienv, Term: ast.Sort{U: 0}}
	tubeClosure := &IClosure{Env: env, IEnv: ienv, Term: ast.Global{Name: "tube"}}
	base := VNeutral{N: Neutral{Head: Head{Glob: "base"}, Sp: nil}}

	got := EvalComp(aClosure, VFaceBot{}, tubeClosure, base)
	if n, ok := got.(VNeutral); ok {
		if n.N.Head.Glob != "base" {
			t.Errorf("got %q, want base", n.N.Head.Glob)
		}
	} else {
		t.Errorf("got %T, want VNeutral", got)
	}
}

func TestEvalComp_Stuck(t *testing.T) {
	env := &Env{}
	ienv := EmptyIEnv()
	aClosure := &IClosure{Env: env, IEnv: ienv, Term: ast.Sort{U: 0}}
	tubeClosure := &IClosure{Env: env, IEnv: ienv, Term: ast.Global{Name: "tube"}}
	base := VNeutral{N: Neutral{Head: Head{Glob: "base"}, Sp: nil}}

	got := EvalComp(aClosure, VFaceEq{ILevel: 0, IsOne: false}, tubeClosure, base)
	if _, ok := got.(VComp); !ok {
		t.Errorf("got %T, want VComp", got)
	}
}

func TestEvalHComp(t *testing.T) {
	env := &Env{}
	ienv := EmptyIEnv()
	a := VSort{Level: 0}
	tubeClosure := &IClosure{Env: env, IEnv: ienv, Term: ast.Global{Name: "tube"}}
	base := VNeutral{N: Neutral{Head: Head{Glob: "base"}, Sp: nil}}

	t.Run("face true", func(t *testing.T) {
		got := EvalHComp(a, VFaceTop{}, tubeClosure, base)
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "tube" {
				t.Errorf("got %q, want tube", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want VNeutral", got)
		}
	})

	t.Run("face false", func(t *testing.T) {
		got := EvalHComp(a, VFaceBot{}, tubeClosure, base)
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "base" {
				t.Errorf("got %q, want base", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want VNeutral", got)
		}
	})

	t.Run("stuck", func(t *testing.T) {
		got := EvalHComp(a, VFaceEq{ILevel: 0, IsOne: false}, tubeClosure, base)
		if _, ok := got.(VHComp); !ok {
			t.Errorf("got %T, want VHComp", got)
		}
	})
}

func TestEvalFill(t *testing.T) {
	env := &Env{}
	ienv := EmptyIEnv()
	aClosure := &IClosure{Env: env, IEnv: ienv, Term: ast.Sort{U: 0}}
	tubeClosure := &IClosure{Env: env, IEnv: ienv, Term: ast.Global{Name: "tube"}}
	base := VNeutral{N: Neutral{Head: Head{Glob: "base"}, Sp: nil}}

	got := EvalFill(aClosure, VFaceTop{}, tubeClosure, base)
	if _, ok := got.(VFill); !ok {
		t.Errorf("got %T, want VFill", got)
	}
}

// =============================================================================
// Phase 7: Glue & Univalence Tests
// =============================================================================

func TestEvalGlue(t *testing.T) {
	a := VSort{Level: 0}
	tVal := VNeutral{N: Neutral{Head: Head{Glob: "T"}, Sp: nil}}
	equiv := VNeutral{N: Neutral{Head: Head{Glob: "e"}, Sp: nil}}

	t.Run("top branch", func(t *testing.T) {
		branches := []VGlueBranch{{Phi: VFaceTop{}, T: tVal, Equiv: equiv}}
		got := EvalGlue(a, branches)
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "T" {
				t.Errorf("got %q, want T", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want VNeutral", got)
		}
	})

	t.Run("no branches", func(t *testing.T) {
		got := EvalGlue(VSort{Level: 42}, []VGlueBranch{})
		if s, ok := got.(VSort); ok {
			if s.Level != 42 {
				t.Errorf("got level %d, want 42", s.Level)
			}
		} else {
			t.Errorf("got %T, want VSort", got)
		}
	})

	t.Run("filters bot", func(t *testing.T) {
		branches := []VGlueBranch{{Phi: VFaceBot{}, T: tVal, Equiv: equiv}}
		got := EvalGlue(a, branches)
		if _, ok := got.(VSort); !ok {
			t.Errorf("got %T, want VSort", got)
		}
	})

	t.Run("stuck", func(t *testing.T) {
		branches := []VGlueBranch{{Phi: VFaceEq{ILevel: 0, IsOne: false}, T: tVal, Equiv: equiv}}
		got := EvalGlue(a, branches)
		if _, ok := got.(VGlue); !ok {
			t.Errorf("got %T, want VGlue", got)
		}
	})
}

func TestEvalGlueElem(t *testing.T) {
	term := VNeutral{N: Neutral{Head: Head{Glob: "t"}, Sp: nil}}
	base := VNeutral{N: Neutral{Head: Head{Glob: "base"}, Sp: nil}}

	t.Run("top branch", func(t *testing.T) {
		branches := []VGlueElemBranch{{Phi: VFaceTop{}, Term: term}}
		got := EvalGlueElem(branches, base)
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "t" {
				t.Errorf("got %q, want t", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want VNeutral", got)
		}
	})

	t.Run("filters bot", func(t *testing.T) {
		branches := []VGlueElemBranch{{Phi: VFaceBot{}, Term: term}}
		got := EvalGlueElem(branches, base)
		if ge, ok := got.(VGlueElem); ok {
			if len(ge.System) != 0 {
				t.Errorf("got %d branches, want 0", len(ge.System))
			}
		} else {
			t.Errorf("got %T, want VGlueElem", got)
		}
	})
}

func TestEvalUnglue(t *testing.T) {
	ty := VSort{Level: 0}
	base := VNeutral{N: Neutral{Head: Head{Glob: "base"}, Sp: nil}}

	t.Run("reduces", func(t *testing.T) {
		glueElem := VGlueElem{System: nil, Base: base}
		got := EvalUnglue(ty, glueElem)
		if n, ok := got.(VNeutral); ok {
			if n.N.Head.Glob != "base" {
				t.Errorf("got %q, want base", n.N.Head.Glob)
			}
		} else {
			t.Errorf("got %T, want VNeutral", got)
		}
	})

	t.Run("stuck", func(t *testing.T) {
		got := EvalUnglue(ty, VNeutral{N: Neutral{Head: Head{Glob: "g"}, Sp: nil}})
		if _, ok := got.(VUnglue); !ok {
			t.Errorf("got %T, want VUnglue", got)
		}
	})
}

func TestEvalUA(t *testing.T) {
	a := VSort{Level: 0}
	b := VSort{Level: 1}
	equiv := VNeutral{N: Neutral{Head: Head{Glob: "equiv"}, Sp: nil}}

	got := EvalUA(a, b, equiv)
	if _, ok := got.(VUA); !ok {
		t.Errorf("got %T, want VUA", got)
	}
}

func TestEvalUABeta(t *testing.T) {
	equiv := VNeutral{N: Neutral{Head: Head{Glob: "equiv"}, Sp: nil}}
	arg := VNeutral{N: Neutral{Head: Head{Glob: "arg"}, Sp: nil}}

	got := EvalUABeta(equiv, arg)
	if _, ok := got.(VUABeta); !ok {
		t.Errorf("got %T, want VUABeta", got)
	}
}

func TestUAPathApply(t *testing.T) {
	a := VSort{Level: 0}
	b := VSort{Level: 1}
	equiv := VNeutral{N: Neutral{Head: Head{Glob: "equiv"}, Sp: nil}}
	ua := VUA{A: a, B: b, Equiv: equiv}

	t.Run("i0", func(t *testing.T) {
		got := UAPathApply(ua, VI0{})
		if s, ok := got.(VSort); ok {
			if s.Level != 0 {
				t.Errorf("got level %d, want 0", s.Level)
			}
		} else {
			t.Errorf("got %T, want VSort", got)
		}
	})

	t.Run("i1", func(t *testing.T) {
		got := UAPathApply(ua, VI1{})
		if s, ok := got.(VSort); ok {
			if s.Level != 1 {
				t.Errorf("got level %d, want 1", s.Level)
			}
		} else {
			t.Errorf("got %T, want VSort", got)
		}
	})

	t.Run("ivar", func(t *testing.T) {
		got := UAPathApply(ua, VIVar{Level: 5})
		if glue, ok := got.(VGlue); ok {
			if len(glue.System) != 1 {
				t.Errorf("got %d branches, want 1", len(glue.System))
			}
		} else {
			t.Errorf("got %T, want VGlue", got)
		}
	})
}

// =============================================================================
// Phase 8: Cubical Reification Tests
// =============================================================================

func TestReifyCubicalAt_Interval(t *testing.T) {
	t.Run("VI0", func(t *testing.T) {
		got := ReifyCubicalAt(0, 0, VI0{})
		if _, ok := got.(ast.I0); !ok {
			t.Errorf("got %T, want ast.I0", got)
		}
	})

	t.Run("VI1", func(t *testing.T) {
		got := ReifyCubicalAt(0, 0, VI1{})
		if _, ok := got.(ast.I1); !ok {
			t.Errorf("got %T, want ast.I1", got)
		}
	})

	t.Run("VIVar", func(t *testing.T) {
		got := ReifyCubicalAt(0, 3, VIVar{Level: 2})
		if iv, ok := got.(ast.IVar); ok {
			if iv.Ix != 0 {
				t.Errorf("got ix=%d, want 0", iv.Ix)
			}
		} else {
			t.Errorf("got %T, want ast.IVar", got)
		}
	})
}

func TestReifyCubicalAt_Path(t *testing.T) {
	t.Run("VPath", func(t *testing.T) {
		v := VPath{A: VSort{Level: 0}, X: VSort{Level: 0}, Y: VSort{Level: 0}}
		got := ReifyCubicalAt(0, 0, v)
		if _, ok := got.(ast.Path); !ok {
			t.Errorf("got %T, want ast.Path", got)
		}
	})

	t.Run("VPathP", func(t *testing.T) {
		v := VPathP{A: &IClosure{Term: ast.Sort{U: 0}}, X: VSort{Level: 0}, Y: VSort{Level: 0}}
		got := ReifyCubicalAt(0, 0, v)
		if _, ok := got.(ast.PathP); !ok {
			t.Errorf("got %T, want ast.PathP", got)
		}
	})

	t.Run("VPathLam", func(t *testing.T) {
		v := VPathLam{Body: &IClosure{Term: ast.Global{Name: "x"}}}
		got := ReifyCubicalAt(0, 0, v)
		if _, ok := got.(ast.PathLam); !ok {
			t.Errorf("got %T, want ast.PathLam", got)
		}
	})
}

func TestReifyCubicalAt_FaceFormulas(t *testing.T) {
	t.Run("VFaceTop", func(t *testing.T) {
		got := ReifyCubicalAt(0, 0, VFaceTop{})
		if _, ok := got.(ast.FaceTop); !ok {
			t.Errorf("got %T, want ast.FaceTop", got)
		}
	})

	t.Run("VFaceBot", func(t *testing.T) {
		got := ReifyCubicalAt(0, 0, VFaceBot{})
		if _, ok := got.(ast.FaceBot); !ok {
			t.Errorf("got %T, want ast.FaceBot", got)
		}
	})

	t.Run("VFaceAnd", func(t *testing.T) {
		got := ReifyCubicalAt(0, 0, VFaceAnd{Left: VFaceTop{}, Right: VFaceBot{}})
		if _, ok := got.(ast.FaceAnd); !ok {
			t.Errorf("got %T, want ast.FaceAnd", got)
		}
	})

	t.Run("VFaceOr", func(t *testing.T) {
		got := ReifyCubicalAt(0, 0, VFaceOr{Left: VFaceTop{}, Right: VFaceBot{}})
		if _, ok := got.(ast.FaceOr); !ok {
			t.Errorf("got %T, want ast.FaceOr", got)
		}
	})
}

func TestReifyCubicalAt_Comp(t *testing.T) {
	v := VComp{
		A:    &IClosure{Term: ast.Sort{U: 0}},
		Phi:  VFaceEq{ILevel: 0, IsOne: false},
		Tube: &IClosure{Term: ast.Global{Name: "tube"}},
		Base: VSort{Level: 0},
	}
	got := ReifyCubicalAt(0, 0, v)
	if _, ok := got.(ast.Comp); !ok {
		t.Errorf("got %T, want ast.Comp", got)
	}
}

func TestReifyCubicalAt_Glue(t *testing.T) {
	v := VGlue{A: VSort{Level: 0}, System: []VGlueBranch{
		{Phi: VFaceTop{}, T: VSort{Level: 1}, Equiv: VSort{Level: 0}},
	}}
	got := ReifyCubicalAt(0, 0, v)
	if _, ok := got.(ast.Glue); !ok {
		t.Errorf("got %T, want ast.Glue", got)
	}
}

func TestReifyCubicalAt_UA(t *testing.T) {
	v := VUA{A: VSort{Level: 0}, B: VSort{Level: 1}, Equiv: VSort{Level: 0}}
	got := ReifyCubicalAt(0, 0, v)
	if _, ok := got.(ast.UA); !ok {
		t.Errorf("got %T, want ast.UA", got)
	}
}

func TestReifyFaceAt(t *testing.T) {
	t.Run("VFaceTop", func(t *testing.T) {
		got := reifyFaceAt(0, 0, VFaceTop{})
		if _, ok := got.(ast.FaceTop); !ok {
			t.Errorf("got %T, want ast.FaceTop", got)
		}
	})

	t.Run("VFaceBot", func(t *testing.T) {
		got := reifyFaceAt(0, 0, VFaceBot{})
		if _, ok := got.(ast.FaceBot); !ok {
			t.Errorf("got %T, want ast.FaceBot", got)
		}
	})

	t.Run("VFaceEq", func(t *testing.T) {
		got := reifyFaceAt(0, 2, VFaceEq{ILevel: 1, IsOne: false})
		if fe, ok := got.(ast.FaceEq); ok {
			if fe.IVar != 0 {
				t.Errorf("got IVar=%d, want 0", fe.IVar)
			}
		} else {
			t.Errorf("got %T, want ast.FaceEq", got)
		}
	})
}

// ============================================================
// Phase 9: Extension Hook Tests
// ============================================================

// TestTryEvalCubical tests the extension hook that evaluates cubical AST terms.
// This hook is called from the main Eval function when it encounters cubical terms.
func TestTryEvalCubical(t *testing.T) {
	env := &Env{}

	// Simple type for testing
	typeU := ast.Sort{U: 0}
	termV := ast.Var{Ix: 0}

	tests := []struct {
		name    string
		term    ast.Term
		checkFn func(Value) bool
	}{
		{"Interval", ast.Interval{}, func(v Value) bool {
			g, ok := v.(VGlobal)
			return ok && g.Name == "I"
		}},
		{"I0", ast.I0{}, func(v Value) bool { _, ok := v.(VI0); return ok }},
		{"I1", ast.I1{}, func(v Value) bool { _, ok := v.(VI1); return ok }},
		{"IVar", ast.IVar{Ix: 0}, func(v Value) bool {
			iv, ok := v.(VIVar)
			return ok && iv.Level == 0
		}},
		{"Path", ast.Path{A: typeU, X: termV, Y: termV}, func(v Value) bool {
			_, ok := v.(VPath)
			return ok
		}},
		{"PathP", ast.PathP{A: typeU, X: termV, Y: termV}, func(v Value) bool {
			_, ok := v.(VPathP)
			return ok
		}},
		{"PathLam", ast.PathLam{Binder: "i", Body: typeU}, func(v Value) bool {
			_, ok := v.(VPathLam)
			return ok
		}},
		{"FaceTop", ast.FaceTop{}, func(v Value) bool { _, ok := v.(VFaceTop); return ok }},
		{"FaceBot", ast.FaceBot{}, func(v Value) bool { _, ok := v.(VFaceBot); return ok }},
		{"FaceEq", ast.FaceEq{IVar: 0, IsOne: true}, func(v Value) bool {
			fe, ok := v.(VFaceEq)
			return ok && fe.IsOne
		}},
		{"FaceAnd_TopTop", ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceTop{}}, func(v Value) bool {
			_, ok := v.(VFaceTop)
			return ok
		}},
		{"FaceOr_BotBot", ast.FaceOr{Left: ast.FaceBot{}, Right: ast.FaceBot{}}, func(v Value) bool {
			_, ok := v.(VFaceBot)
			return ok
		}},
		{"Partial", ast.Partial{Phi: ast.FaceTop{}, A: typeU}, func(v Value) bool {
			_, ok := v.(VPartial)
			return ok
		}},
		{"System", ast.System{Branches: []ast.SystemBranch{
			{Phi: ast.FaceTop{}, Term: typeU},
		}}, func(v Value) bool {
			s, ok := v.(VSystem)
			return ok && len(s.Branches) == 1
		}},
		{"Comp", ast.Comp{
			IBinder: "i",
			A:       typeU,
			Phi:     ast.FaceBot{},
			Tube:    typeU,
			Base:    typeU,
		}, func(v Value) bool {
			// With FaceBot and constant type, transport simplifies to element
			_, ok := v.(VSort)
			return ok
		}},
		{"HComp", ast.HComp{
			A:    typeU,
			Phi:  ast.FaceBot{},
			Tube: typeU,
			Base: typeU,
		}, func(v Value) bool {
			// With FaceBot, hcomp returns base
			_, ok := v.(VSort)
			return ok
		}},
		{"Fill", ast.Fill{
			IBinder: "i",
			A:       typeU,
			Phi:     ast.FaceBot{},
			Tube:    typeU,
			Base:    typeU,
		}, func(v Value) bool {
			_, ok := v.(VFill)
			return ok
		}},
		{"Glue_Empty", ast.Glue{A: typeU, System: nil}, func(v Value) bool {
			// Empty system -> base type
			_, ok := v.(VSort)
			return ok
		}},
		{"GlueElem_Empty", ast.GlueElem{System: nil, Base: typeU}, func(v Value) bool {
			// Empty system -> VGlueElem (not simplified at eval time)
			_, ok := v.(VGlueElem)
			return ok
		}},
		{"Unglue", ast.Unglue{
			Ty: ast.Glue{A: typeU, System: nil},
			G:  typeU,
		}, func(v Value) bool {
			// Should return VUnglue since it's not a VGlueElem
			_, ok := v.(VUnglue)
			return ok
		}},
		{"UA", ast.UA{
			A:     typeU,
			B:     typeU,
			Equiv: typeU,
		}, func(v Value) bool {
			_, ok := v.(VUA)
			return ok
		}},
		{"UABeta", ast.UABeta{
			Equiv: typeU,
			Arg:   typeU,
		}, func(v Value) bool {
			_, ok := v.(VUABeta)
			return ok
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, handled := tryEvalCubical(env, tt.term)
			if !handled {
				t.Errorf("tryEvalCubical did not handle %s", tt.name)
				return
			}
			if !tt.checkFn(val) {
				t.Errorf("tryEvalCubical(%s) returned unexpected value type: %T", tt.name, val)
			}
		})
	}

	// Test that non-cubical terms are not handled
	t.Run("NonCubical_NotHandled", func(t *testing.T) {
		_, handled := tryEvalCubical(env, ast.Sort{U: 0})
		if handled {
			t.Error("tryEvalCubical should not handle non-cubical terms")
		}
	})
}

// TestTryReifyCubical tests the extension hook that reifies cubical values to AST.
// This hook is called from the main reifyAt function when it encounters cubical values.
func TestTryReifyCubical(t *testing.T) {
	level := 0

	// Simple values for testing
	typeV := VSort{Level: 0}
	typeT := ast.Sort{U: 0}

	tests := []struct {
		name    string
		value   Value
		checkFn func(ast.Term) bool
	}{
		{"VI0", VI0{}, func(t ast.Term) bool { _, ok := t.(ast.I0); return ok }},
		{"VI1", VI1{}, func(t ast.Term) bool { _, ok := t.(ast.I1); return ok }},
		{"VIVar", VIVar{Level: 0}, func(t ast.Term) bool {
			iv, ok := t.(ast.IVar)
			return ok && iv.Ix == 0
		}},
		{"VPath", VPath{A: typeV, X: typeV, Y: typeV}, func(t ast.Term) bool {
			_, ok := t.(ast.Path)
			return ok
		}},
		{"VPathP", VPathP{
			A: &IClosure{Env: &Env{}, IEnv: EmptyIEnv(), Term: typeT},
			X: typeV,
			Y: typeV,
		}, func(t ast.Term) bool {
			_, ok := t.(ast.PathP)
			return ok
		}},
		{"VPathLam", VPathLam{
			Body: &IClosure{Env: &Env{}, IEnv: EmptyIEnv(), Term: typeT},
		}, func(t ast.Term) bool {
			_, ok := t.(ast.PathLam)
			return ok
		}},
		{"VTransport", VTransport{
			A: &IClosure{Env: &Env{}, IEnv: EmptyIEnv(), Term: typeT},
			E: typeV,
		}, func(t ast.Term) bool {
			_, ok := t.(ast.Transport)
			return ok
		}},
		{"VFaceTop", VFaceTop{}, func(t ast.Term) bool { _, ok := t.(ast.FaceTop); return ok }},
		{"VFaceBot", VFaceBot{}, func(t ast.Term) bool { _, ok := t.(ast.FaceBot); return ok }},
		{"VFaceEq", VFaceEq{ILevel: 0, IsOne: true}, func(t ast.Term) bool {
			fe, ok := t.(ast.FaceEq)
			return ok && fe.IsOne
		}},
		{"VFaceAnd", VFaceAnd{Left: VFaceTop{}, Right: VFaceTop{}}, func(t ast.Term) bool {
			_, ok := t.(ast.FaceAnd)
			return ok
		}},
		{"VFaceOr", VFaceOr{Left: VFaceBot{}, Right: VFaceBot{}}, func(t ast.Term) bool {
			_, ok := t.(ast.FaceOr)
			return ok
		}},
		{"VPartial", VPartial{Phi: VFaceTop{}, A: typeV}, func(t ast.Term) bool {
			_, ok := t.(ast.Partial)
			return ok
		}},
		{"VSystem", VSystem{Branches: []VSystemBranch{
			{Phi: VFaceTop{}, Term: typeV},
		}}, func(t ast.Term) bool {
			s, ok := t.(ast.System)
			return ok && len(s.Branches) == 1
		}},
		{"VComp", VComp{
			A:    &IClosure{Env: &Env{}, IEnv: EmptyIEnv(), Term: typeT},
			Phi:  VFaceBot{},
			Tube: &IClosure{Env: &Env{}, IEnv: EmptyIEnv(), Term: typeT},
			Base: typeV,
		}, func(t ast.Term) bool {
			_, ok := t.(ast.Comp)
			return ok
		}},
		{"VHComp", VHComp{
			A:    typeV,
			Phi:  VFaceBot{},
			Tube: &IClosure{Env: &Env{}, IEnv: EmptyIEnv(), Term: typeT},
			Base: typeV,
		}, func(t ast.Term) bool {
			_, ok := t.(ast.HComp)
			return ok
		}},
		{"VFill", VFill{
			A:    &IClosure{Env: &Env{}, IEnv: EmptyIEnv(), Term: typeT},
			Phi:  VFaceBot{},
			Tube: &IClosure{Env: &Env{}, IEnv: EmptyIEnv(), Term: typeT},
			Base: typeV,
		}, func(t ast.Term) bool {
			_, ok := t.(ast.Fill)
			return ok
		}},
		{"VGlue", VGlue{A: typeV, System: nil}, func(t ast.Term) bool {
			_, ok := t.(ast.Glue)
			return ok
		}},
		{"VGlueElem", VGlueElem{System: nil, Base: typeV}, func(t ast.Term) bool {
			_, ok := t.(ast.GlueElem)
			return ok
		}},
		{"VUnglue", VUnglue{Ty: typeV, G: typeV}, func(t ast.Term) bool {
			_, ok := t.(ast.Unglue)
			return ok
		}},
		{"VUA", VUA{A: typeV, B: typeV, Equiv: typeV}, func(t ast.Term) bool {
			_, ok := t.(ast.UA)
			return ok
		}},
		{"VUABeta", VUABeta{Equiv: typeV, Arg: typeV}, func(t ast.Term) bool {
			_, ok := t.(ast.UABeta)
			return ok
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			term, handled := tryReifyCubical(level, tt.value)
			if !handled {
				t.Errorf("tryReifyCubical did not handle %s", tt.name)
				return
			}
			if !tt.checkFn(term) {
				t.Errorf("tryReifyCubical(%s) returned unexpected term type: %T", tt.name, term)
			}
		})
	}

	// Test that non-cubical values are not handled
	t.Run("NonCubical_NotHandled", func(t *testing.T) {
		_, handled := tryReifyCubical(level, VSort{Level: 0})
		if handled {
			t.Error("tryReifyCubical should not handle non-cubical values")
		}
	})
}

// TestPathApply_ViaEval tests PathApply through the eval path.
func TestPathApply_ViaEval(t *testing.T) {
	env := &Env{}
	typeU := ast.Sort{U: 0}

	// Test path application via tryEvalCubical
	t.Run("PathApp_I0", func(t *testing.T) {
		// <i> Type0 @ i0 -> Type0
		term := ast.PathApp{
			P: ast.PathLam{Binder: "i", Body: typeU},
			R: ast.I0{},
		}
		val, handled := tryEvalCubical(env, term)
		if !handled {
			t.Error("PathApp not handled")
			return
		}
		if _, ok := val.(VSort); !ok {
			t.Errorf("got %T, want VSort", val)
		}
	})

	t.Run("PathApp_I1", func(t *testing.T) {
		// <i> Type0 @ i1 -> Type0
		term := ast.PathApp{
			P: ast.PathLam{Binder: "i", Body: typeU},
			R: ast.I1{},
		}
		val, handled := tryEvalCubical(env, term)
		if !handled {
			t.Error("PathApp not handled")
			return
		}
		if _, ok := val.(VSort); !ok {
			t.Errorf("got %T, want VSort", val)
		}
	})
}

// TestTransport_ViaEval tests Transport through the eval path.
func TestTransport_ViaEval(t *testing.T) {
	env := &Env{}
	typeU := ast.Sort{U: 0}

	// Transport over constant type should return element unchanged
	t.Run("Transport_Constant", func(t *testing.T) {
		term := ast.Transport{
			A: typeU, // Constant type family
			E: typeU,
		}
		val, handled := tryEvalCubical(env, term)
		if !handled {
			t.Error("Transport not handled")
			return
		}
		if _, ok := val.(VSort); !ok {
			t.Errorf("got %T, want VSort (transport over constant)", val)
		}
	})
}
