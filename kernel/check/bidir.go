package check

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
	tyctx "github.com/watchthelight/HypergraphGo/kernel/ctx"
	"github.com/watchthelight/HypergraphGo/kernel/subst"
)

// synth implements type synthesis for all term constructors.
func (c *Checker) synth(context *tyctx.Ctx, span Span, term ast.Term) (ast.Term, *TypeError) {
	if term == nil {
		return nil, errCannotInfer(span, term)
	}

	switch t := term.(type) {
	case ast.Var:
		return c.synthVar(context, span, t)

	case ast.Sort:
		return c.synthSort(span, t)

	case ast.Global:
		return c.synthGlobal(span, t)

	case ast.Pi:
		return c.synthPi(context, span, t)

	case ast.Sigma:
		return c.synthSigma(context, span, t)

	case ast.Lam:
		return c.synthLam(context, span, t)

	case ast.App:
		return c.synthApp(context, span, t)

	case ast.Pair:
		// Pairs cannot be synthesized without type annotation
		return nil, errCannotInfer(span, term)

	case ast.Fst:
		return c.synthFst(context, span, t)

	case ast.Snd:
		return c.synthSnd(context, span, t)

	case ast.Let:
		return c.synthLet(context, span, t)

	default:
		return nil, errCannotInfer(span, term)
	}
}

// synthVar synthesizes the type of a variable by context lookup.
func (c *Checker) synthVar(context *tyctx.Ctx, span Span, v ast.Var) (ast.Term, *TypeError) {
	ty, ok := context.LookupVar(v.Ix)
	if !ok {
		return nil, errUnboundVar(span, v.Ix)
	}
	// Shift the type to account for the binders between the variable and its use
	return subst.Shift(v.Ix+1, 0, ty), nil
}

// synthSort synthesizes the type of a universe: Sort U : Sort (U+1).
func (c *Checker) synthSort(span Span, s ast.Sort) (ast.Term, *TypeError) {
	return ast.Sort{U: s.U + 1}, nil
}

// synthGlobal synthesizes the type of a global constant.
func (c *Checker) synthGlobal(span Span, g ast.Global) (ast.Term, *TypeError) {
	ty := c.globals.LookupType(g.Name)
	if ty == nil {
		return nil, errUnknownGlobal(span, g.Name)
	}
	return ty, nil
}

// synthPi synthesizes the type of a Pi type.
// Pi (x : A) . B : Sort (max U V) where A : Sort U and B : Sort V under x:A.
func (c *Checker) synthPi(context *tyctx.Ctx, span Span, pi ast.Pi) (ast.Term, *TypeError) {
	// Check A is a type
	levelA, err := c.checkIsType(context, span, pi.A)
	if err != nil {
		return nil, err
	}

	// Extend context with x : A
	context.Extend(pi.Binder, pi.A)
	defer func() { *context = context.Drop() }()

	// Check B is a type under x : A
	levelB, err := c.checkIsType(context, span, pi.B)
	if err != nil {
		return nil, err
	}

	return ast.Sort{U: maxLevel(levelA, levelB)}, nil
}

// synthSigma synthesizes the type of a Sigma type.
// Sigma (x : A) . B : Sort (max U V) where A : Sort U and B : Sort V under x:A.
func (c *Checker) synthSigma(context *tyctx.Ctx, span Span, sigma ast.Sigma) (ast.Term, *TypeError) {
	// Check A is a type
	levelA, err := c.checkIsType(context, span, sigma.A)
	if err != nil {
		return nil, err
	}

	// Extend context with x : A
	context.Extend(sigma.Binder, sigma.A)
	defer func() { *context = context.Drop() }()

	// Check B is a type under x : A
	levelB, err := c.checkIsType(context, span, sigma.B)
	if err != nil {
		return nil, err
	}

	return ast.Sort{U: maxLevel(levelA, levelB)}, nil
}

// synthLam synthesizes the type of an annotated lambda.
// Only annotated lambdas can be synthesized.
func (c *Checker) synthLam(context *tyctx.Ctx, span Span, lam ast.Lam) (ast.Term, *TypeError) {
	if lam.Ann == nil {
		return nil, errCannotInfer(span, lam)
	}

	// Check annotation is a type
	_, err := c.checkIsType(context, span, lam.Ann)
	if err != nil {
		return nil, err
	}

	// Extend context with x : Ann
	context.Extend(lam.Binder, lam.Ann)
	defer func() { *context = context.Drop() }()

	// Synthesize type of body
	bodyTy, err := c.synth(context, span, lam.Body)
	if err != nil {
		return nil, err
	}

	// Result is Pi (x : Ann) . bodyTy
	return ast.Pi{Binder: lam.Binder, A: lam.Ann, B: bodyTy}, nil
}

// synthApp synthesizes the type of a function application.
func (c *Checker) synthApp(context *tyctx.Ctx, span Span, app ast.App) (ast.Term, *TypeError) {
	// Synthesize type of function
	funTy, err := c.synth(context, span, app.T)
	if err != nil {
		return nil, err
	}

	// Ensure it's a Pi type
	pi, err := c.ensurePi(span, funTy)
	if err != nil {
		return nil, err
	}

	// Check argument against domain
	if checkErr := c.check(context, span, app.U, pi.A); checkErr != nil {
		return nil, checkErr
	}

	// Substitute argument into codomain: B[u/x]
	return subst.Subst(0, app.U, pi.B), nil
}

