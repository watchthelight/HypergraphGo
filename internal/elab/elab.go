package elab

import (
	"fmt"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
	"github.com/watchthelight/HypergraphGo/kernel/subst"
)

// normalize normalizes a term using NbE.
func normalize(t ast.Term) ast.Term {
	return eval.EvalNBE(t)
}

// equal checks if two terms are definitionally equal (alpha-equal after normalization).
func equal(a, b ast.Term) bool {
	return eval.AlphaEq(normalize(a), normalize(b))
}

// ElabError represents an elaboration error.
type ElabError struct {
	Span    Span
	Message string
	Context *ElabCtx
}

func (e *ElabError) Error() string {
	if e.Span.File != "" {
		return fmt.Sprintf("%s:%d:%d: %s", e.Span.File, e.Span.Line, e.Span.Col, e.Message)
	}
	return e.Message
}

func errSpan(span Span, format string, args ...any) *ElabError {
	return &ElabError{Span: span, Message: fmt.Sprintf(format, args...)}
}

func errNoSpan(format string, args ...any) *ElabError {
	return &ElabError{Message: fmt.Sprintf(format, args...)}
}

// Elaborator transforms surface syntax into core terms.
type Elaborator struct{}

// NewElaborator creates a new elaborator.
func NewElaborator() *Elaborator {
	return &Elaborator{}
}

// Elaborate elaborates a surface term, inferring its type.
// Returns the elaborated core term and its type.
func (e *Elaborator) Elaborate(ctx *ElabCtx, s STerm) (term ast.Term, ty ast.Term, err *ElabError) {
	return e.synth(ctx, s)
}

// ElaborateCheck elaborates a surface term against an expected type.
// Returns the elaborated core term.
func (e *Elaborator) ElaborateCheck(ctx *ElabCtx, s STerm, expected ast.Term) (term ast.Term, err *ElabError) {
	return e.check(ctx, s, expected)
}

// synth synthesizes the type of a surface term.
// Returns the elaborated core term and its type.
func (e *Elaborator) synth(ctx *ElabCtx, s STerm) (ast.Term, ast.Term, *ElabError) {
	if s == nil {
		return nil, nil, errNoSpan("cannot synthesize type of nil term")
	}

	span := s.Span()

	switch t := s.(type) {
	case *SVar:
		return e.synthVar(ctx, span, t)

	case *SGlobal:
		return e.synthGlobal(ctx, span, t)

	case *SType:
		return e.synthType(span, t)

	case *SPi:
		return e.synthPi(ctx, span, t)

	case *SArrow:
		return e.synthArrow(ctx, span, t)

	case *SLam:
		return e.synthLam(ctx, span, t)

	case *SApp:
		return e.synthApp(ctx, span, t)

	case *SSigma:
		return e.synthSigma(ctx, span, t)

	case *SProd:
		return e.synthProd(ctx, span, t)

	case *SPair:
		// Pairs cannot be synthesized without type annotation
		return nil, nil, errSpan(span, "cannot infer type of pair, add type annotation")

	case *SFst:
		return e.synthFst(ctx, span, t)

	case *SSnd:
		return e.synthSnd(ctx, span, t)

	case *SLet:
		return e.synthLet(ctx, span, t)

	case *SHole:
		return e.synthHole(ctx, span, t)

	case *SId:
		return e.synthId(ctx, span, t)

	case *SRefl:
		return e.synthRefl(ctx, span, t)

	case *SJ:
		return e.synthJ(ctx, span, t)

	case *SPath:
		return e.synthPath(ctx, span, t)

	case *SPathP:
		return e.synthPathP(ctx, span, t)

	case *SPathLam:
		return e.synthPathLam(ctx, span, t)

	case *SPathApp:
		return e.synthPathApp(ctx, span, t)

	case *SI0:
		return ast.I0{}, ast.Interval{}, nil

	case *SI1:
		return ast.I1{}, ast.Interval{}, nil

	case *STransport:
		return e.synthTransport(ctx, span, t)

	case *SIndApp:
		return e.synthIndApp(ctx, span, t)

	case *SCtorApp:
		return e.synthCtorApp(ctx, span, t)

	case *SElim:
		return e.synthElim(ctx, span, t)

	default:
		return nil, nil, errSpan(span, "cannot synthesize type of %T", s)
	}
}

