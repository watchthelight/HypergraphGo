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
type Pi struct {
	Binder string
	A      Term
	B      Term
}

func (Pi) isCoreTerm() {}

// Lam (λ) abstraction with optional type annotation on the binder for printing.
type Lam struct {
	Binder string
	Ann    Term // may be nil
	Body   Term
}

func (Lam) isCoreTerm() {}

// App t u (application).
type App struct {
	T Term
	U Term
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
