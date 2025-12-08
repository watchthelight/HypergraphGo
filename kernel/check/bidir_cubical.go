package check

import (
	"fmt"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
	tyctx "github.com/watchthelight/HypergraphGo/kernel/ctx"
	"github.com/watchthelight/HypergraphGo/kernel/subst"
)

// ICtx tracks interval variable bindings for cubical type checking.
// Interval variables are tracked separately from term variables.
type ICtx struct {
	depth int // number of bound interval variables
}

// NewICtx creates an empty interval context.
func NewICtx() *ICtx {
	return &ICtx{depth: 0}
}

// Extend adds an interval binding and returns the new context.
func (ic *ICtx) Extend() *ICtx {
	return &ICtx{depth: ic.depth + 1}
}

// CheckIVar checks if an interval variable index is valid.
func (ic *ICtx) CheckIVar(ix int) bool {
	return ix >= 0 && ix < ic.depth
}

// synthExtension handles type synthesis for cubical terms.
// Returns (type, nil, true) if handled, or (nil, err, true) if handled with error,
// or (nil, nil, false) if not a cubical term.
func synthExtension(c *Checker, context *tyctx.Ctx, span Span, term ast.Term) (ast.Term, *TypeError, bool) {
	// For full cubical type checking, we'd need to track ICtx.
	// For now, provide basic handling for cubical terms.
	switch t := term.(type) {
	case ast.Interval:
		// I : Type0 (the interval is not really a type, but this suffices for now)
		return ast.Sort{U: 0}, nil, true

	case ast.I0, ast.I1:
		// i0, i1 : I
		return ast.Interval{}, nil, true

	case ast.IVar:
		// Interval variables have type I
		// Validate against interval context
		if !c.CheckIVar(t.Ix) {
			return nil, errUnboundIVar(span, t.Ix), true
		}
		return ast.Interval{}, nil, true

	case ast.Path:
		return synthPath(c, context, span, t)

	case ast.PathP:
		return synthPathP(c, context, span, t)

	case ast.PathLam:
		return synthPathLam(c, context, span, t)

	case ast.PathApp:
		return synthPathApp(c, context, span, t)

	case ast.Transport:
		return synthTransport(c, context, span, t)

	// Face formulas
	case ast.FaceTop, ast.FaceBot:
		// Face formulas are not types themselves, but we allow them in certain contexts
		// They represent face constraints (cofibrations)
		// For now, treat them as having a special "Face" kind
		return ast.Global{Name: "Face"}, nil, true

	case ast.FaceEq:
		// Check interval variable is valid
		if !c.CheckIVar(t.IVar) {
			return nil, errUnboundIVar(span, t.IVar), true
		}
		return ast.Global{Name: "Face"}, nil, true

	case ast.FaceAnd:
		// Check both sides are faces
		if checkErr := c.checkFace(context, span, t.Left); checkErr != nil {
			return nil, checkErr, true
		}
		if checkErr := c.checkFace(context, span, t.Right); checkErr != nil {
			return nil, checkErr, true
		}
		return ast.Global{Name: "Face"}, nil, true

	case ast.FaceOr:
		// Check both sides are faces
		if checkErr := c.checkFace(context, span, t.Left); checkErr != nil {
			return nil, checkErr, true
		}
		if checkErr := c.checkFace(context, span, t.Right); checkErr != nil {
			return nil, checkErr, true
		}
		return ast.Global{Name: "Face"}, nil, true

	// Partial types
	case ast.Partial:
		return synthPartial(c, context, span, t)

	case ast.System:
		return synthSystem(c, context, span, t)

	// Composition operations
	case ast.Comp:
		return synthComp(c, context, span, t)

	case ast.HComp:
		return synthHComp(c, context, span, t)

	case ast.Fill:
		return synthFill(c, context, span, t)

	case ast.Glue:
		return synthGlue(c, context, span, t)

	case ast.GlueElem:
		return synthGlueElem(c, context, span, t)

	case ast.Unglue:
		return synthUnglue(c, context, span, t)

	case ast.UA:
		return synthUA(c, context, span, t)

	case ast.UABeta:
		return synthUABeta(c, context, span, t)

	default:
		return nil, nil, false
	}
}

