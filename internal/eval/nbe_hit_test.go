package eval

import (
	"testing"
)

// ============================================================================
// VHITPathCtor Tests
// ============================================================================

func TestVHITPathCtor_IsValue(t *testing.T) {
	t.Parallel()
	v := VHITPathCtor{
		HITName:  "S1",
		CtorName: "loop",
		Args:     nil,
		IArgs:    []Value{VIVar{Level: 0}},
	}
	// Verify VHITPathCtor implements Value interface
	var _ Value = v
	v.isValue() // Should not panic
}

func TestBoundaryVal_Structure(t *testing.T) {
	t.Parallel()
	bv := BoundaryVal{
		AtZero: VSort{Level: 0},
		AtOne:  VSort{Level: 1},
	}
	if bv.AtZero == nil || bv.AtOne == nil {
		t.Error("BoundaryVal fields should not be nil")
	}
}

// ============================================================================
// evalHITApp Tests
// ============================================================================

func TestEvalHITApp_ReduceAtI0(t *testing.T) {
	t.Parallel()
	// loop @ i0 --> base
	baseVal := VGlobal{Name: "base"}
	boundaries := []BoundaryVal{
		{AtZero: baseVal, AtOne: baseVal},
	}

	result := evalHITApp("S1", "loop", nil, []Value{VI0{}}, boundaries)

	if g, ok := result.(VGlobal); !ok || g.Name != "base" {
		t.Errorf("loop @ i0 should reduce to base, got %T", result)
	}
}

func TestEvalHITApp_ReduceAtI1(t *testing.T) {
	t.Parallel()
	// loop @ i1 --> base
	baseVal := VGlobal{Name: "base"}
	boundaries := []BoundaryVal{
		{AtZero: baseVal, AtOne: baseVal},
	}

	result := evalHITApp("S1", "loop", nil, []Value{VI1{}}, boundaries)

	if g, ok := result.(VGlobal); !ok || g.Name != "base" {
		t.Errorf("loop @ i1 should reduce to base, got %T", result)
	}
}

func TestEvalHITApp_StuckAtIVar(t *testing.T) {
	t.Parallel()
	// loop @ i (variable) --> stuck
	baseVal := VGlobal{Name: "base"}
	boundaries := []BoundaryVal{
		{AtZero: baseVal, AtOne: baseVal},
	}

	result := evalHITApp("S1", "loop", nil, []Value{VIVar{Level: 0}}, boundaries)

	hitPath, ok := result.(VHITPathCtor)
	if !ok {
		t.Fatalf("loop @ i should be stuck VHITPathCtor, got %T", result)
	}
	if hitPath.HITName != "S1" || hitPath.CtorName != "loop" {
		t.Errorf("wrong HIT/ctor: %s/%s", hitPath.HITName, hitPath.CtorName)
	}
	if len(hitPath.IArgs) != 1 {
		t.Errorf("expected 1 IArg, got %d", len(hitPath.IArgs))
	}
}

func TestEvalHITApp_MultipleIArgs_FirstReduces(t *testing.T) {
	t.Parallel()
	// Higher-level path: eq @ i0 @ j --> reduces at first dimension
	leftVal := VGlobal{Name: "left"}
	rightVal := VGlobal{Name: "right"}
	boundaries := []BoundaryVal{
		{AtZero: leftVal, AtOne: rightVal},
		{AtZero: VGlobal{Name: "inner0"}, AtOne: VGlobal{Name: "inner1"}},
	}

	result := evalHITApp("Quot", "eq", nil, []Value{VI0{}, VIVar{Level: 0}}, boundaries)

	if g, ok := result.(VGlobal); !ok || g.Name != "left" {
		t.Errorf("eq @ i0 @ j should reduce to left, got %v", result)
	}
}

func TestEvalHITApp_NoArgs(t *testing.T) {
	t.Parallel()
	// Path with no interval args stays stuck
	result := evalHITApp("S1", "loop", nil, nil, nil)

	hitPath, ok := result.(VHITPathCtor)
	if !ok {
		t.Fatalf("empty iargs should produce VHITPathCtor, got %T", result)
	}
	if hitPath.CtorName != "loop" {
		t.Errorf("wrong ctor: %s", hitPath.CtorName)
	}
}

func TestEvalHITApp_NoBoundaryInfo(t *testing.T) {
	t.Parallel()
	// Even with endpoint, if no boundary info, stays stuck
	result := evalHITApp("S1", "loop", nil, []Value{VI0{}}, nil)

	_, ok := result.(VHITPathCtor)
	if !ok {
		t.Errorf("no boundaries should mean stuck, got %T", result)
	}
}

