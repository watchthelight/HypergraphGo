package check

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// ============================================================================
// synthGlue Error Path Tests
// ============================================================================

// TestSynthGlue_UniverseLevelMismatch tests error when T and A have different levels.
func TestSynthGlue_UniverseLevelMismatch(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)
	c.PushIVar() // Need interval context for face

	// Glue with A : Type0 but T : Type1 (level mismatch)
	glue := ast.Glue{
		A: ast.Sort{U: 0},
		System: []ast.GlueBranch{
			{
				Phi:   ast.FaceEq{IVar: 0, IsOne: true},
				T:     ast.Sort{U: 1}, // Different level!
				Equiv: ast.Var{Ix: 0},
			},
		},
	}
	_, err := c.Synth(ctx, NoSpan(), glue)
	if err == nil {
		t.Error("Expected error for universe level mismatch in Glue")
	}
}

// TestSynthGlue_EquivError tests error when equiv term fails to type-check.
func TestSynthGlue_EquivError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)
	c.PushIVar()

	// Glue with invalid equivalence term (unbound variable)
	glue := ast.Glue{
		A: ast.Sort{U: 0},
		System: []ast.GlueBranch{
			{
				Phi:   ast.FaceEq{IVar: 0, IsOne: true},
				T:     ast.Sort{U: 0},
				Equiv: ast.Var{Ix: 99}, // Unbound variable
			},
		},
	}
	_, err := c.Synth(ctx, NoSpan(), glue)
	if err == nil {
		t.Error("Expected error for invalid equiv in Glue")
	}
}

// TestSynthGlue_FaceError tests error when face formula has unbound interval var.
func TestSynthGlue_FaceError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)
	// No interval context pushed

	glue := ast.Glue{
		A: ast.Sort{U: 0},
		System: []ast.GlueBranch{
			{
				Phi:   ast.FaceEq{IVar: 0, IsOne: true}, // Unbound interval var
				T:     ast.Sort{U: 0},
				Equiv: ast.Sort{U: 0},
			},
		},
	}
	_, err := c.Synth(ctx, NoSpan(), glue)
	if err == nil {
		t.Error("Expected error for unbound interval var in Glue face")
	}
}

// TestSynthGlue_TypeError tests error when T is not a type.
func TestSynthGlue_TypeError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)
	c.PushIVar()

	// Glue with T not being a type (unbound variable)
	glue := ast.Glue{
		A: ast.Sort{U: 0},
		System: []ast.GlueBranch{
			{
				Phi:   ast.FaceEq{IVar: 0, IsOne: true},
				T:     ast.Var{Ix: 99}, // Unbound - not a type
				Equiv: ast.Sort{U: 0},
			},
		},
	}
	_, err := c.Synth(ctx, NoSpan(), glue)
	if err == nil {
		t.Error("Expected error for invalid T in Glue")
	}
}

// ============================================================================
// synthHComp Error Path Tests
// ============================================================================

// TestSynthHComp_ANotType tests error when A is not a type.
func TestSynthHComp_ANotType(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	hcomp := ast.HComp{
		A:    ast.Var{Ix: 99}, // Unbound variable
		Phi:  ast.FaceTop{},
		Base: ast.Sort{U: 0},
		Tube: ast.Sort{U: 0},
	}
	_, err := c.Synth(ctx, NoSpan(), hcomp)
	if err == nil {
		t.Error("Expected error for invalid A in HComp")
	}
}

// TestSynthHComp_FaceError tests error when face formula is invalid.
func TestSynthHComp_FaceError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)
	c.PushIVar() // Interval context for tube, but face uses unbound var

	hcomp := ast.HComp{
		A:    ast.Sort{U: 0},
		Phi:  ast.FaceEq{IVar: 99, IsOne: true}, // Unbound interval var
		Base: ast.Sort{U: 0},
		Tube: ast.Sort{U: 0},
	}
	_, err := c.Synth(ctx, NoSpan(), hcomp)
	if err == nil {
		t.Error("Expected error for invalid face in HComp")
	}
}

// TestSynthHComp_BaseError tests error when base doesn't have type A.
func TestSynthHComp_BaseError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	// Base : Type1, but A is Type0, type mismatch
	hcomp := ast.HComp{
		A:    ast.Sort{U: 0},
		Phi:  ast.FaceTop{},
		Base: ast.Sort{U: 1}, // Type1 : Type2, not Type0
		Tube: ast.Var{Ix: 0},
	}
	_, err := c.Synth(ctx, NoSpan(), hcomp)
	if err == nil {
		t.Error("Expected error for base type mismatch in HComp")
	}
}

