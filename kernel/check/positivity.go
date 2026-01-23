package check

import (
	"fmt"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// Polarity tracks whether we're in a positive or negative position.
type Polarity int

const (
	Positive Polarity = iota
	Negative
)

// Flip returns the opposite polarity.
func (p Polarity) Flip() Polarity {
	if p == Positive {
		return Negative
	}
	return Positive
}

func (p Polarity) String() string {
	if p == Positive {
		return "positive"
	}
	return "negative"
}

// PositivityError represents a strict positivity violation.
type PositivityError struct {
	IndName     string
	Constructor string
	Position    string
	Polarity    Polarity
}

func (e *PositivityError) Error() string {
	return fmt.Sprintf("strict positivity violation: %s occurs in %s position in constructor %s at %s",
		e.IndName, e.Polarity, e.Constructor, e.Position)
}

// CheckPositivity verifies that a single inductive type definition satisfies
// the strict positivity condition. This ensures the inductive is well-founded
// and prevents logical inconsistencies.
//
// A type T occurs strictly positively in X if:
// - X does not mention T, OR
// - X = T (the type itself), OR
// - X = (a : A) -> B where T does NOT occur in A and occurs strictly positively in B
func CheckPositivity(indName string, constructors []Constructor) error {
	for _, c := range constructors {
		if err := checkConstructorPositivity(indName, c.Name, c.Type); err != nil {
			return err
		}
	}
	return nil
}

// CheckMutualPositivity verifies that mutually recursive inductive types
// satisfy the strict positivity condition across all types in the mutual block.
//
// For mutual inductives, each type must occur strictly positively in the
// constructors of ALL types in the mutual block. That is:
// - For each type T in the mutual block
// - For each constructor C of any type in the block
// - T must occur only in strictly positive positions in C
func CheckMutualPositivity(indNames []string, constructors map[string][]Constructor) error {
	// Check positivity across ALL constructors of ALL types
	// Each constructor must have all mutual types occurring only strictly positively
	for _, constrs := range constructors {
		for _, c := range constrs {
			if err := checkConstructorPositivityMulti(indNames, c.Name, c.Type); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkConstructorPositivityMulti checks a constructor for strict positivity
// of all types in a mutual block.
func checkConstructorPositivityMulti(indNames []string, ctorName string, ty ast.Term) error {
	return checkConstructorArgsMulti(indNames, ctorName, ty, 0)
}

// checkConstructorArgsMulti traverses the constructor type's argument structure
// for mutual inductives.
func checkConstructorArgsMulti(indNames []string, ctorName string, ty ast.Term, depth int) error {
	switch t := ty.(type) {
	case ast.Pi:
		// Check that the argument type has all mutual types occurring strictly positively
		if err := checkArgTypePositivityMulti(indNames, ctorName, t.A, Positive, depth); err != nil {
			return err
		}
		// Continue to next argument or result
		return checkConstructorArgsMulti(indNames, ctorName, t.B, depth+1)
	case ast.Global:
		// Result type - any mutual type itself is fine
		return nil
	case ast.App:
		// Result type like (T A) - check arguments don't have mutual types in bad positions
		return checkArgTypePositivityMulti(indNames, ctorName, ty, Positive, depth)
	default:
		// Other result types
		return nil
	}
}

// checkArgTypePositivityMulti checks that within an argument type, all
// mutual inductive types occur only in strictly positive positions.
func checkArgTypePositivityMulti(indNames []string, ctorName string, ty ast.Term, pol Polarity, depth int) error {
	switch t := ty.(type) {
	case ast.Pi:
		// In a function type A -> B:
		// - A is in negative position (no occurrences allowed)
		// - B stays at current polarity
		newPol := Negative
		if err := checkArgTypePositivityMulti(indNames, ctorName, t.A, newPol, depth); err != nil {
			return err
		}
		return checkArgTypePositivityMulti(indNames, ctorName, t.B, pol, depth)

	case ast.Global:
		// Check if this global is one of our mutual inductives
		for _, indName := range indNames {
			if t.Name == indName {
				if pol == Negative {
					return &PositivityError{
						IndName:     indName,
						Constructor: ctorName,
						Position:    fmt.Sprintf("argument %d", depth),
						Polarity:    pol,
					}
				}
			}
		}
		return nil

	case ast.App:
		// Check both parts of an application
		if err := checkArgTypePositivityMulti(indNames, ctorName, t.T, pol, depth); err != nil {
			return err
		}
		return checkArgTypePositivityMulti(indNames, ctorName, t.U, pol, depth)

	case ast.Var:
		return nil

	case ast.Sort:
		return nil

	case ast.Sigma:
		if err := checkArgTypePositivityMulti(indNames, ctorName, t.A, pol, depth); err != nil {
			return err
		}
		return checkArgTypePositivityMulti(indNames, ctorName, t.B, pol, depth)

	case ast.Lam:
		if t.Ann != nil {
			if err := checkArgTypePositivityMulti(indNames, ctorName, t.Ann, pol, depth); err != nil {
				return err
			}
		}
		return checkArgTypePositivityMulti(indNames, ctorName, t.Body, pol, depth)

	case ast.Pair:
		if err := checkArgTypePositivityMulti(indNames, ctorName, t.Fst, pol, depth); err != nil {
			return err
		}
		return checkArgTypePositivityMulti(indNames, ctorName, t.Snd, pol, depth)

	case ast.Fst:
		return checkArgTypePositivityMulti(indNames, ctorName, t.P, pol, depth)

	case ast.Snd:
		return checkArgTypePositivityMulti(indNames, ctorName, t.P, pol, depth)

	case ast.Let:
		if err := checkArgTypePositivityMulti(indNames, ctorName, t.Ann, pol, depth); err != nil {
			return err
		}
		if err := checkArgTypePositivityMulti(indNames, ctorName, t.Val, pol, depth); err != nil {
			return err
		}
		return checkArgTypePositivityMulti(indNames, ctorName, t.Body, pol, depth)

	case ast.Id:
		if err := checkArgTypePositivityMulti(indNames, ctorName, t.A, pol, depth); err != nil {
			return err
		}
		if err := checkArgTypePositivityMulti(indNames, ctorName, t.X, pol, depth); err != nil {
			return err
		}
		return checkArgTypePositivityMulti(indNames, ctorName, t.Y, pol, depth)

	case ast.Refl:
		if err := checkArgTypePositivityMulti(indNames, ctorName, t.A, pol, depth); err != nil {
			return err
		}
		return checkArgTypePositivityMulti(indNames, ctorName, t.X, pol, depth)

	case ast.J:
		if err := checkArgTypePositivityMulti(indNames, ctorName, t.A, pol, depth); err != nil {
			return err
		}
		if err := checkArgTypePositivityMulti(indNames, ctorName, t.C, pol, depth); err != nil {
			return err
		}
		if err := checkArgTypePositivityMulti(indNames, ctorName, t.D, pol, depth); err != nil {
			return err
		}
		if err := checkArgTypePositivityMulti(indNames, ctorName, t.X, pol, depth); err != nil {
			return err
		}
		if err := checkArgTypePositivityMulti(indNames, ctorName, t.Y, pol, depth); err != nil {
			return err
		}
		return checkArgTypePositivityMulti(indNames, ctorName, t.P, pol, depth)

	default:
		// Try extension handlers
		if err, handled := checkArgTypePositivityExtension(indNames[0], ctorName, ty, pol, depth); handled {
			return err
		}

		// Unknown term type - check if any mutual type occurs in negative position
		if pol == Negative {
			for _, indName := range indNames {
				if OccursIn(indName, ty) {
					return &PositivityError{
						IndName:     indName,
						Constructor: ctorName,
						Position:    fmt.Sprintf("argument %d (unknown node type %T)", depth, ty),
						Polarity:    pol,
					}
				}
			}
		}
		return nil
	}
}

// checkConstructorPositivity checks a single constructor type for strict positivity.
// Constructor types have the form (x1 : A1) -> ... -> (xn : An) -> T
// where each Ai must have T occurring only strictly positively.
func checkConstructorPositivity(indName, ctorName string, ty ast.Term) error {
	return checkConstructorArgs(indName, ctorName, ty, 0)
}

// checkConstructorArgs traverses the constructor type's argument structure.
// At the top level of a constructor, we're extracting argument types, not
// entering negative positions yet.
func checkConstructorArgs(indName, ctorName string, ty ast.Term, depth int) error {
	switch t := ty.(type) {
	case ast.Pi:
		// Check that the argument type has T occurring strictly positively
		if err := checkArgTypePositivity(indName, ctorName, t.A, Positive, depth); err != nil {
			return err
		}
		// Continue to next argument or result
		return checkConstructorArgs(indName, ctorName, t.B, depth+1)
	case ast.Global:
		// Result type - T itself is fine
		return nil
	case ast.App:
		// Result type like (T A) - check arguments don't have T in bad positions
		return checkArgTypePositivity(indName, ctorName, ty, Positive, depth)
	default:
		// Other result types
		return nil
	}
}

// checkArgTypePositivity checks that within an argument type, the inductive
// occurs only in strictly positive positions.
//
// Strict positivity for (A -> B) means:
// - The inductive must NOT occur anywhere in A (the domain)
// - The inductive must be strictly positive in B (the codomain)
//
// Once in Negative position (inside a domain), we stay negative - no double-flipping.
func checkArgTypePositivity(indName, ctorName string, ty ast.Term, pol Polarity, depth int) error {
	switch t := ty.(type) {
	case ast.Pi:
		// In a function type A -> B:
		// - A is in negative position (no occurrences allowed)
		// - B stays at current polarity
		//
		// When already negative, entering another domain stays negative.
		// This is because once inside a domain, ANY occurrence is bad.
		newPol := Negative
		if err := checkArgTypePositivity(indName, ctorName, t.A, newPol, depth); err != nil {
			return err
		}
		return checkArgTypePositivity(indName, ctorName, t.B, pol, depth)

	case ast.Global:
		// If we find the inductive type itself
		if t.Name == indName {
			if pol == Negative {
				return &PositivityError{
					IndName:     indName,
					Constructor: ctorName,
					Position:    fmt.Sprintf("argument %d", depth),
					Polarity:    pol,
				}
			}
		}
		return nil

	case ast.App:
		// Check both parts of an application
		if err := checkArgTypePositivity(indName, ctorName, t.T, pol, depth); err != nil {
			return err
		}
		return checkArgTypePositivity(indName, ctorName, t.U, pol, depth)

	case ast.Var:
		// Variables are fine (they refer to bound parameters)
		return nil

	case ast.Sort:
		// Universes are fine
		return nil

	case ast.Sigma:
		// Sigma types: both components must be checked
		if err := checkArgTypePositivity(indName, ctorName, t.A, pol, depth); err != nil {
			return err
		}
		return checkArgTypePositivity(indName, ctorName, t.B, pol, depth)

	case ast.Lam:
		// Lambda bodies are checked (shouldn't normally appear in constructor types)
		if t.Ann != nil {
			if err := checkArgTypePositivity(indName, ctorName, t.Ann, pol, depth); err != nil {
				return err
			}
		}
		return checkArgTypePositivity(indName, ctorName, t.Body, pol, depth)

	case ast.Pair:
		if err := checkArgTypePositivity(indName, ctorName, t.Fst, pol, depth); err != nil {
			return err
		}
		return checkArgTypePositivity(indName, ctorName, t.Snd, pol, depth)

	case ast.Fst:
		return checkArgTypePositivity(indName, ctorName, t.P, pol, depth)

	case ast.Snd:
		return checkArgTypePositivity(indName, ctorName, t.P, pol, depth)

	case ast.Let:
		if err := checkArgTypePositivity(indName, ctorName, t.Ann, pol, depth); err != nil {
			return err
		}
		if err := checkArgTypePositivity(indName, ctorName, t.Val, pol, depth); err != nil {
			return err
		}
		return checkArgTypePositivity(indName, ctorName, t.Body, pol, depth)

	case ast.Id:
		if err := checkArgTypePositivity(indName, ctorName, t.A, pol, depth); err != nil {
			return err
		}
		if err := checkArgTypePositivity(indName, ctorName, t.X, pol, depth); err != nil {
			return err
		}
		return checkArgTypePositivity(indName, ctorName, t.Y, pol, depth)

	case ast.Refl:
		if err := checkArgTypePositivity(indName, ctorName, t.A, pol, depth); err != nil {
			return err
		}
		return checkArgTypePositivity(indName, ctorName, t.X, pol, depth)

	case ast.J:
		// J eliminator - check all components
		if err := checkArgTypePositivity(indName, ctorName, t.A, pol, depth); err != nil {
			return err
		}
		if err := checkArgTypePositivity(indName, ctorName, t.C, pol, depth); err != nil {
			return err
		}
		if err := checkArgTypePositivity(indName, ctorName, t.D, pol, depth); err != nil {
			return err
		}
		if err := checkArgTypePositivity(indName, ctorName, t.X, pol, depth); err != nil {
			return err
		}
		if err := checkArgTypePositivity(indName, ctorName, t.Y, pol, depth); err != nil {
			return err
		}
		return checkArgTypePositivity(indName, ctorName, t.P, pol, depth)

	default:
		// Try extension handlers (e.g., cubical terms when built with -tags cubical)
		if err, handled := checkArgTypePositivityExtension(indName, ctorName, ty, pol, depth); handled {
			return err
		}

		// Unknown term type - be conservative
		// If the inductive occurs in this unknown term and we're in negative position, reject
		if OccursIn(indName, ty) && pol == Negative {
			return &PositivityError{
				IndName:     indName,
				Constructor: ctorName,
				Position:    fmt.Sprintf("argument %d (unknown node type %T)", depth, ty),
				Polarity:    pol,
			}
		}
		// If inductive doesn't occur, or we're in positive position, allow
		return nil
	}
}

// OccursIn checks if a global name occurs anywhere in a term.
//
// This function performs a recursive traversal of the term structure to
// determine whether the specified global name appears. It is used during
// strict positivity checking to detect occurrences of the inductive type
// being defined, and by IsRecursiveArg to identify recursive constructor
// arguments.
//
// The function conservatively returns false for unknown term types, which
// is safe because the positivity checker will catch unknown types separately.
//
// Example: For an inductive type List with constructor cons : A → List A → List A,
// OccursIn("List", consType) returns true because List appears in the argument type.
func OccursIn(name string, ty ast.Term) bool {
	switch t := ty.(type) {
	case ast.Var, ast.Sort:
		// Variables and sorts contain no global references
		return false
	case ast.Meta:
		// Check metavariable arguments (defensive - metas should be zonked before positivity check)
		for _, arg := range t.Args {
			if OccursIn(name, arg) {
				return true
			}
		}
		return false
	case ast.Global:
		return t.Name == name
	case ast.Pi:
		return OccursIn(name, t.A) || OccursIn(name, t.B)
	case ast.Lam:
		return (t.Ann != nil && OccursIn(name, t.Ann)) || OccursIn(name, t.Body)
	case ast.App:
		return OccursIn(name, t.T) || OccursIn(name, t.U)
	case ast.Sigma:
		return OccursIn(name, t.A) || OccursIn(name, t.B)
	case ast.Pair:
		return OccursIn(name, t.Fst) || OccursIn(name, t.Snd)
	case ast.Fst:
		return OccursIn(name, t.P)
	case ast.Snd:
		return OccursIn(name, t.P)
	case ast.Let:
		return OccursIn(name, t.Ann) || OccursIn(name, t.Val) || OccursIn(name, t.Body)
	case ast.Id:
		return OccursIn(name, t.A) || OccursIn(name, t.X) || OccursIn(name, t.Y)
	case ast.Refl:
		return OccursIn(name, t.A) || OccursIn(name, t.X)
	case ast.J:
		return OccursIn(name, t.A) || OccursIn(name, t.C) || OccursIn(name, t.D) ||
			OccursIn(name, t.X) || OccursIn(name, t.Y) || OccursIn(name, t.P)
	default:
		// Try extension handlers (e.g., cubical terms when built with -tags cubical)
		if occurs, handled := occursInExtension(name, ty); handled {
			return occurs
		}
		// Unknown term types default to false (conservative for OccursIn means assuming no occurrence)
		// This is safe because checkArgTypePositivity will catch unknown types separately
		return false
	}
}

// IsRecursiveArg checks if a constructor argument type contains a reference
// to the inductive type being defined.
//
// This function is used during recursor generation to identify which
// constructor arguments are recursive. Recursive arguments require induction
// hypotheses (IH) in the recursor's case types.
//
// For example, given:
//
//	Nat : Type
//	zero : Nat
//	succ : Nat → Nat
//
// In the succ constructor, the Nat argument is recursive (IsRecursiveArg returns true),
// so natElim's step case has type (n : Nat) → P n → P (succ n), where P n is the IH.
//
// Non-recursive arguments (like the A parameter in cons : A → List A → List A)
// simply pass through to the case type without generating an IH.
func IsRecursiveArg(indName string, argType ast.Term) bool {
	return OccursIn(indName, argType)
}
