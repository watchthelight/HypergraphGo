//go:build !cubical

package check

import "github.com/watchthelight/HypergraphGo/internal/ast"

// checkArgTypePositivityExtension is the non-cubical stub that returns false.
func checkArgTypePositivityExtension(indName, ctorName string, ty ast.Term, pol Polarity, depth int) (error, bool) {
	return nil, false
}

// occursInExtension is the non-cubical stub that returns false.
func occursInExtension(name string, ty ast.Term) (bool, bool) {
	return false, false
}
