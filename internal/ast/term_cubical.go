// Package ast provides the abstract syntax tree for the HoTT kernel.
// This file contains cubical type theory extensions.
package ast

// --- Interval Type ---

// Interval represents the abstract interval type I.
// The interval is NOT a proper type in the universe hierarchy;
// it is used only for path abstraction and application.
type Interval struct{}

func (Interval) isCoreTerm() {}

// I0 represents the left endpoint of the interval (0 : I).
type I0 struct{}

func (I0) isCoreTerm() {}

// I1 represents the right endpoint of the interval (1 : I).
type I1 struct{}

func (I1) isCoreTerm() {}

// IVar is an interval variable with its own de Bruijn index space.
// Interval variables are bound by PathLam and PathP type families.
type IVar struct {
	Ix int // de Bruijn index in the interval variable namespace
}

func (IVar) isCoreTerm() {}

// --- Path Types ---

// Path represents the non-dependent path type: Path A x y.
// This is definitionally equal to PathP (λi. A) x y where A is constant.
//
// Formation rule:
//
//	Γ ⊢ A : Type_i    Γ ⊢ x : A    Γ ⊢ y : A
//	─────────────────────────────────────────
//	Γ ⊢ Path A x y : Type_i
type Path struct {
	A Term // Type (constant over interval)
	X Term // Left endpoint: x : A
	Y Term // Right endpoint: y : A
}

func (Path) isCoreTerm() {}

// PathP represents the dependent path type: PathP A x y.
// Here A is a type family over the interval: A : I → Type.
//
// Formation rule:
//
//	Γ, i:I ⊢ A : Type_j    Γ ⊢ x : A[i0/i]    Γ ⊢ y : A[i1/i]
//	──────────────────────────────────────────────────────────
//	Γ ⊢ PathP (λi. A) x y : Type_j
type PathP struct {
	A Term // Type family: I → Type (binds an interval variable)
	X Term // Left endpoint: x : A[i0/i]
	Y Term // Right endpoint: y : A[i1/i]
}

func (PathP) isCoreTerm() {}

// --- Path Operations ---

// PathLam represents path abstraction: <i> t.
// Creates a path by abstracting over an interval variable.
//
// Introduction rule:
//
//	Γ, i:I ⊢ t : A
//	───────────────────────────────────────────
//	Γ ⊢ <i> t : PathP (λi. A) t[i0/i] t[i1/i]
type PathLam struct {
	Binder string // Interval variable name (for printing only)
	Body   Term   // Body with interval variable bound at index 0
}

func (PathLam) isCoreTerm() {}

// PathApp represents path application: p @ r.
// Applies a path to an interval term (i0, i1, or interval variable).
//
// Elimination rule:
//
//	Γ ⊢ p : PathP A x y    Γ ⊢ r : I
//	────────────────────────────────
//	Γ ⊢ p @ r : A[r/i]
//
// Computation rules:
//
//	(<i> t) @ i0  ⟶  t[i0/i]
//	(<i> t) @ i1  ⟶  t[i1/i]
//	(<i> t) @ j   ⟶  t[j/i]
type PathApp struct {
	P Term // Path: PathP A x y or Path A x y
	R Term // Interval argument: I0, I1, or IVar
}

func (PathApp) isCoreTerm() {}

// Transport represents cubical transport: transport A e.
// Transports an element along a type family over the interval.
//
// Typing rule:
//
//	Γ, i:I ⊢ A : Type_j    Γ ⊢ e : A[i0/i]
//	──────────────────────────────────────
//	Γ ⊢ transport (λi. A) e : A[i1/i]
//
// Computation rule:
//
//	transport (λi. A) e  ⟶  e    (when A is constant in i)
type Transport struct {
	A Term // Type family: I → Type (binds an interval variable)
	E Term // Element at i0: e : A[i0/i]
}

func (Transport) isCoreTerm() {}

// --- Face Formulas (Cofibrations) ---

// Face represents a face formula (cofibration).
// Face formulas are boolean expressions over interval endpoints.
// They are used to specify where partial elements are defined.
type Face interface {
	isFace()
}

// FaceTop represents the always-true face formula (⊤ or 1).
// An element defined on FaceTop is defined everywhere.
type FaceTop struct{}

func (FaceTop) isFace()     {}
func (FaceTop) isCoreTerm() {}

// FaceBot represents the always-false face formula (⊥ or 0).
// An element defined on FaceBot is defined nowhere (empty partial element).
type FaceBot struct{}

func (FaceBot) isFace()     {}
func (FaceBot) isCoreTerm() {}

