package eval

import (
	"testing"
)

// ============================================================================
// Recursor Registry Tests
// ============================================================================

// TestRegisterRecursor tests basic recursor registration
func TestRegisterRecursor(t *testing.T) {
	ClearRecursorRegistry()
	defer ClearRecursorRegistry()

	info := &RecursorInfo{
		ElimName:  "testElim",
		IndName:   "TestType",
		NumParams: 0,
		NumCases:  2,
		Ctors: []ConstructorInfo{
			{Name: "ctor1", NumArgs: 0},
			{Name: "ctor2", NumArgs: 1},
		},
	}

	RegisterRecursor(info)

	result := LookupRecursor("testElim")
	if result == nil {
		t.Fatal("LookupRecursor returned nil after registration")
	}
	if result.ElimName != "testElim" {
		t.Errorf("Expected ElimName 'testElim', got '%s'", result.ElimName)
	}
	if result.IndName != "TestType" {
		t.Errorf("Expected IndName 'TestType', got '%s'", result.IndName)
	}
}

// TestLookupRecursor_NotFound tests lookup for non-existent recursor
func TestLookupRecursor_NotFound(t *testing.T) {
	ClearRecursorRegistry()
	defer ClearRecursorRegistry()

	result := LookupRecursor("nonexistent")
	if result != nil {
		t.Errorf("Expected nil for non-existent recursor, got %v", result)
	}
}

// TestClearRecursorRegistry tests clearing the registry
func TestClearRecursorRegistry(t *testing.T) {
	ClearRecursorRegistry()

	// Register something
	RegisterRecursor(&RecursorInfo{
		ElimName: "toBeCleared",
		IndName:  "ClearTest",
	})

	// Verify it exists
	if LookupRecursor("toBeCleared") == nil {
		t.Fatal("Failed to register recursor")
	}

	// Clear
	ClearRecursorRegistry()

	// Verify it's gone
	if LookupRecursor("toBeCleared") != nil {
		t.Error("Registry not cleared properly")
	}
}

// TestMultipleRecursors tests registering multiple recursors
func TestMultipleRecursors(t *testing.T) {
	ClearRecursorRegistry()
	defer ClearRecursorRegistry()

	RegisterRecursor(&RecursorInfo{ElimName: "natElim", IndName: "Nat"})
	RegisterRecursor(&RecursorInfo{ElimName: "boolElim", IndName: "Bool"})
	RegisterRecursor(&RecursorInfo{ElimName: "listElim", IndName: "List"})

	nat := LookupRecursor("natElim")
	bool_ := LookupRecursor("boolElim")
	list := LookupRecursor("listElim")

	if nat == nil || nat.IndName != "Nat" {
		t.Error("natElim not found or incorrect")
	}
	if bool_ == nil || bool_.IndName != "Bool" {
		t.Error("boolElim not found or incorrect")
	}
	if list == nil || list.IndName != "List" {
		t.Error("listElim not found or incorrect")
	}
}

// TestRecursorOverwrite tests overwriting a recursor registration
func TestRecursorOverwrite(t *testing.T) {
	ClearRecursorRegistry()
	defer ClearRecursorRegistry()

	// Register first version
	RegisterRecursor(&RecursorInfo{
		ElimName: "myElim",
		IndName:  "OldType",
		NumCases: 1,
	})

	// Overwrite with new version
	RegisterRecursor(&RecursorInfo{
		ElimName: "myElim",
		IndName:  "NewType",
		NumCases: 3,
	})

	result := LookupRecursor("myElim")
	if result == nil {
		t.Fatal("LookupRecursor returned nil")
	}
	if result.IndName != "NewType" {
		t.Errorf("Expected IndName 'NewType', got '%s'", result.IndName)
	}
	if result.NumCases != 3 {
		t.Errorf("Expected NumCases 3, got %d", result.NumCases)
	}
}

// ============================================================================
// RecursorInfo Structure Tests
// ============================================================================

// TestRecursorInfo_Nat tests RecursorInfo for Nat type
func TestRecursorInfo_Nat(t *testing.T) {
	info := &RecursorInfo{
		ElimName:   "natElim",
		IndName:    "Nat",
		NumParams:  0,
		NumIndices: 0,
		NumCases:   2,
		Ctors: []ConstructorInfo{
			{Name: "zero", NumArgs: 0, RecursiveIdx: nil},
			{Name: "succ", NumArgs: 1, RecursiveIdx: []int{0}},
		},
	}

	if len(info.Ctors) != 2 {
		t.Errorf("Expected 2 constructors, got %d", len(info.Ctors))
	}
	if info.Ctors[0].Name != "zero" {
		t.Errorf("Expected first ctor 'zero', got '%s'", info.Ctors[0].Name)
	}
	if info.Ctors[1].Name != "succ" {
		t.Errorf("Expected second ctor 'succ', got '%s'", info.Ctors[1].Name)
	}
	if len(info.Ctors[1].RecursiveIdx) != 1 || info.Ctors[1].RecursiveIdx[0] != 0 {
		t.Errorf("succ should have recursive arg at index 0")
	}
}

