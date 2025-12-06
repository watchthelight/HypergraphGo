//go:build !cubical

package parser

import "github.com/watchthelight/HypergraphGo/internal/ast"

// parseCubicalAtom is a stub for non-cubical builds.
func (p *SExprParser) parseCubicalAtom(atom string) ast.Term {
	return nil
}

// parseCubicalForm is a stub for non-cubical builds.
func (p *SExprParser) parseCubicalForm(head string) (ast.Term, error) {
	return nil, nil
}

// formatCubicalTerm is a stub for non-cubical builds.
func formatCubicalTerm(t ast.Term) string {
	return ""
}