// FaceEq represents an endpoint constraint: (i = 0) or (i = 1).
// This is satisfied when the interval variable equals the specified endpoint.
type FaceEq struct {
	IVar  int  // Interval variable index (de Bruijn)
	IsOne bool // true for (i = 1), false for (i = 0)
}

func (FaceEq) isFace()     {}
func (FaceEq) isCoreTerm() {}

// FaceAnd represents conjunction of faces: φ ∧ ψ.
// Satisfied when both φ and ψ are satisfied.
type FaceAnd struct {
	Left  Face
	Right Face
}

func (FaceAnd) isFace()     {}
func (FaceAnd) isCoreTerm() {}

// FaceOr represents disjunction of faces: φ ∨ ψ.
// Satisfied when either φ or ψ (or both) are satisfied.
type FaceOr struct {
	Left  Face
	Right Face
}

func (FaceOr) isFace()     {}
func (FaceOr) isCoreTerm() {}

// --- Partial Types and Systems ---

// Partial represents the type of partial elements: Partial φ A.
// A partial element of type A is defined only when face φ is satisfied.
//
// Formation rule:
//
//	Γ ⊢ φ : Face    Γ, φ ⊢ A : Type
//	───────────────────────────────
//	Γ ⊢ Partial φ A : Type
type Partial struct {
	Phi Face // The face constraint
	A   Term // The type (may reference interval variables)
}

func (Partial) isCoreTerm() {}

// System represents a system of partial elements: [ φ₁ ↦ t₁, φ₂ ↦ t₂, ... ].
// Each branch provides an element when its face is satisfied.
// Branches must agree on overlaps.
//
// Typing rule:
//
//	Γ, φᵢ ⊢ tᵢ : A    (∀i,j: φᵢ ∧ φⱼ ⊢ tᵢ = tⱼ)
//	────────────────────────────────────────────
//	Γ ⊢ [φ₁ ↦ t₁, ...] : Partial (φ₁ ∨ ...) A
type System struct {
	Branches []SystemBranch
}

func (System) isCoreTerm() {}

// SystemBranch represents a single branch in a system: φ ↦ t.
type SystemBranch struct {
	Phi  Face // When this face is satisfied...
	Term Term // ...this term is the value
}

// --- Composition Operations ---

// Comp represents heterogeneous composition: comp^i A [φ ↦ u] a₀.
// This is the fundamental operation for transport in cubical type theory.
//
// Typing rule:
//
//	Γ, i:I ⊢ A : Type    Γ, i:I, φ ⊢ u : A    Γ ⊢ a₀ : A[i0/i]    Γ, φ[i0/i] ⊢ u[i0/i] = a₀ : A[i0/i]
//	──────────────────────────────────────────────────────────────────────────────────────────────────
//	Γ ⊢ comp^i A [φ ↦ u] a₀ : A[i1/i]
//
// Computation rules:
//
//	comp^i A [1 ↦ u] a₀  ⟶  u[i1/i]         (face satisfied)
//	comp^i A [0 ↦ _] a₀  ⟶  transport A a₀  (face empty, reduces to transport)
type Comp struct {
	IBinder string // Interval binder name (for printing)
	A       Term   // Type line: I → Type (binds interval variable)
	Phi     Face   // Face constraint
	Tube    Term   // Partial tube: defined when φ holds (binds interval variable)
	Base    Term   // Base element: a₀ : A[i0/i]
}

func (Comp) isCoreTerm() {}

// HComp represents homogeneous composition: hcomp A [φ ↦ u] a₀.
// This is composition where the type A is constant (doesn't depend on i).
//
// Typing rule:
//
//	Γ ⊢ A : Type    Γ, i:I, φ ⊢ u : A    Γ ⊢ a₀ : A    Γ, φ[i0/i] ⊢ u[i0/i] = a₀ : A
//	─────────────────────────────────────────────────────────────────────────────────
//	Γ ⊢ hcomp A [φ ↦ u] a₀ : A
//
// Computation rules:
//
//	hcomp A [1 ↦ u] a₀  ⟶  u[i1/i]   (face satisfied)
//	hcomp A [0 ↦ _] a₀  ⟶  a₀        (face empty, identity)
type HComp struct {
	A    Term // Type (constant, no interval dependency)
	Phi  Face // Face constraint
	Tube Term // Partial tube (binds interval variable)
	Base Term // Base element
}

func (HComp) isCoreTerm() {}

