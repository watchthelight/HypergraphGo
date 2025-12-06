package check

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// GenerateRecursorType generates the type of the eliminator for an inductive type.
//
// For an inductive T with constructors c1, ..., cn, the recursor has type:
//
//	T-elim : (P : T -> Type_j) -> case(c1) -> ... -> case(cn) -> (t : T) -> P t
//
// where case(ci) provides the elimination principle for constructor ci:
//   - Non-recursive args pass through as-is
//   - Recursive args (of type T) get an induction hypothesis (ih : P arg)
func GenerateRecursorType(ind *Inductive) ast.Term {
	// Build from inside out: (t : T) -> P t
	// P is at some de Bruijn index depending on how many cases we have

	numCases := len(ind.Constructors)
	indType := ast.Global{Name: ind.Name}

	// P is at index numCases (under all case binders)
	// t is at index 0 (innermost)
	// P t = App{Var{numCases}, Var{0}}
	pIdx := numCases
	target := ast.Pi{
		Binder: "t",
		A:      indType,
		B:      ast.App{T: ast.Var{Ix: pIdx}, U: ast.Var{Ix: 0}},
	}

	// Wrap with case binders from last to first
	result := ast.Term(target)
	for i := numCases - 1; i >= 0; i-- {
		caseTy := generateCaseType(ind, i, numCases-i-1)
		result = ast.Pi{
			Binder: "case_" + ind.Constructors[i].Name,
			A:      caseTy,
			B:      result,
		}
	}

	// Finally wrap with motive P : T -> Type
	motive := ast.Pi{
		Binder: "_",
		A:      indType,
		B:      ast.Sort{U: 0}, // Type_0 for simplicity
	}

	return ast.Pi{
		Binder: "P",
		A:      motive,
		B:      result,
	}
}

// generateCaseType generates the type of a case for a specific constructor.
//
// For constructor c : (x1 : A1) -> ... -> (xn : An) -> T:
//   - Non-recursive args: pass through
//   - Recursive args (xk : T): add ih_k : P xk
//   - Result: P (c x1 ... xn)
//
// caseIdx is the index of this constructor in the Constructors slice.
// casesAfter is how many case binders come after this one (affects de Bruijn indices).
func generateCaseType(ind *Inductive, caseIdx int, casesAfter int) ast.Term {
	ctor := ind.Constructors[caseIdx]
	indName := ind.Name

	// Extract constructor arguments
	args := extractPiArgs(ctor.Type)

	// Count recursive arguments
	recursiveArgs := countRecursiveArgs(indName, args)

	// P is at index: 1 (motive) + casesAfter (cases after this one)
	// After binding constructor args and IHs, we need to adjust
	pBaseIdx := 1 + casesAfter

	// Build the case type from inside out
	return buildCaseType(indName, args, recursiveArgs, pBaseIdx, ctor.Name)
}

// PiArg represents a single argument in a Pi type chain.
type PiArg struct {
	Name string
	Type ast.Term
}

// extractPiArgs extracts the arguments from a constructor type (Pi chain).
func extractPiArgs(ty ast.Term) []PiArg {
	var args []PiArg
	current := ty
	for {
		if pi, ok := current.(ast.Pi); ok {
			args = append(args, PiArg{Name: pi.Binder, Type: pi.A})
			current = pi.B
		} else {
			break
		}
	}
	return args
}

// countRecursiveArgs counts how many arguments are of the inductive type.
func countRecursiveArgs(indName string, args []PiArg) int {
	count := 0
	for _, arg := range args {
		if isRecursiveArgType(indName, arg.Type) {
			count++
		}
	}
	return count
}

// isRecursiveArgType checks if an argument type is the inductive type (or applied to it).
func isRecursiveArgType(indName string, ty ast.Term) bool {
	switch t := ty.(type) {
	case ast.Global:
		return t.Name == indName
	case ast.App:
		// Check if head is the inductive (e.g., List A where List is our inductive)
		return isAppOfGlobal(t, indName)
	default:
		return false
	}
}

// buildCaseType constructs the full case type for a constructor.
func buildCaseType(indName string, args []PiArg, numRecursive int, pBaseIdx int, ctorName string) ast.Term {
	// Total binders: args + induction hypotheses
	// pIdx at result = pBaseIdx + numArgs + numRecursive

	numArgs := len(args)
	totalBinders := numArgs + numRecursive

	// Build result type: P (c x1 ... xn)
	// c is a global, x1...xn are variables at indices (totalBinders-1)...(numRecursive)
	// Note: innermost (last) arg is at index numRecursive (after all IHs)
	ctorApp := ast.Term(ast.Global{Name: ctorName})
	for i := 0; i < numArgs; i++ {
		// Arg i is at index totalBinders - 1 - i
		argIdx := totalBinders - 1 - i
		ctorApp = ast.App{T: ctorApp, U: ast.Var{Ix: argIdx}}
	}

	// P (c x1 ... xn) where P is at index pBaseIdx + totalBinders
	pIdx := pBaseIdx + totalBinders
	resultTy := ast.App{T: ast.Var{Ix: pIdx}, U: ctorApp}

	// Build from inside out: first IHs, then args
	result := ast.Term(resultTy)

	// Add IH binders for recursive args (in reverse order)
	ihCount := 0
	for i := numArgs - 1; i >= 0; i-- {
		if isRecursiveArgType(indName, args[i].Type) {
			// ih_i : P x_i
			// x_i is at index: ihCount (we're adding IH binders from inner to outer)
			// P is at index: pBaseIdx + (args remaining) + (ihs remaining including this)
			// Actually, let's recompute...

			// When adding this IH binder, we have:
			// - ihCount IH binders already added (inner)
			// - args[i+1:] args still to add (but they come after in outer position)
			// x_i will be at index ihCount when we're at this IH binder
			argOffset := ihCount
			// P is at pBaseIdx + total binders = pBaseIdx + numArgs + numRecursive
			// But we're building from inside, so we need to shift
			ihType := ast.App{T: ast.Var{Ix: pIdx - (numRecursive - ihCount)}, U: ast.Var{Ix: argOffset}}

			result = ast.Pi{
				Binder: "ih_" + args[i].Name,
				A:      ihType,
				B:      result,
			}
			ihCount++
		}
	}

	// Add arg binders (in reverse order, so outer first)
	for i := numArgs - 1; i >= 0; i-- {
		// Shift the arg type to account for binders added so far
		// The type needs to be shifted by the number of binders between original context and here
		argTy := args[i].Type // TODO: may need shifting in complex cases
		result = ast.Pi{
			Binder: args[i].Name,
			A:      argTy,
			B:      result,
		}
	}

	return result
}

// GenerateRecursorTypeSimple generates a simpler recursor type without complex index manipulation.
// This is an alternative implementation that's easier to understand but may have subtle bugs.
func GenerateRecursorTypeSimple(ind *Inductive) ast.Term {
	// For now, use hand-crafted types for common cases
	switch ind.Name {
	case "Nat":
		return mkNatElimType()
	case "Bool":
		return mkBoolElimType()
	default:
		// Fall back to generic generation
		return GenerateRecursorType(ind)
	}
}
