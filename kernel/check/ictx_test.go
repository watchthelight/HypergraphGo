package check

import "testing"

// --- ICtx Deep Nesting Tests ---

// TestICtx_DeepNesting_5Levels verifies ICtx extension chain at 5 levels deep.
func TestICtx_DeepNesting_5Levels(t *testing.T) {
	ictx := NewICtx()

	// Chain 5 levels deep
	ictx1 := ictx.Extend()
	ictx2 := ictx1.Extend()
	ictx3 := ictx2.Extend()
	ictx4 := ictx3.Extend()
	ictx5 := ictx4.Extend()

	// Original should still be depth 0
	if ictx.depth != 0 {
		t.Errorf("Original ICtx depth = %d, want 0", ictx.depth)
	}

	// Each level should have correct depth
	depths := []struct {
		ctx   *ICtx
		depth int
	}{
		{ictx1, 1}, {ictx2, 2}, {ictx3, 3}, {ictx4, 4}, {ictx5, 5},
	}
	for i, d := range depths {
		if d.ctx.depth != d.depth {
			t.Errorf("ictx%d.depth = %d, want %d", i+1, d.ctx.depth, d.depth)
		}
	}

	// Verify all indices are valid at deepest level
	for ix := 0; ix < 5; ix++ {
		if !ictx5.CheckIVar(ix) {
			t.Errorf("ictx5.CheckIVar(%d) = false, want true", ix)
		}
	}
	// Index 5 should be invalid
	if ictx5.CheckIVar(5) {
		t.Error("ictx5.CheckIVar(5) = true, want false")
	}
}

// TestICtx_DeepNesting_10Levels verifies ICtx extension chain at 10 levels.
func TestICtx_DeepNesting_10Levels(t *testing.T) {
	ictx := NewICtx()

	// Build 10 levels
	ctxs := make([]*ICtx, 11)
	ctxs[0] = ictx
	for i := 1; i <= 10; i++ {
		ctxs[i] = ctxs[i-1].Extend()
	}

	// Verify each level has correct depth
	for i, ctx := range ctxs {
		if ctx.depth != i {
			t.Errorf("ctxs[%d].depth = %d, want %d", i, ctx.depth, i)
		}
	}

	// At depth 10, indices 0-9 should be valid
	for ix := 0; ix < 10; ix++ {
		if !ctxs[10].CheckIVar(ix) {
			t.Errorf("ctxs[10].CheckIVar(%d) = false, want true", ix)
		}
	}
	// Index 10 should be invalid
	if ctxs[10].CheckIVar(10) {
		t.Error("ctxs[10].CheckIVar(10) = true, want false")
	}
}

// TestChecker_PushIVar_DeferCleanup verifies that defer cleanup works correctly.
func TestChecker_PushIVar_DeferCleanup(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())

	func() {
		pop := checker.PushIVar()
		defer pop()

		if checker.ICtxDepth() != 1 {
			t.Errorf("Inside func, depth = %d, want 1", checker.ICtxDepth())
		}
	}()

	// After func returns, pop should have been called via defer
	if checker.ICtxDepth() != 0 {
		t.Errorf("After defer pop, depth = %d, want 0", checker.ICtxDepth())
	}
}

// TestChecker_PushIVar_NestedDefers verifies nested defer cleanup.
func TestChecker_PushIVar_NestedDefers(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())

	func() {
		pop1 := checker.PushIVar()
		defer pop1()

		func() {
			pop2 := checker.PushIVar()
			defer pop2()

			func() {
				pop3 := checker.PushIVar()
				defer pop3()

				if checker.ICtxDepth() != 3 {
					t.Errorf("Innermost depth = %d, want 3", checker.ICtxDepth())
				}
			}()

			if checker.ICtxDepth() != 2 {
				t.Errorf("Middle depth = %d, want 2", checker.ICtxDepth())
			}
		}()

		if checker.ICtxDepth() != 1 {
			t.Errorf("Outer depth = %d, want 1", checker.ICtxDepth())
		}
	}()

	if checker.ICtxDepth() != 0 {
		t.Errorf("Final depth = %d, want 0", checker.ICtxDepth())
	}
}

// --- Boundary Condition Tests ---

// TestChecker_CheckIVar_LargNegativeIndex verifies rejection of large negative indices.
func TestChecker_CheckIVar_LargeNegativeIndex(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())
	pop := checker.PushIVar()
	defer pop()

	testCases := []int{-1, -100, -1000000}
	for _, ix := range testCases {
		if checker.CheckIVar(ix) {
			t.Errorf("CheckIVar(%d) = true, want false", ix)
		}
	}
}

