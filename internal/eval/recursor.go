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
			// Check arity matches
			if len(neutral.N.Sp) == ctor.NumArgs {
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
	ctorArgs := neutral.N.Sp

	// Build the result by applying the case function to args and IHs
	// For each argument:
	//   - Apply the argument
	//   - If recursive, also apply the IH (elim P cases... arg)
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
			// IH = elim params... P cases... arg
			// We need all args up to (but not including) scrutinee, plus the arg
			prefixEnd := info.NumParams + 1 + info.NumCases + info.NumIndices
			ih := buildRecursorCall(elimName, sp[:prefixEnd], arg)
			result = Apply(result, ih)
		}
	}

	// Apply any extra arguments
	for _, extra := range extraArgs {
		result = Apply(result, extra)
	}

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
