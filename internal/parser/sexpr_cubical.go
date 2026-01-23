package parser

import (
	"fmt"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

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
	case "HITApp":
		return p.parseHITApp()
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

// parseHITApp parses a HIT path constructor application.
// Syntax: (HITApp hitName ctorName (args...) (iargs...))
func (p *SExprParser) parseHITApp() (ast.Term, error) {
	p.skipWhitespace()
	hitName := p.parseAtom()
	if hitName == "" {
		return nil, &ParseError{Pos: p.pos, Message: "expected HIT name"}
	}

	p.skipWhitespace()
	ctorName := p.parseAtom()
	if ctorName == "" {
		return nil, &ParseError{Pos: p.pos, Message: "expected path constructor name"}
	}

	// Parse term arguments
	p.skipWhitespace()
	args, err := p.parseTermList()
	if err != nil {
		return nil, err
	}

	// Parse interval arguments
	p.skipWhitespace()
	iargs, err := p.parseTermList()
	if err != nil {
		return nil, err
	}

	return ast.HITApp{
		HITName: hitName,
		Ctor:    ctorName,
		Args:    args,
		IArgs:   iargs,
	}, nil
}

// parseTermList parses a list of terms enclosed in parentheses.
// Syntax: (term1 term2 ... termN)
func (p *SExprParser) parseTermList() ([]ast.Term, error) {
	c, ok := p.peek()
	if !ok {
		return nil, &ParseError{Pos: p.pos, Message: "expected '(' for term list"}
	}
	if c != '(' {
		return nil, &ParseError{Pos: p.pos, Message: "expected '(' for term list"}
	}
	p.consume()
	p.skipWhitespace()

	var terms []ast.Term
	for {
		c, ok := p.peek()
		if !ok {
			return nil, &ParseError{Pos: p.pos, Message: "unexpected EOF in term list"}
		}
		if c == ')' {
			p.consume()
			break
		}
		term, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		terms = append(terms, term)
		p.skipWhitespace()
	}
	return terms, nil
}

// formatCubicalTerm formats cubical-specific terms.
// Uses de Bruijn indices for all variables (suitable for round-tripping).
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
	case ast.HITApp:
		argsStr := "("
		for i, arg := range tm.Args {
			if i > 0 {
				argsStr += " "
			}
			argsStr += FormatTerm(arg)
		}
		argsStr += ")"
		iargsStr := "("
		for i, iarg := range tm.IArgs {
			if i > 0 {
				iargsStr += " "
			}
			iargsStr += FormatTerm(iarg)
		}
		iargsStr += ")"
		return fmt.Sprintf("(HITApp %s %s %s %s)", tm.HITName, tm.Ctor, argsStr, iargsStr)
	}
	return ""
}

// formatCubicalTermWithContext formats cubical-specific terms with a variable context.
func formatCubicalTermWithContext(t ast.Term, ctx []string) string {
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
		return fmt.Sprintf("(Path %s %s %s)", FormatTermWithContext(tm.A, ctx), FormatTermWithContext(tm.X, ctx), FormatTermWithContext(tm.Y, ctx))
	case ast.PathP:
		return fmt.Sprintf("(PathP %s %s %s)", FormatTermWithContext(tm.A, ctx), FormatTermWithContext(tm.X, ctx), FormatTermWithContext(tm.Y, ctx))
	case ast.PathLam:
		newCtx := append(ctx, tm.Binder)
		return fmt.Sprintf("(PathLam %s %s)", tm.Binder, FormatTermWithContext(tm.Body, newCtx))
	case ast.PathApp:
		return fmt.Sprintf("(PathApp %s %s)", FormatTermWithContext(tm.P, ctx), FormatTermWithContext(tm.R, ctx))
	case ast.Transport:
		return fmt.Sprintf("(Transport %s %s)", FormatTermWithContext(tm.A, ctx), FormatTermWithContext(tm.E, ctx))
	case ast.HITApp:
		argsStr := "("
		for i, arg := range tm.Args {
			if i > 0 {
				argsStr += " "
			}
			argsStr += FormatTermWithContext(arg, ctx)
		}
		argsStr += ")"
		iargsStr := "("
		for i, iarg := range tm.IArgs {
			if i > 0 {
				iargsStr += " "
			}
			iargsStr += FormatTermWithContext(iarg, ctx)
		}
		iargsStr += ")"
		return fmt.Sprintf("(HITApp %s %s %s %s)", tm.HITName, tm.Ctor, argsStr, iargsStr)
	}
	return ""
}
