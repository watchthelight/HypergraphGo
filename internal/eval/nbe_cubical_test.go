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