func TestEvalHITApp_WithArgs(t *testing.T) {
	t.Parallel()
	// Path constructor with term arguments
	args := []Value{VGlobal{Name: "A"}, VGlobal{Name: "x"}}
	baseVal := VGlobal{Name: "q"}
	boundaries := []BoundaryVal{{AtZero: baseVal, AtOne: baseVal}}

	result := evalHITApp("Quot", "eq", args, []Value{VI0{}}, boundaries)

	if g, ok := result.(VGlobal); !ok || g.Name != "q" {
		t.Errorf("should reduce to boundary, got %T", result)
	}
}

// ============================================================================
// lookupHITBoundaries Tests
// ============================================================================

func TestLookupHITBoundaries_NoRecursor(t *testing.T) {
	// Not parallel: modifies global recursor registry
	ClearRecursorRegistry()
	defer ClearRecursorRegistry()

	result := lookupHITBoundaries("Unknown", "path", nil)
	if result != nil {
		t.Errorf("unknown HIT should return nil boundaries, got %v", result)
	}
}

func TestLookupHITBoundaries_NonHIT(t *testing.T) {
	// Not parallel: modifies global recursor registry
	ClearRecursorRegistry()
	defer ClearRecursorRegistry()

	// Register a non-HIT recursor
	RegisterRecursor(&RecursorInfo{
		ElimName:  "Nat-elim",
		IndName:   "Nat",
		IsHIT:     false,
		NumParams: 0,
		NumCases:  2,
	})

	result := lookupHITBoundaries("Nat", "zero", nil)
	if result != nil {
		t.Errorf("non-HIT should return nil boundaries, got %v", result)
	}
}

func TestLookupHITBoundaries_WithPathCtor(t *testing.T) {
	// Not parallel: modifies global recursor registry
	ClearRecursorRegistry()
	defer ClearRecursorRegistry()

	baseVal := VGlobal{Name: "base"}
	RegisterRecursor(&RecursorInfo{
		ElimName:  "S1-elim",
		IndName:   "S1",
		IsHIT:     true,
		NumParams: 1,
		NumCases:  1,
		PathCtors: []PathConstructorInfo{
			{
				Name:  "loop",
				Level: 1,
				Boundaries: []BoundarySpec{
					{AtZeroVal: baseVal, AtOneVal: baseVal},
				},
			},
		},
	})

	result := lookupHITBoundaries("S1", "loop", nil)
	if result == nil {
		t.Fatal("should find boundaries for loop")
	}
	if len(result) != 1 {
		t.Errorf("expected 1 boundary, got %d", len(result))
	}
	if g, ok := result[0].AtZero.(VGlobal); !ok || g.Name != "base" {
		t.Errorf("wrong AtZero boundary: %v", result[0].AtZero)
	}
}

func TestLookupHITBoundaries_UnknownPathCtor(t *testing.T) {
	// Not parallel: modifies global recursor registry
	ClearRecursorRegistry()
	defer ClearRecursorRegistry()

	RegisterRecursor(&RecursorInfo{
		ElimName:  "S1-elim",
		IndName:   "S1",
		IsHIT:     true,
		NumParams: 1,
		NumCases:  1,
		PathCtors: []PathConstructorInfo{
			{Name: "loop", Level: 1},
		},
	})

	result := lookupHITBoundaries("S1", "unknown", nil)
	if result != nil {
		t.Errorf("unknown ctor should return nil, got %v", result)
	}
}

// ============================================================================
// tryHITPathReduction Tests
// ============================================================================

func TestTryHITPathReduction_UnknownCtor(t *testing.T) {
	t.Parallel()
	info := &RecursorInfo{
		IsHIT:     true,
		PathCtors: []PathConstructorInfo{{Name: "loop"}},
	}
	hitPath := VHITPathCtor{
		HITName:  "S1",
		CtorName: "unknown",
		IArgs:    []Value{VIVar{Level: 0}},
	}

	result := tryHITPathReduction(info, nil, hitPath, nil)
	if result != nil {
		t.Errorf("unknown ctor should return nil, got %v", result)
	}
}

func TestTryHITPathReduction_AtI0(t *testing.T) {
	t.Parallel()
	baseVal := VGlobal{Name: "base"}
	info := &RecursorInfo{
		IsHIT:     true,
		NumParams: 1,
		NumCases:  1,
		PathCtors: []PathConstructorInfo{{Name: "loop", Level: 1}},
	}
	hitPath := VHITPathCtor{
		HITName:    "S1",
		CtorName:   "loop",
		IArgs:      []Value{VI0{}},
		Boundaries: []BoundaryVal{{AtZero: baseVal, AtOne: baseVal}},
	}

	result := tryHITPathReduction(info, nil, hitPath, nil)
	if g, ok := result.(VGlobal); !ok || g.Name != "base" {
		t.Errorf("at i0 should reduce to AtZero, got %v", result)
	}
}

