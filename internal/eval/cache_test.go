package eval

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// Test helper constructors
func tLam(binder string, body ast.Term) ast.Term {
	return ast.Lam{Binder: binder, Body: body}
}

func tApp(f, a ast.Term) ast.Term {
	return ast.App{T: f, U: a}
}

func tGlob(name string) ast.Term {
	return ast.Global{Name: name}
}

func tVr(ix int) ast.Term {
	return ast.Var{Ix: ix}
}

func tPair(fst, snd ast.Term) ast.Term {
	return ast.Pair{Fst: fst, Snd: snd}
}

func tFst(p ast.Term) ast.Term {
	return ast.Fst{P: p}
}

// TestCacheCorrectness verifies that cached evaluation produces the same results
// as non-cached evaluation for a variety of terms.
func TestCacheCorrectness(t *testing.T) {
	tests := []struct {
		name string
		term ast.Term
	}{
		{"identity", tLam("x", tVr(0))},
		{"K combinator", tLam("x", tLam("y", tVr(1)))},
		{"beta reduction", tApp(tLam("x", tVr(0)), tGlob("y"))},
		{"K applied", ast.MkApps(tLam("x", tLam("y", tVr(1))), tGlob("a"), tGlob("b"))},
		{"pair", tPair(tGlob("a"), tGlob("b"))},
		{"fst", tFst(tPair(tGlob("a"), tGlob("b")))},
		{"nested app", tApp(tApp(tGlob("f"), tGlob("x")), tGlob("y"))},
		{"deep lambda", tLam("a", tLam("b", tLam("c", tVr(2))))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Non-cached result
			expected := EvalNBE(tt.term)

			// Cached result
			cache := NewDefaultCache()
			got := EvalNBECached(tt.term, cache)

			// Compare using alpha equality
			if !AlphaEq(expected, got) {
				t.Errorf("cached result differs:\n  expected: %v\n  got: %v", expected, got)
			}
		})
	}
}

// TestCacheHits verifies that the cache actually hits.
func TestCacheHits(t *testing.T) {
	cache := NewDefaultCache()
	env := &Env{Bindings: nil}

	term := ast.MkApps(tLam("x", tLam("y", tVr(1))), tGlob("a"), tGlob("b"))

	// Evaluate same term multiple times
	for i := 0; i < 10; i++ {
		_ = EvalCached(env, term, cache)
	}

	stats := cache.Stats()
	if stats["eval_hits"] == 0 {
		t.Errorf("expected cache hits, got none: %v", stats)
	}
}

// TestCacheSeparation verifies that different environments don't share cache entries.
func TestCacheSeparation(t *testing.T) {
	cache := NewDefaultCache()

	// Same term in different environments should produce different results
	term := tVr(0)

	env1 := &Env{Bindings: []Value{VSort{Level: 1}}}
	env2 := &Env{Bindings: []Value{VSort{Level: 2}}}

	result1 := EvalCached(env1, term, cache)
	result2 := EvalCached(env2, term, cache)

	// Results should be different
	s1, ok1 := result1.(VSort)
	s2, ok2 := result2.(VSort)

	if !ok1 || !ok2 || s1.Level == s2.Level {
		t.Errorf("cache incorrectly shared between environments: %v vs %v", result1, result2)
	}
}

// TestCacheReset verifies that Reset clears the cache.
func TestCacheReset(t *testing.T) {
	cache := NewDefaultCache()
	env := &Env{Bindings: nil}

	term := tGlob("test")
	_ = EvalCached(env, term, cache)

	stats := cache.Stats()
	if stats["eval_size"] == 0 {
		t.Error("expected cache to have entries")
	}

	cache.Reset()

	stats = cache.Stats()
	if stats["eval_size"] != 0 {
		t.Errorf("expected cache to be empty after reset, got size %d", stats["eval_size"])
	}
}

// TestCacheNilSafe verifies that nil cache is handled safely.
func TestCacheNilSafe(t *testing.T) {
	term := tApp(tLam("x", tVr(0)), tGlob("y"))

	// Should not panic
	result := EvalNBECached(term, nil)

	expected := EvalNBE(term)
	if !AlphaEq(expected, result) {
		t.Errorf("nil cache gave different result: %v vs %v", expected, result)
	}
}

// TestCacheMaxSize verifies that cache respects size limits.
func TestCacheMaxSize(t *testing.T) {
	cache := NewCache(10) // Small cache
	env := &Env{Bindings: nil}

	// Add many different terms
	for i := 0; i < 100; i++ {
		term := ast.Global{Name: string(rune('a' + i%26))}
		_ = EvalCached(env, term, cache)
	}

	stats := cache.Stats()
	if stats["eval_size"] > 10 {
		t.Errorf("cache exceeded max size: got %d, max 10", stats["eval_size"])
	}
}

// TestWithCache verifies the WithCache helper.
func TestWithCache(t *testing.T) {
	term := ast.MkApps(tLam("x", tVr(0)), tGlob("y"))
	expected := EvalNBE(term)

	got := WithCache(func(c *Cache) ast.Term {
		return EvalNBECached(term, c)
	})

	if !AlphaEq(expected, got) {
		t.Errorf("WithCache gave different result")
	}
}

// TestNormalizeWithCache verifies NormalizeWithCache matches regular normalization.
func TestNormalizeWithCache(t *testing.T) {
	terms := []ast.Term{
		tApp(tLam("x", tVr(0)), tGlob("y")),
		ast.MkApps(tLam("x", tLam("y", tVr(1))), tGlob("a"), tGlob("b")),
		tFst(tPair(tGlob("a"), tGlob("b"))),
	}

	for _, term := range terms {
		expected := EvalNBE(term)
		got := NormalizeWithCache(term)

		if !AlphaEq(expected, got) {
			t.Errorf("NormalizeWithCache(%v) = %v, want %v", term, got, expected)
		}
	}
}

// BenchmarkCacheOverhead measures the overhead of cache management.
func BenchmarkCacheOverhead(b *testing.B) {
	term := tApp(tLam("x", tVr(0)), tGlob("y"))
	env := &Env{Bindings: nil}

	b.Run("no cache", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = Eval(env, term)
		}
	})

	b.Run("with cache", func(b *testing.B) {
		cache := NewDefaultCache()
		for i := 0; i < b.N; i++ {
			_ = EvalCached(env, term, cache)
		}
	})
}
