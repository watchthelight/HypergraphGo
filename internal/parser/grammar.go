// Package parser provides parsing utilities for the HoTT kernel.
//
// This file documents the S-expression grammar for HoTT terms. The parser
// accepts a Lisp-like syntax where compound forms are parenthesized and
// atoms represent variables, constants, and special keywords.
//
// # Grammar (BNF Notation)
//
// The following grammar specifies the complete syntax:
//
//	term         ::= simple | compound
//	simple       ::= number | keyword | global
//	compound     ::= '(' form-name term* ')'
//
//	number       ::= [0-9]+                    ; de Bruijn index
//	keyword      ::= 'Type' | 'Type0' | 'Type1' | 'Type2'
//	             |   'I' | 'Interval' | 'i0' | 'i1'
//	global       ::= identifier                ; any other atom
//	identifier   ::= [^()\s;]+                 ; non-whitespace, non-paren
//
//	form-name    ::= standard-form | cubical-form
//
// # Standard Type Theory Forms
//
//	standard-form ::= sort-form | var-form | global-form
//	              |   pi-form | lam-form | app-form
//	              |   sigma-form | pair-form | fst-form | snd-form
//	              |   let-form | id-form | refl-form | j-form
//
//	sort-form    ::= 'Sort' [number] | 'Type' [number]
//	var-form     ::= 'Var' number
//	global-form  ::= 'Global' identifier
//
//	pi-form      ::= ('Pi' | '->') binder term term
//	lam-form     ::= ('Lam' | 'λ' | '\\' | 'lambda') binder [term] term
//	app-form     ::= 'App' term term
//
//	sigma-form   ::= ('Sigma' | 'Σ') binder term term
//	pair-form    ::= 'Pair' term term
//	fst-form     ::= 'Fst' term
//	snd-form     ::= 'Snd' term
//
//	let-form     ::= 'Let' binder term term term
//	id-form      ::= 'Id' term term term
//	refl-form    ::= 'Refl' term term
//	j-form       ::= 'J' term term term term term term
//
//	binder       ::= identifier | '_'          ; optional, defaults to '_'
//
// # Cubical Type Theory Forms
//
//	cubical-form ::= ivar-form | path-form | pathp-form
//	             |   pathlam-form | pathapp-form | transport-form
//
//	ivar-form    ::= 'IVar' number
//	path-form    ::= 'Path' term term term
//	pathp-form   ::= 'PathP' term term term
//	pathlam-form ::= ('PathLam' | '<>') binder term
//	pathapp-form ::= ('PathApp' | '@') term term
//	transport-form ::= 'Transport' term term
//
// # Lexical Rules
//
// Whitespace:
//   - Spaces, tabs, and newlines are ignored between tokens
//   - Whitespace separates atoms and is required between adjacent atoms
//
// Comments:
//   - Line comments begin with ';' and extend to end of line
//   - Comments are treated as whitespace
//
// Atoms:
//   - Any sequence of characters excluding '(', ')', whitespace, and ';'
//   - Numeric atoms parse as de Bruijn indices (Var)
//   - Keyword atoms parse as their corresponding term type
//   - All other atoms parse as Global references
//
// # De Bruijn Index Shorthand
//
// Bare numeric atoms are parsed as de Bruijn indexed variables:
//
//	0        → Var{Ix: 0}    ; innermost bound variable
//	1        → Var{Ix: 1}    ; next outer binding
//	42       → Var{Ix: 42}   ; etc.
//
// This provides concise notation for variable references.
//
// # Form Aliases
//
// Several forms accept alternative keywords for convenience:
//
//	Pi:      'Pi', '->'
//	Lam:     'Lam', 'λ', '\', 'lambda'
//	Sigma:   'Sigma', 'Σ'
//	PathLam: 'PathLam', '<>'
//	PathApp: 'PathApp', '@'
//
// # Examples
//
// Identity function on Type:
//
//	(Lam A (Lam x 0))
//	; or equivalently:
//	(λ A (λ x 0))
//
// Dependent function type (A : Type) → A → A:
//
//	(Pi A Type (Pi _ 0 1))
//	; or equivalently:
//	(-> A Type (-> _ 0 1))
//
// Pair of natural numbers:
//
//	(Pair zero (succ zero))
//
// Path type Path Nat zero zero:
//
//	(Path Nat zero zero)
//
// Path abstraction <i> x:
//
//	(PathLam i x)
//	; or equivalently:
//	(<> i x)
//
// Transport along a type family:
//
//	(Transport (PathLam i A) e)
//
// Identity type with reflexivity:
//
//	(Id Nat zero zero)
//	(Refl Nat zero)
//
// J eliminator application:
//
//	(J A C d x y p)
//
// # Error Handling
//
// Parse errors include position information:
//
//	type ParseError struct {
//	    Pos     int      // byte offset in input
//	    Message string   // description of error
//	}
//
// Common error conditions:
//   - Unexpected EOF while parsing
//   - Missing closing parenthesis
//   - Unknown form name
//   - Invalid de Bruijn index (non-numeric where number expected)
//
// # Formatting
//
// The FormatTerm function converts AST terms back to S-expression syntax,
// providing round-trip support. The output uses canonical form names
// (Sort, Pi, Lam, etc.) rather than aliases.
package parser
