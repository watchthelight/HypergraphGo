package ast

// Level represents a universe level (Type^u).
type Level uint

// Sort is a universe: Type^U.
type Sort struct{ U Level }

func (Sort) isCoreTerm() {}

// Var is a de Bruijn variable: index (0 = innermost binder).
type Var struct{ Ix int }

func (Var) isCoreTerm() {}

// Global refers to a globally named constant (post-elaboration).
type Global struct{ Name string }

func (Global) isCoreTerm() {}

// Pi (Π) type with domain A and codomain B in a binder (x:A).B.
// We keep the binder name only for pretty-printing; kernel uses de Bruijn.
// Implicit indicates whether this is an implicit argument: {x : A} -> B.
type Pi struct {
	Binder   string
	A        Term
	B        Term
	Implicit bool // If true, argument is implicit (inferred by unification)
}

func (Pi) isCoreTerm() {}

// Lam (λ) abstraction with optional type annotation on the binder for printing.
// Implicit indicates whether this is an implicit lambda: λ{x}. body.
type Lam struct {
	Binder   string
	Ann      Term // may be nil
	Body     Term
	Implicit bool // If true, this lambda binds an implicit argument
}

func (Lam) isCoreTerm() {}

// App t u (application).
// Implicit indicates whether this is an implicit application: f {arg}.
type App struct {
	T        Term
	U        Term
	Implicit bool // If true, this application was inserted by elaboration
}

func (App) isCoreTerm() {}

// Sigma (Σ) type.
type Sigma struct {
	Binder string
	A      Term
	B      Term
}

func (Sigma) isCoreTerm() {}

// Pair (a , b)
type Pair struct {
	Fst Term
	Snd Term
}

func (Pair) isCoreTerm() {}

// Fst and Snd projections.
type Fst struct{ P Term }

func (Fst) isCoreTerm() {}

type Snd struct{ P Term }

func (Snd) isCoreTerm() {}

// Let x = v : A in body. (Convenience; kernel may desugar away later.)
type Let struct {
	Binder string
	Ann    Term // may be nil
	Val    Term
	Body   Term
}

func (Let) isCoreTerm() {}

// Id represents the identity type: Id A x y
// "x and y are propositionally equal in type A"
type Id struct {
	A Term // Type
	X Term // Left endpoint
	Y Term // Right endpoint
}

func (Id) isCoreTerm() {}

// Refl is the reflexivity constructor: refl A x : Id A x x
type Refl struct {
	A Term // Type
	X Term // The term being proven equal to itself
}

func (Refl) isCoreTerm() {}

// J is the identity eliminator (path induction).
// J A C d x y p : C y p
// where:
//
//	A : Type
//	C : (y : A) -> Id A x y -> Type   (motive)
//	D : C x (refl A x)                (base case)
//	X : A                             (left endpoint)
//	Y : A                             (right endpoint)
//	P : Id A x y                      (proof of equality)
type J struct {
	A Term // Type
	C Term // Motive: (y : A) -> Id A x y -> Type
	D Term // Base case: C x (refl A x)
	X Term // Left endpoint
	Y Term // Right endpoint
	P Term // Proof: Id A x y
}

func (J) isCoreTerm() {}

// Term is the interface for all core terms.
type Term interface{ isCoreTerm() }

// Helpers

// IsZeroLevel returns true if this is Sort 0 (Type0).
func (s Sort) IsZeroLevel() bool { return s.U == 0 }

// MkApps applies t to us left-associatively.
func MkApps(t Term, us ...Term) Term {
	for _, u := range us {
		t = App{T: t, U: u}
	}
	return t
}

// Meta represents a metavariable (hole) to be solved during elaboration.
// Metavariables are placeholders for unknown terms that are filled in
// by unification. Each metavariable has a unique ID and may be applied
// to a spine of arguments representing the local context.
type Meta struct {
	ID   int    // Unique identifier for this metavariable
	Args []Term // Spine of arguments (local context variables)
}

func (Meta) isCoreTerm() {}