// synthPath synthesizes the type of a non-dependent path type.
// Path A x y : Sort_i where A : Sort_i, x : A, y : A
func synthPath(c *Checker, context *tyctx.Ctx, span Span, path ast.Path) (ast.Term, *TypeError, bool) {
	// Check A is a type
	level, err := c.checkIsType(context, span, path.A)
	if err != nil {
		return nil, err, true
	}

	// Check x : A
	if checkErr := c.check(context, span, path.X, path.A); checkErr != nil {
		return nil, checkErr, true
	}

	// Check y : A
	if checkErr := c.check(context, span, path.Y, path.A); checkErr != nil {
		return nil, checkErr, true
	}

	return ast.Sort{U: level}, nil, true
}

// synthPathP synthesizes the type of a dependent path type.
// PathP A x y : Sort_j where A : I → Type_j, x : A[i0/i], y : A[i1/i]
func synthPathP(c *Checker, context *tyctx.Ctx, span Span, pathp ast.PathP) (ast.Term, *TypeError, bool) {
	// A should be a type family: when we substitute i0 or i1, we get a type
	// Check A[i0/i] is a type
	aAtI0 := subst.ISubst(0, ast.I0{}, pathp.A)
	level, err := c.checkIsType(context, span, aAtI0)
	if err != nil {
		return nil, err, true
	}

	// Check A[i1/i] is a type (should be at same level)
	aAtI1 := subst.ISubst(0, ast.I1{}, pathp.A)
	_, err = c.checkIsType(context, span, aAtI1)
	if err != nil {
		return nil, err, true
	}

	// Check x : A[i0/i]
	if checkErr := c.check(context, span, pathp.X, aAtI0); checkErr != nil {
		return nil, checkErr, true
	}

	// Check y : A[i1/i]
	if checkErr := c.check(context, span, pathp.Y, aAtI1); checkErr != nil {
		return nil, checkErr, true
	}

	return ast.Sort{U: level}, nil, true
}

// synthPathLam synthesizes the type of a path abstraction.
// <i> t : PathP (λi. A) t[i0/i] t[i1/i] where Γ, i:I ⊢ t : A
func synthPathLam(c *Checker, context *tyctx.Ctx, span Span, plam ast.PathLam) (ast.Term, *TypeError, bool) {
	// Extend interval context for the body
	popIVar := c.PushIVar()
	defer popIVar()

	// Synthesize type of body with extended interval context
	bodyTy, err := c.synth(context, span, plam.Body)
	if err != nil {
		return nil, err, true
	}

	// Compute endpoints by substituting i0 and i1
	leftEnd := subst.ISubst(0, ast.I0{}, plam.Body)
	rightEnd := subst.ISubst(0, ast.I1{}, plam.Body)

	// Normalize endpoints using cubical evaluation and reification
	// This ensures we return AST terms, not values
	leftNorm := normalizeCubical(leftEnd)
	rightNorm := normalizeCubical(rightEnd)

	// Result is PathP (λi. bodyTy) leftEnd rightEnd
	// X and Y must be AST terms (normalized)
	return ast.PathP{A: bodyTy, X: leftNorm, Y: rightNorm}, nil, true
}

// normalizeCubical normalizes a term using cubical NbE.
// Returns a normalized AST term.
func normalizeCubical(t ast.Term) ast.Term {
	val := eval.EvalCubical(nil, eval.EmptyIEnv(), t)
	return eval.ReifyCubicalAt(0, 0, val)
}

// synthPathApp synthesizes the type of a path application.
// p @ r : A[r/i] where p : PathP A x y and r : I
func synthPathApp(c *Checker, context *tyctx.Ctx, span Span, papp ast.PathApp) (ast.Term, *TypeError, bool) {
	// Synthesize type of path
	pathTy, err := c.synth(context, span, papp.P)
	if err != nil {
		return nil, err, true
	}

	// Normalize to get PathP or Path
	nf := c.whnf(pathTy)

	switch pt := nf.(type) {
	case ast.PathP:
		// Check r : I
		if checkErr := c.check(context, span, papp.R, ast.Interval{}); checkErr != nil {
			return nil, checkErr, true
		}
		// Result type is A[r/i]
		return subst.ISubst(0, papp.R, pt.A), nil, true

	case ast.Path:
		// Check r : I
		if checkErr := c.check(context, span, papp.R, ast.Interval{}); checkErr != nil {
			return nil, checkErr, true
		}
		// For non-dependent path, result is just A
		return pt.A, nil, true

	default:
		return nil, errNotAPath(span, "expected Path or PathP type in path application"), true
	}
}

