// Example inductive demonstrates inductive types and their eliminators.
//
// This example shows how to:
//   - Define and use inductive types (Nat, Bool)
//   - Use eliminators (recursors) for dependent elimination
//   - Compute with inductive types via NbE
//   - Understand the structure of recursor types
//
// Inductive types are the foundation of dependent type theory, providing
// a way to define datatypes with guaranteed termination through structural
// recursion.
package main

import (
	"fmt"
	"os"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
	"github.com/watchthelight/HypergraphGo/internal/parser"
	"github.com/watchthelight/HypergraphGo/kernel/check"
)

func main() {
	// Create a type checker with built-in Nat and Bool
	checker := check.NewCheckerWithPrimitives()

	fmt.Println("=== Inductive Types Examples ===")
	fmt.Println()

	// Example 1: Natural numbers definition
	fmt.Println("1. Natural numbers (Nat):")
	fmt.Println("   Nat : Type")
	fmt.Println("   zero : Nat")
	fmt.Println("   succ : Nat → Nat")
	fmt.Println()

	// Type check constructors
	zeroType, _ := checker.Synth(nil, check.NoSpan(), parser.MustParse("zero"))
	succType, _ := checker.Synth(nil, check.NoSpan(), parser.MustParse("succ"))
	fmt.Printf("   zero : %s\n", parser.FormatTerm(zeroType))
	fmt.Printf("   succ : %s\n", parser.FormatTerm(succType))
	fmt.Println()

	// Example 2: Building natural numbers
	fmt.Println("2. Building natural numbers:")
	numbers := []struct {
		name string
		term string
	}{
		{"0", "zero"},
		{"1", "(App succ zero)"},
		{"2", "(App succ (App succ zero))"},
		{"3", "(App succ (App succ (App succ zero)))"},
	}
	for _, n := range numbers {
		term := parser.MustParse(n.term)
		nf := eval.EvalNBE(term)
		fmt.Printf("   %s = %s\n", n.name, parser.FormatTerm(nf))
	}
	fmt.Println()

	// Example 3: Natural number eliminator
	fmt.Println("3. Natural number eliminator (natElim):")
	fmt.Println("   natElim : (P : Nat → Type)")
	fmt.Println("           → P zero")
	fmt.Println("           → ((n : Nat) → P n → P (succ n))")
	fmt.Println("           → (n : Nat) → P n")
	fmt.Println()

	natElimType, _ := checker.Synth(nil, check.NoSpan(), parser.MustParse("natElim"))
	fmt.Printf("   natElim : %s\n", parser.FormatTerm(natElimType))
	fmt.Println()

	// Example 4: Computing with natElim
	fmt.Println("4. Computing with natElim (isZero predicate):")
	fmt.Println("   isZero = natElim (λn. Bool) true (λn. λih. false)")

	// Build the isZero function step by step
	// P = λn. Bool (the motive)
	motive := ast.Lam{Binder: "n", Body: ast.Global{Name: "Bool"}}
	// pz = true (base case)
	baseCase := ast.Global{Name: "true"}
	// ps = λn. λih. false (step case ignores IH)
	stepCase := ast.Lam{
		Binder: "n",
		Body: ast.Lam{
			Binder: "ih",
			Body:   ast.Global{Name: "false"},
		},
	}

	// Apply natElim to build isZero
	natElim := ast.Global{Name: "natElim"}
	isZero := ast.MkApps(natElim, motive, baseCase, stepCase)

	fmt.Println()
	fmt.Println("   Testing isZero:")

	// Apply to zero
	isZeroOfZero := ast.App{T: isZero, U: ast.Global{Name: "zero"}}
	resultZero := eval.EvalNBE(isZeroOfZero)
	fmt.Printf("   isZero zero = %s\n", parser.FormatTerm(resultZero))

	// Apply to one
	one := ast.App{T: ast.Global{Name: "succ"}, U: ast.Global{Name: "zero"}}
	isZeroOfOne := ast.App{T: isZero, U: one}
	resultOne := eval.EvalNBE(isZeroOfOne)
	fmt.Printf("   isZero (succ zero) = %s\n", parser.FormatTerm(resultOne))

	// Apply to two
	two := ast.App{T: ast.Global{Name: "succ"}, U: one}
	isZeroOfTwo := ast.App{T: isZero, U: two}
	resultTwo := eval.EvalNBE(isZeroOfTwo)
	fmt.Printf("   isZero (succ (succ zero)) = %s\n", parser.FormatTerm(resultTwo))
	fmt.Println()

	// Example 5: Boolean type
	fmt.Println("5. Boolean type (Bool):")
	fmt.Println("   Bool : Type")
	fmt.Println("   true : Bool")
	fmt.Println("   false : Bool")
	fmt.Println()

	trueType, _ := checker.Synth(nil, check.NoSpan(), parser.MustParse("true"))
	falseType, _ := checker.Synth(nil, check.NoSpan(), parser.MustParse("false"))
	fmt.Printf("   true : %s\n", parser.FormatTerm(trueType))
	fmt.Printf("   false : %s\n", parser.FormatTerm(falseType))
	fmt.Println()

	// Example 6: Boolean eliminator
	fmt.Println("6. Boolean eliminator (boolElim):")
	fmt.Println("   boolElim : (P : Bool → Type)")
	fmt.Println("            → P true")
	fmt.Println("            → P false")
	fmt.Println("            → (b : Bool) → P b")
	fmt.Println()

	boolElimType, _ := checker.Synth(nil, check.NoSpan(), parser.MustParse("boolElim"))
	fmt.Printf("   boolElim : %s\n", parser.FormatTerm(boolElimType))
	fmt.Println()

	// Example 7: Computing with boolElim (not function)
	fmt.Println("7. Computing with boolElim (not function):")
	fmt.Println("   not = boolElim (λb. Bool) false true")

	// Build not function
	boolMotive := ast.Lam{Binder: "b", Body: ast.Global{Name: "Bool"}}
	boolElimGlobal := ast.Global{Name: "boolElim"}
	notFn := ast.MkApps(boolElimGlobal, boolMotive, ast.Global{Name: "false"}, ast.Global{Name: "true"})

	fmt.Println()
	notTrue := ast.App{T: notFn, U: ast.Global{Name: "true"}}
	notFalse := ast.App{T: notFn, U: ast.Global{Name: "false"}}
	fmt.Printf("   not true  = %s\n", parser.FormatTerm(eval.EvalNBE(notTrue)))
	fmt.Printf("   not false = %s\n", parser.FormatTerm(eval.EvalNBE(notFalse)))
	fmt.Println()

	// Example 8: Dependent elimination (conceptual)
	fmt.Println("8. Dependent elimination (proof example):")
	fmt.Println("   Eliminators support dependent types, allowing proofs.")
	fmt.Println("   Example: prove that succ n ≠ zero for all n")
	fmt.Println("   The motive P can be a proposition that varies with n.")
	fmt.Println()

	fmt.Println("=== All inductive examples completed ===")
	os.Exit(0)
}
