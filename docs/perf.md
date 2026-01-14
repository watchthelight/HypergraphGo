# Performance Optimization

This document describes the performance optimizations implemented in HoTTGo's type checking and normalization pipeline.

## Overview

The core performance-sensitive operations in HoTTGo are:

1. **Normalization by Evaluation (NbE)** - Converting terms to semantic values and back
2. **Conversion Checking** - Determining if two terms are definitionally equal
3. **Alpha-Equality** - Comparing syntactic terms for structural equality

## Caching Infrastructure

### Evaluation Cache

The `eval.Cache` provides memoization for NbE evaluation. It maps `(term, env)` pairs to evaluated `Value` results.

```go
cache := eval.NewDefaultCache()
val := eval.EvalCached(env, term, cache)
```

**Key design decisions:**

1. **Pointer identity keys** - Cache keys use Go interface comparison (pointer identity for most AST node types). This means:
   - Same term node evaluated multiple times → cache hit
   - Structurally equal but different nodes → no sharing

2. **Evaluation-only caching** - Only evaluation is cached, not reification. Values can contain slices (`VNeutral.Sp`) which are not hashable as map keys.

3. **Scoped lifetime** - Caches are intended for single operations (e.g., one type-checking pass) to avoid unbounded memory growth.

4. **Size limits** - `DefaultCacheSize = 10000` entries with simple eviction (skip storing when full).

### Conversion Context

For multiple related conversion checks, `ConvContext` maintains a shared cache:

```go
ctx := NewConvContext()
for _, pair := range termsToCheck {
    if !ctx.Conv(env, pair[0], pair[1], opts) {
        return false
    }
}
```

## Benchmarks

All benchmarks run on Apple M4, Go 1.24.

### Allocation Reduction

The most dramatic improvement is allocation reduction when the same terms are evaluated repeatedly:

| Benchmark | Time | Allocations |
|-----------|------|-------------|
| `Eval_RepeatedTerm100` (no cache) | 11,741 ns | 800 allocs |
| `EvalCached_RepeatedTerm100` | 10,613 ns | 11 allocs |

**Result: 99% allocation reduction** for repeated evaluation of the same term.

### Conversion Checking

Basic conversion operations (comparing self to self):

| Benchmark | Time | Memory |
|-----------|------|--------|
| `Conv_Simple` | 103 ns | 192 B |
| `Conv_Beta` | 181 ns | 320 B |
| `Conv_Projections` | 92 ns | 208 B |
| `Conv_EtaFunction` | 198 ns | 400 B |
| `Conv_EtaPair` | 208 ns | 416 B |
| `Conv_Neutral` | 340 ns | 736 B |

### Deep Term Performance

Church numeral benchmarks show linear scaling:

| Depth | Time |
|-------|------|
| Church(10) | 1,536 ns |
| Church(50) | 8,462 ns |
| Church(100) | 18,146 ns |

### Alpha-Equality

Zero-allocation comparison for equal terms:

| Benchmark | Time | Allocations |
|-----------|------|-------------|
| `AlphaEq_LargeTerm` (50 nodes) | 726 ns | 0 |
| `AlphaEq_MismatchEarly` | 5.7 ns | 0 |
| `AlphaEq_MismatchLate` | 100 ns | 0 |

## API Reference

### Basic Usage

```go
// Single conversion check with fresh cache
result := ConvCached(env, term1, term2, ConvOptions{})

// Multiple checks with shared cache
ctx := NewConvContext()
for _, pair := range pairs {
    ctx.Conv(env, pair[0], pair[1], opts)
}

// Direct NbE with caching
normalized := eval.NormalizeWithCache(term)

// Custom cache usage
cache := eval.NewDefaultCache()
val := eval.EvalCached(env, term, cache)
stats := cache.Stats() // {"eval_hits": N, "eval_misses": M, "eval_size": S}
```

### When Caching Helps

Caching provides the most benefit when:

1. **Same AST nodes are traversed multiple times** - Common in type checking where terms are checked against multiple types
2. **Terms share subexpressions** - The same subterm pointer evaluated in different contexts
3. **Batch operations** - Multiple conversion checks on related terms

Caching provides less benefit when:

1. **Terms are structurally equal but pointer-different** - Each `makeChurchNumeral(5)` call creates new AST nodes
2. **Single one-off evaluations** - Cache overhead exceeds benefit
3. **Very small terms** - Overhead of cache lookup may exceed saved work

## Files

| File | Description |
|------|-------------|
| `internal/eval/cache.go` | Cache implementation with size limits |
| `internal/eval/nbe_cached.go` | Cached NbE functions |
| `internal/eval/cache_test.go` | Regression tests for cache correctness |
| `internal/core/conv_cached.go` | Cached conversion checking |
| `internal/core/perf_bench_test.go` | Performance benchmarks |

## Running Benchmarks

```bash
# Run all performance benchmarks
go test -bench=. -benchmem ./internal/core/...

# Run specific benchmark
go test -bench=BenchmarkConv_DeepChurch -benchmem ./internal/core/...

# Compare cached vs non-cached
go test -bench='(Cached|NonCached)' -benchmem ./internal/core/...

# Profile with pprof
go test -bench=BenchmarkConv_DeepChurch100 -cpuprofile=cpu.prof ./internal/core/...
go tool pprof cpu.prof
```

## Future Optimizations

Potential areas for further improvement:

1. **Hash-consing for AST nodes** - Would enable cache sharing for structurally equal terms
2. **LRU eviction** - Current implementation just stops storing when full
3. **Parallel evaluation** - Cache is thread-safe but evaluation is single-threaded
4. **Specialized caches** - Different cache strategies for different term types
