// Package eval implements Normalization by Evaluation (NbE) for the HoTT kernel.
//
// NbE is a technique for computing normal forms and definitional equality that
// works by interpreting syntax into a semantic domain, then reading back the
// result as syntax. This two-phase approach provides efficient normalization
// while naturally handling open terms and binders.
//
// # The NbE Pipeline
//
// The normalization process consists of two phases:
//
//  1. Evaluation (Eval): Interprets syntax (ast.Term) into the semantic domain
//     (Value), computing all possible beta and iota reductions.
//
//  2. Reification (Reify): Reads back semantic values as normalized syntax,
//     using eta-expansion where appropriate.
//
// The complete pipeline is:
//
//	term → Eval(env, term) → value → Reify(value) → normal form
//
// The convenience function EvalNBE combines both phases.
//
// # Semantic Domain
//
// Values represent terms in weak head normal form (WHNF):
//
//   - VSort: Universe sorts (Type₀, Type₁, ...)
//   - VLam: Lambda closures capturing environment and body
//   - VPi: Pi type closures with evaluated domain
//   - VSigma: Sigma type closures with evaluated domain
//   - VPair: Pairs in WHNF
//   - VNeutral: Stuck computations (variable applications)
//   - VGlobal: Global constant references
//   - VId, VRefl: Identity type values
//
// Cubical values (VPath, VPathP, VPathLam, etc.) extend the domain for
// path types and composition operations.
//
// # Closures and Lazy Evaluation
//
// Closures (Closure type) capture an environment together with a term,
// deferring evaluation until the closure is applied. This enables:
//
//   - Lazy evaluation under binders (lambdas, Pi/Sigma types)
//   - Efficient handling of substitution via environment extension
//   - Natural treatment of higher-order functions
//
// When a lambda is applied, its closure is extended with the argument and
// the body is evaluated in the new environment.
//
// # Neutral Terms
//
// Neutral terms (Neutral type) represent computations that are "stuck"
// because the head is a variable or an applied constant that cannot reduce:
//
//   - A variable applied to arguments: x arg1 arg2 ...
//   - A global that's not a redex: natElim P pz ps n (when n is a variable)
//
// Neutral terms consist of a Head (variable or global) and a Spine (list
// of applied arguments). During reification, neutrals are converted back
// to application chains.
//
// # Environment
//
// The evaluation environment (Env type) maps de Bruijn indices to values.
// Index 0 is the most recently bound variable. Environments grow by
// prepending new bindings:
//
//	env.Extend(val) → new environment with val at index 0
//
// # Recursor Reduction
//
// The evaluator supports computation rules for inductive type eliminators:
//
// Built-in:
//   - natElim: Reduces on zero and succ constructors
//   - boolElim: Reduces on true and false constructors
//
// User-defined:
//   - RegisterRecursor: Registers elimination rules for custom inductives
//   - Generic recursor reduction matches constructor patterns in the spine
//
// Recursor reduction enables dependent elimination to compute, which is
// essential for proving properties of inductive types.
//
// # Debug Mode
//
// Set HOTTGO_DEBUG=1 to enable strict error handling. In debug mode,
// internal errors panic immediately rather than returning fallback values.
// This is useful for catching bugs during development.
//
// # Cubical Extension
//
// The package includes support for cubical type theory:
//
//   - Interval values and variables (separate namespace)
//   - Path type evaluation (VPath, VPathP)
//   - Path lambda and application
//   - Transport and composition operations
//   - Face formula evaluation and simplification
//   - Glue type operations
//
// These extensions enable computational univalence, where transport along
// equivalences computes to function application.
//
// # References
//
//   - Abel, A. (2013). Normalization by Evaluation: Dependent Types and
//     Impredicativity. Habilitation thesis.
//   - Coquand, T. (1996). An Algorithm for Type-Checking Dependent Types.
//     Science of Computer Programming.
//   - Gratzer, D. & Sterling, J. (2020). Implementing a Modal Dependent
//     Type Theory. ICFP 2020.
package eval