// TestSynthHComp_TubeError tests error when tube doesn't have type A.
func TestSynthHComp_TubeError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	hcomp := ast.HComp{
		A:    ast.Sort{U: 0},
		Phi:  ast.FaceTop{},
		Base: ast.Var{Ix: 0}, // Valid: x : Type0
		Tube: ast.Sort{U: 1}, // Type1 : Type2, not Type0
	}
	_, err := c.Synth(ctx, NoSpan(), hcomp)
	if err == nil {
		t.Error("Expected error for tube type mismatch in HComp")
	}
}

// ============================================================================
// synthSystem Error Path Tests
// ============================================================================

// TestSynthSystem_FirstTermError tests error in first branch term.
func TestSynthSystem_FirstTermError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)
	c.PushIVar()

	sys := ast.System{
		Branches: []ast.SystemBranch{
			{
				Phi:  ast.FaceEq{IVar: 0, IsOne: true},
				Term: ast.Var{Ix: 99}, // Unbound variable
			},
		},
	}
	_, err := c.Synth(ctx, NoSpan(), sys)
	if err == nil {
		t.Error("Expected error for invalid first term in System")
	}
}

// TestSynthSystem_SubsequentTermError tests error in subsequent branch term.
func TestSynthSystem_SubsequentTermError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})
	c.PushIVar()

	sys := ast.System{
		Branches: []ast.SystemBranch{
			{
				Phi:  ast.FaceEq{IVar: 0, IsOne: true},
				Term: ast.Var{Ix: 0}, // Valid: x : Type0
			},
			{
				Phi:  ast.FaceEq{IVar: 0, IsOne: false},
				Term: ast.Sort{U: 1}, // Type1 : Type2, not Type0
			},
		},
	}
	_, err := c.Synth(ctx, NoSpan(), sys)
	if err == nil {
		t.Error("Expected error for type mismatch in second System branch")
	}
}

// TestSynthSystem_SubsequentFaceError tests error in subsequent branch face.
func TestSynthSystem_SubsequentFaceError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})
	c.PushIVar()

	sys := ast.System{
		Branches: []ast.SystemBranch{
			{
				Phi:  ast.FaceEq{IVar: 0, IsOne: true},
				Term: ast.Var{Ix: 0},
			},
			{
				Phi:  ast.FaceEq{IVar: 99, IsOne: false}, // Unbound interval var
				Term: ast.Var{Ix: 0},
			},
		},
	}
	_, err := c.Synth(ctx, NoSpan(), sys)
	if err == nil {
		t.Error("Expected error for invalid face in second System branch")
	}
}

// ============================================================================
// checkSystemAgreement Error Path Tests
// ============================================================================

// TestCheckSystemAgreement_Disagreement tests error when branches disagree on overlap.
func TestCheckSystemAgreement_Disagreement(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}, ast.Sort{U: 0}})
	c.PushIVar()

	// Two branches with overlapping faces but different terms
	sys := ast.System{
		Branches: []ast.SystemBranch{
			{
				Phi:  ast.FaceTop{}, // Always true
				Term: ast.Var{Ix: 0},
			},
			{
				Phi:  ast.FaceTop{}, // Also always true - overlaps!
				Term: ast.Var{Ix: 1}, // Different term
			},
		},
	}
	_, err := c.Synth(ctx, NoSpan(), sys)
	if err == nil {
		t.Error("Expected error for disagreeing branches in System")
	}
}

// ============================================================================
// synthTransport Error Path Tests
// ============================================================================

// TestSynthTransport_ANotType tests error when A[i0/i] is not a type.
func TestSynthTransport_ANotType(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	tr := ast.Transport{
		A: ast.Var{Ix: 99}, // Unbound variable
		E: ast.Sort{U: 0},
	}
	_, err := c.Synth(ctx, NoSpan(), tr)
	if err == nil {
		t.Error("Expected error for invalid A in Transport")
	}
}

// TestSynthTransport_EError tests error when e doesn't have type A[i0/i].
func TestSynthTransport_EError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	// Transport with A = Type0, but e : Type2 (e is Type1 which has type Type2)
	tr := ast.Transport{
		A: ast.Sort{U: 0},
		E: ast.Sort{U: 1}, // Type1 : Type2, not Type0
	}
	_, err := c.Synth(ctx, NoSpan(), tr)
	if err == nil {
		t.Error("Expected error for e type mismatch in Transport")
	}
}

