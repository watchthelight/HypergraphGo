// Example cubical demonstrates cubical type theory features.
//
// This example shows how to work with:
//   - The interval type I and its endpoints i0, i1
//   - Path types Path A x y
//   - Dependent path types PathP
//   - Path abstractions and applications
//   - Transport along type families
//
// Cubical type theory provides a computational interpretation of the
// univalence axiom, making it possible to transport along equivalences.
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
	// Create a type checker with built-in primitives
	checker := check.NewCheckerWithPrimitives()

	fmt.Println("=== Cubical Type Theory Examples ===")
	fmt.Println()

	// Example 1: The interval type and its endpoints
	fmt.Println("1. Interval type I and endpoints:")
	fmt.Println("   I  - the abstract interval (not a proper type)")
	fmt.Println("   i0 - the left endpoint (0 : I)")
	fmt.Println("   i1 - the right endpoint (1 : I)")
	fmt.Println()

	// Example 2: Non-dependent path type
	fmt.Println("2. Path type (Path Nat zero zero):")
	pathType := parser.MustParse("(Path Nat zero zero)")
	pathTypeType, err := checker.Synth(nil, check.NoSpan(), pathType)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Path Nat zero zero : %s\n", parser.FormatTerm(pathTypeType))
	}
	fmt.Println()

	// Example 3: Reflexivity as a path
	fmt.Println("3. Reflexivity path <i> zero : Path Nat zero zero")
	reflPath := ast.PathLam{
		Binder: "i",
		Body:   ast.Global{Name: "zero"},
	}
	fmt.Printf("   Term: %s\n", parser.FormatTerm(reflPath))
	reflPathType, err := checker.Synth(nil, check.NoSpan(), reflPath)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   Type: %s\n", parser.FormatTerm(reflPathType))
	}
	fmt.Println()

	// Example 4: Path application
	fmt.Println("4. Path application (evaluating path at endpoints):")
	pathLam := parser.MustParse("(PathLam i zero)")
	fmt.Printf("   Path:   %s\n", parser.FormatTerm(pathLam))

	atI0 := ast.PathApp{P: pathLam, R: ast.I0{}}
	atI1 := ast.PathApp{P: pathLam, R: ast.I1{}}
	fmt.Printf("   @ i0 = %s\n", parser.FormatTerm(eval.EvalNBE(atI0)))
	fmt.Printf("   @ i1 = %s\n", parser.FormatTerm(eval.EvalNBE(atI1)))
	fmt.Println()

	// Example 5: Dependent path type (PathP)
	fmt.Println("5. Dependent path type (PathP):")
	fmt.Println("   PathP allows the type to vary along the path")
	fmt.Println("   PathP (λi. A) x y  where x : A[i0/i], y : A[i1/i]")
	fmt.Println()

	// Example 6: Transport
	fmt.Println("6. Transport operation:")
	fmt.Println("   transport (λi. A) e : A[i1/i]  given e : A[i0/i]")
	fmt.Println("   When A is constant: transport (λi. A) e ≡ e")

	// Build: transport (λi. Nat) zero
	transportTerm := ast.Transport{
		A: ast.PathLam{
			Binder: "i",
			Body:   ast.Global{Name: "Nat"},
		},
		E: ast.Global{Name: "zero"},
	}
	fmt.Printf("   Term:   %s\n", parser.FormatTerm(transportTerm))
	result := eval.EvalNBE(transportTerm)
	fmt.Printf("   Result: %s (constant type → identity)\n", parser.FormatTerm(result))
	fmt.Println()

	// Example 7: Path between Booleans
	fmt.Println("7. Path between boolean values:")
	boolPathType := parser.MustParse("(Path Bool true true)")
	fmt.Printf("   Path Bool true true : Type\n")
	boolRefl := ast.PathLam{
		Binder: "i",
		Body:   ast.Global{Name: "true"},
	}
	boolReflType, err := checker.Synth(nil, check.NoSpan(), boolRefl)
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   <i> true : %s\n", parser.FormatTerm(boolReflType))
	}
	_ = boolPathType
	fmt.Println()

	// Example 8: Function extensionality (conceptual)
	fmt.Println("8. Function extensionality (funext):")
	fmt.Println("   In cubical type theory, funext is derivable:")
	fmt.Println("   Given h : (x : A) → f x = g x")
	fmt.Println("   We get <i> λx. h x @ i : Path (A → B) f g")
	fmt.Println()

	fmt.Println("=== All cubical examples completed ===")
	os.Exit(0)
}
