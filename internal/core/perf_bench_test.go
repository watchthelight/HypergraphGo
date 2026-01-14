package core

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
)

// Performance benchmarks for identifying optimization opportunities.
// These benchmarks focus on scenarios where caching would be beneficial.

// === Deep Term Benchmarks ===

// makeChurchNumeral creates a Church-encoded natural number.
// church(n) = λf. λx. f^n x
func makeChurchNumeral(n int) ast.Term {
	// Build f^n x
	var body ast.Term = ast.Var{Ix: 0} // x
	for i := 0; i < n; i++ {
		body = ast.App{T: ast.Var{Ix: 1}, U: body} // f body
	}
	return ast.Lam{Binder: "f", Body: ast.Lam{Binder: "x", Body: body}}
}

// BenchmarkConv_DeepChurch benchmarks conversion on Church numerals.
func BenchmarkConv_DeepChurch10(b *testing.B) {
	env := NewEnv()
	term := makeChurchNumeral(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Conv(env, term, term, ConvOptions{})
	}
}

func BenchmarkConv_DeepChurch50(b *testing.B) {
	env := NewEnv()
	term := makeChurchNumeral(50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Conv(env, term, term, ConvOptions{})
	}
}

func BenchmarkConv_DeepChurch100(b *testing.B) {
	env := NewEnv()
	term := makeChurchNumeral(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Conv(env, term, term, ConvOptions{})
	}
}

// === Repeated Conversion Benchmarks (shows caching benefit) ===

// BenchmarkConv_RepeatedSame benchmarks repeated conversion of identical terms.
// This demonstrates potential benefit of caching.
func BenchmarkConv_RepeatedSame100(b *testing.B) {
	env := NewEnv()
	// K combinator
	k := lam("x", lam("y", vr(1)))
	term1 := ast.MkApps(k, glob("a"), glob("b"))
	term2 := glob("a")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Repeat same conversion check 100 times
		for j := 0; j < 100; j++ {
			_ = Conv(env, term1, term2, ConvOptions{})
		}
	}
}

// BenchmarkConv_SharedSubterms benchmarks terms with shared structure.
func BenchmarkConv_SharedSubterms(b *testing.B) {
	env := NewEnv()
	// Build a term with shared subexpressions
	shared := lam("z", app(vr(0), vr(0)))
	term := app(lam("f", pair(app(vr(0), glob("a")), app(vr(0), glob("b")))), shared)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Conv(env, term, term, ConvOptions{})
	}
}

// === Normalization Benchmarks ===

// BenchmarkNormalize_DeepNested benchmarks normalization of deeply nested lambdas.
func BenchmarkNormalize_DeepNested10(b *testing.B) {
	// λx1.λx2...λx10. x1
	var term ast.Term = ast.Var{Ix: 9}
	for i := 0; i < 10; i++ {
		term = ast.Lam{Binder: "_", Body: term}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eval.EvalNBE(term)
	}
}

func BenchmarkNormalize_DeepNested50(b *testing.B) {
	var term ast.Term = ast.Var{Ix: 49}
	for i := 0; i < 50; i++ {
		term = ast.Lam{Binder: "_", Body: term}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eval.EvalNBE(term)
	}
}

// BenchmarkNormalize_PairChain benchmarks normalization of chained pair projections.
func BenchmarkNormalize_PairChain10(b *testing.B) {
	// fst (fst (fst ... (pair (pair ... a b) c) ...))
	var term ast.Term = pair(glob("a"), glob("b"))
	for i := 0; i < 10; i++ {
		term = pair(term, glob("c"))
	}
	for i := 0; i < 10; i++ {
		term = fst(term)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eval.EvalNBE(term)
	}
}

// BenchmarkNormalize_BetaChain benchmarks chained beta reductions.
func BenchmarkNormalize_BetaChain20(b *testing.B) {
	// (λx. (λy. (λz. ... z) c) b) a
	var term ast.Term = vr(0)
	for i := 0; i < 20; i++ {
		term = app(lam("_", term), glob("arg"))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eval.EvalNBE(term)
	}
}

// === Pi/Sigma Type Benchmarks ===

// BenchmarkConv_DependentType benchmarks conversion of dependent types.
func BenchmarkConv_DependentType(b *testing.B) {
	env := NewEnv()
	// (x : A) -> (y : B x) -> C x y
	typeA := sort(0)
	typeB := pi("x", typeA, sort(0))
	typeC := pi("x", typeA, pi("y", app(glob("B"), vr(0)), sort(0)))

	term := pi("x", typeA, pi("y", app(glob("B"), vr(0)), app(app(glob("C"), vr(1)), vr(0))))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Conv(env, term, term, ConvOptions{})
		_ = typeB
		_ = typeC
	}
}

// === Alpha-Equality Benchmarks ===