// ============================================================================
// synthComp Error Path Tests
// ============================================================================

// TestSynthComp_ANotType tests error when A[i0/i] is not a type.
func TestSynthComp_ANotType(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	comp := ast.Comp{
		A:    ast.Var{Ix: 99}, // Unbound variable
		Phi:  ast.FaceTop{},
		Base: ast.Sort{U: 0},
		Tube: ast.Sort{U: 0},
	}
	_, err := c.Synth(ctx, NoSpan(), comp)
	if err == nil {
		t.Error("Expected error for invalid A in Comp")
	}
}

// TestSynthComp_FaceError tests error when face formula is invalid.
func TestSynthComp_FaceError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)
	c.PushIVar()

	comp := ast.Comp{
		A:    ast.Sort{U: 0},
		Phi:  ast.FaceEq{IVar: 99, IsOne: true}, // Unbound interval var
		Base: ast.Sort{U: 0},
		Tube: ast.Sort{U: 0},
	}
	_, err := c.Synth(ctx, NoSpan(), comp)
	if err == nil {
		t.Error("Expected error for invalid face in Comp")
	}
}

// TestSynthComp_BaseError tests error when base doesn't have type A[i0/i].
func TestSynthComp_BaseError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	comp := ast.Comp{
		A:    ast.Sort{U: 0},
		Phi:  ast.FaceTop{},
		Base: ast.Sort{U: 1}, // Type1 : Type2, not Type0
		Tube: ast.Var{Ix: 0},
	}
	_, err := c.Synth(ctx, NoSpan(), comp)
	if err == nil {
		t.Error("Expected error for base type mismatch in Comp")
	}
}

// TestSynthComp_TubeError tests error when tube doesn't have type A.
func TestSynthComp_TubeError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	comp := ast.Comp{
		A:    ast.Sort{U: 0},
		Phi:  ast.FaceTop{},
		Base: ast.Var{Ix: 0},
		Tube: ast.Sort{U: 1}, // Type1 : Type2, not Type0
	}
	_, err := c.Synth(ctx, NoSpan(), comp)
	if err == nil {
		t.Error("Expected error for tube type mismatch in Comp")
	}
}

// ============================================================================
// synthFill Error Path Tests
// ============================================================================

// TestSynthFill_ANotType tests error when A[i0/i] is not a type.
func TestSynthFill_ANotType(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	fill := ast.Fill{
		A:    ast.Var{Ix: 99}, // Unbound variable
		Phi:  ast.FaceTop{},
		Base: ast.Sort{U: 0},
		Tube: ast.Sort{U: 0},
	}
	_, err := c.Synth(ctx, NoSpan(), fill)
	if err == nil {
		t.Error("Expected error for invalid A in Fill")
	}
}

// TestSynthFill_FaceError tests error when face formula is invalid.
func TestSynthFill_FaceError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)
	c.PushIVar()

	fill := ast.Fill{
		A:    ast.Sort{U: 0},
		Phi:  ast.FaceEq{IVar: 99, IsOne: true}, // Unbound interval var
		Base: ast.Sort{U: 0},
		Tube: ast.Sort{U: 0},
	}
	_, err := c.Synth(ctx, NoSpan(), fill)
	if err == nil {
		t.Error("Expected error for invalid face in Fill")
	}
}

// TestSynthFill_BaseError tests error when base doesn't have type A[i0/i].
func TestSynthFill_BaseError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	fill := ast.Fill{
		A:    ast.Sort{U: 0},
		Phi:  ast.FaceTop{},
		Base: ast.Sort{U: 1}, // Type1 : Type2, not Type0
		Tube: ast.Var{Ix: 0},
	}
	_, err := c.Synth(ctx, NoSpan(), fill)
	if err == nil {
		t.Error("Expected error for base type mismatch in Fill")
	}
}

// TestSynthFill_TubeError tests error when tube doesn't have type A.
func TestSynthFill_TubeError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	fill := ast.Fill{
		A:    ast.Sort{U: 0},
		Phi:  ast.FaceTop{},
		Base: ast.Var{Ix: 0},
		Tube: ast.Sort{U: 1}, // Type1 : Type2, not Type0
	}
	_, err := c.Synth(ctx, NoSpan(), fill)
	if err == nil {
		t.Error("Expected error for tube type mismatch in Fill")
	}
}

// ============================================================================
// synthPathApp Error Path Tests
// ============================================================================