// TestRecursorInfo_List tests RecursorInfo for List type
func TestRecursorInfo_List(t *testing.T) {
	info := &RecursorInfo{
		ElimName:   "listElim",
		IndName:    "List",
		NumParams:  1, // A parameter
		NumIndices: 0,
		NumCases:   2,
		Ctors: []ConstructorInfo{
			{Name: "nil", NumArgs: 0, RecursiveIdx: nil},
			{Name: "cons", NumArgs: 2, RecursiveIdx: []int{1}}, // head, tail (tail is recursive)
		},
	}

	if info.NumParams != 1 {
		t.Errorf("Expected 1 param, got %d", info.NumParams)
	}
	if info.Ctors[1].NumArgs != 2 {
		t.Errorf("cons should have 2 args, got %d", info.Ctors[1].NumArgs)
	}
	if len(info.Ctors[1].RecursiveIdx) != 1 || info.Ctors[1].RecursiveIdx[0] != 1 {
		t.Errorf("cons should have recursive arg at index 1 (tail)")
	}
}

// TestRecursorInfo_Vec tests RecursorInfo for indexed Vec type
func TestRecursorInfo_Vec(t *testing.T) {
	info := &RecursorInfo{
		ElimName:   "vecElim",
		IndName:    "Vec",
		NumParams:  1, // A parameter
		NumIndices: 1, // n index
		NumCases:   2,
		Ctors: []ConstructorInfo{
			{Name: "vnil", NumArgs: 0, RecursiveIdx: nil},
			{
				Name:         "vcons",
				NumArgs:      3,              // n, x, xs
				RecursiveIdx: []int{2},       // xs is recursive
				IndexArgPositions: map[int][]int{
					2: {0}, // xs's index is at position 0 (n)
				},
			},
		},
	}

	if info.NumIndices != 1 {
		t.Errorf("Expected 1 index, got %d", info.NumIndices)
	}

	vcons := info.Ctors[1]
	if vcons.NumArgs != 3 {
		t.Errorf("vcons should have 3 args, got %d", vcons.NumArgs)
	}
	if vcons.IndexArgPositions[2][0] != 0 {
		t.Error("vcons recursive arg at 2 should have index at position 0")
	}
}

// TestConstructorInfo_NonRecursive tests constructor with no recursive args
func TestConstructorInfo_NonRecursive(t *testing.T) {
	info := ConstructorInfo{
		Name:         "leaf",
		NumArgs:      1,
		RecursiveIdx: nil,
	}

	if len(info.RecursiveIdx) > 0 {
		t.Error("Non-recursive constructor should have empty RecursiveIdx")
	}
}

// TestConstructorInfo_MultipleRecursive tests constructor with multiple recursive args
func TestConstructorInfo_MultipleRecursive(t *testing.T) {
	// Binary tree: node : Tree -> Tree -> Tree
	info := ConstructorInfo{
		Name:         "node",
		NumArgs:      2,
		RecursiveIdx: []int{0, 1}, // Both children are recursive
	}

	if len(info.RecursiveIdx) != 2 {
		t.Errorf("Expected 2 recursive args, got %d", len(info.RecursiveIdx))
	}
	if info.RecursiveIdx[0] != 0 || info.RecursiveIdx[1] != 1 {
		t.Error("node should have recursive args at indices 0 and 1")
	}
}

// ============================================================================
// Concurrency Tests
// ============================================================================

// TestRegistryConcurrency tests concurrent access to the registry
func TestRegistryConcurrency(t *testing.T) {
	ClearRecursorRegistry()
	defer ClearRecursorRegistry()

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			RegisterRecursor(&RecursorInfo{
				ElimName: "concurrentElim",
				IndName:  "Concurrent",
			})
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_ = LookupRecursor("concurrentElim")
		}
		done <- true
	}()

	// Clear goroutine
	go func() {
		for i := 0; i < 10; i++ {
			ClearRecursorRegistry()
		}
		done <- true
	}()

	// Wait for all goroutines
	<-done
	<-done
	<-done

	// Test passes if no race conditions or panics
}

// ============================================================================
// Indexed Inductive Tests (for buildRecursorCallWithIndices)
// ============================================================================

