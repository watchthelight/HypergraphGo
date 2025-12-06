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
//
// The universe j for the motive is derived from the inductive's type.
func GenerateRecursorType(ind *Inductive) ast.Term {
	// Extract the universe level from the inductive's type
	motiveUniverse := extractUniverseLevel(ind.Type)

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

	// Finally wrap with motive P : T -> Type_j
	// The motive universe is at least the inductive's universe
	motive := ast.Pi{
		Binder: "_",
		A:      indType,
		B:      ast.Sort{U: motiveUniverse},
	}

	return ast.Pi{
		Binder: "P",
		A:      motive,
		B:      result,
	}
}

// extractUniverseLevel extracts the universe level from a type.
// For Sort{U: n}, returns n. For other types, returns 0 as a fallback.
func extractUniverseLevel(ty ast.Term) ast.Level {
	switch t := ty.(type) {
	case ast.Sort:
		return t.U
	default:
		// Fallback for non-Sort types (shouldn't happen for validated inductives)
		return 0
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
//
// For constructor c : (x1 : A1) -> ... -> (xn : An) -> T where some Ai = T (recursive):
//
// Case type structure (built outside-in for clarity):
//
//	(x1 : A1) -> [ih1 : P x1] -> ... -> (xn : An) -> [ihn : P xn] -> P (c x1 ... xn)
//
// where [ih_i : P x_i] is present only for recursive arguments (Ai = T or App of T).
//
// De Bruijn indices are computed relative to the depth at each position:
//   - pBaseIdx is P's index before any case binders are added
//   - At depth d (under d binders), P is at index pBaseIdx + d
//   - Argument x_i bound at depth d_i is at index (current_depth - d_i - 1) when referenced
func buildCaseType(indName string, args []PiArg, numRecursive int, pBaseIdx int, ctorName string) ast.Term {
	numArgs := len(args)

	// Build the binder structure: for each arg, we add the arg binder,
	// and if recursive, also an IH binder immediately after.
	// Track which depth each original argument is bound at.
	type binderInfo struct {
		name        string
		ty          ast.Term
		isIH        bool
		argIdx      int // which original arg this refers to (-1 for IH)
		boundAtArg  int // for IH: which arg's IH this is
	}

	var binders []binderInfo
	argDepths := make([]int, numArgs) // depth at which arg i is bound

	depth := 0
	for i, arg := range args {
		// Add argument binder
		argDepths[i] = depth
		binders = append(binders, binderInfo{
			name:   arg.Name,
			ty:     arg.Type,
			isIH:   false,
			argIdx: i,
		})
		depth++

		// If recursive, add IH binder immediately after
		if isRecursiveArgType(indName, arg.Type) {
			// IH type: P x_i
			// At this point, x_i was just bound (at depth-1), so it's at index 0
			// P is at pBaseIdx + depth (we're about to add the IH binder)
			pIdx := pBaseIdx + depth
			ihType := ast.App{T: ast.Var{Ix: pIdx}, U: ast.Var{Ix: 0}}
			binders = append(binders, binderInfo{
				name:       "ih_" + arg.Name,
				ty:         ihType,
				isIH:       true,
				boundAtArg: i,
			})
			depth++
		}
	}

	totalBinders := depth

	// Build result type: P (c x1 ... xn)
	// P is at index pBaseIdx + totalBinders
	// Each x_i is at index (totalBinders - argDepths[i] - 1)
	pIdxAtResult := pBaseIdx + totalBinders

	ctorApp := ast.Term(ast.Global{Name: ctorName})
	for i := 0; i < numArgs; i++ {
		argVarIdx := totalBinders - argDepths[i] - 1
		ctorApp = ast.App{T: ctorApp, U: ast.Var{Ix: argVarIdx}}
	}

	var result ast.Term = ast.App{T: ast.Var{Ix: pIdxAtResult}, U: ctorApp}

	// Now wrap with binders from inside out (reverse order)
	for i := len(binders) - 1; i >= 0; i-- {
		b := binders[i]
		result = ast.Pi{
			Binder: b.name,
			A:      b.ty,
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