// synthVar synthesizes the type of a variable reference.
func (e *Elaborator) synthVar(ctx *ElabCtx, span Span, v *SVar) (ast.Term, ast.Term, *ElabError) {
	// First try local variables
	ix, ty, _, ok := ctx.LookupName(v.Name)
	if ok {
		// Shift the type to account for binders
		shiftedTy := subst.Shift(ix+1, 0, ty)
		return ast.Var{Ix: ix}, shiftedTy, nil
	}

	// Try interval variables
	iix, iok := ctx.LookupIName(v.Name)
	if iok {
		return ast.IVar{Ix: iix}, ast.Interval{}, nil
	}

	// Try globals
	if ctx.Globals != nil {
		gty, _, gok := ctx.Globals.LookupGlobal(v.Name)
		if gok {
			return ast.Global{Name: v.Name}, gty, nil
		}
	}

	return nil, nil, errSpan(span, "unbound variable: %s", v.Name)
}

// synthGlobal synthesizes the type of a global reference.
func (e *Elaborator) synthGlobal(ctx *ElabCtx, span Span, g *SGlobal) (ast.Term, ast.Term, *ElabError) {
	if ctx.Globals == nil {
		return nil, nil, errSpan(span, "no global environment for: %s", g.Name)
	}

	ty, _, ok := ctx.Globals.LookupGlobal(g.Name)
	if !ok {
		return nil, nil, errSpan(span, "unknown global: %s", g.Name)
	}

	return ast.Global{Name: g.Name}, ty, nil
}

// synthType synthesizes the type of a universe.
func (e *Elaborator) synthType(span Span, t *SType) (ast.Term, ast.Term, *ElabError) {
	return ast.Sort{U: ast.Level(t.Level)}, ast.Sort{U: ast.Level(t.Level + 1)}, nil
}

// synthPi synthesizes the type of a Pi type.
func (e *Elaborator) synthPi(ctx *ElabCtx, span Span, pi *SPi) (ast.Term, ast.Term, *ElabError) {
	// Elaborate domain and check it's a type
	domTerm, domTy, err := e.synth(ctx, pi.Dom)
	if err != nil {
		return nil, nil, err
	}

	levelA, sortErr := e.ensureSort(span, domTy)
	if sortErr != nil {
		return nil, nil, sortErr
	}

	// Extend context with the binder
	extCtx := ctx.Extend(pi.Binder, domTerm, pi.Icity)

	// Elaborate codomain and check it's a type
	codTerm, codTy, err := e.synth(extCtx, pi.Cod)
	if err != nil {
		return nil, nil, err
	}

	levelB, sortErr := e.ensureSort(span, codTy)
	if sortErr != nil {
		return nil, nil, sortErr
	}

	return ast.Pi{Binder: pi.Binder, A: domTerm, B: codTerm, Implicit: pi.Icity == Implicit},
		ast.Sort{U: maxLevel(levelA, levelB)},
		nil
}

// synthArrow synthesizes the type of a non-dependent function type.
func (e *Elaborator) synthArrow(ctx *ElabCtx, span Span, arr *SArrow) (ast.Term, ast.Term, *ElabError) {
	// Convert to Pi with unused binder
	pi := &SPi{
		base:   arr.base,
		Binder: "_",
		Icity:  Explicit,
		Dom:    arr.Dom,
		Cod:    arr.Cod,
	}
	return e.synthPi(ctx, span, pi)
}

// synthLam synthesizes the type of a lambda.
func (e *Elaborator) synthLam(ctx *ElabCtx, span Span, lam *SLam) (ast.Term, ast.Term, *ElabError) {
	if lam.Ann == nil {
		return nil, nil, errSpan(span, "cannot infer type of unannotated lambda, add type annotation")
	}

	// Elaborate annotation and check it's a type
	annTerm, annTy, err := e.synth(ctx, lam.Ann)
	if err != nil {
		return nil, nil, err
	}

	_, sortErr := e.ensureSort(span, annTy)
	if sortErr != nil {
		return nil, nil, sortErr
	}

	// Extend context with the binder
	extCtx := ctx.Extend(lam.Binder, annTerm, lam.Icity)

	// Synthesize type of body
	bodyTerm, bodyTy, err := e.synth(extCtx, lam.Body)
	if err != nil {
		return nil, nil, err
	}

	isImplicit := lam.Icity == Implicit
	return ast.Lam{Binder: lam.Binder, Ann: annTerm, Body: bodyTerm, Implicit: isImplicit},
		ast.Pi{Binder: lam.Binder, A: annTerm, B: bodyTy, Implicit: isImplicit},
		nil
}

