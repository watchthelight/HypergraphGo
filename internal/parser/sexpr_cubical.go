//go:build cubical

package parser

import (
	"fmt"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

func init() {
	cubicalEnabled = true
}

// parseCubicalAtom parses cubical-specific atoms.
func (p *SExprParser) parseCubicalAtom(atom string) ast.Term {
	switch atom {
	case "I", "Interval":
		return ast.Interval{}
	case "i0":
		return ast.I0{}
	case "i1":
		return ast.I1{}
	}
	return nil
}

// parseCubicalForm parses cubical-specific compound forms.
func (p *SExprParser) parseCubicalForm(head string) (ast.Term, error) {
	switch head {
	case "IVar":
		return p.parseIVar()
	case "Path":
		return p.parsePath()
	case "PathP":
		return p.parsePathP()
	case "PathLam", "<>":
		return p.parsePathLam()
	case "PathApp", "@":
		return p.parsePathApp()
	case "Transport":
		return p.parseTransport()
	}
	return nil, nil
}

func (p *SExprParser) parseIVar() (ast.Term, error) {
	p.skipWhitespace()
	atom := p.parseAtom()
	var ix int
	if _, err := fmt.Sscanf(atom, "%d", &ix); err != nil {
		return nil, &ParseError{Pos: p.pos, Message: "expected interval variable index"}
	}
	return ast.IVar{Ix: ix}, nil
}

func (p *SExprParser) parsePath() (ast.Term, error) {
	p.skipWhitespace()
	a, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	x, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	y, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	return ast.Path{A: a, X: x, Y: y}, nil
}

func (p *SExprParser) parsePathP() (ast.Term, error) {
	p.skipWhitespace()
	a, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	x, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	y, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	return ast.PathP{A: a, X: x, Y: y}, nil
}

func (p *SExprParser) parsePathLam() (ast.Term, error) {
	p.skipWhitespace()
	binder := p.parseAtom()
	if binder == "" {
		binder = "i"
	}

	p.skipWhitespace()
	body, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	return ast.PathLam{Binder: binder, Body: body}, nil
}

func (p *SExprParser) parsePathApp() (ast.Term, error) {
	p.skipWhitespace()
	path, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	arg, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	return ast.PathApp{P: path, R: arg}, nil
}

func (p *SExprParser) parseTransport() (ast.Term, error) {
	p.skipWhitespace()
	a, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	e, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	return ast.Transport{A: a, E: e}, nil
}

// formatCubicalTerm formats cubical-specific terms.
func formatCubicalTerm(t ast.Term) string {
	switch tm := t.(type) {
	case ast.Interval:
		return "I"
	case ast.I0:
		return "i0"
	case ast.I1:
		return "i1"
	case ast.IVar:
		return fmt.Sprintf("(IVar %d)", tm.Ix)
	case ast.Path:
		return fmt.Sprintf("(Path %s %s %s)", FormatTerm(tm.A), FormatTerm(tm.X), FormatTerm(tm.Y))
	case ast.PathP:
		return fmt.Sprintf("(PathP %s %s %s)", FormatTerm(tm.A), FormatTerm(tm.X), FormatTerm(tm.Y))
	case ast.PathLam:
		return fmt.Sprintf("(PathLam %s %s)", tm.Binder, FormatTerm(tm.Body))
	case ast.PathApp:
		return fmt.Sprintf("(PathApp %s %s)", FormatTerm(tm.P), FormatTerm(tm.R))
	case ast.Transport:
		return fmt.Sprintf("(Transport %s %s)", FormatTerm(tm.A), FormatTerm(tm.E))
	}
	return ""
}
