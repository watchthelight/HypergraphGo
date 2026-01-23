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

	// Face formulas - no type variables
	case ast.FaceTop, ast.FaceBot, ast.FaceEq:
		return nil, true

	case ast.FaceAnd:
		if err := checkArgTypePositivityFace(indName, ctorName, t.Left, pol, depth); err != nil {
			return err, true
		}
		return checkArgTypePositivityFace(indName, ctorName, t.Right, pol, depth), true

	case ast.FaceOr:
		if err := checkArgTypePositivityFace(indName, ctorName, t.Left, pol, depth); err != nil {
			return err, true
		}
		return checkArgTypePositivityFace(indName, ctorName, t.Right, pol, depth), true

	// Partial types
	case ast.Partial:
		if err := checkArgTypePositivityFace(indName, ctorName, t.Phi, pol, depth); err != nil {
			return err, true
		}
		return checkArgTypePositivity(indName, ctorName, t.A, pol, depth), true

	case ast.System:
		for _, br := range t.Branches {
			if err := checkArgTypePositivityFace(indName, ctorName, br.Phi, pol, depth); err != nil {
				return err, true
			}
			if err := checkArgTypePositivity(indName, ctorName, br.Term, pol, depth); err != nil {
				return err, true
			}
		}
		return nil, true

	// Composition operations
	case ast.Comp:
		if err := checkArgTypePositivity(indName, ctorName, t.A, pol, depth); err != nil {
			return err, true
		}
		if err := checkArgTypePositivityFace(indName, ctorName, t.Phi, pol, depth); err != nil {
			return err, true
		}
		if err := checkArgTypePositivity(indName, ctorName, t.Tube, pol, depth); err != nil {
			return err, true
		}
		return checkArgTypePositivity(indName, ctorName, t.Base, pol, depth), true

	case ast.HComp:
		if err := checkArgTypePositivity(indName, ctorName, t.A, pol, depth); err != nil {
			return err, true
		}
		if err := checkArgTypePositivityFace(indName, ctorName, t.Phi, pol, depth); err != nil {
			return err, true
		}
		if err := checkArgTypePositivity(indName, ctorName, t.Tube, pol, depth); err != nil {
			return err, true
		}
		return checkArgTypePositivity(indName, ctorName, t.Base, pol, depth), true

	case ast.Fill:
		if err := checkArgTypePositivity(indName, ctorName, t.A, pol, depth); err != nil {
			return err, true
		}
		if err := checkArgTypePositivityFace(indName, ctorName, t.Phi, pol, depth); err != nil {
			return err, true
		}
		if err := checkArgTypePositivity(indName, ctorName, t.Tube, pol, depth); err != nil {
			return err, true
		}
		return checkArgTypePositivity(indName, ctorName, t.Base, pol, depth), true

	// Glue types
	case ast.Glue:
		if err := checkArgTypePositivity(indName, ctorName, t.A, pol, depth); err != nil {
			return err, true
		}
		for _, br := range t.System {
			if err := checkArgTypePositivityFace(indName, ctorName, br.Phi, pol, depth); err != nil {
				return err, true
			}
			if err := checkArgTypePositivity(indName, ctorName, br.T, pol, depth); err != nil {
				return err, true
			}
			if err := checkArgTypePositivity(indName, ctorName, br.Equiv, pol, depth); err != nil {
				return err, true
			}
		}
		return nil, true

	case ast.GlueElem:
		for _, br := range t.System {
			if err := checkArgTypePositivityFace(indName, ctorName, br.Phi, pol, depth); err != nil {
				return err, true
			}
			if err := checkArgTypePositivity(indName, ctorName, br.Term, pol, depth); err != nil {
				return err, true
			}
		}
		return checkArgTypePositivity(indName, ctorName, t.Base, pol, depth), true

	case ast.Unglue:
		if err := checkArgTypePositivity(indName, ctorName, t.Ty, pol, depth); err != nil {
			return err, true
		}
		return checkArgTypePositivity(indName, ctorName, t.G, pol, depth), true

	// Univalence
	case ast.UA:
		if err := checkArgTypePositivity(indName, ctorName, t.A, pol, depth); err != nil {
			return err, true
		}
		if err := checkArgTypePositivity(indName, ctorName, t.B, pol, depth); err != nil {
			return err, true
		}
		return checkArgTypePositivity(indName, ctorName, t.Equiv, pol, depth), true

	case ast.UABeta:
		if err := checkArgTypePositivity(indName, ctorName, t.Equiv, pol, depth); err != nil {
			return err, true
		}
		return checkArgTypePositivity(indName, ctorName, t.Arg, pol, depth), true

	default:
		return nil, false
	}
}