// BenchmarkAlphaEq_Large benchmarks alpha-equality on large terms.
func BenchmarkAlphaEq_LargeTerm(b *testing.B) {
	// Build a large term with many subterms
	var term ast.Term = glob("base")
	for i := 0; i < 50; i++ {
		term = app(lam("_", term), glob("arg"))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AlphaEq(term, term)
	}
}

// BenchmarkAlphaEq_Mismatch benchmarks alpha-equality that fails early vs late.
func BenchmarkAlphaEq_MismatchEarly(b *testing.B) {
	term1 := lam("x", app(glob("f"), vr(0)))
	term2 := lam("x", app(glob("g"), vr(0))) // Different head

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AlphaEq(term1, term2)
	}
}

func BenchmarkAlphaEq_MismatchLate(b *testing.B) {
	// Build deep terms that differ only at the end
	var term1, term2 ast.Term = glob("same"), glob("same")
	for i := 0; i < 20; i++ {
		term1 = lam("_", term1)
		term2 = lam("_", term2)
	}
	// Now make them different at leaf
	term1 = lam("_", glob("a"))
	term2 = lam("_", glob("b"))
	for i := 0; i < 20; i++ {
		term1 = lam("_", term1)
		term2 = lam("_", term2)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = AlphaEq(term1, term2)
	}
}

// === Eval-heavy Benchmarks ===

// BenchmarkEval_RepeatedTerm benchmarks repeated evaluation of the same term.
func BenchmarkEval_RepeatedTerm100(b *testing.B) {
	env := &eval.Env{Bindings: nil}
	term := ast.MkApps(lam("x", lam("y", vr(1))), glob("a"), glob("b"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			_ = eval.Eval(env, term)
		}
	}
}

// === Cached vs Non-Cached Comparison Benchmarks ===

// BenchmarkConvCached_RepeatedSame100 benchmarks cached repeated conversion.
func BenchmarkConvCached_RepeatedSame100(b *testing.B) {
	env := NewEnv()
	k := lam("x", lam("y", vr(1)))
	term1 := ast.MkApps(k, glob("a"), glob("b"))
	term2 := glob("a")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := NewConvContext()
		for j := 0; j < 100; j++ {
			_ = ctx.Conv(env, term1, term2, ConvOptions{})
		}
	}
}

// BenchmarkConvCached_DeepChurch10 benchmarks cached conversion on Church numerals.
func BenchmarkConvCached_DeepChurch10(b *testing.B) {
	env := NewEnv()
	term := makeChurchNumeral(10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConvCached(env, term, term, ConvOptions{})
	}
}

func BenchmarkConvCached_DeepChurch50(b *testing.B) {
	env := NewEnv()
	term := makeChurchNumeral(50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConvCached(env, term, term, ConvOptions{})
	}
}

func BenchmarkConvCached_DeepChurch100(b *testing.B) {
	env := NewEnv()
	term := makeChurchNumeral(100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConvCached(env, term, term, ConvOptions{})
	}
}

// BenchmarkEvalCached_RepeatedTerm100 benchmarks cached repeated evaluation.
func BenchmarkEvalCached_RepeatedTerm100(b *testing.B) {
	env := &eval.Env{Bindings: nil}
	term := ast.MkApps(lam("x", lam("y", vr(1))), glob("a"), glob("b"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache := eval.NewDefaultCache()
		for j := 0; j < 100; j++ {
			_ = eval.EvalCached(env, term, cache)
		}
	}
}

// BenchmarkNormalizeCached_DeepNested benchmarks cached deep nested normalization.
func BenchmarkNormalizeCached_DeepNested50(b *testing.B) {
	var term ast.Term = ast.Var{Ix: 49}
	for i := 0; i < 50; i++ {
		term = ast.Lam{Binder: "_", Body: term}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eval.NormalizeWithCache(term)
	}
}

// BenchmarkConvContext_MultipleChecks benchmarks using ConvContext for multiple checks.
func BenchmarkConvContext_MultipleChecks(b *testing.B) {
	env := NewEnv()
	// Create several term pairs
	pairs := make([][2]ast.Term, 50)
	for i := 0; i < 50; i++ {
		church := makeChurchNumeral(i % 10)
		pairs[i] = [2]ast.Term{church, church}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConvAllCached(env, pairs, ConvOptions{})
	}
}

// BenchmarkConv_MultipleChecksNonCached benchmarks non-cached multiple checks.
func BenchmarkConv_MultipleChecksNonCached(b *testing.B) {
	env := NewEnv()
	pairs := make([][2]ast.Term, 50)
	for i := 0; i < 50; i++ {
		church := makeChurchNumeral(i % 10)
		pairs[i] = [2]ast.Term{church, church}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pair := range pairs {
			_ = Conv(env, pair[0], pair[1], ConvOptions{})
		}
	}
}