// synthApp synthesizes the type of a function application.
func (e *Elaborator) synthApp(ctx *ElabCtx, span Span, app *SApp) (ast.Term, ast.Term, *ElabError) {
	// Synthesize type of function
	fnTerm, fnTy, err := e.synth(ctx, app.Fn)
	if err != nil {
		return nil, nil, err
	}

	// Insert implicit arguments if needed
	fnTerm, fnTy, err = e.insertImplicits(ctx, span, fnTerm, fnTy, app.Icity)
	if err != nil {
		return nil, nil, err
	}

	// Ensure function type is a Pi
	pi, piErr := e.ensurePi(span, fnTy)
	if piErr != nil {
		return nil, nil, piErr
	}

	// Check/elaborate argument against domain
	argTerm, argErr := e.check(ctx, app.Arg, pi.A)
	if argErr != nil {
		return nil, nil, argErr
	}

	// Result type is codomain with argument substituted
	resultTy := subst.Subst(0, argTerm, pi.B)

	return ast.App{T: fnTerm, U: argTerm, Implicit: false}, resultTy, nil
}

// insertImplicits inserts metavariables for implicit arguments.
// When the function type has implicit Pi arguments and the application is explicit,
// we insert metavariables for each implicit argument until we reach an explicit Pi.
func (e *Elaborator) insertImplicits(ctx *ElabCtx, span Span, fn ast.Term, fnTy ast.Term, argIcity Icity) (ast.Term, ast.Term, *ElabError) {
	// Normalize the function type
	fnTy = normalize(fnTy)

	// If the argument is implicit, don't insert implicits - the user is providing it explicitly
	if argIcity == Implicit {
		return fn, fnTy, nil
	}

	// Insert metavariables for implicit arguments
	for {
		pi, ok := fnTy.(ast.Pi)
		if !ok {
			// Not a Pi type, stop
			break
		}
		if !pi.Implicit {
			// Explicit Pi, stop - user will provide this argument
			break
		}

		// Insert implicit argument as a metavariable
		meta := ctx.FreshMeta(pi.A, span)
		fn = ast.App{T: fn, U: meta, Implicit: true}
		fnTy = normalize(subst.Subst(0, meta, pi.B))
	}

	return fn, fnTy, nil
}

// synthSigma synthesizes the type of a Sigma type.
func (e *Elaborator) synthSigma(ctx *ElabCtx, span Span, sigma *SSigma) (ast.Term, ast.Term, *ElabError) {
	// Elaborate first component type
	fstTerm, fstTy, err := e.synth(ctx, sigma.Fst)
	if err != nil {
		return nil, nil, err
	}

	levelA, sortErr := e.ensureSort(span, fstTy)
	if sortErr != nil {
		return nil, nil, sortErr
	}

	// Extend context
	extCtx := ctx.Extend(sigma.Binder, fstTerm, Explicit)

	// Elaborate second component type
	sndTerm, sndTy, err := e.synth(extCtx, sigma.Snd)
	if err != nil {
		return nil, nil, err
	}

	levelB, sortErr := e.ensureSort(span, sndTy)
	if sortErr != nil {
		return nil, nil, sortErr
	}

	return ast.Sigma{Binder: sigma.Binder, A: fstTerm, B: sndTerm},
		ast.Sort{U: maxLevel(levelA, levelB)},
		nil
}

// synthProd synthesizes the type of a non-dependent product type.
func (e *Elaborator) synthProd(ctx *ElabCtx, span Span, prod *SProd) (ast.Term, ast.Term, *ElabError) {
	sigma := &SSigma{
		base:   prod.base,
		Binder: "_",
		Fst:    prod.Fst,
		Snd:    prod.Snd,
	}
	return e.synthSigma(ctx, span, sigma)
}

// synthFst synthesizes the type of a first projection.
func (e *Elaborator) synthFst(ctx *ElabCtx, span Span, fst *SFst) (ast.Term, ast.Term, *ElabError) {
	pairTerm, pairTy, err := e.synth(ctx, fst.Pair)
	if err != nil {
		return nil, nil, err
	}

	sigma, sigmaErr := e.ensureSigma(span, pairTy)
	if sigmaErr != nil {
		return nil, nil, sigmaErr
	}

	return ast.Fst{P: pairTerm}, sigma.A, nil
}