// TestChecker_CheckIVar_BoundaryExact tests exact boundary conditions.
func TestChecker_CheckIVar_BoundaryExact(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())

	// Push 3 interval variables
	pop1 := checker.PushIVar()
	pop2 := checker.PushIVar()
	pop3 := checker.PushIVar()
	defer pop1()
	defer pop2()
	defer pop3()

	// Test exact boundaries
	tests := []struct {
		ix    int
		valid bool
	}{
		{-1, false}, // below range
		{0, true},   // first valid
		{1, true},   // middle
		{2, true},   // last valid
		{3, false},  // just over boundary
		{4, false},  // beyond boundary
	}

	for _, tt := range tests {
		got := checker.CheckIVar(tt.ix)
		if got != tt.valid {
			t.Errorf("CheckIVar(%d) = %v, want %v", tt.ix, got, tt.valid)
		}
	}
}

// TestChecker_ICtxDepth_Accuracy verifies depth tracking after multiple push/pop.
func TestChecker_ICtxDepth_Accuracy(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())

	// Push 5 in LIFO order (stack-like)
	pop1 := checker.PushIVar()
	if checker.ICtxDepth() != 1 {
		t.Errorf("After push 1: depth = %d, want 1", checker.ICtxDepth())
	}

	pop2 := checker.PushIVar()
	if checker.ICtxDepth() != 2 {
		t.Errorf("After push 2: depth = %d, want 2", checker.ICtxDepth())
	}

	pop3 := checker.PushIVar()
	if checker.ICtxDepth() != 3 {
		t.Errorf("After push 3: depth = %d, want 3", checker.ICtxDepth())
	}

	pop4 := checker.PushIVar()
	if checker.ICtxDepth() != 4 {
		t.Errorf("After push 4: depth = %d, want 4", checker.ICtxDepth())
	}

	pop5 := checker.PushIVar()
	if checker.ICtxDepth() != 5 {
		t.Errorf("After push 5: depth = %d, want 5", checker.ICtxDepth())
	}

	// Pop in LIFO order
	pop5()
	if checker.ICtxDepth() != 4 {
		t.Errorf("After pop5: depth = %d, want 4", checker.ICtxDepth())
	}

	pop4()
	if checker.ICtxDepth() != 3 {
		t.Errorf("After pop4: depth = %d, want 3", checker.ICtxDepth())
	}

	pop3()
	if checker.ICtxDepth() != 2 {
		t.Errorf("After pop3: depth = %d, want 2", checker.ICtxDepth())
	}

	pop2()
	if checker.ICtxDepth() != 1 {
		t.Errorf("After pop2: depth = %d, want 1", checker.ICtxDepth())
	}

	pop1()
	if checker.ICtxDepth() != 0 {
		t.Errorf("After pop1: depth = %d, want 0", checker.ICtxDepth())
	}
}

// --- Complex Push/Pop Patterns ---

// TestChecker_PushIVar_InterleavedPushPop tests non-stack order pop behavior.
func TestChecker_PushIVar_InterleavedPushPop(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())

	// Push three
	pop1 := checker.PushIVar() // depth 1
	pop2 := checker.PushIVar() // depth 2
	pop3 := checker.PushIVar() // depth 3

	if checker.ICtxDepth() != 3 {
		t.Fatalf("After 3 pushes: depth = %d, want 3", checker.ICtxDepth())
	}

	// Pop in non-LIFO order (pop2 first)
	// This should restore to depth at time of push (depth 1)
	pop2()
	// After pop2, depth should be restored to what it was when push2 was called: 1
	if checker.ICtxDepth() != 1 {
		t.Errorf("After pop2: depth = %d, want 1 (restored to push2 state)", checker.ICtxDepth())
	}

	// Now pop1 and pop3 can be called but behavior depends on implementation
	pop3()
	pop1()

	// Final state
	if checker.ICtxDepth() != 0 {
		t.Errorf("After all pops: depth = %d, want 0", checker.ICtxDepth())
	}
}