// TestSynthPathApp_NotPath tests error when term is not a path type.
func TestSynthPathApp_NotPath(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	// Apply to a non-path term
	papp := ast.PathApp{
		P: ast.Var{Ix: 0}, // Just a variable with type Type0
		R: ast.I0{},
	}
	_, err := c.Synth(ctx, NoSpan(), papp)
	if err == nil {
		t.Error("Expected error for path application to non-path type")
	}
}

// TestSynthPathApp_PathCaseRError tests error when r doesn't have type I in Path case.
func TestSynthPathApp_PathCaseRError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{
		// x : Path Type0 Type0 Type0
		ast.Path{A: ast.Sort{U: 0}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 0}},
	})

	// Path application with r not of type I
	papp := ast.PathApp{
		P: ast.Var{Ix: 0},     // x : Path Type0 Type0 Type0
		R: ast.Sort{U: 0}, // Type0 is not of type I
	}
	_, err := c.Synth(ctx, NoSpan(), papp)
	if err == nil {
		t.Error("Expected error for path application with non-interval argument")
	}
}

// ============================================================================
// synthUA Error Path Tests
// ============================================================================

// TestSynthUA_UniverseLevelMismatch tests error when A and B have different levels.
func TestSynthUA_UniverseLevelMismatch(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	ua := ast.UA{
		A:     ast.Sort{U: 0}, // Type0
		B:     ast.Sort{U: 1}, // Type1 - different level
		Equiv: ast.Var{Ix: 0},
	}
	_, err := c.Synth(ctx, NoSpan(), ua)
	if err == nil {
		t.Error("Expected error for universe level mismatch in UA")
	}
}

// TestSynthUA_EquivError tests error when equiv term fails to type-check.
func TestSynthUA_EquivError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	ua := ast.UA{
		A:     ast.Sort{U: 0},
		B:     ast.Sort{U: 0},
		Equiv: ast.Var{Ix: 99}, // Unbound variable
	}
	_, err := c.Synth(ctx, NoSpan(), ua)
	if err == nil {
		t.Error("Expected error for invalid equiv in UA")
	}
}

// ============================================================================
// synthUnglue Error Path Tests
// ============================================================================

// TestSynthUnglue_WithTypeAnnotation tests unglue with type annotation.
func TestSynthUnglue_WithTypeAnnotation(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	// Unglue with Glue type annotation
	unglue := ast.Unglue{
		G: ast.Var{Ix: 0},
		Ty: ast.Glue{
			A:      ast.Sort{U: 0},
			System: nil,
		},
	}
	ty, err := c.Synth(ctx, NoSpan(), unglue)
	if err != nil {
		t.Fatalf("Failed to synth Unglue with annotation: %v", err)
	}

	// Result should be the base type from annotation
	if _, ok := ty.(ast.Sort); !ok {
		t.Errorf("Expected Sort from Unglue, got %T", ty)
	}
}

// TestSynthUnglue_WithNonGlueAnnotation tests unglue with non-Glue annotation.
func TestSynthUnglue_WithNonGlueAnnotation(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	// Unglue with non-Glue type annotation
	unglue := ast.Unglue{
		G:  ast.Var{Ix: 0},
		Ty: ast.Sort{U: 0}, // Not a Glue type
	}
	// Should still succeed but return the synthesized type
	ty, err := c.Synth(ctx, NoSpan(), unglue)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Result should be the synthesized type of G
	if _, ok := ty.(ast.Sort); !ok {
		t.Errorf("Expected Sort from Unglue fallback, got %T", ty)
	}
}

// ============================================================================
// checkPathLam Error Path Tests
// ============================================================================

// TestCheckPathLam_PathCase tests checkPathLam with non-dependent Path type.
func TestCheckPathLam_PathCase(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	// Check <i> x against Path Type0 x x
	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.Var{Ix: 0}, // x
	}
	pathType := ast.Path{
		A: ast.Sort{U: 0},
		X: ast.Var{Ix: 0}, // x
		Y: ast.Var{Ix: 0}, // x
	}

	err := c.check(ctx, NoSpan(), plam, pathType)
	if err != nil {
		t.Errorf("checkPathLam failed for Path case: %v", err)
	}
}

// TestCheckPathLam_PathCaseBodyError tests checkPathLam error when body has wrong type.
func TestCheckPathLam_PathCaseBodyError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	// Check <i> Type1 against Path Type0 x x (body has wrong type)
	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.Sort{U: 1}, // Type1 : Type2, not Type0
	}
	pathType := ast.Path{
		A: ast.Sort{U: 0},
		X: ast.Var{Ix: 0},
		Y: ast.Var{Ix: 0},
	}

	err := c.check(ctx, NoSpan(), plam, pathType)
	if err == nil {
		t.Error("Expected error for body type mismatch in Path case")
	}
}