// synthSnd synthesizes the type of a second projection.
func (e *Elaborator) synthSnd(ctx *ElabCtx, span Span, snd *SSnd) (ast.Term, ast.Term, *ElabError) {
	pairTerm, pairTy, err := e.synth(ctx, snd.Pair)
	if err != nil {
		return nil, nil, err
	}

	sigma, sigmaErr := e.ensureSigma(span, pairTy)
	if sigmaErr != nil {
		return nil, nil, sigmaErr
	}

	// Type is B[fst p / x]
	resultTy := subst.Subst(0, ast.Fst{P: pairTerm}, sigma.B)

	return ast.Snd{P: pairTerm}, resultTy, nil
}

// synthLet synthesizes the type of a let expression.
func (e *Elaborator) synthLet(ctx *ElabCtx, span Span, let *SLet) (ast.Term, ast.Term, *ElabError) {
	var valTerm ast.Term
	var valTy ast.Term

	if let.Ann != nil {
		// Elaborate annotation
		annTerm, annTy, err := e.synth(ctx, let.Ann)
		if err != nil {
			return nil, nil, err
		}

		_, sortErr := e.ensureSort(span, annTy)
		if sortErr != nil {
			return nil, nil, sortErr
		}

		// Check value against annotation
		valTerm, err = e.check(ctx, let.Val, annTerm)
		if err != nil {
			return nil, nil, err
		}
		valTy = annTerm
	} else {
		// Synthesize value type
		var err *ElabError
		valTerm, valTy, err = e.synth(ctx, let.Val)
		if err != nil {
			return nil, nil, err
		}
	}

	// Extend context with definition
	extCtx := ctx.ExtendDef(let.Binder, valTy, valTerm)

	// Elaborate body
	bodyTerm, bodyTy, err := e.synth(extCtx, let.Body)
	if err != nil {
		return nil, nil, err
	}

	// Result type has value substituted
	resultTy := subst.Subst(0, valTerm, bodyTy)

	return ast.Let{Binder: let.Binder, Ann: valTy, Val: valTerm, Body: bodyTerm},
		resultTy,
		nil
}

// synthHole synthesizes a fresh metavariable.
func (e *Elaborator) synthHole(ctx *ElabCtx, span Span, hole *SHole) (ast.Term, ast.Term, *ElabError) {
	// We need a type for the metavariable, but we don't know it yet.
	// Create a metavariable for the type first.
	// The type of a type is a Sort, but we don't know which level.
	// For now, use Type0 (can be refined with universe polymorphism later).
	typeMeta := ctx.FreshMeta(ast.Sort{U: 0}, span)

	// Create the hole metavariable with this type
	var holeMeta ast.Term
	if hole.Name != "" {
		id := ctx.FreshNamed(typeMeta, span, hole.Name)
		holeMeta = MkMeta(id)
	} else {
		holeMeta = ctx.FreshMeta(typeMeta, span)
	}

	return holeMeta, typeMeta, nil
}

// synthId synthesizes the type of an identity type.
func (e *Elaborator) synthId(ctx *ElabCtx, span Span, id *SId) (ast.Term, ast.Term, *ElabError) {
	// Elaborate A and check it's a type
	aTerm, aTy, err := e.synth(ctx, id.A)
	if err != nil {
		return nil, nil, err
	}

	level, sortErr := e.ensureSort(span, aTy)
	if sortErr != nil {
		return nil, nil, sortErr
	}

	// Check x : A
	xTerm, err := e.check(ctx, id.X, aTerm)
	if err != nil {
		return nil, nil, err
	}

	// Check y : A
	yTerm, err := e.check(ctx, id.Y, aTerm)
	if err != nil {
		return nil, nil, err
	}

	return ast.Id{A: aTerm, X: xTerm, Y: yTerm}, ast.Sort{U: level}, nil
}

// synthRefl synthesizes the type of a reflexivity proof.
func (e *Elaborator) synthRefl(ctx *ElabCtx, span Span, refl *SRefl) (ast.Term, ast.Term, *ElabError) {
	if refl.A == nil || refl.X == nil {
		return nil, nil, errSpan(span, "cannot infer type of partial refl, provide A and x")
	}

	// Elaborate A and check it's a type
	aTerm, aTy, err := e.synth(ctx, refl.A)
	if err != nil {
		return nil, nil, err
	}

	_, sortErr := e.ensureSort(span, aTy)
	if sortErr != nil {
		return nil, nil, sortErr
	}

	// Check x : A
	xTerm, err := e.check(ctx, refl.X, aTerm)
	if err != nil {
		return nil, nil, err
	}

	return ast.Refl{A: aTerm, X: xTerm}, ast.Id{A: aTerm, X: xTerm, Y: xTerm}, nil
}

