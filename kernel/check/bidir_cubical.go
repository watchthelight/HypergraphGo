//go:build cubical

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