// synthTransport synthesizes the type of transport.
// transport A e : A[i1/i] where A : I → Type and e : A[i0/i]
func synthTransport(c *Checker, context *tyctx.Ctx, span Span, tr ast.Transport) (ast.Term, *TypeError, bool) {
	// Check A[i0/i] is a type
	aAtI0 := subst.ISubst(0, ast.I0{}, tr.A)
	_, err := c.checkIsType(context, span, aAtI0)
	if err != nil {
		return nil, err, true
	}

	// Check e : A[i0/i]
	if checkErr := c.check(context, span, tr.E, aAtI0); checkErr != nil {
		return nil, checkErr, true
	}

	// Result type is A[i1/i]
	return subst.ISubst(0, ast.I1{}, tr.A), nil, true
}

// checkExtension handles type checking for cubical terms.
// Returns (nil, true) if handled successfully, (err, true) if handled with error,
// or (nil, false) if not a cubical term.
func checkExtension(c *Checker, context *tyctx.Ctx, span Span, term ast.Term, expected ast.Term) (*TypeError, bool) {
	switch t := term.(type) {
	case ast.PathLam:
		return checkPathLam(c, context, span, t, expected)
	default:
		return nil, false
	}
}

// checkPathLam checks a path lambda against an expected PathP type.
func checkPathLam(c *Checker, context *tyctx.Ctx, span Span, plam ast.PathLam, expected ast.Term) (*TypeError, bool) {
	// Extend interval context for the body
	popIVar := c.PushIVar()
	defer popIVar()

	// Normalize expected type
	nf := c.whnf(expected)

	switch pt := nf.(type) {
	case ast.PathP:
		// Check body has type A (under interval binder)
		// The body should have type pt.A when interval variable is bound
		if checkErr := c.check(context, span, plam.Body, pt.A); checkErr != nil {
			return checkErr, true
		}

		// Check endpoints match
		// t[i0/i] should be convertible with pt.X
		leftEnd := subst.ISubst(0, ast.I0{}, plam.Body)
		if !c.conv(leftEnd, pt.X) {
			return errPathEndpointMismatch(span, "path left endpoint mismatch"), true
		}

		// t[i1/i] should be convertible with pt.Y
		rightEnd := subst.ISubst(0, ast.I1{}, plam.Body)
		if !c.conv(rightEnd, pt.Y) {
			return errPathEndpointMismatch(span, "path right endpoint mismatch"), true
		}

		return nil, true

	case ast.Path:
		// For non-dependent path, check body type and endpoints
		if checkErr := c.check(context, span, plam.Body, pt.A); checkErr != nil {
			return checkErr, true
		}

		// Check endpoints
		leftEnd := subst.ISubst(0, ast.I0{}, plam.Body)
		if !c.conv(leftEnd, pt.X) {
			return errPathEndpointMismatch(span, "path left endpoint mismatch"), true
		}

		rightEnd := subst.ISubst(0, ast.I1{}, plam.Body)
		if !c.conv(rightEnd, pt.Y) {
			return errPathEndpointMismatch(span, "path right endpoint mismatch"), true
		}

		return nil, true

	default:
		// Not a path type - fall through to synthesis
		return nil, false
	}
}

// errNotAPath creates an error for when a path type was expected.
func errNotAPath(span Span, msg string) *TypeError {
	return &TypeError{
		Span:    span,
		Kind:    ErrNotAFunction, // reuse existing kind
		Message: msg,
	}
}

// errPathEndpointMismatch creates an error for path endpoint mismatches.
func errPathEndpointMismatch(span Span, msg string) *TypeError {
	return &TypeError{
		Span:    span,
		Kind:    ErrTypeMismatch, // reuse existing kind
		Message: msg,
	}
}

// errUnboundIVar creates an error for unbound interval variables.
func errUnboundIVar(span Span, ix int) *TypeError {
	return &TypeError{
		Span:    span,
		Kind:    ErrUnboundVariable, // reuse existing kind
		Message: fmt.Sprintf("unbound interval variable %d", ix),
	}
}

