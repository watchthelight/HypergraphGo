// Package eval implements Normalization by Evaluation (NbE) for the HoTT kernel.
//
// NbE is a technique for normalizing lambda calculus terms by evaluating
// syntax into a semantic domain and then reifying values back to normal form.
// This approach is more efficient than repeated syntactic reduction and
// naturally handles evaluation under binders.
//
// # Overview
//
// The normalization process has three phases:
//
//  1. Eval: syntax → semantic values (Values)
//  2. Apply: handle applications (performs beta reduction)
//  3. Reify: semantic values → syntax in normal form
//
// # Semantic Domain
//
// The [Value] interface represents semantic values:
//
//   - [VLam] - lambda closures capturing environment
//   - [VPi] - Pi type with domain value and codomain closure
//   - [VSigma] - Sigma type with domain value and codomain closure
//   - [VPair] - pairs in weak head normal form
//   - [VSort] - universe sorts
//   - [VGlobal] - global constant references
//   - [VNeutral] - stuck computations (see below)
//   - [VId] - identity type values
//   - [VRefl] - reflexivity proof values
//
// # Closures
//
// [Closure] captures an environment and a term for lazy evaluation under
// binders. When a lambda is evaluated, it becomes a VLam containing a closure.
// When applied, the closure's term is evaluated in an extended environment.
//
// # Neutral Terms
//
// [Neutral] represents stuck computations - terms that cannot reduce further
// because they're blocked on a variable. A neutral term has:
//
//   - [Head] - either a de Bruijn variable index or global name
//   - Spine - list of arguments already applied
//
// For example, (x 1 2) where x is a variable becomes a neutral with
// head x and spine [1, 2].
//
// # Environment
//
// [Env] maps de Bruijn indices to Values:
//
//   - Lookup(ix) - retrieves value at index, returns neutral if unbound
//   - Extend(v) - adds binding at index 0, shifting others up
//
// # Key Functions
//
//   - [Eval] - evaluates a term to weak head normal form
//   - [Apply] - applies a function value to an argument
//   - [Reify] - converts a value back to syntax in normal form
//   - [EvalNBE] - convenience: Eval + Reify in one call
//   - [Fst], [Snd] - projections in the semantic domain
//
// # Debug Mode
//
// The [DebugMode] variable (set via HOTTGO_DEBUG=1) controls error handling:
//
//   - When true: internal errors cause panics for easier debugging
//   - When false: errors return diagnostic fallback values
//
// Well-typed inputs should never trigger these error paths.
//
// # Recursor Reduction
//
// The evaluator handles computation rules for eliminators:
//
//   - natElim P pz ps zero → pz
//   - natElim P pz ps (succ n) → ps n (natElim P pz ps n)
//   - boolElim P pt pf true → pt
//   - boolElim P pt pf false → pf
//   - J A C d x x (refl A x) → d
//
// User-defined recursors are handled via the recursor registry.
//
// # Cubical Extensions
//
// When cubical features are enabled, additional Value types and evaluation
// rules are defined in nbe_cubical.go for paths, transport, composition, etc.
//
// # References
//
//   - Abel, A. "Normalization by Evaluation: Dependent Types and Impredicativity"
//   - Coquand, T. "An Algorithm for Type-Checking Dependent Types"
package eval