// synthJ synthesizes the type of J elimination.
func (e *Elaborator) synthJ(ctx *ElabCtx, span Span, j *SJ) (ast.Term, ast.Term, *ElabError) {
	// Elaborate A
	aTerm, aTy, err := e.synth(ctx, j.A)
	if err != nil {
		return nil, nil, err
	}

	levelA, sortErr := e.ensureSort(span, aTy)
	if sortErr != nil {
		return nil, nil, sortErr
	}

	// Check x : A
	xTerm, err := e.check(ctx, j.X, aTerm)
	if err != nil {
		return nil, nil, err
	}

	// Check y : A
	yTerm, err := e.check(ctx, j.Y, aTerm)
	if err != nil {
		return nil, nil, err
	}

	// Build motive type: (y : A) -> Id A x y -> Type
	motiveType := e.mkJMotiveType(aTerm, xTerm, levelA)

	// Check C : motive type
	cTerm, err := e.check(ctx, j.C, motiveType)
	if err != nil {
		return nil, nil, err
	}

	// Check d : C x (refl A x)
	reflXX := ast.Refl{A: aTerm, X: xTerm}
	dType := ast.MkApps(cTerm, xTerm, reflXX)
	dTerm, err := e.check(ctx, j.D, dType)
	if err != nil {
		return nil, nil, err
	}

	// Check p : Id A x y
	idType := ast.Id{A: aTerm, X: xTerm, Y: yTerm}
	pTerm, err := e.check(ctx, j.P, idType)
	if err != nil {
		return nil, nil, err
	}

	// Result type: C y p
	resultTy := ast.MkApps(cTerm, yTerm, pTerm)

	return ast.J{A: aTerm, C: cTerm, D: dTerm, X: xTerm, Y: yTerm, P: pTerm},
		resultTy,
		nil
}

// mkJMotiveType builds the motive type for J: (y : A) -> Id A x y -> Type
func (e *Elaborator) mkJMotiveType(a, x ast.Term, level ast.Level) ast.Term {
	aShifted := subst.Shift(1, 0, a)
	xShifted := subst.Shift(1, 0, x)

	return ast.Pi{
		Binder: "y",
		A:      a,
		B: ast.Pi{
			Binder: "p",
			A:      ast.Id{A: aShifted, X: xShifted, Y: ast.Var{Ix: 0}},
			B:      ast.Sort{U: level},
		},
	}
}

// synthPath synthesizes the type of a path type.
func (e *Elaborator) synthPath(ctx *ElabCtx, span Span, path *SPath) (ast.Term, ast.Term, *ElabError) {
	// Elaborate A
	aTerm, aTy, err := e.synth(ctx, path.A)
	if err != nil {
		return nil, nil, err
	}

	level, sortErr := e.ensureSort(span, aTy)
	if sortErr != nil {
		return nil, nil, sortErr
	}

	// Check endpoints
	xTerm, err := e.check(ctx, path.X, aTerm)
	if err != nil {
		return nil, nil, err
	}

	yTerm, err := e.check(ctx, path.Y, aTerm)
	if err != nil {
		return nil, nil, err
	}

	return ast.Path{A: aTerm, X: xTerm, Y: yTerm}, ast.Sort{U: level}, nil
}

// synthPathP synthesizes the type of a dependent path type.
func (e *Elaborator) synthPathP(ctx *ElabCtx, span Span, pathP *SPathP) (ast.Term, ast.Term, *ElabError) {
	// Elaborate A (should be I -> Type)
	aTerm, _, err := e.synth(ctx, pathP.A)
	if err != nil {
		return nil, nil, err
	}

	// For now, just elaborate the endpoints without full checking
	xTerm, _, err := e.synth(ctx, pathP.X)
	if err != nil {
		return nil, nil, err
	}

	yTerm, _, err := e.synth(ctx, pathP.Y)
	if err != nil {
		return nil, nil, err
	}

	// Assume result is a type (simplified for now)
	return ast.PathP{A: aTerm, X: xTerm, Y: yTerm}, ast.Sort{U: 0}, nil
}