// --- Face and Partial Type Checking ---

// checkFace validates a face formula.
func (c *Checker) checkFace(context *tyctx.Ctx, span Span, face ast.Face) *TypeError {
	if face == nil {
		return nil
	}
	switch f := face.(type) {
	case ast.FaceTop, ast.FaceBot:
		return nil

	case ast.FaceEq:
		if !c.CheckIVar(f.IVar) {
			return errUnboundIVar(span, f.IVar)
		}
		return nil

	case ast.FaceAnd:
		if err := c.checkFace(context, span, f.Left); err != nil {
			return err
		}
		return c.checkFace(context, span, f.Right)

	case ast.FaceOr:
		if err := c.checkFace(context, span, f.Left); err != nil {
			return err
		}
		return c.checkFace(context, span, f.Right)

	default:
		return &TypeError{
			Span:    span,
			Kind:    ErrTypeMismatch,
			Message: "invalid face formula",
		}
	}
}

// synthPartial synthesizes the type of a partial type.
// Partial φ A : Type_i where φ : Face and A : Type_i
func synthPartial(c *Checker, context *tyctx.Ctx, span Span, partial ast.Partial) (ast.Term, *TypeError, bool) {
	// Check the face formula is valid
	if err := c.checkFace(context, span, partial.Phi); err != nil {
		return nil, err, true
	}

	// Check A is a type
	level, err := c.checkIsType(context, span, partial.A)
	if err != nil {
		return nil, err, true
	}

	// Partial φ A : Type_i (same level as A)
	return ast.Sort{U: level}, nil, true
}

// synthSystem synthesizes the type of a system.
// [φ₁ ↦ t₁, ...] : Partial (φ₁ ∨ ...) A
// where each t_i has type A when φ_i holds
func synthSystem(c *Checker, context *tyctx.Ctx, span Span, sys ast.System) (ast.Term, *TypeError, bool) {
	if len(sys.Branches) == 0 {
		// Empty system: has type Partial ⊥ A for any A
		// We can't infer A without an annotation, so error
		return nil, &TypeError{
			Span:    span,
			Kind:    ErrCannotInfer,
			Message: "cannot infer type of empty system; add annotation",
		}, true
	}

	// Check and synthesize from first branch to get the type
	first := sys.Branches[0]
	if err := c.checkFace(context, span, first.Phi); err != nil {
		return nil, err, true
	}

	// Synthesize type from first branch term
	termTy, err := c.synth(context, span, first.Term)
	if err != nil {
		return nil, err, true
	}

	// Build the combined face formula
	combinedFace := first.Phi

	// Check remaining branches
	for i := 1; i < len(sys.Branches); i++ {
		br := sys.Branches[i]

		// Check face formula
		if err := c.checkFace(context, span, br.Phi); err != nil {
			return nil, err, true
		}

		// Check term has the same type
		if checkErr := c.check(context, span, br.Term, termTy); checkErr != nil {
			return nil, checkErr, true
		}

		// Add to combined face
		combinedFace = ast.FaceOr{Left: combinedFace, Right: br.Phi}
	}

	// TODO: Check agreement on overlaps (φ_i ∧ φ_j ⊢ t_i = t_j)
	// This requires evaluating under face constraints, which we defer for now

	return ast.Partial{Phi: combinedFace, A: termTy}, nil, true
}

// --- Composition Type Checking ---

// synthComp synthesizes the type of heterogeneous composition.
// comp^i A [φ ↦ u] a₀ : A[i1/i]
func synthComp(c *Checker, context *tyctx.Ctx, span Span, comp ast.Comp) (ast.Term, *TypeError, bool) {
	// Push interval variable for type family and tube
	popIVar := c.PushIVar()
	defer popIVar()

	// Check A[i0/i] is a type
	aAtI0 := subst.ISubst(0, ast.I0{}, comp.A)
	_, err := c.checkIsType(context, span, aAtI0)
	if err != nil {
		return nil, err, true
	}

	// Check the face formula (under interval binder)
	if checkErr := c.checkFace(context, span, comp.Phi); checkErr != nil {
		return nil, checkErr, true
	}

	// Check base : A[i0/i]
	if checkErr := c.check(context, span, comp.Base, aAtI0); checkErr != nil {
		return nil, checkErr, true
	}

	// TODO: Check tube has type A when φ holds
	// TODO: Check tube[i0/i] = base when φ[i0/i] holds

	// Result type is A[i1/i]
	return subst.ISubst(0, ast.I1{}, comp.A), nil, true
}

