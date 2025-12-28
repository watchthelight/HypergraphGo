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
// For indexed inductives like Vec : Type -> Nat -> Type:
//
//	vecElim : (A : Type) -> (P : (n : Nat) -> Vec A n -> Type) ->
//	          case_vnil -> case_vcons -> (n : Nat) -> (xs : Vec A n) -> P n xs
//
// where case(ci) provides the elimination principle for constructor ci:
//   - Non-recursive args pass through as-is
//   - Recursive args (of type T) get an induction hypothesis (ih : P indices arg)
//
// The universe j for the motive is derived from the inductive's type.
func GenerateRecursorType(ind *Inductive) ast.Term {
	// Extract the universe level from the inductive's type
	motiveUniverse := extractUniverseLevel(ind.Type)
	numParams := ind.NumParams
	numIndices := ind.NumIndices
	numCases := len(ind.Constructors)

	// Build from inside out: target + index binders + case binders + motive + param binders

	// 1. Build innermost: P indices... t
	// At innermost (result of target Pi):
	// - t is at index 0
	// - indices are at indices 1..numIndices
	// - P is at numIndices + numCases + 1
	pIdxInner := numIndices + numCases + 1
	targetResult := ast.Term(ast.Var{Ix: pIdxInner})
	// Apply indices to P
	for i := 0; i < numIndices; i++ {
		// index_i is at position (numIndices - i) from innermost
		idxVar := numIndices - i
		targetResult = ast.App{T: targetResult, U: ast.Var{Ix: idxVar}}
	}
	// Apply t to P
	targetResult = ast.App{T: targetResult, U: ast.Var{Ix: 0}}

	// 2. Build target: (t : T params indices) -> P indices t
	appliedIndAtTarget := buildAppliedInductiveFull(ind.Name, numParams, numIndices, numCases)
	target := ast.Pi{
		Binder: "t",
		A:      appliedIndAtTarget,
		B:      targetResult,
	}

	// 3. Wrap with index binders (if any)
	result := ast.Term(target)
	for i := numIndices - 1; i >= 0; i-- {
		// Index types need shifting. At index binder i position:
		// We're adding binders from inside out, so shift by (numIndices - i - 1)
		// plus account for being under numCases + 1 (P) binders
		shiftAmount := numCases + 1 + (numIndices - i - 1)
		shiftedIdxType := subst.Shift(shiftAmount, 0, ind.IndexTypes[i])
		result = ast.Pi{
			Binder: indexName(i),
			A:      shiftedIdxType,
			B:      result,
		}
	}

	// 4. Wrap with case binders from last to first
	for i := numCases - 1; i >= 0; i-- {
		caseTy := generateCaseType(ind, i, numCases-i-1)
		result = ast.Pi{
			Binder: "case_" + ind.Constructors[i].Name,
			A:      caseTy,
			B:      result,
		}
	}

	// 5. Build motive type: P : (indices...) -> T params indices -> Type_j
	motiveType := buildMotiveTypeFull(ind, motiveUniverse)
	result = ast.Pi{
		Binder: "P",
		A:      motiveType,
		B:      result,
	}

	// 6. Wrap with parameter binders (outermost)
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

// indexName returns the name for index i (n, m, k, ...)
func indexName(i int) string {
	names := []string{"n", "m", "k", "j", "i"}
	if i < len(names) {
		return names[i]
	}
	return "i" + string(rune('0'+i))
}

// buildAppliedInductiveFull builds T params... indices...
// where params are de Bruijn variables and indices are bound in the target section.
// At target domain position (under numParams, 1 motive, numCases cases, numIndices index binders):
//
//	param_i is at index: numCases + 1 + numIndices + (numParams - i - 1)
//	index_j is at index: numIndices - j - 1
func buildAppliedInductiveFull(indName string, numParams int, numIndices int, numCases int) ast.Term {
	result := ast.Term(ast.Global{Name: indName})

	// Apply params first
	for i := 0; i < numParams; i++ {
		// param_i is at: numCases + 1 + numIndices + (numParams - i - 1)
		paramIdx := numCases + 1 + numIndices + (numParams - i - 1)
		result = ast.App{T: result, U: ast.Var{Ix: paramIdx}}
	}

	// Apply indices
	for j := 0; j < numIndices; j++ {
		// index_j is at: numIndices - j - 1 (indices are bound inside cases)
		idxIdx := numIndices - j - 1
		result = ast.App{T: result, U: ast.Var{Ix: idxIdx}}
	}

	return result
}

// buildMotiveTypeFull builds P : (indices...) -> T params indices -> Type_j
// For indexed inductives, the motive takes index arguments before the target.
func buildMotiveTypeFull(ind *Inductive, universe ast.Level) ast.Term {
	numParams := ind.NumParams
	numIndices := ind.NumIndices

	// Build innermost: Type_j
	result := ast.Term(ast.Sort{U: universe})

	// Build T params indices -> Type_j (under index binders)
	// At this position, indices are at 0..numIndices-1, params at numIndices..numIndices+numParams-1
	domain := ast.Term(ast.Global{Name: ind.Name})
	for i := 0; i < numParams; i++ {
		// param_i is at numIndices + (numParams - i - 1)
		paramIdx := numIndices + (numParams - i - 1)
		domain = ast.App{T: domain, U: ast.Var{Ix: paramIdx}}
	}
	for j := 0; j < numIndices; j++ {
		// index_j is at numIndices - j - 1
		idxIdx := numIndices - j - 1
		domain = ast.App{T: domain, U: ast.Var{Ix: idxIdx}}
	}
	result = ast.Pi{
		Binder: "_",
		A:      domain,
		B:      result,
	}

	// Wrap with index binders from inside out
	for j := numIndices - 1; j >= 0; j-- {
		// Index types need to be shifted. At index binder j:
		// - We're under (numIndices - j - 1) already-added index binders
		// - Plus we're under numParams param binders
		shiftAmount := numParams + (numIndices - j - 1)
		shiftedIdxType := subst.Shift(shiftAmount, 0, ind.IndexTypes[j])
		result = ast.Pi{
			Binder: indexName(j),
			A:      shiftedIdxType,
			B:      result,
		}
	}

	return result
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
// For constructor c : (params...) -> (x1 : A1) -> ... -> (xn : An) -> T params indices:
//   - Parameter args: skipped (not rebind in case, reference eliminator's params)
//   - Non-recursive data args: pass through
//   - Recursive data args (xk : T params idx): add ih_k : P idx xk
//   - Result: P ctorIndices (c params x1 ... xn)
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

	// Extract index values from constructor result type
	ctorResultIndices := extractConstructorIndices(ctor.Type, numParams, ind.NumIndices)

	// Build the case type from inside out
	return buildCaseTypeFull(ind, dataArgs, recursiveArgs, pBaseIdx, ctor.Name, ctorResultIndices)
}

// extractConstructorIndices extracts the index expressions from a constructor's result type.
// For vcons : (A : Type) -> A -> (n : Nat) -> Vec A n -> Vec A (succ n)
// Returns [succ n] (the index expressions in the result Vec A (succ n))
func extractConstructorIndices(ctorType ast.Term, numParams int, numIndices int) []ast.Term {
	resultTy := constructorResultType(ctorType)
	args := extractAppArgs(resultTy)

	// args = [param_0, ..., param_k-1, idx_0, ..., idx_m-1]
	// We want the last numIndices args
	if len(args) < numParams+numIndices {
		return nil
	}
	return args[numParams:]
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
	return isRecursiveArgTypeMulti([]string{indName}, ty)
}

// isRecursiveArgTypeMulti checks if an argument type refers to any of the given
// inductive type names. This is used for mutual inductives where a constructor
// may have recursive arguments of a different type in the mutual group.
func isRecursiveArgTypeMulti(indNames []string, ty ast.Term) bool {
	switch t := ty.(type) {
	case ast.Global:
		for _, name := range indNames {
			if t.Name == name {
				return true
			}
		}
		return false
	case ast.App:
		// Check if head is one of the inductives (e.g., List A where List is our inductive)
		for _, name := range indNames {
			if isAppOfGlobal(t, name) {
				return true
			}
		}
		return false
	case ast.Pi:
		// Higher-order recursive: check if the codomain contains any inductive.
		// For example, (A -> T) where T is one of our inductive types.
		for _, name := range indNames {
			if OccursIn(name, t.B) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// buildCaseTypeFull constructs the full case type for a constructor with index support.
//
// For indexed constructor c : (params...) -> (x1 : A1) -> ... -> (xn : An) -> T params indices:
//
// Case type structure:
//
//	(x1 : A1) -> [ih1 : P idx1 x1] -> ... -> (xn : An) -> [ihn : P idxn xn] -> P ctorIndices (c params x1 ... xn)
//
// where [ih_i : P idx_i x_i] is present only for recursive arguments.
// ctorIndices are the index expressions from the constructor's result type.
//
// For IH types, we need to extract the index values from the recursive arg's type.
func buildCaseTypeFull(ind *Inductive, args []PiArg, numRecursive int, pBaseIdx int, ctorName string, ctorResultIndices []ast.Term) ast.Term {
	indName := ind.Name
	numParams := ind.NumParams
	numIndices := ind.NumIndices
	numArgs := len(args)

	// Build the binder structure: for each arg, we add the arg binder,
	// and if recursive, also an IH binder immediately after.
	type binderInfo struct {
		name string
		ty   ast.Term
		isIH bool
	}

	var binders []binderInfo
	argDepths := make([]int, numArgs) // depth at which arg i is bound

	ihCount := 0
	depth := 0

	for i, arg := range args {
		// Shift argument type by ihCount (extra IH binders added so far)
		shiftedType := subst.Shift(ihCount, 0, arg.Type)

		argDepths[i] = depth
		binders = append(binders, binderInfo{
			name: arg.Name,
			ty:   shiftedType,
			isIH: false,
		})
		depth++

		// If recursive, add IH binder
		if isRecursiveArgType(indName, arg.Type) {
			// IH type: P indices x_i
			// x_i was just bound, so it's at index 0
			// P is at pBaseIdx + depth
			pIdx := pBaseIdx + depth

			// For indexed inductives, extract indices from the recursive arg's type
			ihType := buildIHType(pIdx, arg.Type, numParams, numIndices, ihCount)

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

	// Build result type: P ctorIndices (c params x1 ... xn)
	pIdxAtResult := pBaseIdx + totalBinders

	// Start with P
	resultType := ast.Term(ast.Var{Ix: pIdxAtResult})

	// Apply constructor's result indices to P (shifted for current depth)
	for _, idxExpr := range ctorResultIndices {
		// Shift the index expression: it's written under constructor's binders,
		// we need to adjust for our case binders
		shiftedIdx := shiftIndexExpr(idxExpr, totalBinders, argDepths, numParams, ihCount)
		resultType = ast.App{T: resultType, U: shiftedIdx}
	}

	// Build constructor application: c params x1 ... xn
	ctorApp := ast.Term(ast.Global{Name: ctorName})

	// Apply parameters
	for i := 0; i < numParams; i++ {
		paramIdx := pIdxAtResult + numParams - i
		ctorApp = ast.App{T: ctorApp, U: ast.Var{Ix: paramIdx}}
	}

	// Apply data arguments
	for i := 0; i < numArgs; i++ {
		argVarIdx := totalBinders - argDepths[i] - 1
		ctorApp = ast.App{T: ctorApp, U: ast.Var{Ix: argVarIdx}}
	}

	// Apply constructor to P indices
	resultType = ast.App{T: resultType, U: ctorApp}

	// Wrap with binders from inside out
	for i := len(binders) - 1; i >= 0; i-- {
		b := binders[i]
		resultType = ast.Pi{
			Binder: b.name,
			A:      b.ty,
			B:      resultType,
		}
	}

	return resultType
}

// buildIHType builds the type for an induction hypothesis: P indices x
// where x is the recursive argument at index 0, and indices come from x's type.
func buildIHType(pIdx int, argType ast.Term, numParams int, numIndices int, ihCount int) ast.Term {
	// Simple case: no indices, just P x
	if numIndices == 0 {
		return ast.App{T: ast.Var{Ix: pIdx}, U: ast.Var{Ix: 0}}
	}

	// Extract indices from the argument type (e.g., Vec A n -> indices are [n])
	argIndices := extractIndicesFromType(argType, numParams, numIndices)

	// Build P indices x
	result := ast.Term(ast.Var{Ix: pIdx})
	for _, idx := range argIndices {
		// Shift index by 1 (for the x binder) + ihCount (previous IHs)
		shiftedIdx := subst.Shift(1+ihCount, 0, idx)
		result = ast.App{T: result, U: shiftedIdx}
	}
	// Apply x (at index 0)
	result = ast.App{T: result, U: ast.Var{Ix: 0}}

	return result
}

// extractIndicesFromType extracts index values from a type like Vec A n.
// Returns [n] for Vec A n (the index values after params).
func extractIndicesFromType(ty ast.Term, numParams int, numIndices int) []ast.Term {
	args := extractAppArgs(ty)
	if len(args) < numParams+numIndices {
		return nil
	}
	return args[numParams:]
}

// shiftIndexExpr shifts an index expression from constructor context to case type context.
// The index expression was written under numParams param binders in the constructor.
// In the case type, we're under totalBinders case binders.
func shiftIndexExpr(expr ast.Term, totalBinders int, argDepths []int, numParams int, ihCount int) ast.Term {
	// Variables in the index expression refer to constructor args.
	// We need to map them to case type args (accounting for IH binders).
	return subst.Shift(ihCount, 0, expr)
}

// buildCaseType constructs the full case type for a constructor (legacy, no indices).
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

// GenerateRecursorTypeSimple generates a recursor (eliminator) type for an inductive definition.
//
// A recursor encodes the induction principle for an inductive type, allowing
// dependent elimination (proving properties) and recursion (defining functions).
// The generated type has the form:
//
//	elimT : (P : T → Type) → case₁ → ... → caseₙ → (t : T) → P t
//
// where each caseᵢ corresponds to a constructor with appropriate induction hypotheses
// for recursive arguments.
//
// For example, the recursor for Nat (with zero and succ) has type:
//
//	natElim : (P : Nat → Type)
//	        → P zero                           ; base case
//	        → ((n : Nat) → P n → P (succ n))   ; step case with IH
//	        → (n : Nat) → P n
//
// This function provides hand-crafted types for common inductives (Nat, Bool) for
// clarity and correctness, falling back to GenerateRecursorType for other inductives.
//
// See also: GenerateRecursorType for the general algorithm handling indexed types.
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