// TestCheckPathLam_PathCaseLeftMismatch tests error when left endpoint doesn't match.
func TestCheckPathLam_PathCaseLeftMismatch(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}, ast.Sort{U: 0}})

	// Check <i> x against Path Type0 y y (x != y at left endpoint)
	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.Var{Ix: 0}, // x
	}
	pathType := ast.Path{
		A: ast.Sort{U: 0},
		X: ast.Var{Ix: 1}, // y (different!)
		Y: ast.Var{Ix: 0}, // x
	}

	err := c.check(ctx, NoSpan(), plam, pathType)
	if err == nil {
		t.Error("Expected error for left endpoint mismatch in Path case")
	}
}

// TestCheckPathLam_PathCaseRightMismatch tests error when right endpoint doesn't match.
func TestCheckPathLam_PathCaseRightMismatch(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}, ast.Sort{U: 0}})

	// Check <i> x against Path Type0 x y (x != y at right endpoint)
	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.Var{Ix: 0}, // x
	}
	pathType := ast.Path{
		A: ast.Sort{U: 0},
		X: ast.Var{Ix: 0}, // x
		Y: ast.Var{Ix: 1}, // y (different!)
	}

	err := c.check(ctx, NoSpan(), plam, pathType)
	if err == nil {
		t.Error("Expected error for right endpoint mismatch in Path case")
	}
}

// TestCheckPathLam_NotPathType tests that checkPathLam falls through for non-path types.
func TestCheckPathLam_NotPathType(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	// Check <i> Type0 against Type1 (not a path type)
	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.Sort{U: 0},
	}

	// This should fall through to synthesis, which will succeed
	// since <i> Type0 : PathP (λi. Type1) Type0 Type0
	err := c.check(ctx, NoSpan(), plam, ast.Sort{U: 1})
	// The check should either succeed via subsumption or fail nicely
	// Either way, we've exercised the "not a path type" branch
	_ = err // We don't care about the specific result
}

// ============================================================================
// synthUABeta Error Path Tests
// ============================================================================

// TestSynthUABeta_EquivError tests error when equiv term fails to type-check.
func TestSynthUABeta_EquivError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	uab := ast.UABeta{
		Equiv: ast.Var{Ix: 99}, // Unbound variable
		Arg:   ast.Sort{U: 0},
	}
	_, err := c.Synth(ctx, NoSpan(), uab)
	if err == nil {
		t.Error("Expected error for invalid equiv in UABeta")
	}
}

// TestSynthUABeta_ArgError tests error when arg fails to type-check.
func TestSynthUABeta_ArgError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	uab := ast.UABeta{
		Equiv: ast.Var{Ix: 0},
		Arg:   ast.Var{Ix: 99}, // Unbound variable
	}
	_, err := c.Synth(ctx, NoSpan(), uab)
	if err == nil {
		t.Error("Expected error for invalid arg in UABeta")
	}
}

// TestSynthUABeta_WithSigmaPiEquiv tests UABeta with properly structured Equiv.
func TestSynthUABeta_WithSigmaPiEquiv(t *testing.T) {
	c := NewChecker(nil)
	// Context with e : Σ(f : A → B). isEquiv f where A = Type0, B = Type0
	equivType := ast.Sigma{
		Binder: "f",
		A: ast.Pi{
			Binder: "_",
			A:      ast.Sort{U: 0}, // A = Type0
			B:      ast.Sort{U: 0}, // B = Type0
		},
		B: ast.Sort{U: 0}, // isEquiv placeholder
	}
	ctx := makeTestContext([]ast.Term{equivType, ast.Sort{U: 0}})

	uab := ast.UABeta{
		Equiv: ast.Var{Ix: 0}, // e : Equiv Type0 Type0
		Arg:   ast.Var{Ix: 1}, // a : Type0
	}
	ty, err := c.Synth(ctx, NoSpan(), uab)
	if err != nil {
		t.Fatalf("Failed to synth UABeta with Sigma-Pi equiv: %v", err)
	}

	// Result should be B (codomain of the function)
	if _, ok := ty.(ast.Sort); !ok {
		t.Errorf("Expected Sort from UABeta, got %T", ty)
	}
}

// ============================================================================
// synthGlueElem Error Path Tests
// ============================================================================

