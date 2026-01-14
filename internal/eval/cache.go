package eval

import (
	"sync"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// Cache provides memoization for NbE evaluation.
// It is designed to be scoped to a single operation (like a conversion check)
// to avoid memory leaks while still providing performance benefits.
//
// IMPORTANT: This cache is ONLY safe when used with terms that have stable
// pointer identity within a single operation. It uses pointer comparison
// for cache keys, which means:
//   - Within one conversion check, the same term node will hit the cache
//   - Structurally equal but separately constructed terms will NOT share cache entries
//
// Safety invariants:
//   - Cache is cleared after each operation (via Reset or by creating new instances)
//   - Cache only stores values, not references that could cause cycles
//   - Thread-safe for concurrent access within an operation
//
// Note: Reification is NOT cached because Values may contain slices which are
// not hashable in Go maps. Evaluation caching provides the main performance benefit.
type Cache struct {
	mu sync.RWMutex

	// evalCache maps (term interface value, env) to evaluated Value
	// We use the term itself as key (interface comparison uses pointer equality for most types)
	evalCache map[evalCacheKey]Value

	// Stats for debugging/profiling
	EvalHits   int64
	EvalMisses int64

	// MaxSize limits cache growth. 0 means unlimited.
	MaxSize int
}

// evalCacheKey uniquely identifies an evaluation context.
// We use the term interface and env pointer for identity.
type evalCacheKey struct {
	term ast.Term
	env  *Env
}

// NewCache creates a new cache with optional size limit.
// A size of 0 means unlimited (use with caution).
func NewCache(maxSize int) *Cache {
	return &Cache{
		evalCache: make(map[evalCacheKey]Value),
		MaxSize:   maxSize,
	}
}

// DefaultCacheSize is a reasonable default for most operations.
const DefaultCacheSize = 10000

// NewDefaultCache creates a cache with the default size limit.
func NewDefaultCache() *Cache {
	return NewCache(DefaultCacheSize)
}

// Reset clears the cache for reuse.
func (c *Cache) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.evalCache = make(map[evalCacheKey]Value)
	c.EvalHits = 0
	c.EvalMisses = 0
}

// LookupEval checks if a term has been evaluated in this environment context.
// Uses pointer identity for the term and env.
func (c *Cache) LookupEval(t ast.Term, env *Env) (Value, bool) {
	if c == nil {
		return nil, false
	}
	key := evalCacheKey{term: t, env: env}
	c.mu.RLock()
	val, ok := c.evalCache[key]
	c.mu.RUnlock()
	if ok {
		c.mu.Lock()
		c.EvalHits++
		c.mu.Unlock()
	}
	return val, ok
}

// StoreEval stores an evaluation result.
func (c *Cache) StoreEval(t ast.Term, env *Env, val Value) {
	if c == nil {
		return
	}
	// Check size limit
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.MaxSize > 0 && len(c.evalCache) >= c.MaxSize {
		// Simple eviction: just skip storing (could implement LRU instead)
		return
	}

	c.EvalMisses++
	key := evalCacheKey{term: t, env: env}
	c.evalCache[key] = val
}

// Stats returns cache statistics as a map.
func (c *Cache) Stats() map[string]int64 {
	if c == nil {
		return nil
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	return map[string]int64{
		"eval_hits":   c.EvalHits,
		"eval_misses": c.EvalMisses,
		"eval_size":   int64(len(c.evalCache)),
	}
}