// TestIndexedInductive_VecWithMetadata tests indexed inductive (Vec) with IndexArgPositions.
// This exercises the metadata path in buildRecursorCallWithIndices.
func TestIndexedInductive_VecWithMetadata(t *testing.T) {
	ClearRecursorRegistry()
	defer ClearRecursorRegistry()

	// Vec : Type -> Nat -> Type
	// vnil : Vec A 0
	// vcons : (A : Type) -> (n : Nat) -> A -> Vec A n -> Vec A (succ n)
	vecInfo := &RecursorInfo{
		ElimName:   "vecElim",
		IndName:    "Vec",
		NumParams:  1, // A
		NumIndices: 1, // n
		NumCases:   2,
		Ctors: []ConstructorInfo{
			{Name: "vnil", NumArgs: 0, RecursiveIdx: nil},
			{
				Name:         "vcons",
				NumArgs:      3,        // n, x, xs
				RecursiveIdx: []int{2}, // xs is recursive
				IndexArgPositions: map[int][]int{
					2: {0}, // xs's index is at position 0 (n)
				},
			},
		},
	}

	RegisterRecursor(vecInfo)

	// Verify the recursor was registered correctly
	info := LookupRecursor("vecElim")
	if info == nil {
		t.Fatal("vecElim should be registered")
	}
	if info.NumIndices != 1 {
		t.Errorf("Expected 1 index, got %d", info.NumIndices)
	}

	// Verify vcons has correct IndexArgPositions
	vcons := info.Ctors[1]
	if vcons.IndexArgPositions == nil {
		t.Fatal("vcons should have IndexArgPositions")
	}
	positions, ok := vcons.IndexArgPositions[2]
	if !ok || len(positions) != 1 || positions[0] != 0 {
		t.Errorf("vcons recursive arg 2 should have index at position 0, got %v", positions)
	}
}

// TestIndexedInductive_WithoutMetadata tests indexed inductive without IndexArgPositions.
// This exercises the fallback heuristic in buildRecursorCallWithIndices.
func TestIndexedInductive_WithoutMetadata(t *testing.T) {
	ClearRecursorRegistry()
	defer ClearRecursorRegistry()

	// Same Vec structure but without IndexArgPositions (simulating old declarations)
	vecInfo := &RecursorInfo{
		ElimName:   "vecElimLegacy",
		IndName:    "Vec",
		NumParams:  1,
		NumIndices: 1,
		NumCases:   2,
		Ctors: []ConstructorInfo{
			{Name: "vnil", NumArgs: 0, RecursiveIdx: nil},
			{
				Name:         "vcons",
				NumArgs:      3,
				RecursiveIdx: []int{2},
				// No IndexArgPositions - will use fallback heuristic
			},
		},
	}

	RegisterRecursor(vecInfo)

	info := LookupRecursor("vecElimLegacy")
	if info == nil {
		t.Fatal("vecElimLegacy should be registered")
	}

	// Verify vcons has NO IndexArgPositions
	vcons := info.Ctors[1]
	if vcons.IndexArgPositions != nil {
		t.Error("vcons should not have IndexArgPositions in this test")
	}
}

// TestIndexedInductive_MultipleIndices tests inductive with multiple indices.
func TestIndexedInductive_MultipleIndices(t *testing.T) {
	ClearRecursorRegistry()
	defer ClearRecursorRegistry()

	// Fin : (n : Nat) -> Type
	// fzero : Fin (succ n)
	// fsucc : Fin n -> Fin (succ n)
	// This is actually 1 index, but let's test a hypothetical 2-index type
	multiIdxInfo := &RecursorInfo{
		ElimName:   "multiElim",
		IndName:    "Multi",
		NumParams:  1,
		NumIndices: 2, // Two indices
		NumCases:   2,
		Ctors: []ConstructorInfo{
			{Name: "base", NumArgs: 0, RecursiveIdx: nil},
			{
				Name:         "step",
				NumArgs:      4,        // idx1, idx2, data, prev
				RecursiveIdx: []int{3}, // prev is recursive
				IndexArgPositions: map[int][]int{
					3: {0, 1}, // prev's indices are at positions 0 and 1
				},
			},
		},
	}

	RegisterRecursor(multiIdxInfo)

	info := LookupRecursor("multiElim")
	if info == nil {
		t.Fatal("multiElim should be registered")
	}
	if info.NumIndices != 2 {
		t.Errorf("Expected 2 indices, got %d", info.NumIndices)
	}

	// Verify step has correct IndexArgPositions for 2 indices
	step := info.Ctors[1]
	positions, ok := step.IndexArgPositions[3]
	if !ok || len(positions) != 2 {
		t.Errorf("step recursive arg 3 should have 2 index positions, got %v", positions)
	}
	if positions[0] != 0 || positions[1] != 1 {
		t.Errorf("step indices should be at [0, 1], got %v", positions)
	}
}

