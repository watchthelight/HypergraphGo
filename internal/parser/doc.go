// Package parser provides S-expression parsing for the HoTT kernel.
//
// This package converts textual S-expression syntax into [ast.Term] values
// that can be type-checked and evaluated. It supports both core type theory
// constructs and cubical type theory extensions.
//
// # Parser Types
//
// The main types are:
//
//   - [SExprParser] - stateful parser with position tracking
//   - [ParseError] - error type with position information
//
// # Key Functions
//
//   - [ParseTerm] - convenience function to parse a single term
//   - [NewSExprParser] - create parser for more control
//   - [SExprParser.Parse] - parse single term, ensure no trailing content
//   - [ParseMultiple] - parse a sequence of terms
//   - [FormatTerm] - serialize term back to S-expression
//
// # Syntax Overview
//
// The parser accepts Lisp-style S-expressions:
//
//	; Comments start with semicolon
//	Type                       ; Universe Type₀
//	Type0, Type1, Type2        ; Universe shortcuts
//	(Sort 3)                   ; Universe Type₃
//	(Var 0)                    ; De Bruijn variable (0 = innermost)
//	0, 1, 2                    ; Variable shortcuts
//	name                       ; Global reference
//	(Pi x A B)                 ; Dependent function type
//	(Lam x body)               ; Lambda abstraction
//	(Lam x [A] body)           ; Lambda with type annotation
//	(App f x)                  ; Application
//	(Sigma x A B)              ; Dependent pair type
//	(Pair a b)                 ; Pair constructor
//	(Fst p), (Snd p)           ; Projections
//	(Let x A v body)           ; Let binding
//	(Id A x y)                 ; Identity type
//	(Refl A x)                 ; Reflexivity
//	(J A C d x y p)            ; J eliminator
//
// # Cubical Extensions
//
// When cubical features are needed:
//
//	I, Interval                ; Interval type
//	i0, i1                     ; Interval endpoints
//	(IVar 0)                   ; Interval variable
//	(Path A x y)               ; Non-dependent path type
//	(PathP A x y)              ; Dependent path type
//	(PathLam i body)           ; Path abstraction
//	(PathApp p r)              ; Path application
//	(Transport A e)            ; Transport
//
// # Alternative Syntax
//
// The parser accepts various alternative notations:
//
//	λ, \, fn        → Lam
//	Π, Pi, ->       → Pi
//	Σ, Sigma        → Sigma
//	×               → Sigma (non-dependent)
//
// # Error Handling
//
// Parse errors include position information for diagnostics:
//
//	term, err := parser.ParseTerm(input)
//	if err != nil {
//	    if pe, ok := err.(*parser.ParseError); ok {
//	        fmt.Printf("Error at position %d: %s\n", pe.Pos, pe.Message)
//	    }
//	}
//
// # Round-Trip
//
// Terms can be formatted back to S-expressions:
//
//	term, _ := parser.ParseTerm("(Pi x Type (Var 0))")
//	s := parser.FormatTerm(term)
//	// s == "(Pi x Type0 (Var 0))"
//
// Note: Formatting may normalize some constructs (e.g., Type → Type0).
package parser