// TestChecker_PushIVar_DoublePop tests calling same pop function twice.
func TestChecker_PushIVar_DoublePop(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())

	pop := checker.PushIVar()
	if checker.ICtxDepth() != 1 {
		t.Fatalf("After push: depth = %d, want 1", checker.ICtxDepth())
	}

	// First pop
	pop()
	if checker.ICtxDepth() != 0 {
		t.Errorf("After first pop: depth = %d, want 0", checker.ICtxDepth())
	}

	// Second pop (should be idempotent or safe)
	pop()
	if checker.ICtxDepth() != 0 {
		t.Errorf("After second pop: depth = %d, want 0", checker.ICtxDepth())
	}
}

// TestChecker_PushIVar_FirstPushCreatesContext verifies context creation on first push.
func TestChecker_PushIVar_FirstPushCreatesContext(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())

	// Initially no context
	if checker.ictx != nil {
		t.Error("Initial ictx should be nil")
	}
	if checker.ICtxDepth() != 0 {
		t.Errorf("Initial ICtxDepth = %d, want 0", checker.ICtxDepth())
	}

	// Push creates context
	pop := checker.PushIVar()
	if checker.ictx == nil {
		t.Error("After push, ictx should not be nil")
	}
	if checker.ICtxDepth() != 1 {
		t.Errorf("After push, ICtxDepth = %d, want 1", checker.ICtxDepth())
	}

	// Pop destroys context when returning to 0
	pop()
	if checker.ictx != nil {
		t.Error("After pop to 0, ictx should be nil")
	}
	if checker.ICtxDepth() != 0 {
		t.Errorf("After pop, ICtxDepth = %d, want 0", checker.ICtxDepth())
	}
}

// TestChecker_CheckIVar_WithNilContext confirms behavior when ictx is nil.
func TestChecker_CheckIVar_WithNilContext(t *testing.T) {
	checker := NewChecker(NewGlobalEnv())

	// ictx is nil initially
	if checker.ictx != nil {
		t.Error("Expected nil ictx initially")
	}

	// All indices should be invalid
	for ix := -5; ix <= 5; ix++ {
		if checker.CheckIVar(ix) {
			t.Errorf("CheckIVar(%d) with nil ictx = true, want false", ix)
		}
	}
}

// --- ICtx Struct Method Tests ---

// TestICtx_CheckIVar_AllBoundaryConditions comprehensively tests CheckIVar.
func TestICtx_CheckIVar_AllBoundaryConditions(t *testing.T) {
	tests := []struct {
		name  string
		depth int
		tests []struct {
			ix    int
			valid bool
		}
	}{
		{
			name:  "depth=0 (empty)",
			depth: 0,
			tests: []struct {
				ix    int
				valid bool
			}{
				{-1, false},
				{0, false},
				{1, false},
			},
		},
		{
			name:  "depth=1",
			depth: 1,
			tests: []struct {
				ix    int
				valid bool
			}{
				{-1, false},
				{0, true},
				{1, false},
				{2, false},
			},
		},
		{
			name:  "depth=5",
			depth: 5,
			tests: []struct {
				ix    int
				valid bool
			}{
				{-1, false},
				{0, true},
				{2, true},
				{4, true},
				{5, false},
				{10, false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ictx := &ICtx{depth: tt.depth}
			for _, tc := range tt.tests {
				got := ictx.CheckIVar(tc.ix)
				if got != tc.valid {
					t.Errorf("ICtx{depth=%d}.CheckIVar(%d) = %v, want %v",
						tt.depth, tc.ix, got, tc.valid)
				}
			}
		})
	}
}

// TestICtx_Extend_ImmutabilityChain verifies Extend doesn't modify originals.
func TestICtx_Extend_ImmutabilityChain(t *testing.T) {
	base := NewICtx()

	// Create a chain
	ext1 := base.Extend()
	ext2 := ext1.Extend()
	ext3 := ext2.Extend()

	// Verify all original depths are unchanged
	if base.depth != 0 {
		t.Errorf("base.depth = %d after extensions, want 0", base.depth)
	}
	if ext1.depth != 1 {
		t.Errorf("ext1.depth = %d, want 1", ext1.depth)
	}
	if ext2.depth != 2 {
		t.Errorf("ext2.depth = %d, want 2", ext2.depth)
	}
	if ext3.depth != 3 {
		t.Errorf("ext3.depth = %d, want 3", ext3.depth)
	}

	// Create another branch from ext1
	branch := ext1.Extend()
	if branch.depth != 2 {
		t.Errorf("branch.depth = %d, want 2", branch.depth)
	}
	// Original ext1 should still be depth 1
	if ext1.depth != 1 {
		t.Errorf("ext1.depth changed to %d after branch, want 1", ext1.depth)
	}
}