// synthPathLam synthesizes the type of a path abstraction.
func (e *Elaborator) synthPathLam(ctx *ElabCtx, span Span, plam *SPathLam) (ast.Term, ast.Term, *ElabError) {
	// Extend context with interval variable
	extCtx := ctx.ExtendI(plam.Binder)

	// Synthesize body
	bodyTerm, bodyTy, err := e.synth(extCtx, plam.Body)
	if err != nil {
		return nil, nil, err
	}

	// Compute endpoints by substituting i0 and i1 for the interval variable
	// The interval variable has de Bruijn index 0 in the body
	xEndpoint := subst.Subst(0, ast.I0{}, bodyTerm)
	yEndpoint := subst.Subst(0, ast.I1{}, bodyTerm)

	// Build path type: PathP (λi. bodyTy) x y
	// where x = bodyTerm[i0/i] and y = bodyTerm[i1/i]
	return ast.PathLam{Binder: plam.Binder, Body: bodyTerm},
		ast.PathP{A: ast.Lam{Binder: plam.Binder, Body: bodyTy}, X: xEndpoint, Y: yEndpoint},
		nil
}

// synthPathApp synthesizes the type of a path application.
func (e *Elaborator) synthPathApp(ctx *ElabCtx, span Span, papp *SPathApp) (ast.Term, ast.Term, *ElabError) {
	pathTerm, pathTy, err := e.synth(ctx, papp.Path)
	if err != nil {
		return nil, nil, err
	}

	argTerm, _, err := e.synth(ctx, papp.Arg)
	if err != nil {
		return nil, nil, err
	}

	// Get the result type by looking at the path type
	var resultTy ast.Term
	switch pt := pathTy.(type) {
	case ast.Path:
		resultTy = pt.A
	case ast.PathP:
		// Apply the type family to the interval argument
		resultTy = ast.App{T: pt.A, U: argTerm, Implicit: false}
	default:
		return nil, nil, errSpan(span, "expected path type, got %T", pathTy)
	}

	return ast.PathApp{P: pathTerm, R: argTerm}, resultTy, nil
}

// synthTransport synthesizes the type of transport.
func (e *Elaborator) synthTransport(ctx *ElabCtx, span Span, tr *STransport) (ast.Term, ast.Term, *ElabError) {
	// Elaborate type family
	aTerm, _, err := e.synth(ctx, tr.A)
	if err != nil {
		return nil, nil, err
	}

	// Elaborate element
	eTerm, _, err := e.synth(ctx, tr.E)
	if err != nil {
		return nil, nil, err
	}

	// Result type is A applied to i1 (simplified)
	resultTy := ast.App{T: aTerm, U: ast.I1{}, Implicit: false}

	return ast.Transport{A: aTerm, E: eTerm}, resultTy, nil
}

// synthIndApp synthesizes the type of an inductive type application.
// Example: (Nat) or (List Nat) where List is a parameterized inductive.
func (e *Elaborator) synthIndApp(ctx *ElabCtx, span Span, ind *SIndApp) (ast.Term, ast.Term, *ElabError) {
	if ctx.Globals == nil {
		return nil, nil, errSpan(span, "no global environment for inductive: %s", ind.Name)
	}

	// Look up the inductive type
	info, ok := ctx.Globals.LookupInductive(ind.Name)
	if !ok {
		return nil, nil, errSpan(span, "unknown inductive type: %s", ind.Name)
	}

	// Elaborate all arguments
	args := make([]ast.Term, 0, len(ind.Args))
	for _, sarg := range ind.Args {
		argTerm, _, err := e.synth(ctx, sarg)
		if err != nil {
			return nil, nil, err
		}
		args = append(args, argTerm)
	}

	// Build the inductive type application: (Global IndName) arg1 arg2 ...
	result := ast.Term(ast.Global{Name: ind.Name})
	for _, arg := range args {
		result = ast.App{T: result, U: arg, Implicit: false}
	}

	// The type of an inductive type is its declared type applied to arguments
	// For a fully applied inductive, this is typically a Sort
	resultTy := info.Type
	for _, arg := range args {
		if pi, ok := resultTy.(ast.Pi); ok {
			resultTy = subst.Subst(0, arg, pi.B)
		}
	}

	return result, resultTy, nil
}

