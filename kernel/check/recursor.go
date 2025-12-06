package check

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/kernel/subst"
)

// GenerateRecursorType generates the type of the eliminator for an inductive type.
//
// For an inductive T with constructors c1, ..., cn, the recursor has type:
//
//	T-elim : (P : T -> Type_j) -> case(c1) -> ... -> case(cn) -> (t : T) -> P t
//
// For parameterized inductives like List : Type -> Type:
//
//	listElim : (A : Type) -> (P : List A -> Type) -> case_nil -> case_cons -> (xs : List A) -> P xs
//
// where case(ci) provides the elimination principle for constructor ci:
//   - Non-recursive args pass through as-is
//   - Recursive args (of type T) get an induction hypothesis (ih : P arg)
//
// The universe j for the motive is derived from the inductive's type.
func GenerateRecursorType(ind *Inductive) ast.Term {
	// Extract the universe level from the inductive's type
	motiveUniverse := extractUniverseLevel(ind.Type)
	numParams := ind.NumParams
	numCases := len(ind.Constructors)

	// Build the applied inductive type: T A_1 ... A_k where A_i are param variables
	// At the target position (after params, P, and cases), param_i is at index:
	//   numCases + 1 + (numParams - i - 1) from the target body
	appliedIndAtTarget := buildAppliedInductive(ind.Name, numParams, numCases)

	// Build from inside out: (t : T params...) -> P t
	// P is at index numCases (under all case binders, before target)
	// t is at index 0 (innermost)
	pIdx := numCases
	target := ast.Pi{
		Binder: "t",
		A:      appliedIndAtTarget,
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

	// Build motive type: P : T params... -> Type_j
	// At the motive position, params are at indices 0..numParams-1 (just bound)
	motiveType := buildMotiveType(ind.Name, numParams, motiveUniverse)
	result = ast.Pi{
		Binder: "P",
		A:      motiveType,
		B:      result,
	}

	// Wrap with parameter binders (outermost)
	for i := numParams - 1; i >= 0; i-- {
		// Parameter types need shifting since we're adding outer binders
		// When we add param_i, there are (numParams - i - 1) params already bound outside
		shiftedParamType := subst.Shift(numParams-i-1, 0, ind.ParamTypes[i])
		result = ast.Pi{
			Binder: paramName(i),
			A:      shiftedParamType,
			B:      result,
		}
	}

	return result
}

// paramName returns the name for parameter i (A, B, C, ...)
func paramName(i int) string {
	if i < 26 {
		return string(rune('A' + i))
	}
	return "P" + string(rune('0'+i-26))
}

// buildAppliedInductive builds T param_{k-1} ... param_0
// where params are de Bruijn variables at the appropriate indices for the target position.
// At target body position (under numParams params, 1 motive, numCases cases, 1 target):
//   param_i is at index: numCases + 1 + (numParams - i - 1) = numCases + numParams - i
func buildAppliedInductive(indName string, numParams int, numCases int) ast.Term {
	result := ast.Term(ast.Global{Name: indName})
	for i := 0; i < numParams; i++ {
		// param_i (bound at position i) is at index: numCases + numParams - i
		// But we're building the domain of the target Pi, not the body
		// So at domain position (before target binder), param_i is at:
		// numCases + 1 + (numParams - i - 1) = numCases + numParams - i
		paramIdx := numCases + numParams - i
		result = ast.App{T: result, U: ast.Var{Ix: paramIdx}}
	}
	return result
}

// buildMotiveType builds P : T params... -> Type_j
// At the motive position, params are at indices (numParams - i - 1) for param_i
func buildMotiveType(indName string, numParams int, universe ast.Level) ast.Term {
	// Domain is T applied to params
	domain := ast.Term(ast.Global{Name: indName})
	for i := 0; i < numParams; i++ {
		// param_i is at index (numParams - i - 1) at motive position
		paramIdx := numParams - i - 1
		domain = ast.App{T: domain, U: ast.Var{Ix: paramIdx}}
	}
	return ast.Pi{
		Binder: "_",
		A:      domain,
		B:      ast.Sort{U: universe},
	}
}

// extractUniverseLevel extracts the universe level from an inductive type.
// For Sort{U: n}, returns n.
// For Pi chains like (A : Type) -> Type, extracts from the final Sort.
func extractUniverseLevel(ty ast.Term) ast.Level {
	current := ty
	for {
		switch t := current.(type) {
		case ast.Sort:
			return t.U
		case ast.Pi:
			current = t.B
		default:
			// Fallback for unexpected types
			return 0
		}
	}
}

// generateCaseType generates the type of a case for a specific constructor.
//
// For constructor c : (params...) -> (x1 : A1) -> ... -> (xn : An) -> T params:
//   - Parameter args: skipped (not rebind in case, reference eliminator's params)
//   - Non-recursive data args: pass through
//   - Recursive data args (xk : T params): add ih_k : P xk
//   - Result: P (c params x1 ... xn)
//
// caseIdx is the index of this constructor in the Constructors slice.
// casesAfter is how many case binders come after this one (affects de Bruijn indices).
func generateCaseType(ind *Inductive, caseIdx int, casesAfter int) ast.Term {
	ctor := ind.Constructors[caseIdx]
	indName := ind.Name
	numParams := ind.NumParams

	// Extract all constructor arguments
	allArgs := extractPiArgs(ctor.Type)

	// Skip parameter args - case types don't rebind parameters
	dataArgs := allArgs
	if numParams > 0 && len(allArgs) >= numParams {
		dataArgs = allArgs[numParams:]
	}

	// Count recursive arguments among data args
	recursiveArgs := countRecursiveArgs(indName, dataArgs)

	// P is at index: numParams (params) + 1 (motive) + casesAfter (cases after this one)
	// After binding constructor args and IHs, we need to adjust
	pBaseIdx := numParams + 1 + casesAfter

	// Build the case type from inside out
	return buildCaseType(indName, dataArgs, recursiveArgs, pBaseIdx, ctor.Name, numParams)
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
// Also detects higher-order recursive arguments like (A -> T) where T is the inductive.
func isRecursiveArgType(indName string, ty ast.Term) bool {
	switch t := ty.(type) {
	case ast.Global:
		return t.Name == indName
	case ast.App:
		// Check if head is the inductive (e.g., List A where List is our inductive)
		return isAppOfGlobal(t, indName)
	case ast.Pi:
		// Higher-order recursive: check if the codomain contains the inductive.
		// For example, (A -> T) where T is our inductive type.
		// We use OccursIn to check if indName appears in the codomain.
		return OccursIn(indName, t.B)
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
//	(x1 : A1) -> [ih1 : P x1] -> ... -> (xn : An) -> [ihn : P xn] -> P (c params x1 ... xn)
//
// where [ih_i : P x_i] is present only for recursive arguments (Ai = T or App of T).
//
// For parameterized inductives, the constructor application includes parameters:
//   - P (c A x1 ... xn) where A is the parameter
//
// De Bruijn indices are computed using proper shifting:
//   - Argument types from the constructor need to be shifted when IH binders are interleaved
//   - pBaseIdx is P's index before any case binders are added
//   - At depth d (under d binders), P is at index pBaseIdx + d
//   - Argument x_i bound at depth d_i is at index (current_depth - d_i - 1) when referenced
func buildCaseType(indName string, args []PiArg, numRecursive int, pBaseIdx int, ctorName string, numParams int) ast.Term {
	numArgs := len(args)

	// Build the binder structure: for each arg, we add the arg binder,
	// and if recursive, also an IH binder immediately after.
	// Track which depth each original argument is bound at.
	type binderInfo struct {
		name string
		ty   ast.Term
		isIH bool
	}

	var binders []binderInfo
	argDepths := make([]int, numArgs) // depth at which arg i is bound

	// Track how many IH binders have been added before each position.
	// This is used to shift the argument types correctly.
	ihCount := 0
	depth := 0

	for i, arg := range args {
		// The argument type comes from the constructor, where it's at depth i
		// (under i binders for x1...xi-1). In our case type, we're at depth `depth`,
		// but we've added `ihCount` extra IH binders. So we need to shift by `ihCount`.
		shiftedType := subst.Shift(ihCount, 0, arg.Type)

		// Add argument binder
		argDepths[i] = depth
		binders = append(binders, binderInfo{
			name: arg.Name,
			ty:   shiftedType,
			isIH: false,
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
				name: "ih_" + arg.Name,
				ty:   ihType,
				isIH: true,
			})
			depth++
			ihCount++
		}
	}

	totalBinders := depth

	// Build result type: P (c params x1 ... xn)
	// P is at index pBaseIdx + totalBinders
	// Each param_i is at index pBaseIdx + totalBinders + 1 + (numParams - i - 1)
	// Each x_i is at index (totalBinders - argDepths[i] - 1)
	pIdxAtResult := pBaseIdx + totalBinders

	// Start constructor application
	ctorApp := ast.Term(ast.Global{Name: ctorName})

	// Apply parameters first (for parameterized inductives)
	for i := 0; i < numParams; i++ {
		// param_i is at index: pIdxAtResult + 1 + (numParams - i - 1) = pIdxAtResult + numParams - i
		paramIdx := pIdxAtResult + numParams - i
		ctorApp = ast.App{T: ctorApp, U: ast.Var{Ix: paramIdx}}
	}

	// Apply data arguments
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
