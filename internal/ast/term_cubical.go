//go:build cubical

// Package ast provides the abstract syntax tree for the HoTT kernel.
// This file contains cubical type theory extensions (gated by build tag).
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
