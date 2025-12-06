// Package parser provides parsing utilities for the HoTT kernel.
package parser

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// cubicalEnabled is set to true in sexpr_cubical.go when built with cubical tag.
var cubicalEnabled = false

// ParseError represents a parsing error with position information.
type ParseError struct {
	Pos     int
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at position %d: %s", e.Pos, e.Message)
}

// SExprParser parses S-expressions into AST terms.
type SExprParser struct {
	input string
	pos   int
}

// NewSExprParser creates a new S-expression parser.
func NewSExprParser(input string) *SExprParser {
	return &SExprParser{input: input, pos: 0}
}

// Parse parses the input and returns an AST term.
func (p *SExprParser) Parse() (ast.Term, error) {
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

// ParseTerm parses the input string as an AST term.
func ParseTerm(input string) (ast.Term, error) {
	return NewSExprParser(input).Parse()
}

func (p *SExprParser) skipWhitespace() {
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if c == ';' {
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

func (p *SExprParser) peek() (byte, bool) {
	if p.pos >= len(p.input) {
		return 0, false
	}
	return p.input[p.pos], true
}

func (p *SExprParser) consume() byte {
	c := p.input[p.pos]
	p.pos++
	return c
}

func (p *SExprParser) expect(expected byte) error {
	if c, ok := p.peek(); !ok {
		return &ParseError{Pos: p.pos, Message: fmt.Sprintf("expected '%c', got EOF", expected)}
	} else if c != expected {
		return &ParseError{Pos: p.pos, Message: fmt.Sprintf("expected '%c', got '%c'", expected, c)}
	}
	p.consume()
	return nil
}

func (p *SExprParser) parseAtom() string {
	start := p.pos
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if c == '(' || c == ')' || unicode.IsSpace(rune(c)) || c == ';' {
			break
		}
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *SExprParser) parseTerm() (ast.Term, error) {
	p.skipWhitespace()
	c, ok := p.peek()
	if !ok {
		return nil, &ParseError{Pos: p.pos, Message: "unexpected EOF"}
	}

	if c == '(' {
		return p.parseCompound()
	}
	return p.parseSimple()
}

func (p *SExprParser) parseSimple() (ast.Term, error) {
	atom := p.parseAtom()
	if atom == "" {
		return nil, &ParseError{Pos: p.pos, Message: "expected atom"}
	}

	// Check for numeric literal (de Bruijn index shorthand)
	if n, err := strconv.Atoi(atom); err == nil && n >= 0 {
		return ast.Var{Ix: n}, nil
	}

	// Special atoms
	switch atom {
	case "Type", "Type0":
		return ast.Sort{U: 0}, nil
	case "Type1":
		return ast.Sort{U: 1}, nil
	case "Type2":
		return ast.Sort{U: 2}, nil
	default:
		// Check cubical atoms
		if term := p.parseCubicalAtom(atom); term != nil {
			return term, nil
		}
		// Treat as global reference
		return ast.Global{Name: atom}, nil
	}
}

func (p *SExprParser) parseCompound() (ast.Term, error) {
	if err := p.expect('('); err != nil {
		return nil, err
	}
	p.skipWhitespace()

	head := p.parseAtom()
	if head == "" {
		return nil, &ParseError{Pos: p.pos, Message: "expected form name"}
	}

	var term ast.Term
	var err error

	switch head {
	case "Sort", "Type":
		term, err = p.parseSort()
	case "Var":
		term, err = p.parseVar()
	case "Global":
		term, err = p.parseGlobal()
	case "Pi", "->":
		term, err = p.parsePi()
	case "Lam", "λ", "\\", "lambda":
		term, err = p.parseLam()
	case "App":
		term, err = p.parseApp()
	case "Sigma", "Σ":
		term, err = p.parseSigma()
	case "Pair":
		term, err = p.parsePair()
	case "Fst":
		term, err = p.parseFst()
	case "Snd":
		term, err = p.parseSnd()
	case "Let":
		term, err = p.parseLet()
	case "Id":
		term, err = p.parseId()
	case "Refl":
		term, err = p.parseRefl()
	case "J":
		term, err = p.parseJ()
	default:
		// Try cubical forms
		term, err = p.parseCubicalForm(head)
		if err != nil {
			return nil, err
		}
		if term == nil {
			return nil, &ParseError{Pos: p.pos, Message: fmt.Sprintf("unknown form: %s", head)}
		}
	}

	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	if err := p.expect(')'); err != nil {
		return nil, err
	}

	return term, nil
}

func (p *SExprParser) parseSort() (ast.Term, error) {
	p.skipWhitespace()
	atom := p.parseAtom()
	if atom == "" {
		return ast.Sort{U: 0}, nil
	}
	level, err := strconv.Atoi(atom)
	if err != nil {
		return nil, &ParseError{Pos: p.pos, Message: "expected universe level"}
	}
	return ast.Sort{U: ast.Level(level)}, nil
}

func (p *SExprParser) parseVar() (ast.Term, error) {
	p.skipWhitespace()
	atom := p.parseAtom()
	ix, err := strconv.Atoi(atom)
	if err != nil {
		return nil, &ParseError{Pos: p.pos, Message: "expected de Bruijn index"}
	}
	return ast.Var{Ix: ix}, nil
}

func (p *SExprParser) parseGlobal() (ast.Term, error) {
	p.skipWhitespace()
	name := p.parseAtom()
	if name == "" {
		return nil, &ParseError{Pos: p.pos, Message: "expected global name"}
	}
	return ast.Global{Name: name}, nil
}

func (p *SExprParser) parsePi() (ast.Term, error) {
	p.skipWhitespace()
	binder := p.parseAtom()
	if binder == "" {
		binder = "_"
	}

	p.skipWhitespace()
	domain, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	codomain, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	return ast.Pi{Binder: binder, A: domain, B: codomain}, nil
}

func (p *SExprParser) parseLam() (ast.Term, error) {
	p.skipWhitespace()
	binder := p.parseAtom()
	if binder == "" {
		binder = "_"
	}

	p.skipWhitespace()
	// Check if there's an annotation
	var ann ast.Term
	c, ok := p.peek()
	if ok && c == '(' {
		// Could be annotation or body - peek ahead
		savePos := p.pos
		p.consume() // consume '('
		p.skipWhitespace()
		head := p.parseAtom()
		p.pos = savePos // restore

		// If it looks like a type form, treat as annotation
		if head == "Sort" || head == "Type" || head == "Pi" || head == "Sigma" || head == "Global" || head == "Id" || head == "Path" || head == "PathP" {
			ann, _ = p.parseTerm()
			p.skipWhitespace()
		}
	}

	body, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	return ast.Lam{Binder: binder, Ann: ann, Body: body}, nil
}

func (p *SExprParser) parseApp() (ast.Term, error) {
	p.skipWhitespace()
	fn, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	arg, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	return ast.App{T: fn, U: arg}, nil
}

func (p *SExprParser) parseSigma() (ast.Term, error) {
	p.skipWhitespace()
	binder := p.parseAtom()
	if binder == "" {
		binder = "_"
	}

	p.skipWhitespace()
	fstType, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	sndType, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	return ast.Sigma{Binder: binder, A: fstType, B: sndType}, nil
}

func (p *SExprParser) parsePair() (ast.Term, error) {
	p.skipWhitespace()
	fst, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	snd, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	return ast.Pair{Fst: fst, Snd: snd}, nil
}

func (p *SExprParser) parseFst() (ast.Term, error) {
	p.skipWhitespace()
	pair, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	return ast.Fst{P: pair}, nil
}

func (p *SExprParser) parseSnd() (ast.Term, error) {
	p.skipWhitespace()
	pair, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	return ast.Snd{P: pair}, nil
}

func (p *SExprParser) parseLet() (ast.Term, error) {
	p.skipWhitespace()
	binder := p.parseAtom()
	if binder == "" {
		binder = "_"
	}

	p.skipWhitespace()
	ann, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	val, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	body, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	return ast.Let{Binder: binder, Ann: ann, Val: val, Body: body}, nil
}

func (p *SExprParser) parseId() (ast.Term, error) {
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

	return ast.Id{A: a, X: x, Y: y}, nil
}

func (p *SExprParser) parseRefl() (ast.Term, error) {
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

	return ast.Refl{A: a, X: x}, nil
}

func (p *SExprParser) parseJ() (ast.Term, error) {
	p.skipWhitespace()
	a, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	c, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	p.skipWhitespace()
	d, err := p.parseTerm()
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

	p.skipWhitespace()
	prf, err := p.parseTerm()
	if err != nil {
		return nil, err
	}

	return ast.J{A: a, C: c, D: d, X: x, Y: y, P: prf}, nil
}


// FormatTerm formats an AST term as an S-expression string.
func FormatTerm(t ast.Term) string {
	if t == nil {
		return "nil"
	}

	switch tm := t.(type) {
	case ast.Sort:
		if tm.U == 0 {
			return "Type"
		}
		return fmt.Sprintf("(Sort %d)", tm.U)
	case ast.Var:
		return fmt.Sprintf("(Var %d)", tm.Ix)
	case ast.Global:
		return tm.Name
	case ast.Pi:
		return fmt.Sprintf("(Pi %s %s %s)", tm.Binder, FormatTerm(tm.A), FormatTerm(tm.B))
	case ast.Lam:
		if tm.Ann != nil {
			return fmt.Sprintf("(Lam %s %s %s)", tm.Binder, FormatTerm(tm.Ann), FormatTerm(tm.Body))
		}
		return fmt.Sprintf("(Lam %s %s)", tm.Binder, FormatTerm(tm.Body))
	case ast.App:
		return fmt.Sprintf("(App %s %s)", FormatTerm(tm.T), FormatTerm(tm.U))
	case ast.Sigma:
		return fmt.Sprintf("(Sigma %s %s %s)", tm.Binder, FormatTerm(tm.A), FormatTerm(tm.B))
	case ast.Pair:
		return fmt.Sprintf("(Pair %s %s)", FormatTerm(tm.Fst), FormatTerm(tm.Snd))
	case ast.Fst:
		return fmt.Sprintf("(Fst %s)", FormatTerm(tm.P))
	case ast.Snd:
		return fmt.Sprintf("(Snd %s)", FormatTerm(tm.P))
	case ast.Let:
		return fmt.Sprintf("(Let %s %s %s %s)", tm.Binder, FormatTerm(tm.Ann), FormatTerm(tm.Val), FormatTerm(tm.Body))
	case ast.Id:
		return fmt.Sprintf("(Id %s %s %s)", FormatTerm(tm.A), FormatTerm(tm.X), FormatTerm(tm.Y))
	case ast.Refl:
		return fmt.Sprintf("(Refl %s %s)", FormatTerm(tm.A), FormatTerm(tm.X))
	case ast.J:
		return fmt.Sprintf("(J %s %s %s %s %s %s)", FormatTerm(tm.A), FormatTerm(tm.C), FormatTerm(tm.D), FormatTerm(tm.X), FormatTerm(tm.Y), FormatTerm(tm.P))
	default:
		// Try cubical formatting
		if s := formatCubicalTerm(t); s != "" {
			return s
		}
		return fmt.Sprintf("<%T>", t)
	}
}

// MustParse parses a term or panics.
func MustParse(input string) ast.Term {
	term, err := ParseTerm(input)
	if err != nil {
		panic(err)
	}
	return term
}

// ParseMultiple parses multiple terms separated by whitespace.
func ParseMultiple(input string) ([]ast.Term, error) {
	var terms []ast.Term
	p := NewSExprParser(input)

	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) {
			break
		}
		term, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		terms = append(terms, term)
	}

	return terms, nil
}

// Normalize removes extra whitespace and standardizes formatting.
func Normalize(input string) string {
	return strings.TrimSpace(input)
}