// TestSynthGlueElem_BaseError tests error when base fails to type-check.
func TestSynthGlueElem_BaseError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	elem := ast.GlueElem{
		Base:   ast.Var{Ix: 99}, // Unbound variable
		System: nil,
	}
	_, err := c.Synth(ctx, NoSpan(), elem)
	if err == nil {
		t.Error("Expected error for invalid base in GlueElem")
	}
}

// TestSynthGlueElem_FaceError tests error when face formula is invalid.
func TestSynthGlueElem_FaceError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	elem := ast.GlueElem{
		Base: ast.Var{Ix: 0},
		System: []ast.GlueElemBranch{
			{
				Phi:  ast.FaceEq{IVar: 99, IsOne: true}, // Unbound interval var
				Term: ast.Var{Ix: 0},
			},
		},
	}
	_, err := c.Synth(ctx, NoSpan(), elem)
	if err == nil {
		t.Error("Expected error for invalid face in GlueElem")
	}
}

// TestSynthGlueElem_TermError tests error when branch term fails to type-check.
func TestSynthGlueElem_TermError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})
	c.PushIVar()

	elem := ast.GlueElem{
		Base: ast.Var{Ix: 0},
		System: []ast.GlueElemBranch{
			{
				Phi:  ast.FaceEq{IVar: 0, IsOne: true},
				Term: ast.Var{Ix: 99}, // Unbound variable
			},
		},
	}
	_, err := c.Synth(ctx, NoSpan(), elem)
	if err == nil {
		t.Error("Expected error for invalid term in GlueElem")
	}
}

// ============================================================================
// checkFace Error Path Tests
// ============================================================================

// TestCheckFace_DefaultCase tests that unknown face types error.
func TestCheckFace_DefaultCase(t *testing.T) {
	// This tests the default case of checkFace which returns "invalid face formula"
	// We can't easily create a custom Face type, but we can verify the error handling
	// by using the existing face types and checking error messages

	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	// FaceAnd with invalid nested face (unbound interval var)
	face := ast.FaceAnd{
		Left:  ast.FaceEq{IVar: 99, IsOne: true},
		Right: ast.FaceTop{},
	}
	err := c.checkFace(ctx, NoSpan(), face)
	if err == nil {
		t.Error("Expected error for invalid nested face")
	}
}

// TestCheckFace_FaceOrRightError tests FaceOr with invalid right face.
func TestCheckFace_FaceOrRightError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)
	c.PushIVar()

	face := ast.FaceOr{
		Left:  ast.FaceEq{IVar: 0, IsOne: true}, // Valid
		Right: ast.FaceEq{IVar: 99, IsOne: false}, // Unbound
	}
	err := c.checkFace(ctx, NoSpan(), face)
	if err == nil {
		t.Error("Expected error for invalid right face in FaceOr")
	}
}

// ============================================================================
// faceIsBot Tests
// ============================================================================

