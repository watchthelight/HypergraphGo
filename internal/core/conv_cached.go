package core

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
)

// This file provides cached versions of conversion checking functions.
// The caching is scoped to avoid memory leaks - each conversion check
// uses its own cache that is discarded after the check completes.

// ConvCached reports whether t and u are definitionally equal under env,
// using memoization for improved performance on repeated subterms.
//
// This is a drop-in replacement for Conv that provides the same semantics
// but with better performance for terms with shared structure.
func ConvCached(env *Env, t, u ast.Term, opts ConvOptions) bool {
	if env == nil {
		env = NewEnv()
	}

	// Create a scoped cache for this conversion check
	cache := eval.NewDefaultCache()

	// Evaluate both terms using cached NbE
	valT := eval.EvalCached(env.evalEnv, t, cache)
	valU := eval.EvalCached(env.evalEnv, u, cache)

	// Reify to normal forms with caching
	nfT := eval.ReifyCached(valT, cache)
	nfU := eval.ReifyCached(valU, cache)

	// Compare normal forms with η-equality if enabled
	if opts.EnableEta {
		return etaEqual(nfT, nfU)
	}

	// Structural comparison of normal forms
	return AlphaEq(nfT, nfU)
}

// ConvContext provides a reusable context for multiple conversion checks.
// This is more efficient than ConvCached when doing many related checks,
// as the cache persists across checks and can benefit from shared subterms.
//
// Usage:
//
//	ctx := NewConvContext()
//	defer ctx.Reset() // optional: clear cache if context will be reused
//	for _, pair := range termsToCheck {
//	    if !ctx.Conv(env, pair.left, pair.right, opts) {
//	        return false
//	    }
//	}
type ConvContext struct {
	cache *eval.Cache
}

// NewConvContext creates a new conversion context with a fresh cache.
func NewConvContext() *ConvContext {
	return &ConvContext{
		cache: eval.NewDefaultCache(),
	}
}

// NewConvContextWithSize creates a new conversion context with a custom cache size.
func NewConvContextWithSize(maxSize int) *ConvContext {
	return &ConvContext{
		cache: eval.NewCache(maxSize),
	}
}

// Conv checks definitional equality using the context's cache.
func (ctx *ConvContext) Conv(env *Env, t, u ast.Term, opts ConvOptions) bool {
	if env == nil {
		env = NewEnv()
	}

	// Evaluate both terms using the shared cache
	valT := eval.EvalCached(env.evalEnv, t, ctx.cache)
	valU := eval.EvalCached(env.evalEnv, u, ctx.cache)

	// Reify to normal forms
	nfT := eval.ReifyCached(valT, ctx.cache)
	nfU := eval.ReifyCached(valU, ctx.cache)

	// Compare normal forms
	if opts.EnableEta {
		return etaEqual(nfT, nfU)
	}
	return AlphaEq(nfT, nfU)
}

// Reset clears the context's cache for reuse.
// Call this when starting a new batch of unrelated conversion checks.
func (ctx *ConvContext) Reset() {
	ctx.cache.Reset()
}

// Stats returns cache statistics for debugging/profiling.
func (ctx *ConvContext) Stats() map[string]int64 {
	return ctx.cache.Stats()
}

// ConvAllCached checks if all pairs of terms are convertible.
// Uses a shared cache for efficiency.
func ConvAllCached(env *Env, pairs [][2]ast.Term, opts ConvOptions) bool {
	ctx := NewConvContext()
	for _, pair := range pairs {
		if !ctx.Conv(env, pair[0], pair[1], opts) {
			return false
		}
	}
	return true
}
