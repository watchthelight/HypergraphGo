// Package parser provides parsing utilities for the HoTT kernel.
// This file implements parsing for surface syntax with holes and implicits.

package parser

import (
	"fmt"
	"strconv"
	"unicode"

	"github.com/watchthelight/HypergraphGo/internal/elab"
)

// SurfaceParser parses surface syntax terms.
type SurfaceParser struct {
	input string
	pos   int
}

// NewSurfaceParser creates a new surface syntax parser.
func NewSurfaceParser(input string) *SurfaceParser {
	return &SurfaceParser{input: input, pos: 0}
}

// ParseSurface parses the input and returns a surface term.
func (p *SurfaceParser) ParseSurface() (elab.STerm, error) {
	p.skipWhitespace()
	term, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	if p.pos < len(p.input) {
		return nil, &ParseError{Pos: p.pos, Message: "unexpected characters after term"}
	}
	return term, nil
}

// ParseSurfaceTerm parses the input string as a surface term.
func ParseSurfaceTerm(input string) (elab.STerm, error) {
	return NewSurfaceParser(input).ParseSurface()
}

func (p *SurfaceParser) skipWhitespace() {
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if c == ';' || (c == '-' && p.pos+1 < len(p.input) && p.input[p.pos+1] == '-') {
			// Skip comment to end of line
			for p.pos < len(p.input) && p.input[p.pos] != '\n' {
				p.pos++
			}
		} else if unicode.IsSpace(rune(c)) {
			p.pos++
		} else {
			break
		}
	}
}

func (p *SurfaceParser) peek() (byte, bool) {
	if p.pos >= len(p.input) {
		return 0, false
	}
	return p.input[p.pos], true
}

func (p *SurfaceParser) peekN(n int) string {
	if p.pos+n > len(p.input) {
		return p.input[p.pos:]
	}
	return p.input[p.pos : p.pos+n]
}

func (p *SurfaceParser) consume() byte {
	c := p.input[p.pos]
	p.pos++
	return c
}

func (p *SurfaceParser) expect(expected byte) error {
	if c, ok := p.peek(); !ok {
		return &ParseError{Pos: p.pos, Message: fmt.Sprintf("expected '%c', got EOF", expected)}
	} else if c != expected {
		return &ParseError{Pos: p.pos, Message: fmt.Sprintf("expected '%c', got '%c'", expected, c)}
	}
	p.consume()
	return nil
}

func (p *SurfaceParser) isAtomChar(c byte) bool {
	return c != '(' && c != ')' && c != '{' && c != '}' && c != '\\' && c != '.' && c != ':' && !unicode.IsSpace(rune(c)) && c != ';'
}

func (p *SurfaceParser) parseAtom() string {
	start := p.pos
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if !p.isAtomChar(c) {
			break
		}
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *SurfaceParser) parseTerm() (elab.STerm, error) {
	p.skipWhitespace()

	// Parse a base term, then check for arrows and applications
	term, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	// Check for application spine or arrow
	for {
		p.skipWhitespace()
		c, ok := p.peek()
		if !ok {
			break
		}

		// Check for arrow
		if c == '-' && p.peekN(2) == "->" {
			p.pos += 2 // consume "->"
			p.skipWhitespace()
			cod, err := p.parseTerm()
			if err != nil {
				return nil, err
			}
			// Non-dependent function type: A -> B
			term = &elab.SPi{Binder: "_", Icity: elab.Explicit, Dom: term, Cod: cod}
			continue
		}

		// Check for implicit application {arg}
		if c == '{' {
			p.consume()
			p.skipWhitespace()
			arg, err := p.parseTerm()
			if err != nil {
				return nil, err
			}
			p.skipWhitespace()
			if err := p.expect('}'); err != nil {
				return nil, err
			}
			term = &elab.SApp{Fn: term, Arg: arg, Icity: elab.Implicit}
			continue
		}

		// Check for explicit application
		// Application: next term starts with '(' or identifier
		if c == '(' || p.isAtomStart(c) {
			savePos := p.pos
			arg, err := p.parseBase()
			if err != nil {
				p.pos = savePos
				break
			}
			term = &elab.SApp{Fn: term, Arg: arg, Icity: elab.Explicit}
			continue
		}

		break
	}

	return term, nil
}

func (p *SurfaceParser) isAtomStart(c byte) bool {
	return p.isAtomChar(c) && c != '?' && c != '_' || c == '?' || c == '_'
}