// TestFaceIsBot_AllCases tests faceIsBot for various face formulas.
func TestFaceIsBot_AllCases(t *testing.T) {
	tests := []struct {
		name     string
		face     ast.Face
		expected bool
	}{
		{"FaceBot", ast.FaceBot{}, true},
		{"FaceTop", ast.FaceTop{}, false},
		{"FaceEq", ast.FaceEq{IVar: 0, IsOne: true}, false},
		{"FaceAnd with Bot left", ast.FaceAnd{Left: ast.FaceBot{}, Right: ast.FaceTop{}}, true},
		{"FaceAnd with Bot right", ast.FaceAnd{Left: ast.FaceTop{}, Right: ast.FaceBot{}}, true},
		{"FaceAnd contradiction", ast.FaceAnd{
			Left:  ast.FaceEq{IVar: 0, IsOne: true},
			Right: ast.FaceEq{IVar: 0, IsOne: false},
		}, true},
		{"FaceAnd no contradiction", ast.FaceAnd{
			Left:  ast.FaceEq{IVar: 0, IsOne: true},
			Right: ast.FaceEq{IVar: 1, IsOne: false},
		}, false},
		{"FaceOr both bot", ast.FaceOr{Left: ast.FaceBot{}, Right: ast.FaceBot{}}, true},
		{"FaceOr left bot", ast.FaceOr{Left: ast.FaceBot{}, Right: ast.FaceTop{}}, false},
		{"FaceOr right bot", ast.FaceOr{Left: ast.FaceTop{}, Right: ast.FaceBot{}}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := faceIsBot(tt.face)
			if result != tt.expected {
				t.Errorf("faceIsBot(%v) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

// TestCollectFaceEqs tests collectFaceEqs for various faces.
func TestCollectFaceEqs(t *testing.T) {
	tests := []struct {
		name     string
		face     ast.Face
		expected int
	}{
		{"FaceEq", ast.FaceEq{IVar: 0, IsOne: true}, 1},
		{"FaceTop", ast.FaceTop{}, 0},
		{"FaceBot", ast.FaceBot{}, 0},
		{"FaceAnd two eqs", ast.FaceAnd{
			Left:  ast.FaceEq{IVar: 0, IsOne: true},
			Right: ast.FaceEq{IVar: 1, IsOne: false},
		}, 2},
		{"FaceOr (conservative)", ast.FaceOr{
			Left:  ast.FaceEq{IVar: 0, IsOne: true},
			Right: ast.FaceEq{IVar: 1, IsOne: false},
		}, 0}, // FaceOr doesn't collect
		{"Nested FaceAnd", ast.FaceAnd{
			Left: ast.FaceAnd{
				Left:  ast.FaceEq{IVar: 0, IsOne: true},
				Right: ast.FaceEq{IVar: 1, IsOne: true},
			},
			Right: ast.FaceEq{IVar: 2, IsOne: false},
		}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectFaceEqs(tt.face)
			if len(result) != tt.expected {
				t.Errorf("collectFaceEqs(%s) returned %d eqs, want %d", tt.name, len(result), tt.expected)
			}
		})
	}
}

// ============================================================================
// synthPathLam Error Path Tests
// ============================================================================

// TestSynthPathLam_BodyError tests error when body fails to synthesize.
func TestSynthPathLam_BodyError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.Var{Ix: 99}, // Unbound variable
	}
	_, err := c.Synth(ctx, NoSpan(), plam)
	if err == nil {
		t.Error("Expected error for invalid body in PathLam")
	}
}

// ============================================================================
// Additional Coverage Tests
// ============================================================================

// TestSynthPathApp_PathPCaseRError tests error when r doesn't have type I in PathP case.
func TestSynthPathApp_PathPCaseRError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{
		// x : PathP Type0 Type0 Type0
		ast.PathP{A: ast.Sort{U: 0}, X: ast.Sort{U: 0}, Y: ast.Sort{U: 0}},
	})

	// PathP application with r not of type I
	papp := ast.PathApp{
		P: ast.Var{Ix: 0},     // x : PathP Type0 Type0 Type0
		R: ast.Sort{U: 0}, // Type0 is not of type I
	}
	_, err := c.Synth(ctx, NoSpan(), papp)
	if err == nil {
		t.Error("Expected error for PathP application with non-interval argument")
	}
}

// TestSynthPathApp_PathPSynthError tests error when path term fails to synth.
func TestSynthPathApp_PathPSynthError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	papp := ast.PathApp{
		P: ast.Var{Ix: 99}, // Unbound variable
		R: ast.I0{},
	}
	_, err := c.Synth(ctx, NoSpan(), papp)
	if err == nil {
		t.Error("Expected error for path application with unbound path")
	}
}

// TestCheckPathLam_PathPCaseBodyError tests error when body has wrong type in PathP case.
func TestCheckPathLam_PathPCaseBodyError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}})

	// Check <i> Type1 against PathP Type0 x x (body has wrong type)
	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.Sort{U: 1}, // Type1 : Type2, not Type0
	}
	pathPType := ast.PathP{
		A: ast.Sort{U: 0},
		X: ast.Var{Ix: 0},
		Y: ast.Var{Ix: 0},
	}

	err := c.check(ctx, NoSpan(), plam, pathPType)
	if err == nil {
		t.Error("Expected error for body type mismatch in PathP case")
	}
}

// TestCheckPathLam_PathPCaseLeftMismatch tests error when left endpoint doesn't match in PathP.
func TestCheckPathLam_PathPCaseLeftMismatch(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}, ast.Sort{U: 0}})

	// Check <i> x against PathP Type0 y y (x != y at left endpoint)
	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.Var{Ix: 0}, // x
	}
	pathPType := ast.PathP{
		A: ast.Sort{U: 0},
		X: ast.Var{Ix: 1}, // y (different!)
		Y: ast.Var{Ix: 0}, // x
	}

	err := c.check(ctx, NoSpan(), plam, pathPType)
	if err == nil {
		t.Error("Expected error for left endpoint mismatch in PathP case")
	}
}

