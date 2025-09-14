package core

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// BenchmarkConv_Simple benchmarks conversion checking on small terms.
func BenchmarkConv_Simple(b *testing.B) {
	env := NewEnv()

	// Simple beta reduction: (\x. x) y ≡ y
	id := lam("x", vr(0))
	term1 := app(id, glob("y"))
	term2 := glob("y")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Conv(env, term1, term2, ConvOptions{})
	}
}

// BenchmarkConv_Beta benchmarks beta reduction normalization.
func BenchmarkConv_Beta(b *testing.B) {
	env := NewEnv()

	// K combinator: (\x. \y. x) a b ≡ a
	k := lam("x", lam("y", vr(1)))
	term1 := ast.MkApps(k, glob("a"), glob("b"))
	term2 := glob("a")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Conv(env, term1, term2, ConvOptions{})
	}
}

// BenchmarkConv_Projections benchmarks projection normalization.
func BenchmarkConv_Projections(b *testing.B) {
	env := NewEnv()

	// fst (pair a b) ≡ a
	p := pair(glob("a"), glob("b"))
	term1 := fst(p)
	term2 := glob("a")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Conv(env, term1, term2, ConvOptions{})
	}
}

// BenchmarkConv_EtaFunction benchmarks eta conversion for functions.
func BenchmarkConv_EtaFunction(b *testing.B) {
	env := NewEnv()

	// f ≡ \x. f x (with eta)
	f := glob("f")
	etaExpanded := lam("x", app(f, vr(0)))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Conv(env, f, etaExpanded, ConvOptions{EnableEta: true})
	}
}

// BenchmarkConv_EtaPair benchmarks eta conversion for pairs.
func BenchmarkConv_EtaPair(b *testing.B) {
	env := NewEnv()

	// p ≡ (fst p, snd p) (with eta)
	p := glob("p")
	etaExpanded := pair(fst(p), snd(p))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Conv(env, p, etaExpanded, ConvOptions{EnableEta: true})
	}
}

// BenchmarkConv_ComplexTerm benchmarks conversion on more complex terms.
func BenchmarkConv_ComplexTerm(b *testing.B) {
	env := NewEnv()

	// (\x. \y. fst (pair x y)) a b ≡ a
	inner := lam("y", fst(pair(vr(1), vr(0))))
	outer := lam("x", inner)
	term1 := ast.MkApps(outer, glob("a"), glob("b"))
	term2 := glob("a")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Conv(env, term1, term2, ConvOptions{})
	}
}

// BenchmarkConv_Neutral benchmarks conversion on neutral terms.
func BenchmarkConv_Neutral(b *testing.B) {
	env := NewEnv()

	// f x y ≡ f x y (neutral terms)
	f := glob("f")
	term1 := ast.MkApps(f, vr(0), vr(1))
	term2 := ast.MkApps(f, vr(0), vr(1))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Conv(env, term1, term2, ConvOptions{})
	}
}

// BenchmarkConv_PiSigma benchmarks conversion on Pi and Sigma types.
func BenchmarkConv_PiSigma(b *testing.B) {
	env := NewEnv()

	// Pi x : A . B ≡ Pi x : A . B
	typeA := sort(0)
	typeB := sort(0)
	term1 := pi("x", typeA, typeB)
	term2 := pi("x", typeA, typeB)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Conv(env, term1, term2, ConvOptions{})
	}
}

// BenchmarkConv_WithoutEta benchmarks conversion without eta rules.
func BenchmarkConv_WithoutEta(b *testing.B) {
	env := NewEnv()

	// Simple terms without eta conversion
	term1 := glob("a")
	term2 := glob("a")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Conv(env, term1, term2, ConvOptions{EnableEta: false})
	}
}

// BenchmarkConv_WithEta benchmarks conversion with eta rules enabled.
func BenchmarkConv_WithEta(b *testing.B) {
	env := NewEnv()

	// Same terms but with eta enabled
	term1 := glob("a")
	term2 := glob("a")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Conv(env, term1, term2, ConvOptions{EnableEta: true})
	}
}