func TestTryHITPathReduction_AtI1(t *testing.T) {
	t.Parallel()
	leftVal := VGlobal{Name: "left"}
	rightVal := VGlobal{Name: "right"}
	info := &RecursorInfo{
		IsHIT:     true,
		NumParams: 1,
		NumCases:  1,
		PathCtors: []PathConstructorInfo{{Name: "eq", Level: 1}},
	}
	hitPath := VHITPathCtor{
		HITName:    "Quot",
		CtorName:   "eq",
		IArgs:      []Value{VI1{}},
		Boundaries: []BoundaryVal{{AtZero: leftVal, AtOne: rightVal}},
	}

	result := tryHITPathReduction(info, nil, hitPath, nil)
	if g, ok := result.(VGlobal); !ok || g.Name != "right" {
		t.Errorf("at i1 should reduce to AtOne, got %v", result)
	}
}

func TestTryHITPathReduction_AtEndpointWithExtraArgs(t *testing.T) {
	t.Parallel()
	// Reducing at endpoint then applying extra args
	// Use a simple value that Apply can work with
	baseVal := VGlobal{Name: "baseCase"}
	info := &RecursorInfo{
		IsHIT:     true,
		NumParams: 1,
		NumCases:  1,
		PathCtors: []PathConstructorInfo{{Name: "loop", Level: 1}},
	}
	hitPath := VHITPathCtor{
		HITName:    "S1",
		CtorName:   "loop",
		IArgs:      []Value{VI0{}},
		Boundaries: []BoundaryVal{{AtZero: baseVal, AtOne: baseVal}},
	}

	// When we have extra args, Apply is called on the boundary value
	// Since baseVal is VGlobal, Apply will create a neutral term
	result := tryHITPathReduction(info, nil, hitPath, []Value{VSort{Level: 0}})

	// Should apply extra arg to the boundary value
	if result == nil {
		t.Error("should reduce and apply extra args")
	}
}

func TestTryHITPathReduction_Stuck_ApplyPathCase(t *testing.T) {
	t.Parallel()
	// Not at endpoint - apply path case to interval
	// Use a VGlobal as the path case - it will create a neutral NPathApp
	ploop := VGlobal{Name: "ploop"}
	info := &RecursorInfo{
		IsHIT:     true,
		NumParams: 1,
		NumCases:  1, // pbase
		PathCtors: []PathConstructorInfo{{Name: "loop", Level: 1}},
	}
	// sp structure: [param, P (motive), pbase (point case), ploop (path case)]
	// pathCasesStart = NumParams(1) + 1(motive) + NumCases(1) = 3
	// pathCaseIdx = 3 + 0 = 3, so sp[3] = ploop
	sp := []Value{VSort{Level: 0}, VSort{Level: 0}, VGlobal{Name: "pbase"}, ploop}

	hitPath := VHITPathCtor{
		HITName:  "S1",
		CtorName: "loop",
		Args:     nil,
		IArgs:    []Value{VIVar{Level: 0}},
	}

	result := tryHITPathReduction(info, sp, hitPath, nil)

	// Should apply ploop @ i (returns a neutral)
	if result == nil {
		t.Error("should apply path case to interval")
	}
}

func TestTryHITPathReduction_NotEnoughArgs(t *testing.T) {
	t.Parallel()
	info := &RecursorInfo{
		IsHIT:     true,
		NumParams: 1,
		NumCases:  1,
		PathCtors: []PathConstructorInfo{{Name: "loop", Level: 1}},
	}
	hitPath := VHITPathCtor{
		HITName:  "S1",
		CtorName: "loop",
		IArgs:    []Value{VIVar{Level: 0}},
	}

	// sp is too short
	result := tryHITPathReduction(info, []Value{VSort{Level: 0}}, hitPath, nil)
	if result != nil {
		t.Errorf("should return nil when not enough args, got %v", result)
	}
}

func TestTryHITPathReduction_WithTermArgs(t *testing.T) {
	t.Parallel()
	// Path constructor with term arguments
	// Use a VGlobal as the path case - it will create neutral values when applied
	pEq := VGlobal{Name: "pEq"}
	info := &RecursorInfo{
		IsHIT:     true,
		NumParams: 2, // A, R
		NumCases:  1, // q
		PathCtors: []PathConstructorInfo{{Name: "eq", Level: 1}},
	}
	// sp structure: [A, R (params), P (motive), q (point case), pEq (path case)]
	// pathCasesStart = NumParams(2) + 1(motive) + NumCases(1) = 4
	// pathCaseIdx = 4 + 0 = 4, so sp[4] = pEq
	sp := []Value{VSort{Level: 0}, VSort{Level: 0}, VSort{Level: 0}, VGlobal{Name: "q"}, pEq}

	hitPath := VHITPathCtor{
		HITName:  "Quot",
		CtorName: "eq",
		Args:     []Value{VGlobal{Name: "x"}, VGlobal{Name: "y"}},
		IArgs:    []Value{VIVar{Level: 0}},
	}

	result := tryHITPathReduction(info, sp, hitPath, nil)
	// Should apply pEq to term args then interval
	if result == nil {
		t.Error("should reduce with term args")
	}
}
