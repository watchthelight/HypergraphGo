// Package script provides parsing and execution of tactic scripts.
// This file contains fuzz tests for the script parser.
package script

import (
	"strings"
	"testing"
)

// FuzzParseString fuzzes the script parser with random inputs.
// It verifies that the parser never panics on arbitrary input.
func FuzzParseString(f *testing.F) {
	// Seed corpus with valid scripts
	seeds := []string{
		// Empty and whitespace
		"",
		"   ",
		"\n\n\n",
		"\t\t\t",

		// Comments only
		"-- comment",
		"-- comment\n-- another",

		// Simple theorem
		`Theorem test : Type
Proof
Qed`,

		// Multiple tactics
		`Theorem id : (Pi A Type (Pi x (Var 0) (Var 1)))
Proof
  intro A
  intro x
  exact (Var 0)
Qed`,

		// Multiple theorems
		`Theorem t1 : Unit
Proof
  constructor
Qed

Theorem t2 : Unit
Proof
  trivial
Qed`,

		// With comments
		`-- Main theorem
Theorem test : Type
Proof
  -- Apply intro
  intro x
  -- Finish proof
  exact (Var 0)
Qed`,

		// Various tactics
		`Theorem test : (Pi A Type A)
Proof
  intros
  assumption
Qed`,

		`Theorem test : (Id Nat zero zero)
Proof
  reflexivity
Qed`,

		`Theorem test : (Sigma A Type A)
Proof
  split
  exact Type
  exact Type
Qed`,

		// Tactic with term argument
		`Theorem test : (Pi A Type A)
Proof
  intro A
  exact (Var 0)
Qed`,

		// Apply tactic
		`Theorem test : (Pi A Type (Pi B Type A))
Proof
  intro A
  intro B
  apply (Var 1)
Qed`,

		// Rewrite tactic
		`Theorem test : (Pi A Type (Pi B Type (Pi eq (Id Type A B) A)))
Proof
  intro A
  intro B
  intro eq
  rewrite eq
  exact (Var 2)
Qed`,

		// Exists tactic
		`Theorem test : (Sigma A Type A)
Proof
  exists Type
  exact Type
Qed`,

		// Destruct and cases
		`Theorem test : (Pi s (Sum Unit Unit) Unit)
Proof
  intro s
  destruct s
  trivial
  trivial
Qed`,

		// Induction
		`Theorem test : (Pi n Nat Nat)
Proof
  intro n
  induction n
  exact zero
  exact (succ (Var 0))
Qed`,

		// Contradiction
		`Theorem test : (Pi e Empty Unit)
Proof
  intro e
  contradiction
Qed`,

		// Left and right
		`Theorem test : (Sum Unit Unit)
Proof
  left
  constructor
Qed`,

		// Simpl and auto
		`Theorem test : Unit
Proof
  simpl
  auto
Qed`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// The parser should never panic
		_, _ = ParseString(input)
	})
}

// FuzzMalformedScript tests malformed script inputs.
func FuzzMalformedScript(f *testing.F) {
	malformed := []string{
		// Missing components
		"Theorem",
		"Theorem test",
		"Theorem test :",
		"Theorem : Type",
		"Proof",
		"Qed",

		// Incomplete theorem
		"Theorem test : Type",
		"Theorem test : Type\nProof",

		// Missing Proof block
		"Theorem test : Type\nQed",

		// Double Proof/Qed
		"Theorem test : Type\nProof\nProof\nQed",
		"Theorem test : Type\nProof\nQed\nQed",

		// Invalid type expression
		"Theorem test : (\nProof\nQed",
		"Theorem test : (Pi\nProof\nQed",
		"Theorem test : (Pi x\nProof\nQed",

		// Unexpected tokens
		"Proof\nTheorem test : Type\nQed",
		"Qed\nTheorem test : Type\nProof",

		// Very long name
		"Theorem " + strings.Repeat("x", 10000) + " : Type\nProof\nQed",

		// Very long type
		"Theorem test : " + strings.Repeat("(Pi x Type ", 100) + "Type" + strings.Repeat(")", 100) + "\nProof\nQed",

		// Control characters
		"Theorem\x00test : Type\nProof\nQed",
		"Theorem test : Type\nProof\n\x00tactic\nQed",

		// Unicode in names
		"Theorem αβγ : Type\nProof\nQed",
		"Theorem test : Type\nProof\nδεζ\nQed",

		// Tab characters
		"Theorem\ttest\t:\tType\nProof\nQed",

		// Nested comments (not supported)
		"-- outer -- inner\nTheorem test : Type\nProof\nQed",

		// Empty lines everywhere
		"\n\n\nTheorem test : Type\n\n\nProof\n\n\n\n\nQed\n\n\n",

		// Colon in different positions
		"Theorem : test Type\nProof\nQed",
		"Theorem test Type :\nProof\nQed",

		// Multiple theorems with same name (valid but edge case)
		"Theorem x : Type\nProof\nQed\nTheorem x : Type\nProof\nQed",

		// Deeply nested parens in type
		"Theorem test : " + strings.Repeat("(", 1000) + "Type" + strings.Repeat(")", 1000) + "\nProof\nQed",

		// Empty tactic line
		"Theorem test : Type\nProof\n\nQed",

		// Only whitespace as tactic
		"Theorem test : Type\nProof\n   \nQed",

		// Case sensitivity tests
		"THEOREM test : Type\nPROOF\nQED",
		"theorem test : Type\nproof\nqed",
	}

	for _, m := range malformed {
		f.Add(m)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// The parser should never panic, even on malformed input
		_, _ = ParseString(input)
	})
}