// checkArgTypePositivityFace handles positivity checking for face formulas.
func checkArgTypePositivityFace(indName, ctorName string, f ast.Face, pol Polarity, depth int) error {
	if f == nil {
		return nil
	}
	switch t := f.(type) {
	case ast.FaceTop, ast.FaceBot, ast.FaceEq:
		return nil
	case ast.FaceAnd:
		if err := checkArgTypePositivityFace(indName, ctorName, t.Left, pol, depth); err != nil {
			return err
		}
		return checkArgTypePositivityFace(indName, ctorName, t.Right, pol, depth)
	case ast.FaceOr:
		if err := checkArgTypePositivityFace(indName, ctorName, t.Left, pol, depth); err != nil {
			return err
		}
		return checkArgTypePositivityFace(indName, ctorName, t.Right, pol, depth)
	default:
		return nil
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

	// Face formulas - no type references
	case ast.FaceTop, ast.FaceBot, ast.FaceEq:
		return false, true

	case ast.FaceAnd:
		return occursInFace(name, t.Left) || occursInFace(name, t.Right), true

	case ast.FaceOr:
		return occursInFace(name, t.Left) || occursInFace(name, t.Right), true

	// Partial types
	case ast.Partial:
		return occursInFace(name, t.Phi) || OccursIn(name, t.A), true

	case ast.System:
		for _, br := range t.Branches {
			if occursInFace(name, br.Phi) || OccursIn(name, br.Term) {
				return true, true
			}
		}
		return false, true

	// Composition operations
	case ast.Comp:
		return OccursIn(name, t.A) || occursInFace(name, t.Phi) ||
			OccursIn(name, t.Tube) || OccursIn(name, t.Base), true

	case ast.HComp:
		return OccursIn(name, t.A) || occursInFace(name, t.Phi) ||
			OccursIn(name, t.Tube) || OccursIn(name, t.Base), true

	case ast.Fill:
		return OccursIn(name, t.A) || occursInFace(name, t.Phi) ||
			OccursIn(name, t.Tube) || OccursIn(name, t.Base), true

	// Glue types
	case ast.Glue:
		if OccursIn(name, t.A) {
			return true, true
		}
		for _, br := range t.System {
			if occursInFace(name, br.Phi) || OccursIn(name, br.T) || OccursIn(name, br.Equiv) {
				return true, true
			}
		}
		return false, true

	case ast.GlueElem:
		for _, br := range t.System {
			if occursInFace(name, br.Phi) || OccursIn(name, br.Term) {
				return true, true
			}
		}
		return OccursIn(name, t.Base), true

	case ast.Unglue:
		return OccursIn(name, t.Ty) || OccursIn(name, t.G), true

	// Univalence
	case ast.UA:
		return OccursIn(name, t.A) || OccursIn(name, t.B) || OccursIn(name, t.Equiv), true

	case ast.UABeta:
		return OccursIn(name, t.Equiv) || OccursIn(name, t.Arg), true

	// HIT path application
	case ast.HITApp:
		for _, arg := range t.Args {
			if OccursIn(name, arg) {
				return true, true
			}
		}
		for _, iarg := range t.IArgs {
			if OccursIn(name, iarg) {
				return true, true
			}
		}
		return false, true

	default:
		return false, false
	}
}

// occursInFace checks if a name occurs in a face formula.
func occursInFace(name string, f ast.Face) bool {
	if f == nil {
		return false
	}
	switch t := f.(type) {
	case ast.FaceTop, ast.FaceBot, ast.FaceEq:
		return false
	case ast.FaceAnd:
		return occursInFace(name, t.Left) || occursInFace(name, t.Right)
	case ast.FaceOr:
		return occursInFace(name, t.Left) || occursInFace(name, t.Right)
	default:
		return false
	}
}