// synthFst synthesizes the type of a first projection.
func (c *Checker) synthFst(context *tyctx.Ctx, span Span, fst ast.Fst) (ast.Term, *TypeError) {
	// Synthesize type of pair
	pairTy, err := c.synth(context, span, fst.P)
	if err != nil {
		return nil, err
	}

	// Ensure it's a Sigma type
	sigma, err := c.ensureSigma(span, pairTy)
	if err != nil {
		return nil, err
	}

	// Return the first component type
	return sigma.A, nil
}

// synthSnd synthesizes the type of a second projection.
func (c *Checker) synthSnd(context *tyctx.Ctx, span Span, snd ast.Snd) (ast.Term, *TypeError) {
	// Synthesize type of pair
	pairTy, err := c.synth(context, span, snd.P)
	if err != nil {
		return nil, err
	}

	// Ensure it's a Sigma type
	sigma, err := c.ensureSigma(span, pairTy)
	if err != nil {
		return nil, err
	}

	// Substitute fst p into the second component type: B[fst p/x]
	return subst.Subst(0, ast.Fst{P: snd.P}, sigma.B), nil
}

// synthLet synthesizes the type of a let expression.
func (c *Checker) synthLet(context *tyctx.Ctx, span Span, let ast.Let) (ast.Term, *TypeError) {
	var valTy ast.Term

	if let.Ann != nil {
		// Check annotation is a type
		_, err := c.checkIsType(context, span, let.Ann)
		if err != nil {
			return nil, err
		}

		// Check value against annotation
		if checkErr := c.check(context, span, let.Val, let.Ann); checkErr != nil {
			return nil, checkErr
		}
		valTy = let.Ann
	} else {
		// Synthesize type of value
		ty, err := c.synth(context, span, let.Val)
		if err != nil {
			return nil, err
		}
		valTy = ty
	}

	// Extend context with x : valTy
	context.Extend(let.Binder, valTy)
	defer func() { *context = context.Drop() }()

	// Synthesize type of body
	bodyTy, err := c.synth(context, span, let.Body)
	if err != nil {
		return nil, err
	}

	// Substitute value into body type: bodyTy[val/x]
	return subst.Subst(0, let.Val, bodyTy), nil
}

// check implements type checking mode.
func (c *Checker) check(context *tyctx.Ctx, span Span, term ast.Term, expected ast.Term) *TypeError {
	if term == nil {
		return errCannotInfer(span, term)
	}

	switch t := term.(type) {
	case ast.Lam:
		// Unannotated lambda checks against Pi type
		if t.Ann == nil {
			return c.checkLam(context, span, t, expected)
		}
		// Annotated lambda: synthesize and compare
		return c.checkBySynth(context, span, term, expected)

	case ast.Pair:
		// Pair checks against Sigma type
		return c.checkPair(context, span, t, expected)

	default:
		// Default: synthesize and compare
		return c.checkBySynth(context, span, term, expected)
	}
}

// checkBySynth checks a term by synthesizing its type and comparing.
func (c *Checker) checkBySynth(context *tyctx.Ctx, span Span, term ast.Term, expected ast.Term) *TypeError {
	inferred, err := c.synth(context, span, term)
	if err != nil {
		return err
	}

	if !c.conv(inferred, expected) {
		return errTypeMismatch(span, expected, inferred)
	}
	return nil
}

// checkLam checks an unannotated lambda against a Pi type.
func (c *Checker) checkLam(context *tyctx.Ctx, span Span, lam ast.Lam, expected ast.Term) *TypeError {
	// Ensure expected type is a Pi
	pi, err := c.ensurePi(span, expected)
	if err != nil {
		return err
	}

	// Extend context with x : A (domain of Pi)
	context.Extend(lam.Binder, pi.A)
	defer func() { *context = context.Drop() }()

	// Check body against codomain B
	return c.check(context, span, lam.Body, pi.B)
}

// checkPair checks a pair against a Sigma type.
func (c *Checker) checkPair(context *tyctx.Ctx, span Span, pair ast.Pair, expected ast.Term) *TypeError {
	// Ensure expected type is a Sigma
	sigma, err := c.ensureSigma(span, expected)
	if err != nil {
		return err
	}

	// Check first component against A
	if checkErr := c.check(context, span, pair.Fst, sigma.A); checkErr != nil {
		return checkErr
	}

	// Substitute first component into B: B[fst/x]
	sndTy := subst.Subst(0, pair.Fst, sigma.B)

	// Check second component against B[fst/x]
	return c.check(context, span, pair.Snd, sndTy)
}

// checkIsType checks that a term is a valid type and returns its universe level.
func (c *Checker) checkIsType(context *tyctx.Ctx, span Span, term ast.Term) (ast.Level, *TypeError) {
	ty, err := c.synth(context, span, term)
	if err != nil {
		return 0, err
	}

	sort, sortErr := c.ensureSort(span, ty)
	if sortErr != nil {
		return 0, sortErr
	}

	return sort.U, nil
}
