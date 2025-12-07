package eval

import (
	"sync"
)

// RecursorInfo contains the information needed to reduce a recursor application.
type RecursorInfo struct {
	ElimName   string            // Name of the eliminator (e.g., "natElim")
	IndName    string            // Name of the inductive type (e.g., "Nat")
	NumParams  int               // Number of type parameters (e.g., 1 for List A)
	NumIndices int               // Number of indices (e.g., 1 for Vec A n)
	NumCases   int               // Number of constructor cases
	Ctors      []ConstructorInfo // Info about each constructor
}

// ConstructorInfo contains information about a single constructor.
type ConstructorInfo struct {
	Name         string // Constructor name (e.g., "zero", "succ")
	NumArgs      int    // Total number of arguments
	RecursiveIdx []int  // Indices of recursive arguments (need IH)
	// IndexArgPositions maps each recursive arg to the positions of its index args.
	// For Vec's vcons with data args [n, x, xs] where xs : Vec A n:
	//   IndexArgPositions[2] = []int{0}  (xs at position 2 uses n at position 0)
	// For non-indexed inductives or non-recursive args, entries are nil.
	IndexArgPositions map[int][]int
}

// recursorRegistry maps eliminator names to their RecursorInfo.
var recursorRegistry = struct {
	sync.RWMutex
	data map[string]*RecursorInfo
}{
	data: make(map[string]*RecursorInfo),
}

// RegisterRecursor registers an eliminator for generic reduction.
// This should be called when declaring an inductive type.
func RegisterRecursor(info *RecursorInfo) {
	recursorRegistry.Lock()
	defer recursorRegistry.Unlock()
	recursorRegistry.data[info.ElimName] = info
}

// LookupRecursor returns the RecursorInfo for an eliminator, or nil if not found.
func LookupRecursor(elimName string) *RecursorInfo {
	recursorRegistry.RLock()
	defer recursorRegistry.RUnlock()
	return recursorRegistry.data[elimName]
}

// ClearRecursorRegistry clears all registered recursors (useful for testing).
func ClearRecursorRegistry() {
	recursorRegistry.Lock()
	defer recursorRegistry.Unlock()
	recursorRegistry.data = make(map[string]*RecursorInfo)
}

// tryGenericRecursorReduction attempts to reduce a registered recursor.
// Returns nil if the recursor is not registered or if the scrutinee is not a constructor.
//
// For a recursor elim : (P : T -> Type) -> case_c1 -> ... -> case_cn -> (t : T) -> P t
// applied to P, cases..., and a constructor (ci args...):
//
//	elim P case_c1 ... case_cn (ci a1 ... am) -->
//	  case_ci a1 [ih1] a2 [ih2] ... am [ihm]
//
// where [ihi] is included only if ai is a recursive argument.
func tryGenericRecursorReduction(elimName string, sp []Value) Value {
	info := LookupRecursor(elimName)
	if info == nil {
		return nil
	}

	// Arguments to eliminator for parameterized/indexed inductives:
	//   params..., P (motive), case_c1, ..., case_cn, indices..., scrutinee
	// For non-parameterized, non-indexed: P, case_c1, ..., case_cn, scrutinee
	//
	// Need at least: NumParams + 1 (motive) + NumCases + NumIndices + 1 (scrutinee)
	minArgs := info.NumParams + 1 + info.NumCases + info.NumIndices + 1
	if len(sp) < minArgs {
		return nil // Not fully applied yet
	}

	// Extract arguments accounting for parameters
	// params := sp[:info.NumParams]                                      // type parameters (unused directly)
	// p := sp[info.NumParams]                                            // motive (unused directly)
	casesStart := info.NumParams + 1
	casesEnd := casesStart + info.NumCases
	cases := sp[casesStart:casesEnd]           // case for each constructor
	scrutineeIdx := casesEnd + info.NumIndices // skip indices
	scrutinee := sp[scrutineeIdx]              // the value being eliminated
	extraArgs := sp[scrutineeIdx+1:]           // additional arguments after scrutinee

	// Check if scrutinee is a constructor
	neutral, ok := scrutinee.(VNeutral)
	if !ok {
		return nil // Not a neutral value
	}

	// Find which constructor matches
	ctorIdx := -1
	for i, ctor := range info.Ctors {
		if neutral.N.Head.Glob == ctor.Name {
			// Constructor spine includes params + data args
			// Check arity matches: spine length = NumParams + NumArgs
			expectedSpineLen := info.NumParams + ctor.NumArgs
			if len(neutral.N.Sp) == expectedSpineLen {
				ctorIdx = i
				break
			}
		}
	}

	if ctorIdx < 0 {
		return nil // Scrutinee doesn't match any constructor
	}

	ctor := info.Ctors[ctorIdx]
	caseFunc := cases[ctorIdx]
	// Skip parameter args in constructor spine - only use data args for case application
	ctorArgs := neutral.N.Sp[info.NumParams:]

	// Build the result by applying the case function to args and IHs
	// For each argument:
	//   - Apply the argument
	//   - If recursive, also apply the IH (elim P cases... indices_from_ctor arg)
	result := caseFunc

	isRecursive := make(map[int]bool)
	for _, idx := range ctor.RecursiveIdx {
		isRecursive[idx] = true
	}

	for i, arg := range ctorArgs {
		// Apply the argument
		result = Apply(result, arg)

		// If recursive, also apply the IH
		if isRecursive[i] {
			// IH = elim params... P cases... indices_from_recursive_arg... arg
			//
			// For indexed inductives, the indices for the IH come from the recursive
			// argument's type, which is encoded in the constructor's data args.
			// For a recursive arg at position i, its indices are the preceding
			// data args that appear in its type.
			//
			// For non-indexed inductives (NumIndices == 0), we just use params + P + cases.
			// For indexed inductives, we extract indices from constructor args using
			// precomputed IndexArgPositions metadata.
			ih := buildRecursorCallWithIndices(elimName, sp, ctorArgs, i, info, &ctor)
			result = Apply(result, ih)
		}
	}

	// Apply any extra arguments
	for _, extra := range extraArgs {
		result = Apply(result, extra)
	}

	return result
}

