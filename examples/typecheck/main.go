// Example typecheck demonstrates basic type checking with the HoTT kernel.
//
// This example shows how to:
//   - Parse terms from S-expression syntax
//   - Create a type checker with primitive types
//   - Synthesize types for terms
//   - Check terms against expected types
//   - Handle type errors
package main

import (
	"fmt"
	"os"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/parser"
	"github.com/watchthelight/HypergraphGo/kernel/check"
)

func main() {
	// Create a type checker with built-in Nat and Bool types
	checker := check.NewCheckerWithPrimitives()

	fmt.Println("=== HoTT Type Checking Examples ===")
	fmt.Println()

	// Example 1: Type synthesis for identity function
	fmt.Println("1. Identity function on Type:")
	idTerm := parser.MustParse("(Lam A (Lam x 0))")
	idType, err := checker.Synth(nil, check.NoSpan(), idTerm)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Term: %s\n", parser.FormatTerm(idTerm))
		fmt.Printf("   Type: %s\n", parser.FormatTerm(idType))
	}
	fmt.Println()

	// Example 2: Type synthesis for a function type
	fmt.Println("2. Function type (A : Type) -> A -> A:")
	piTerm := parser.MustParse("(Pi A Type (Pi _ 0 1))")
	piType, err := checker.Synth(nil, check.NoSpan(), piTerm)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Term: %s\n", parser.FormatTerm(piTerm))
		fmt.Printf("   Type: %s (i.e., Type at universe level %d)\n",
			parser.FormatTerm(piType), piType.(ast.Sort).U)
	}
	fmt.Println()

	// Example 3: Type checking with natural numbers
	fmt.Println("3. Natural number zero:")
	zeroTerm := parser.MustParse("zero")
	zeroType, err := checker.Synth(nil, check.NoSpan(), zeroTerm)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Term: %s\n", parser.FormatTerm(zeroTerm))
		fmt.Printf("   Type: %s\n", parser.FormatTerm(zeroType))
	}
	fmt.Println()

	// Example 4: Successor applied to zero
	fmt.Println("4. Successor of zero:")
	succZero := parser.MustParse("(App succ zero)")
	succType, err := checker.Synth(nil, check.NoSpan(), succZero)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Term: %s\n", parser.FormatTerm(succZero))
		fmt.Printf("   Type: %s\n", parser.FormatTerm(succType))
	}
	fmt.Println()

	// Example 5: Identity type and reflexivity
	fmt.Println("5. Identity type and reflexivity:")
	idTypeTerm := parser.MustParse("(Id Nat zero zero)")
	idTypeType, err := checker.Synth(nil, check.NoSpan(), idTypeTerm)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Id type:  %s\n", parser.FormatTerm(idTypeTerm))
		fmt.Printf("   Has type: %s\n", parser.FormatTerm(idTypeType))
	}

	reflTerm := parser.MustParse("(Refl Nat zero)")
	reflType, err := checker.Synth(nil, check.NoSpan(), reflTerm)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Refl:     %s\n", parser.FormatTerm(reflTerm))
		fmt.Printf("   Has type: %s\n", parser.FormatTerm(reflType))
	}
	fmt.Println()

	// Example 6: Type error - applying non-function
	fmt.Println("6. Type error example (applying zero to zero):")
	badTerm := parser.MustParse("(App zero zero)")
	_, err = checker.Synth(nil, check.NoSpan(), badTerm)
	if err != nil {
		fmt.Printf("   Error (expected): %v\n", err)
	} else {
		fmt.Println("   Unexpectedly succeeded!")
	}
	fmt.Println()

	// Example 7: Checking against expected type
	fmt.Println("7. Checking term against expected type:")
	term := parser.MustParse("(Lam n 0)")
	expected := parser.MustParse("(Pi n Nat Nat)")
	checkErr := checker.Check(nil, check.NoSpan(), term, expected)
	if checkErr != nil {
		fmt.Printf("   Error: %v\n", checkErr)
	} else {
		fmt.Printf("   Term: %s\n", parser.FormatTerm(term))
		fmt.Printf("   Checks against: %s\n", parser.FormatTerm(expected))
		fmt.Println("   Result: OK")
	}
	fmt.Println()

	// Example 8: Boolean type
	fmt.Println("8. Boolean values:")
	trueTerm := parser.MustParse("true")
	trueType, err := checker.Synth(nil, check.NoSpan(), trueTerm)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   true : %s\n", parser.FormatTerm(trueType))
	}

	falseTerm := parser.MustParse("false")
	falseType, err := checker.Synth(nil, check.NoSpan(), falseTerm)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   false : %s\n", parser.FormatTerm(falseType))
	}

	fmt.Println("\n=== All examples completed ===")
	os.Exit(0)
}