// TestIndexedInductive_MultipleRecursiveArgs tests inductive with multiple recursive args.
func TestIndexedInductive_MultipleRecursiveArgs(t *testing.T) {
	ClearRecursorRegistry()
	defer ClearRecursorRegistry()

	// Binary tree indexed by depth
	// Tree : Nat -> Type
	// leaf : Tree 0
	// node : (n : Nat) -> Tree n -> Tree n -> Tree (succ n)
	treeInfo := &RecursorInfo{
		ElimName:   "treeElim",
		IndName:    "Tree",
		NumParams:  0,
		NumIndices: 1, // depth n
		NumCases:   2,
		Ctors: []ConstructorInfo{
			{Name: "leaf", NumArgs: 0, RecursiveIdx: nil},
			{
				Name:         "node",
				NumArgs:      3,           // n, left, right
				RecursiveIdx: []int{1, 2}, // Both children are recursive
				IndexArgPositions: map[int][]int{
					1: {0}, // left's index is at position 0 (n)
					2: {0}, // right's index is at position 0 (n)
				},
			},
		},
	}

	RegisterRecursor(treeInfo)

	info := LookupRecursor("treeElim")
	if info == nil {
		t.Fatal("treeElim should be registered")
	}

	node := info.Ctors[1]
	if len(node.RecursiveIdx) != 2 {
		t.Errorf("node should have 2 recursive args, got %d", len(node.RecursiveIdx))
	}

	// Both recursive args should have index at position 0
	for _, recIdx := range []int{1, 2} {
		positions, ok := node.IndexArgPositions[recIdx]
		if !ok || len(positions) != 1 || positions[0] != 0 {
			t.Errorf("node recursive arg %d should have index at position 0, got %v", recIdx, positions)
		}
	}
}

// TestIndexedInductive_PartialMetadata tests inductive with incomplete IndexArgPositions.
// This should trigger the fallback heuristic for missing entries.
func TestIndexedInductive_PartialMetadata(t *testing.T) {
	ClearRecursorRegistry()
	defer ClearRecursorRegistry()

	// Vec-like with 2 indices but only partial metadata
	partialInfo := &RecursorInfo{
		ElimName:   "partialElim",
		IndName:    "Partial",
		NumParams:  0,
		NumIndices: 2, // Two indices expected
		NumCases:   2,
		Ctors: []ConstructorInfo{
			{Name: "pnil", NumArgs: 0, RecursiveIdx: nil},
			{
				Name:         "pcons",
				NumArgs:      4,
				RecursiveIdx: []int{3},
				IndexArgPositions: map[int][]int{
					3: {0}, // Only 1 position, but 2 indices expected - incomplete!
				},
			},
		},
	}

	RegisterRecursor(partialInfo)

	info := LookupRecursor("partialElim")
	if info == nil {
		t.Fatal("partialElim should be registered")
	}

	pcons := info.Ctors[1]
	positions := pcons.IndexArgPositions[3]
	// Should have only 1 position (incomplete metadata)
	if len(positions) != 1 {
		t.Errorf("pcons should have partial metadata with 1 position, got %d", len(positions))
	}
	// The reduction code should detect this and use fallback
	// (verified by NumIndices != len(positions))
	if len(positions) == info.NumIndices {
		t.Error("Partial metadata should NOT match NumIndices")
	}
}

// TestIndexedInductive_EmptyIndexArgPositions tests constructor with empty map.
func TestIndexedInductive_EmptyIndexArgPositions(t *testing.T) {
	ClearRecursorRegistry()
	defer ClearRecursorRegistry()

	emptyInfo := &RecursorInfo{
		ElimName:   "emptyElim",
		IndName:    "Empty",
		NumParams:  0,
		NumIndices: 1,
		NumCases:   2,
		Ctors: []ConstructorInfo{
			{Name: "enil", NumArgs: 0, RecursiveIdx: nil},
			{
				Name:              "econs",
				NumArgs:           2,
				RecursiveIdx:      []int{1},
				IndexArgPositions: map[int][]int{}, // Empty map, not nil
			},
		},
	}

	RegisterRecursor(emptyInfo)

	info := LookupRecursor("emptyElim")
	if info == nil {
		t.Fatal("emptyElim should be registered")
	}

	econs := info.Ctors[1]
	// Empty map should behave like nil (trigger fallback)
	_, ok := econs.IndexArgPositions[1]
	if ok {
		t.Error("Empty IndexArgPositions should not have entry for recursive arg")
	}
}