// Fill represents the filler for composition: fill^i A [φ ↦ u] a₀.
// This produces a path from a₀ to comp^i A [φ ↦ u] a₀.
//
// Typing rule:
//
//	Γ, i:I ⊢ A : Type    Γ, i:I, φ ⊢ u : A    Γ ⊢ a₀ : A[i0/i]
//	──────────────────────────────────────────────────────────
//	Γ, j:I ⊢ fill^i A [φ ↦ u] a₀ @ j : A[j/i]
//
// The fill operation is defined as:
//
//	fill^i A [φ ↦ u] a₀ = λj. comp^i A[j∧i/i] [φ ∨ (j=0) ↦ ...] a₀
//
// Endpoints:
//
//	fill^i A [φ ↦ u] a₀ @ i0 = a₀
//	fill^i A [φ ↦ u] a₀ @ i1 = comp^i A [φ ↦ u] a₀
type Fill struct {
	IBinder string // Interval binder name (for printing)
	A       Term   // Type line: I → Type (binds interval variable)
	Phi     Face   // Face constraint
	Tube    Term   // Partial tube (binds interval variable)
	Base    Term   // Base element
}

func (Fill) isCoreTerm() {}

// --- Glue Types (for univalence) ---

// Glue represents the Glue type: Glue A [φ ↦ (T, e)].
// This is used to construct paths between types via equivalences.
//
// Formation rule:
//
//	Γ ⊢ A : Type    Γ, φ ⊢ T : Type    Γ, φ ⊢ e : Equiv T A
//	────────────────────────────────────────────────────────
//	Γ ⊢ Glue A [φ ↦ (T, e)] : Type
//
// Computation rules:
//
//	Glue A [1 ↦ (T, e)] = T              (face satisfied)
//	Glue A [0 ↦ (T, e)] = A              (face empty)
type Glue struct {
	A      Term         // Base type
	System []GlueBranch // System of equivalences
}

func (Glue) isCoreTerm() {}

// GlueBranch represents a branch in a Glue type: φ ↦ (T, e).
type GlueBranch struct {
	Phi   Face // Face constraint
	T     Term // Fiber type (when φ holds)
	Equiv Term // Equivalence: Equiv T A (when φ holds)
}

// GlueElem represents a Glue element constructor: glue [φ ↦ t] a.
// Creates an element of a Glue type.
//
// Typing rule:
//
//	Γ ⊢ a : A    Γ, φ ⊢ t : T    Γ, φ ⊢ e.fst t = a : A
//	────────────────────────────────────────────────────
//	Γ ⊢ glue [φ ↦ t] a : Glue A [φ ↦ (T, e)]
//
// Computation rules:
//
//	glue [1 ↦ t] a = t                   (face satisfied)
type GlueElem struct {
	System []GlueElemBranch // Partial element in fiber
	Base   Term             // Base element: a : A
}

func (GlueElem) isCoreTerm() {}

// GlueElemBranch represents a branch in a GlueElem: φ ↦ t.
type GlueElemBranch struct {
	Phi  Face // Face constraint
	Term Term // Element in T (when φ holds)
}

// Unglue extracts the base from a Glue element: unglue g.
//
// Typing rule:
//
//	Γ ⊢ g : Glue A [φ ↦ (T, e)]
//	───────────────────────────
//	Γ ⊢ unglue g : A
//
// Computation rules:
//
//	unglue (glue [φ ↦ t] a) = a          (definitional)
//	unglue g = e.fst g                   (when φ holds)
type Unglue struct {
	Ty Term // The Glue type (needed for computation)
	G  Term // The Glue element
}

func (Unglue) isCoreTerm() {}

// --- Univalence ---

// UA represents the univalence axiom: ua e : Path Type A B.
// Given an equivalence e : Equiv A B, ua produces a path between types.
//
// Typing rule:
//
//	Γ ⊢ A : Type_i    Γ ⊢ B : Type_i    Γ ⊢ e : Equiv A B
//	─────────────────────────────────────────────────────
//	Γ ⊢ ua e : Path Type_i A B
//
// Definition via Glue:
//
//	ua e = <i> Glue B [(i=0) ↦ (A, e)]
//
// At i=0: Glue B [⊤ ↦ (A, e)] = A
// At i=1: Glue B [⊥ ↦ (A, e)] = B
//
// Key computation (transport computes!):
//
//	transport (ua e) a = e.fst a
type UA struct {
	A     Term // Source type
	B     Term // Target type
	Equiv Term // Equivalence: Equiv A B
}

func (UA) isCoreTerm() {}

// UABeta represents the computation rule for transport along ua.
// This is used when we need to explicitly mark that transport (ua e) a = e.fst a.
//
//	transport (ua e) a  ⟶  e.fst a
type UABeta struct {
	Equiv Term // The equivalence e : Equiv A B
	Arg   Term // The argument a : A
}

func (UABeta) isCoreTerm() {}