// buildRecursorCallWithIndices constructs the IH call for indexed inductives.
// For a recursive arg at position recArgIdx in ctorArgs, extracts the correct indices.
//
// For Vec with vcons : (A : Type) -> (n : Nat) -> A -> Vec A n -> Vec A (succ n):
//   - data args: [n, x, xs]
//   - xs is recursive at index 2
//   - xs : Vec A n, so its index is n (ctorArgs[0])
//
// The IH for xs should be: vecElim A P pvnil pvcons n xs
func buildRecursorCallWithIndices(elimName string, sp []Value, ctorArgs []Value, recArgIdx int, info *RecursorInfo, ctor *ConstructorInfo) Value {
	result := vGlobal(elimName)

	// Apply params
	for i := 0; i < info.NumParams; i++ {
		result = Apply(result, sp[i])
	}

	// Apply motive P
	result = Apply(result, sp[info.NumParams])

	// Apply cases
	casesStart := info.NumParams + 1
	for i := 0; i < info.NumCases; i++ {
		result = Apply(result, sp[casesStart+i])
	}

	// Apply indices from constructor args using precomputed metadata
	if info.NumIndices > 0 {
		// Use precomputed IndexArgPositions if available and COMPLETE
		// (i.e., we have position info for ALL indices, not just some)
		useMetadata := false
		if ctor.IndexArgPositions != nil {
			if indexPositions, ok := ctor.IndexArgPositions[recArgIdx]; ok && len(indexPositions) == info.NumIndices {
				// Use the exact positions computed at declaration time
				// Only if we have complete metadata (all indices are variable references)
				useMetadata = true
				for _, pos := range indexPositions {
					if pos >= 0 && pos < len(ctorArgs) {
						result = Apply(result, ctorArgs[pos])
					}
				}
			}
		}
		// If no metadata or incomplete metadata, fall back to heuristic
		// This handles:
		// - Inductives declared before IndexArgPositions feature
		// - Indices that are computed expressions (not variable references)
		// Note: The heuristic assumes indices precede the recursive arg, which
		// may not hold for all indexed inductives with computed index expressions.
		if !useMetadata {
			indicesExtracted := 0
			for j := 0; j < recArgIdx && indicesExtracted < info.NumIndices; j++ {
				result = Apply(result, ctorArgs[j])
				indicesExtracted++
			}
			for indicesExtracted < info.NumIndices && indicesExtracted < len(ctorArgs) {
				result = Apply(result, ctorArgs[indicesExtracted])
				indicesExtracted++
			}
		}
	}

	// Apply the recursive argument
	result = Apply(result, ctorArgs[recArgIdx])

	return result
}

// buildRecursorCall constructs: elim P case_c1 ... case_cn arg
func buildRecursorCall(elimName string, prefix []Value, arg Value) Value {
	result := vGlobal(elimName)
	for _, v := range prefix {
		result = Apply(result, v)
	}
	return Apply(result, arg)
}