// synthHComp synthesizes the type of homogeneous composition.
// hcomp A [φ ↦ u] a₀ : A
func synthHComp(c *Checker, context *tyctx.Ctx, span Span, hcomp ast.HComp) (ast.Term, *TypeError, bool) {
	// Push interval variable for tube
	popIVar := c.PushIVar()
	defer popIVar()

	// Check A is a type
	_, err := c.checkIsType(context, span, hcomp.A)
	if err != nil {
		return nil, err, true
	}

	// Check the face formula (under interval binder)
	if checkErr := c.checkFace(context, span, hcomp.Phi); checkErr != nil {
		return nil, checkErr, true
	}

	// Check base : A
	if checkErr := c.check(context, span, hcomp.Base, hcomp.A); checkErr != nil {
		return nil, checkErr, true
	}

	// TODO: Check tube has type A when φ holds
	// TODO: Check tube[i0/i] = base when φ[i0/i] holds

	// Result type is A (same as input, since A is constant)
	return hcomp.A, nil, true
}

// synthFill synthesizes the type of fill.
// fill^i A [φ ↦ u] a₀ : A (under interval binder j, result is A[j/i])
func synthFill(c *Checker, context *tyctx.Ctx, span Span, fill ast.Fill) (ast.Term, *TypeError, bool) {
	// Push interval variable for type family and tube
	popIVar := c.PushIVar()
	defer popIVar()

	// Check A[i0/i] is a type
	aAtI0 := subst.ISubst(0, ast.I0{}, fill.A)
	_, err := c.checkIsType(context, span, aAtI0)
	if err != nil {
		return nil, err, true
	}

	// Check the face formula (under interval binder)
	if checkErr := c.checkFace(context, span, fill.Phi); checkErr != nil {
		return nil, checkErr, true
	}

	// Check base : A[i0/i]
	if checkErr := c.check(context, span, fill.Base, aAtI0); checkErr != nil {
		return nil, checkErr, true
	}

	// TODO: Check tube and agreement

	// Fill produces a value in A (the type family itself, since fill @ j has type A[j/i])
	// For now, return the type family - caller should apply to interval
	return fill.A, nil, true
}

// synthGlue synthesizes the type of Glue A [φ ↦ (T, e)].
// Formation rule:
//
//	Γ ⊢ A : Type_i    Γ, φ ⊢ T : Type_i    Γ, φ ⊢ e : Equiv T A
//	────────────────────────────────────────────────────────────
//	Γ ⊢ Glue A [φ ↦ (T, e)] : Type_i
func synthGlue(c *Checker, context *tyctx.Ctx, span Span, glue ast.Glue) (ast.Term, *TypeError, bool) {
	// Check A is a type
	sortA, err := c.checkIsType(context, span, glue.A)
	if err != nil {
		return nil, err, true
	}

	// Check each branch
	for _, br := range glue.System {
		// Check face formula
		if checkErr := c.checkFace(context, span, br.Phi); checkErr != nil {
			return nil, checkErr, true
		}

		// Check T is a type (at same level as A)
		sortT, err := c.checkIsType(context, span, br.T)
		if err != nil {
			return nil, err, true
		}

		// Universe levels should match
		if sortA != sortT {
			return nil, errTypeMismatch(span, ast.Sort{U: sortA}, ast.Sort{U: sortT}), true
		}

		// TODO: Check e : Equiv T A
		// For now, we just check that Equiv is a type
		_, err = c.Synth(context, span, br.Equiv)
		if err != nil {
			return nil, err, true
		}
	}

	// Result type is same universe as A
	return ast.Sort{U: sortA}, nil, true
}

