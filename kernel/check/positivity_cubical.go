//go:build cubical

package check

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// checkArgTypePositivityExtension handles positivity checking for cubical terms.
// Returns (error, handled) where handled indicates if this was a cubical term.
func checkArgTypePositivityExtension(indName, ctorName string, ty ast.Term, pol Polarity, depth int) (error, bool) {
	switch t := ty.(type) {
	case ast.Interval:
		// Interval type contains no type variables
		return nil, true

	case ast.I0:
		// Interval endpoint
		return nil, true

	case ast.I1:
		// Interval endpoint
		return nil, true

	case ast.IVar:
		// Interval variable - doesn't contain type references
		return nil, true

	case ast.Path:
		// Path A x y - check A, x, y
		if err := checkArgTypePositivity(indName, ctorName, t.A, pol, depth); err != nil {
			return err, true
		}
		if err := checkArgTypePositivity(indName, ctorName, t.X, pol, depth); err != nil {
			return err, true
		}
		return checkArgTypePositivity(indName, ctorName, t.Y, pol, depth), true

	case ast.PathP:
		// PathP A x y - check A, x, y
		if err := checkArgTypePositivity(indName, ctorName, t.A, pol, depth); err != nil {
			return err, true
		}
		if err := checkArgTypePositivity(indName, ctorName, t.X, pol, depth); err != nil {
			return err, true
		}
		return checkArgTypePositivity(indName, ctorName, t.Y, pol, depth), true

	case ast.PathLam:
		// Path lambda - check the body
		return checkArgTypePositivity(indName, ctorName, t.Body, pol, depth), true

	case ast.PathApp:
		// Path application p @ r - check path and interval
		if err := checkArgTypePositivity(indName, ctorName, t.P, pol, depth); err != nil {
			return err, true
		}
		return checkArgTypePositivity(indName, ctorName, t.R, pol, depth), true

	case ast.Transport:
		// transport A e - check type family and element
		if err := checkArgTypePositivity(indName, ctorName, t.A, pol, depth); err != nil {
			return err, true
		}
		return checkArgTypePositivity(indName, ctorName, t.E, pol, depth), true

	default:
		return nil, false
	}
}

// occursInExtension handles occurrence checking for cubical terms.
// Returns (occurs, handled) where handled indicates if this was a cubical term.
func occursInExtension(name string, ty ast.Term) (bool, bool) {
	switch t := ty.(type) {
	case ast.Interval:
		return false, true

	case ast.I0:
		return false, true

	case ast.I1:
		return false, true

	case ast.IVar:
		return false, true

	case ast.Path:
		return OccursIn(name, t.A) || OccursIn(name, t.X) || OccursIn(name, t.Y), true

	case ast.PathP:
		return OccursIn(name, t.A) || OccursIn(name, t.X) || OccursIn(name, t.Y), true

	case ast.PathLam:
		return OccursIn(name, t.Body), true

	case ast.PathApp:
		return OccursIn(name, t.P) || OccursIn(name, t.R), true

	case ast.Transport:
		return OccursIn(name, t.A) || OccursIn(name, t.E), true

	default:
		return false, false
	}
}
