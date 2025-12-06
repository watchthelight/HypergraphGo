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

// CheckPositivity verifies that an inductive type definition satisfies
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
		// Unknown term type - be conservative and allow
		return nil
	}
}

// OccursIn checks if a global name occurs anywhere in a term.
func OccursIn(name string, ty ast.Term) bool {
	switch t := ty.(type) {
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
		return false
	}
}

// IsRecursiveArg checks if a constructor argument type contains a reference
// to the inductive type being defined.
func IsRecursiveArg(indName string, argType ast.Term) bool {
	return OccursIn(indName, argType)
}