func (p *SurfaceParser) parseBase() (elab.STerm, error) {
	p.skipWhitespace()
	c, ok := p.peek()
	if !ok {
		return nil, &ParseError{Pos: p.pos, Message: "unexpected EOF"}
	}

	// Hole: _ or ?name
	if c == '_' {
		p.consume()
		return &elab.SHole{Name: ""}, nil
	}
	if c == '?' {
		p.consume()
		name := p.parseAtom()
		return &elab.SHole{Name: name}, nil
	}

	// Lambda: \x. body or \{x}. body
	if c == '\\' {
		return p.parseLambda()
	}

	// Parenthesized or compound form
	if c == '(' {
		return p.parseParenOrCompound()
	}

	// Implicit binder: {x : A} -> B
	if c == '{' {
		return p.parseImplicitPi()
	}

	// Simple atom
	return p.parseSimple()
}

func (p *SurfaceParser) parseLambda() (elab.STerm, error) {
	if err := p.expect('\\'); err != nil {
		return nil, err
	}
	p.skipWhitespace()

	// Check for implicit binder \{x}
	icity := elab.Explicit
	c, _ := p.peek()
	if c == '{' {
		icity = elab.Implicit
		p.consume()
		p.skipWhitespace()
	}

	binder := p.parseAtom()
	if binder == "" {
		binder = "_"
	}

	// Check for type annotation
	var ann elab.STerm
	p.skipWhitespace()
	c, _ = p.peek()
	if c == ':' {
		p.consume()
		p.skipWhitespace()
		var err error
		ann, err = p.parseTerm()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
	}

	// Close implicit brace if needed
	if icity == elab.Implicit {
		if err := p.expect('}'); err != nil {
			return nil, err
		}
		p.skipWhitespace()
	}

	// Expect dot
	if err := p.expect('.'); err != nil {
		return nil, err
	}

	p.skipWhitespace()
	body, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	return &elab.SLam{Binder: binder, Icity: icity, Ann: ann, Body: body}, nil
}

func (p *SurfaceParser) parseImplicitPi() (elab.STerm, error) {
	if err := p.expect('{'); err != nil {
		return nil, err
	}
	p.skipWhitespace()

	binder := p.parseAtom()
	if binder == "" {
		binder = "_"
	}

	p.skipWhitespace()
	if err := p.expect(':'); err != nil {
		return nil, err
	}

	p.skipWhitespace()
	dom, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if err := p.expect('}'); err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if err := p.expect('-'); err != nil {
		return nil, err
	}
	if err := p.expect('>'); err != nil {
		return nil, err
	}

	p.skipWhitespace()
	cod, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	return &elab.SPi{Binder: binder, Icity: elab.Implicit, Dom: dom, Cod: cod}, nil
}

func (p *SurfaceParser) parseParenOrCompound() (elab.STerm, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	p.skipWhitespace()

	// Check for explicit Pi: (x : A) -> B
	savePos := p.pos
	binder := p.parseAtom()
	p.skipWhitespace()
	c, _ := p.peek()
	if c == ':' && binder != "" {
		// Looks like dependent type
		p.consume()
		p.skipWhitespace()
		dom, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		if err := p.expect(')'); err != nil {
			return nil, err
		}
		p.skipWhitespace()
		if p.peekN(2) == "->" {
			p.pos += 2
			p.skipWhitespace()
			cod, err := p.parseTerm()
			if err != nil {
				return nil, err
			}
			return &elab.SPi{Binder: binder, Icity: elab.Explicit, Dom: dom, Cod: cod}, nil
		}
		// Not a Pi, reparse as something else
	}

	// Restore and check for S-expression form
	p.pos = savePos
	head := p.parseAtom()

	switch head {
	case "Type", "Sort":
		return p.parseSType()
	case "Pi", "->":
		return p.parseSPi()
	case "Lam", "lambda", "\\":
		return p.parseSLam()
	case "Sigma":
		return p.parseSSigma()
	case "Pair":
		return p.parseSPair()
	case "Fst":
		return p.parseSFst()
	case "Snd":
		return p.parseSSnd()
	case "Let":
		return p.parseSLet()
	case "Id":
		return p.parseSId()
	case "Refl":
		return p.parseSRefl()
	case "J":
		return p.parseSJ()
	case "Path":
		return p.parseSPath()
	case "":
		// Just parenthesized expression
		term, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		p.skipWhitespace()
		if err := p.expect(')'); err != nil {
			return nil, err
		}
		return term, nil
	default:
		// Treat as application with head as function
		fn := &elab.SVar{Name: head}
		var result elab.STerm = fn
		for {
			p.skipWhitespace()
			c, ok := p.peek()
			if !ok || c == ')' {
				break
			}
			arg, err := p.parseTerm()
			if err != nil {
				return nil, err
			}
			result = &elab.SApp{Fn: result, Arg: arg, Icity: elab.Explicit}
		}
		p.skipWhitespace()
		if err := p.expect(')'); err != nil {
			return nil, err
		}
		return result, nil
	}
}