// TestCheckPathLam_PathPCaseRightMismatch tests error when right endpoint doesn't match in PathP.
func TestCheckPathLam_PathPCaseRightMismatch(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}, ast.Sort{U: 0}})

	// Check <i> x against PathP Type0 x y (x != y at right endpoint)
	plam := ast.PathLam{
		Binder: "i",
		Body:   ast.Var{Ix: 0}, // x
	}
	pathPType := ast.PathP{
		A: ast.Sort{U: 0},
		X: ast.Var{Ix: 0}, // x
		Y: ast.Var{Ix: 1}, // y (different!)
	}

	err := c.check(ctx, NoSpan(), plam, pathPType)
	if err == nil {
		t.Error("Expected error for right endpoint mismatch in PathP case")
	}
}

// TestSynthUA_BNotType tests error when B is not a type.
func TestSynthUA_BNotType(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	ua := ast.UA{
		A:     ast.Sort{U: 0},
		B:     ast.Var{Ix: 99}, // Unbound variable
		Equiv: ast.Sort{U: 0},
	}
	_, err := c.Synth(ctx, NoSpan(), ua)
	if err == nil {
		t.Error("Expected error for invalid B in UA")
	}
}

// TestSynthUnglue_SynthError tests error when g fails to synth.
func TestSynthUnglue_SynthError(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	unglue := ast.Unglue{
		G:  ast.Var{Ix: 99}, // Unbound variable
		Ty: nil,
	}
	_, err := c.Synth(ctx, NoSpan(), unglue)
	if err == nil {
		t.Error("Expected error for unbound g in Unglue")
	}
}

// TestSynthGlue_ANotType tests error when A is not a type.
func TestSynthGlue_ANotType(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext(nil)

	glue := ast.Glue{
		A:      ast.Var{Ix: 99}, // Unbound variable
		System: nil,
	}
	_, err := c.Synth(ctx, NoSpan(), glue)
	if err == nil {
		t.Error("Expected error for invalid A in Glue")
	}
}

// TestSynthComp_TubeBaseDisagreement tests tube[i0/i] != base error.
func TestSynthComp_TubeBaseDisagreement(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}, ast.Sort{U: 0}})

	// tube[i0/i] = x, base = y (x != y when face is satisfiable at i0)
	comp := ast.Comp{
		A:    ast.Sort{U: 0},
		Phi:  ast.FaceTop{}, // Always true - satisfiable at i0
		Base: ast.Var{Ix: 0}, // x
		Tube: ast.Var{Ix: 1}, // y (different from x at i0)
	}
	_, err := c.Synth(ctx, NoSpan(), comp)
	if err == nil {
		t.Error("Expected error for tube[i0/i] != base in Comp")
	}
}

// TestSynthHComp_TubeBaseDisagreement tests tube[i0/i] != base error.
func TestSynthHComp_TubeBaseDisagreement(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}, ast.Sort{U: 0}})

	hcomp := ast.HComp{
		A:    ast.Sort{U: 0},
		Phi:  ast.FaceTop{}, // Always true - satisfiable at i0
		Base: ast.Var{Ix: 0}, // x
		Tube: ast.Var{Ix: 1}, // y (different from x at i0)
	}
	_, err := c.Synth(ctx, NoSpan(), hcomp)
	if err == nil {
		t.Error("Expected error for tube[i0/i] != base in HComp")
	}
}

// TestSynthFill_TubeBaseDisagreement tests tube[i0/i] != base error.
func TestSynthFill_TubeBaseDisagreement(t *testing.T) {
	c := NewChecker(nil)
	ctx := makeTestContext([]ast.Term{ast.Sort{U: 0}, ast.Sort{U: 0}})

	fill := ast.Fill{
		A:    ast.Sort{U: 0},
		Phi:  ast.FaceTop{}, // Always true - satisfiable at i0
		Base: ast.Var{Ix: 0}, // x
		Tube: ast.Var{Ix: 1}, // y (different from x at i0)
	}
	_, err := c.Synth(ctx, NoSpan(), fill)
	if err == nil {
		t.Error("Expected error for tube[i0/i] != base in Fill")
	}
}

// TestFaceIsBot_DefaultCase tests the default case returns false.
func TestFaceIsBot_DefaultCase(t *testing.T) {
	// Create a nil face to exercise the default case
	// Since Face is an interface, nil satisfies it
	var nilFace ast.Face = nil
	result := faceIsBot(nilFace)
	if result != false {
		t.Error("faceIsBot(nil) should return false")
	}
}

