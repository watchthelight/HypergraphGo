// Package ast defines the abstract syntax tree for the HoTT kernel.
//
// This package provides the core term representation used throughout the type
// checker and evaluator. Terms use de Bruijn indices for variables, ensuring
// that alpha-equivalent terms have identical representations.
//
// # Term Hierarchy
//
// The Term interface is implemented by all core term types:
//
// Base terms:
//   - Sort: Universe types (Type₀, Type₁, Type₂, ...)
//   - Var: De Bruijn indexed variables
//   - Global: References to globally defined constants
//
// Function types (Π-types):
//   - Pi: Dependent function type (x : A) → B
//   - Lam: Lambda abstraction λx. t
//   - App: Function application f x
//
// Pair types (Σ-types):
//   - Sigma: Dependent pair type (x : A) × B
//   - Pair: Pair constructor (a, b)
//   - Fst, Snd: Projections
//
// Identity types:
//   - Id: Identity type Id A x y
//   - Refl: Reflexivity proof refl : Id A x x
//   - J: Path induction eliminator
//
// Let bindings:
//   - Let: Local definitions let x = v : A in t
//
// # Cubical Extensions
//
// The package includes cubical type theory extensions for computational
// univalence, as described in Cohen et al. (2018):
//
// Interval type:
//   - Interval: The abstract interval type I
//   - I0, I1: Interval endpoints (0 : I) and (1 : I)
//   - IVar: Interval variables (separate de Bruijn namespace)
//
// Path types:
//   - Path: Non-dependent path type Path A x y
//   - PathP: Dependent path type PathP (λi. A) x y
//   - PathLam: Path abstraction <i> t
//   - PathApp: Path application p @ r
//
// Transport and composition:
//   - Transport: transport (λi. A) e
//   - Comp: Heterogeneous composition comp^i A [φ ↦ u] a₀
//   - HComp: Homogeneous composition hcomp A [φ ↦ u] a₀
//   - Fill: Composition filler fill^i A [φ ↦ u] a₀
//
// Face formulas:
//   - Face interface with FaceTop, FaceBot, FaceEq, FaceAnd, FaceOr
//   - Partial: Partial types Partial φ A
//   - System: Systems of partial elements [φ ↦ t, ψ ↦ u, ...]
//
// Glue types (for univalence):
//   - Glue: Glue type Glue A [φ ↦ (T, e)]
//   - GlueElem: Glue element constructor
//   - Unglue: Glue element projector
//   - UA: Univalence axiom ua : Equiv A B → Path Type A B
//   - UABeta: Computation rule transport (ua e) a ≡ e.fst a
//
// # De Bruijn Representation
//
// Variables use de Bruijn indices where index 0 refers to the innermost
// (most recently bound) variable. This applies to both term variables (Var)
// and interval variables (IVar), which occupy separate namespaces.
//
// Binder names (in Pi, Lam, Sigma, Let, PathLam) are preserved for
// pretty-printing but are not semantically significant. The kernel treats
// terms as equivalent if they differ only in binder names.
//
// # Name Resolution
//
// Raw terms (with named variables) are converted to core terms via the
// Resolve function, which handles scope tracking and de Bruijn index
// assignment. This separation allows parsing to produce user-friendly
// error messages while the kernel operates on the canonical representation.
//
// # References
//
//   - Pierce, B. C. (2002). Types and Programming Languages. MIT Press.
//     Chapter 6 (Nameless Representation of Terms).
//   - Cohen, C., Coquand, T., Huber, S., & Mörtberg, A. (2018). Cubical Type
//     Theory: A Constructive Interpretation of the Univalence Axiom. TYPES 2015.
//   - The Univalent Foundations Program (2013). Homotopy Type Theory:
//     Univalent Foundations of Mathematics.
package ast