// synthCtorApp synthesizes the type of a constructor application.
// Example: (Nat.Z) or (Nat.S n) where Z and S are constructors of Nat.
func (e *Elaborator) synthCtorApp(ctx *ElabCtx, span Span, ctor *SCtorApp) (ast.Term, ast.Term, *ElabError) {
	if ctx.Globals == nil {
		return nil, nil, errSpan(span, "no global environment for constructor: %s.%s", ctor.Ind, ctor.Ctor)
	}

	// Look up the constructor - try fully qualified name first
	fullName := ctor.Ind + "_" + ctor.Ctor
	info, ok := ctx.Globals.LookupConstructor(fullName)
	if !ok {
		// Try just the constructor name
		info, ok = ctx.Globals.LookupConstructor(ctor.Ctor)
		if !ok {
			return nil, nil, errSpan(span, "unknown constructor: %s.%s", ctor.Ind, ctor.Ctor)
		}
	}

	// Elaborate all arguments
	args := make([]ast.Term, 0, len(ctor.Args))
	for _, sarg := range ctor.Args {
		argTerm, _, err := e.synth(ctx, sarg)
		if err != nil {
			return nil, nil, err
		}
		args = append(args, argTerm)
	}

	// Build the constructor application: (Global CtorName) arg1 arg2 ...
	result := ast.Term(ast.Global{Name: info.Name})
	for _, arg := range args {
		result = ast.App{T: result, U: arg, Implicit: false}
	}

	// The type is the constructor's type with arguments substituted
	resultTy := info.Type
	for _, arg := range args {
		if pi, ok := resultTy.(ast.Pi); ok {
			resultTy = subst.Subst(0, arg, pi.B)
		}
	}

	return result, resultTy, nil
}

// synthElim synthesizes the type of an eliminator (recursor) application.
// Example: (Nat_elim motive baseCase stepCase target)
func (e *Elaborator) synthElim(ctx *ElabCtx, span Span, elim *SElim) (ast.Term, ast.Term, *ElabError) {
	if ctx.Globals == nil {
		return nil, nil, errSpan(span, "no global environment for eliminator: %s", elim.Name)
	}

	// Look up the eliminator as a global definition
	elimTy, _, ok := ctx.Globals.LookupGlobal(elim.Name)
	if !ok {
		return nil, nil, errSpan(span, "unknown eliminator: %s", elim.Name)
	}

	// Elaborate the motive
	motiveTerm, _, err := e.synth(ctx, elim.Motive)
	if err != nil {
		return nil, nil, err
	}

	// Elaborate all methods
	methods := make([]ast.Term, 0, len(elim.Methods))
	for _, smethod := range elim.Methods {
		methodTerm, _, err := e.synth(ctx, smethod)
		if err != nil {
			return nil, nil, err
		}
		methods = append(methods, methodTerm)
	}

	// Elaborate the target
	targetTerm, _, err := e.synth(ctx, elim.Target)
	if err != nil {
		return nil, nil, err
	}

	// Build the eliminator application: (Global ElimName) motive method1 method2 ... target
	result := ast.Term(ast.Global{Name: elim.Name})
	result = ast.App{T: result, U: motiveTerm, Implicit: false}
	for _, method := range methods {
		result = ast.App{T: result, U: method, Implicit: false}
	}
	result = ast.App{T: result, U: targetTerm, Implicit: false}

	// Compute the result type by applying the motive to the target
	// For a well-typed eliminator call, the result is (motive target)
	resultTy := elimTy
	// Apply type to each argument to get final result type
	allArgs := append([]ast.Term{motiveTerm}, methods...)
	allArgs = append(allArgs, targetTerm)
	for _, arg := range allArgs {
		if pi, ok := resultTy.(ast.Pi); ok {
			resultTy = subst.Subst(0, arg, pi.B)
		}
	}

	return result, resultTy, nil
}

// check checks a surface term against an expected type.
func (e *Elaborator) check(ctx *ElabCtx, s STerm, expected ast.Term) (ast.Term, *ElabError) {
	if s == nil {
		return nil, errNoSpan("cannot check nil term")
	}

	span := s.Span()

	// Normalize expected type
	expected = normalize(expected)

	// Insert implicit lambdas: if expected is {x : A} -> B and term is not an implicit lambda,
	// wrap the term in an implicit lambda
	if pi, ok := expected.(ast.Pi); ok && pi.Implicit {
		// Check if term is already an implicit lambda
		if lam, isLam := s.(*SLam); isLam && lam.Icity == Implicit {
			// Already an implicit lambda, check normally
			return e.checkLam(ctx, span, lam, expected)
		}
		// Not an implicit lambda - insert one
		// Extend context with the implicit binder
		extCtx := ctx.Extend(pi.Binder, pi.A, Implicit)
		// Check the term against the codomain (the term can reference the implicit var)
		bodyTerm, err := e.check(extCtx, s, pi.B)
		if err != nil {
			return nil, err
		}
		// Wrap in an implicit lambda
		return ast.Lam{Binder: pi.Binder, Ann: pi.A, Body: bodyTerm, Implicit: true}, nil
	}

	switch t := s.(type) {
	case *SLam:
		// Unannotated lambda checks against Pi type
		if t.Ann == nil {
			return e.checkLam(ctx, span, t, expected)
		}
		// Annotated lambda: synthesize and compare
		return e.checkBySynth(ctx, span, s, expected)

	case *SPair:
		// Pair checks against Sigma type
		return e.checkPair(ctx, span, t, expected)

	case *SHole:
		// Hole with known type: create metavariable with that type
		return e.checkHole(ctx, span, t, expected)

	default:
		// Default: synthesize and compare
		return e.checkBySynth(ctx, span, s, expected)
	}
}