func (p *SurfaceParser) parseSimple() (elab.STerm, error) {
	atom := p.parseAtom()
	if atom == "" {
		return nil, &ParseError{Pos: p.pos, Message: "expected atom"}
	}

	// Check for Type literals
	switch atom {
	case "Type", "Type0":
		return &elab.SType{Level: 0}, nil
	case "Type1":
		return &elab.SType{Level: 1}, nil
	case "Type2":
		return &elab.SType{Level: 2}, nil
	default:
		// Regular variable
		return &elab.SVar{Name: atom}, nil
	}
}

func (p *SurfaceParser) parseSType() (elab.STerm, error) {
	p.skipWhitespace()
	atom := p.parseAtom()
	var level uint
	if atom == "" {
		level = 0
	} else if n, err := strconv.Atoi(atom); err == nil && n >= 0 {
		level = uint(n)
	} else {
		level = 0
	}
	p.skipWhitespace()
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	return &elab.SType{Level: level}, nil
}

func (p *SurfaceParser) parseSPi() (elab.STerm, error) {
	p.skipWhitespace()
	binder := p.parseAtom()
	if binder == "" {
		binder = "_"
	}

	p.skipWhitespace()
	dom, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	cod, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return &elab.SPi{Binder: binder, Icity: elab.Explicit, Dom: dom, Cod: cod}, nil
}

func (p *SurfaceParser) parseSLam() (elab.STerm, error) {
	p.skipWhitespace()
	binder := p.parseAtom()
	if binder == "" {
		binder = "_"
	}

	p.skipWhitespace()
	body, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return &elab.SLam{Binder: binder, Icity: elab.Explicit, Ann: nil, Body: body}, nil
}

func (p *SurfaceParser) parseSSigma() (elab.STerm, error) {
	p.skipWhitespace()
	binder := p.parseAtom()
	if binder == "" {
		binder = "_"
	}

	p.skipWhitespace()
	fst, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	snd, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return &elab.SSigma{Binder: binder, Fst: fst, Snd: snd}, nil
}

func (p *SurfaceParser) parseSPair() (elab.STerm, error) {
	p.skipWhitespace()
	fst, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	snd, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return &elab.SPair{Fst: fst, Snd: snd}, nil
}

func (p *SurfaceParser) parseSFst() (elab.STerm, error) {
	p.skipWhitespace()
	pair, err := p.parseBase()
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	return &elab.SFst{Pair: pair}, nil
}

func (p *SurfaceParser) parseSSnd() (elab.STerm, error) {
	p.skipWhitespace()
	pair, err := p.parseBase()
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	if err := p.expect(')'); err != nil {
		return nil, err
	}
	return &elab.SSnd{Pair: pair}, nil
}

func (p *SurfaceParser) parseSLet() (elab.STerm, error) {
	p.skipWhitespace()
	binder := p.parseAtom()
	if binder == "" {
		binder = "_"
	}

	p.skipWhitespace()
	ann, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	val, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	body, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return &elab.SLet{Binder: binder, Ann: ann, Val: val, Body: body}, nil
}

func (p *SurfaceParser) parseSId() (elab.STerm, error) {
	p.skipWhitespace()
	a, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	x, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	y, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return &elab.SId{A: a, X: x, Y: y}, nil
}

func (p *SurfaceParser) parseSRefl() (elab.STerm, error) {
	p.skipWhitespace()
	a, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	x, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return &elab.SRefl{A: a, X: x}, nil
}

func (p *SurfaceParser) parseSJ() (elab.STerm, error) {
	p.skipWhitespace()
	a, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	c, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	d, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	x, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	y, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	prf, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return &elab.SJ{A: a, C: c, D: d, X: x, Y: y, P: prf}, nil
}

