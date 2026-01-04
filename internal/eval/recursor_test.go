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
