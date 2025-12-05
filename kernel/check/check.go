package check

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/core"
	"github.com/watchthelight/HypergraphGo/internal/eval"
	tyctx "github.com/watchthelight/HypergraphGo/kernel/ctx"
)

// Checker performs bidirectional type checking.
type Checker struct {
	globals  *GlobalEnv
	convOpts core.ConvOptions
}

// NewChecker creates a new type checker with the given global environment.
func NewChecker(globals *GlobalEnv) *Checker {
	if globals == nil {
		globals = NewGlobalEnv()
	}
	return &Checker{
		globals:  globals,
		convOpts: core.ConvOptions{EnableEta: false},
	}
}

// NewCheckerWithEta creates a checker with eta-equality enabled.
func NewCheckerWithEta(globals *GlobalEnv) *Checker {
	if globals == nil {
		globals = NewGlobalEnv()
	}
	return &Checker{
		globals:  globals,
		convOpts: core.ConvOptions{EnableEta: true},
	}
}

// Globals returns the global environment.
func (c *Checker) Globals() *GlobalEnv {
	return c.globals
}

// Synth synthesizes (infers) the type of a term.
// Returns the inferred type and nil error on success.
// If ctx is nil, an empty context is used.
func (c *Checker) Synth(ctx *tyctx.Ctx, span Span, term ast.Term) (ast.Term, *TypeError) {
	if ctx == nil {
		ctx = &tyctx.Ctx{}
	}
	return c.synth(ctx, span, term)
}

// Check verifies that a term has the expected type.
// Returns nil on success.
// If ctx is nil, an empty context is used.
func (c *Checker) Check(ctx *tyctx.Ctx, span Span, term ast.Term, expected ast.Term) *TypeError {
	if ctx == nil {
		ctx = &tyctx.Ctx{}
	}
	return c.check(ctx, span, term, expected)
}

// CheckIsType verifies that a term is a well-formed type.
// Returns the universe level and nil error on success.
// If ctx is nil, an empty context is used.
func (c *Checker) CheckIsType(ctx *tyctx.Ctx, span Span, term ast.Term) (ast.Level, *TypeError) {
	if ctx == nil {
		ctx = &tyctx.Ctx{}
	}
	return c.checkIsType(ctx, span, term)
}

// InferAndCheck is a convenience that synthesizes a type and checks it against expected.
// If ctx is nil, an empty context is used.
func (c *Checker) InferAndCheck(ctx *tyctx.Ctx, span Span, term ast.Term, expected ast.Term) *TypeError {
	if ctx == nil {
		ctx = &tyctx.Ctx{}
	}
	inferred, err := c.synth(ctx, span, term)
	if err != nil {
		return err
	}
	if !c.conv(inferred, expected) {
		return errTypeMismatch(span, expected, inferred)
	}
	return nil
}

// conv checks definitional equality using the core conversion checker.
func (c *Checker) conv(t, u ast.Term) bool {
	return core.Conv(core.NewEnv(), t, u, c.convOpts)
}

// whnf reduces a term to weak head normal form.
func (c *Checker) whnf(t ast.Term) ast.Term {
	// For now, full normalization. Could optimize to WHNF later.
	return eval.EvalNBE(t)
}

// ensurePi checks that a term is a Pi type, normalizing only if needed.
func (c *Checker) ensurePi(span Span, ty ast.Term) (ast.Pi, *TypeError) {
	// Fast path: already syntactically a Pi (preserves de Bruijn indices)
	if pi, ok := ty.(ast.Pi); ok {
		return pi, nil
	}
	// Normalize and check
	nf := c.whnf(ty)
	if pi, ok := nf.(ast.Pi); ok {
		return pi, nil
	}
	return ast.Pi{}, errNotAFunction(span, ty)
}

// ensureSigma checks that a term is a Sigma type, normalizing only if needed.
func (c *Checker) ensureSigma(span Span, ty ast.Term) (ast.Sigma, *TypeError) {
	// Fast path: already syntactically a Sigma (preserves de Bruijn indices)
	if sigma, ok := ty.(ast.Sigma); ok {
		return sigma, nil
	}
	// Normalize and check
	nf := c.whnf(ty)
	if sigma, ok := nf.(ast.Sigma); ok {
		return sigma, nil
	}
	return ast.Sigma{}, errNotAPair(span, ty)
}

// ensureSort normalizes a term and checks it's a Sort.
func (c *Checker) ensureSort(span Span, ty ast.Term) (ast.Sort, *TypeError) {
	nf := c.whnf(ty)
	if sort, ok := nf.(ast.Sort); ok {
		return sort, nil
	}
	return ast.Sort{}, errNotAType(span, ty)
}

// maxLevel returns the maximum of two universe levels.
func maxLevel(a, b ast.Level) ast.Level {
	if a > b {
		return a
	}
	return b
}