func (p *SurfaceParser) parseSPath() (elab.STerm, error) {
	p.skipWhitespace()
	a, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	x, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	y, err := p.parseBase()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return &elab.SPath{A: a, X: x, Y: y}, nil
}

// FormatSurfaceTerm formats a surface term as a string.
func FormatSurfaceTerm(t elab.STerm) string {
	if t == nil {
		return "nil"
	}

	switch tm := t.(type) {
	case *elab.SVar:
		return tm.Name
	case *elab.SType:
		if tm.Level == 0 {
			return "Type"
		}
		return fmt.Sprintf("Type%d", tm.Level)
	case *elab.SPi:
		if tm.Icity == elab.Implicit {
			return fmt.Sprintf("{%s : %s} -> %s", tm.Binder, FormatSurfaceTerm(tm.Dom), FormatSurfaceTerm(tm.Cod))
		}
		if tm.Binder == "_" {
			return fmt.Sprintf("%s -> %s", FormatSurfaceTerm(tm.Dom), FormatSurfaceTerm(tm.Cod))
		}
		return fmt.Sprintf("(%s : %s) -> %s", tm.Binder, FormatSurfaceTerm(tm.Dom), FormatSurfaceTerm(tm.Cod))
	case *elab.SLam:
		if tm.Icity == elab.Implicit {
			if tm.Ann != nil {
				return fmt.Sprintf("\\{%s : %s}. %s", tm.Binder, FormatSurfaceTerm(tm.Ann), FormatSurfaceTerm(tm.Body))
			}
			return fmt.Sprintf("\\{%s}. %s", tm.Binder, FormatSurfaceTerm(tm.Body))
		}
		if tm.Ann != nil {
			return fmt.Sprintf("\\%s : %s. %s", tm.Binder, FormatSurfaceTerm(tm.Ann), FormatSurfaceTerm(tm.Body))
		}
		return fmt.Sprintf("\\%s. %s", tm.Binder, FormatSurfaceTerm(tm.Body))
	case *elab.SApp:
		if tm.Icity == elab.Implicit {
			return fmt.Sprintf("%s {%s}", FormatSurfaceTerm(tm.Fn), FormatSurfaceTerm(tm.Arg))
		}
		return fmt.Sprintf("(%s %s)", FormatSurfaceTerm(tm.Fn), FormatSurfaceTerm(tm.Arg))
	case *elab.SSigma:
		return fmt.Sprintf("(Sigma %s %s %s)", tm.Binder, FormatSurfaceTerm(tm.Fst), FormatSurfaceTerm(tm.Snd))
	case *elab.SPair:
		return fmt.Sprintf("(Pair %s %s)", FormatSurfaceTerm(tm.Fst), FormatSurfaceTerm(tm.Snd))
	case *elab.SFst:
		return fmt.Sprintf("(Fst %s)", FormatSurfaceTerm(tm.Pair))
	case *elab.SSnd:
		return fmt.Sprintf("(Snd %s)", FormatSurfaceTerm(tm.Pair))
	case *elab.SLet:
		return fmt.Sprintf("(Let %s %s %s %s)", tm.Binder, FormatSurfaceTerm(tm.Ann), FormatSurfaceTerm(tm.Val), FormatSurfaceTerm(tm.Body))
	case *elab.SHole:
		if tm.Name == "" {
			return "_"
		}
		return "?" + tm.Name
	case *elab.SId:
		return fmt.Sprintf("(Id %s %s %s)", FormatSurfaceTerm(tm.A), FormatSurfaceTerm(tm.X), FormatSurfaceTerm(tm.Y))
	case *elab.SRefl:
		return fmt.Sprintf("(Refl %s %s)", FormatSurfaceTerm(tm.A), FormatSurfaceTerm(tm.X))
	case *elab.SJ:
		return fmt.Sprintf("(J %s %s %s %s %s %s)", FormatSurfaceTerm(tm.A), FormatSurfaceTerm(tm.C),
			FormatSurfaceTerm(tm.D), FormatSurfaceTerm(tm.X), FormatSurfaceTerm(tm.Y), FormatSurfaceTerm(tm.P))
	case *elab.SPath:
		return fmt.Sprintf("(Path %s %s %s)", FormatSurfaceTerm(tm.A), FormatSurfaceTerm(tm.X), FormatSurfaceTerm(tm.Y))
	default:
		return fmt.Sprintf("<%T>", t)
	}
}