// checkBySynth checks by synthesizing and comparing types.
func (e *Elaborator) checkBySynth(ctx *ElabCtx, span Span, s STerm, expected ast.Term) (ast.Term, *ElabError) {
	term, inferred, err := e.synth(ctx, s)
	if err != nil {
		return nil, err
	}

	// Check equality (both terms are normalized and compared)
	if !equal(inferred, expected) {
		return nil, errSpan(span, "type mismatch: expected %v, got %v", expected, inferred)
	}

	return term, nil
}

// checkLam checks an unannotated lambda against a Pi type.
func (e *Elaborator) checkLam(ctx *ElabCtx, span Span, lam *SLam, expected ast.Term) (ast.Term, *ElabError) {
	pi, piErr := e.ensurePi(span, expected)
	if piErr != nil {
		return nil, piErr
	}

	// Check icity compatibility
	lamIsImplicit := lam.Icity == Implicit
	if pi.Implicit != lamIsImplicit {
		if pi.Implicit {
			return nil, errSpan(span, "expected implicit lambda {%s}, got explicit lambda", pi.Binder)
		}
		return nil, errSpan(span, "expected explicit lambda (%s), got implicit lambda", pi.Binder)
	}

	// Extend context with binder
	extCtx := ctx.Extend(lam.Binder, pi.A, lam.Icity)

	// Check body against codomain
	bodyTerm, err := e.check(extCtx, lam.Body, pi.B)
	if err != nil {
		return nil, err
	}

	return ast.Lam{Binder: lam.Binder, Ann: pi.A, Body: bodyTerm, Implicit: pi.Implicit}, nil
}

// checkPair checks a pair against a Sigma type.
func (e *Elaborator) checkPair(ctx *ElabCtx, span Span, pair *SPair, expected ast.Term) (ast.Term, *ElabError) {
	sigma, sigmaErr := e.ensureSigma(span, expected)
	if sigmaErr != nil {
		return nil, sigmaErr
	}

	// Check first component
	fstTerm, err := e.check(ctx, pair.Fst, sigma.A)
	if err != nil {
		return nil, err
	}

	// Substitute first component into second component type
	sndTy := subst.Subst(0, fstTerm, sigma.B)

	// Check second component
	sndTerm, err := e.check(ctx, pair.Snd, sndTy)
	if err != nil {
		return nil, err
	}

	return ast.Pair{Fst: fstTerm, Snd: sndTerm}, nil
}

// checkHole creates a metavariable with the expected type.
func (e *Elaborator) checkHole(ctx *ElabCtx, span Span, hole *SHole, expected ast.Term) (ast.Term, *ElabError) {
	if hole.Name != "" {
		id := ctx.FreshNamed(expected, span, hole.Name)
		return MkMeta(id), nil
	}
	return ctx.FreshMeta(expected, span), nil
}

// Helper functions

func (e *Elaborator) ensureSort(span Span, ty ast.Term) (ast.Level, *ElabError) {
	switch t := ty.(type) {
	case ast.Sort:
		return t.U, nil
	default:
		return 0, errSpan(span, "expected Type, got %T", ty)
	}
}

func (e *Elaborator) ensurePi(span Span, ty ast.Term) (ast.Pi, *ElabError) {
	switch t := ty.(type) {
	case ast.Pi:
		return t, nil
	default:
		return ast.Pi{}, errSpan(span, "expected function type, got %T", ty)
	}
}

func (e *Elaborator) ensureSigma(span Span, ty ast.Term) (ast.Sigma, *ElabError) {
	switch t := ty.(type) {
	case ast.Sigma:
		return t, nil
	default:
		return ast.Sigma{}, errSpan(span, "expected pair type, got %T", ty)
	}
}

func maxLevel(a, b ast.Level) ast.Level {
	if a > b {
		return a
	}
	return b
}