// synthGlueElem synthesizes the type of glue [φ ↦ t] a.
// Typing rule:
//
//	Γ ⊢ a : A    Γ, φ ⊢ t : T    Γ, φ ⊢ e.fst t = a : A
//	────────────────────────────────────────────────────
//	Γ ⊢ glue [φ ↦ t] a : Glue A [φ ↦ (T, e)]
func synthGlueElem(c *Checker, context *tyctx.Ctx, span Span, elem ast.GlueElem) (ast.Term, *TypeError, bool) {
	// Synthesize base type
	baseType, err := c.Synth(context, span, elem.Base)
	if err != nil {
		return nil, err, true
	}

	// Build Glue type from the branches
	glueBranches := make([]ast.GlueBranch, len(elem.System))
	for i, br := range elem.System {
		// Check face formula
		if checkErr := c.checkFace(context, span, br.Phi); checkErr != nil {
			return nil, checkErr, true
		}

		// Synthesize term type
		termType, err := c.Synth(context, span, br.Term)
		if err != nil {
			return nil, err, true
		}

		// For now, use the term type as T and leave Equiv as placeholder
		// A full implementation would require tracking the equivalence
		glueBranches[i] = ast.GlueBranch{
			Phi:   br.Phi,
			T:     termType,
			Equiv: ast.Global{Name: "idEquiv"}, // Placeholder
		}
	}

	// Result is Glue type
	return ast.Glue{A: baseType, System: glueBranches}, nil, true
}

// synthUnglue synthesizes the type of unglue g.
// Typing rule:
//
//	Γ ⊢ g : Glue A [φ ↦ (T, e)]
//	───────────────────────────
//	Γ ⊢ unglue g : A
func synthUnglue(c *Checker, context *tyctx.Ctx, span Span, unglue ast.Unglue) (ast.Term, *TypeError, bool) {
	// Synthesize type of g
	gType, err := c.Synth(context, span, unglue.G)
	if err != nil {
		return nil, err, true
	}

	// If g has Glue type, return the base type A
	if glueType, ok := gType.(ast.Glue); ok {
		return glueType.A, nil, true
	}

	// If we have a stored Glue type annotation, use it
	if unglue.Ty != nil {
		if glueType, ok := unglue.Ty.(ast.Glue); ok {
			return glueType.A, nil, true
		}
	}

	// Otherwise stuck - can't determine base type
	// Return the synthesized type for now
	return gType, nil, true
}

// synthUA synthesizes the type of ua e : Path Type A B.
// Typing rule:
//
//	Γ ⊢ A : Type_i    Γ ⊢ B : Type_i    Γ ⊢ e : Equiv A B
//	─────────────────────────────────────────────────────
//	Γ ⊢ ua e : Path Type_i A B
func synthUA(c *Checker, context *tyctx.Ctx, span Span, ua ast.UA) (ast.Term, *TypeError, bool) {
	// Check A is a type
	sortA, err := c.checkIsType(context, span, ua.A)
	if err != nil {
		return nil, err, true
	}

	// Check B is a type at same level
	sortB, err := c.checkIsType(context, span, ua.B)
	if err != nil {
		return nil, err, true
	}

	// Universe levels should match
	if sortA != sortB {
		return nil, errTypeMismatch(span, ast.Sort{U: sortA}, ast.Sort{U: sortB}), true
	}

	// Check equivalence term (for now just verify it type-checks)
	_, err = c.Synth(context, span, ua.Equiv)
	if err != nil {
		return nil, err, true
	}

	// Result type: Path Type_i A B
	return ast.Path{
		A: ast.Sort{U: sortA},
		X: ua.A,
		Y: ua.B,
	}, nil, true
}

// synthUABeta synthesizes the type of ua-β e a.
// This represents the computation: transport (ua e) a = e.fst a.
// Typing rule:
//
//	Γ ⊢ e : Equiv A B    Γ ⊢ a : A
//	──────────────────────────────
//	Γ ⊢ ua-β e a : B
func synthUABeta(c *Checker, context *tyctx.Ctx, span Span, uab ast.UABeta) (ast.Term, *TypeError, bool) {
	// Synthesize type of equivalence
	equivType, err := c.Synth(context, span, uab.Equiv)
	if err != nil {
		return nil, err, true
	}

	// Check argument type-checks
	_, err = c.Synth(context, span, uab.Arg)
	if err != nil {
		return nil, err, true
	}

	// The result type is B (the target type of the equivalence)
	// For now, return the synthesized type of the equivalence
	// A full implementation would extract B from Equiv A B
	return equivType, nil, true
}
