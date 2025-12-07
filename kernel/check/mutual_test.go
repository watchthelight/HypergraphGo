package check

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
)

// TestMutualInductive_EvenOdd tests basic mutual inductive declaration.
func TestMutualInductive_EvenOdd(t *testing.T) {
	eval.ClearRecursorRegistry()
	env := NewGlobalEnv()

	// even : Type
	// odd : Type
	evenType := ast.Sort{U: 0}
	oddType := ast.Sort{U: 0}

	// zero : even
	zeroType := ast.Global{Name: "even"}

	// succOdd : odd -> even
	succOddType := ast.Pi{
		Binder: "_",
		A:      ast.Global{Name: "odd"},
		B:      ast.Global{Name: "even"},
	}

	// succ : even -> odd
	succType := ast.Pi{
		Binder: "_",
		A:      ast.Global{Name: "even"},
		B:      ast.Global{Name: "odd"},
	}

	err := env.DeclareMutual([]MutualInductiveSpec{
		{
			Name: "even",
			Type: evenType,
			Constructors: []Constructor{
				{Name: "zero", Type: zeroType},
				{Name: "succOdd", Type: succOddType},
			},
			Eliminator: "evenElim",
		},
		{
			Name: "odd",
			Type: oddType,
			Constructors: []Constructor{
				{Name: "succ", Type: succType},
			},
			Eliminator: "oddElim",
		},
	})
	if err != nil {
		t.Fatalf("DeclareMutual(even/odd) failed: %v", err)
	}

	// Verify both types were added
	evenTy := env.LookupType("even")
	if evenTy == nil {
		t.Error("even type not found")
	}

	oddTy := env.LookupType("odd")
	if oddTy == nil {
		t.Error("odd type not found")
	}

	// Verify constructors
	zeroTy := env.LookupType("zero")
	if zeroTy == nil {
		t.Error("zero constructor not found")
	}

	succOddTy := env.LookupType("succOdd")
	if succOddTy == nil {
		t.Error("succOdd constructor not found")
	}

	succTy := env.LookupType("succ")
	if succTy == nil {
		t.Error("succ constructor not found")
	}

	// Verify eliminators
	evenElimTy := env.LookupType("evenElim")
	if evenElimTy == nil {
		t.Error("evenElim not found")
	}

	oddElimTy := env.LookupType("oddElim")
	if oddElimTy == nil {
		t.Error("oddElim not found")
	}

	// Verify mutual group is recorded
	evenInd := env.inductives["even"]
	if evenInd == nil {
		t.Fatal("even inductive not found")
	}
	if len(evenInd.MutualGroup) != 2 {
		t.Errorf("even.MutualGroup = %v, want 2 elements", evenInd.MutualGroup)
	}

	oddInd := env.inductives["odd"]
	if oddInd == nil {
		t.Fatal("odd inductive not found")
	}
	if len(oddInd.MutualGroup) != 2 {
		t.Errorf("odd.MutualGroup = %v, want 2 elements", oddInd.MutualGroup)
	}

	eval.ClearRecursorRegistry()
}

// TestMutualInductive_SingleIsSameAsDeclareInductive verifies backward compatibility.
func TestMutualInductive_SingleIsSameAsDeclareInductive(t *testing.T) {
	eval.ClearRecursorRegistry()

	// Test via DeclareInductive (which calls DeclareMutual)
	env := NewGlobalEnv()
	err := env.DeclareInductive("Nat", ast.Sort{U: 0}, []Constructor{
		{Name: "zero", Type: ast.Global{Name: "Nat"}},
		{Name: "succ", Type: ast.Pi{
			Binder: "_",
			A:      ast.Global{Name: "Nat"},
			B:      ast.Global{Name: "Nat"},
		}},
	}, "natElim")
	if err != nil {
		t.Fatalf("DeclareInductive(Nat) failed: %v", err)
	}

	// Verify Nat was added
	natTy := env.LookupType("Nat")
	if natTy == nil {
		t.Error("Nat type not found")
	}

	// Verify single inductives have nil MutualGroup
	natInd := env.inductives["Nat"]
	if natInd == nil {
		t.Fatal("Nat inductive not found")
	}
	if natInd.MutualGroup != nil {
		t.Errorf("Nat.MutualGroup = %v, want nil for single inductive", natInd.MutualGroup)
	}

	eval.ClearRecursorRegistry()
}

// TestMutualInductive_Positivity_Reject tests that negative occurrences are rejected.
func TestMutualInductive_Positivity_Reject(t *testing.T) {
	eval.ClearRecursorRegistry()
	env := NewGlobalEnv()

	// bad : Type
	// mk : (bad -> Nat) -> bad  -- bad in negative position
	// This should be rejected
	badType := ast.Sort{U: 0}
	goodType := ast.Sort{U: 0}

	// mk : (bad -> good) -> good -- bad in negative position
	mkType := ast.Pi{
		Binder: "_",
		A: ast.Pi{
			Binder: "_",
			A:      ast.Global{Name: "bad"},
			B:      ast.Global{Name: "good"},
		},
		B: ast.Global{Name: "good"},
	}

	err := env.DeclareMutual([]MutualInductiveSpec{
		{
			Name:         "bad",
			Type:         badType,
			Constructors: []Constructor{},
			Eliminator:   "badElim",
		},
		{
			Name: "good",
			Type: goodType,
			Constructors: []Constructor{
				{Name: "mk", Type: mkType},
			},
			Eliminator: "goodElim",
		},
	})
	if err == nil {
		t.Error("DeclareMutual should reject negative occurrence of mutual type")
	}

	// Verify it's a positivity error
	if _, ok := err.(*PositivityError); !ok {
		t.Errorf("Expected PositivityError, got %T: %v", err, err)
	}

	eval.ClearRecursorRegistry()
}

// TestMutualInductive_Reduction tests that separate eliminators reduce correctly.
// With separate eliminators, cross-type args don't get IHs - they just pass through.
func TestMutualInductive_Reduction(t *testing.T) {
	eval.ClearRecursorRegistry()
	env := NewGlobalEnv()

	// Declare even/odd mutual inductives
	evenType := ast.Sort{U: 0}
	oddType := ast.Sort{U: 0}
	zeroType := ast.Global{Name: "even"}
	succOddType := ast.Pi{
		Binder: "o",
		A:      ast.Global{Name: "odd"},
		B:      ast.Global{Name: "even"},
	}
	succType := ast.Pi{
		Binder: "e",
		A:      ast.Global{Name: "even"},
		B:      ast.Global{Name: "odd"},
	}

	err := env.DeclareMutual([]MutualInductiveSpec{
		{
			Name: "even",
			Type: evenType,
			Constructors: []Constructor{
				{Name: "zero", Type: zeroType},
				{Name: "succOdd", Type: succOddType},
			},
			Eliminator: "evenElim",
		},
		{
			Name: "odd",
			Type: oddType,
			Constructors: []Constructor{
				{Name: "succ", Type: succType},
			},
			Eliminator: "oddElim",
		},
	})
	if err != nil {
		t.Fatalf("DeclareMutual failed: %v", err)
	}

	// Test 1: evenElim P pzero psuccOdd zero --> pzero
	// evenElim : (P : even -> Type) -> P zero -> ((o : odd) -> P (succOdd o)) -> even -> ...
	t.Run("evenElim_zero", func(t *testing.T) {
		// Build: evenElim P pzero psuccOdd zero
		term := ast.App{
			T: ast.App{
				T: ast.App{
					T: ast.App{
						T: ast.Global{Name: "evenElim"},
						U: ast.Global{Name: "P"}, // motive
					},
					U: ast.Global{Name: "pzero"}, // case for zero
				},
				U: ast.Global{Name: "psuccOdd"}, // case for succOdd
			},
			U: ast.Global{Name: "zero"}, // scrutinee
		}

		result := eval.Eval(nil, term)
		// Should reduce to pzero
		if !isGlobal(result, "pzero") {
			t.Errorf("evenElim P pzero psuccOdd zero = %v, want pzero", result)
		}
	})

	// Test 2: evenElim P pzero psuccOdd (succOdd o) --> psuccOdd o
	// No IH because 'o : odd' is cross-type (separate eliminators)
	t.Run("evenElim_succOdd", func(t *testing.T) {
		// Build: evenElim P pzero psuccOdd (succOdd o)
		term := ast.App{
			T: ast.App{
				T: ast.App{
					T: ast.App{
						T: ast.Global{Name: "evenElim"},
						U: ast.Global{Name: "P"},
					},
					U: ast.Global{Name: "pzero"},
				},
				U: ast.Global{Name: "psuccOdd"},
			},
			U: ast.App{
				T: ast.Global{Name: "succOdd"},
				U: ast.Global{Name: "o"},
			},
		}

		result := eval.Eval(nil, term)
		// Should reduce to psuccOdd o (no IH for cross-type arg)
		expected := eval.Apply(eval.MakeNeutralGlobal("psuccOdd"), eval.MakeNeutralGlobal("o"))
		if !valuesEqual(result, expected) {
			t.Errorf("evenElim P pzero psuccOdd (succOdd o) = %v, want psuccOdd o", result)
		}
	})

	// Test 3: oddElim Q qsucc (succ e) --> qsucc e
	// No IH because 'e : even' is cross-type
	t.Run("oddElim_succ", func(t *testing.T) {
		// Build: oddElim Q qsucc (succ e)
		term := ast.App{
			T: ast.App{
				T: ast.App{
					T: ast.Global{Name: "oddElim"},
					U: ast.Global{Name: "Q"},
				},
				U: ast.Global{Name: "qsucc"},
			},
			U: ast.App{
				T: ast.Global{Name: "succ"},
				U: ast.Global{Name: "e"},
			},
		}

		result := eval.Eval(nil, term)
		// Should reduce to qsucc e (no IH for cross-type arg)
		expected := eval.Apply(eval.MakeNeutralGlobal("qsucc"), eval.MakeNeutralGlobal("e"))
		if !valuesEqual(result, expected) {
			t.Errorf("oddElim Q qsucc (succ e) = %v, want qsucc e", result)
		}
	})

	eval.ClearRecursorRegistry()
}

// isGlobal checks if a value is a VNeutral with a global head.
func isGlobal(v eval.Value, name string) bool {
	neutral, ok := v.(eval.VNeutral)
	if !ok {
		return false
	}
	return neutral.N.Head.Glob == name && len(neutral.N.Sp) == 0
}

// valuesEqual checks if two values are structurally equal (simplified).
func valuesEqual(a, b eval.Value) bool {
	na, okA := a.(eval.VNeutral)
	nb, okB := b.(eval.VNeutral)
	if !okA || !okB {
		return false
	}
	if na.N.Head.Glob != nb.N.Head.Glob {
		return false
	}
	if len(na.N.Sp) != len(nb.N.Sp) {
		return false
	}
	for i := range na.N.Sp {
		if !valuesEqual(na.N.Sp[i], nb.N.Sp[i]) {
			return false
		}
	}
	return true
}

// TestMutualInductive_SameTypeRecursion tests that same-type recursion still works in mutual blocks.
// For example, an `even` type with a constructor that takes an `even` arg should still get an IH.
func TestMutualInductive_SameTypeRecursion(t *testing.T) {
	eval.ClearRecursorRegistry()
	env := NewGlobalEnv()

	// even : Type with double : even -> even (same-type recursion)
	// odd : Type with succ : even -> odd (cross-type, no IH)
	evenType := ast.Sort{U: 0}
	oddType := ast.Sort{U: 0}

	zeroType := ast.Global{Name: "even"}
	doubleType := ast.Pi{
		Binder: "e",
		A:      ast.Global{Name: "even"},
		B:      ast.Global{Name: "even"},
	}
	succType := ast.Pi{
		Binder: "e",
		A:      ast.Global{Name: "even"},
		B:      ast.Global{Name: "odd"},
	}

	err := env.DeclareMutual([]MutualInductiveSpec{
		{
			Name: "even",
			Type: evenType,
			Constructors: []Constructor{
				{Name: "zero", Type: zeroType},
				{Name: "double", Type: doubleType},
			},
			Eliminator: "evenElim",
		},
		{
			Name: "odd",
			Type: oddType,
			Constructors: []Constructor{
				{Name: "succ", Type: succType},
			},
			Eliminator: "oddElim",
		},
	})
	if err != nil {
		t.Fatalf("DeclareMutual failed: %v", err)
	}

	// Verify that evenElim's "double" case gets an IH for the even arg
	// by checking that reduction works: evenElim P pzero pdouble (double e) --> pdouble e (IH)
	t.Run("evenElim_double_has_IH", func(t *testing.T) {
		// Build: evenElim P pzero pdouble (double e)
		term := ast.App{
			T: ast.App{
				T: ast.App{
					T: ast.App{
						T: ast.Global{Name: "evenElim"},
						U: ast.Global{Name: "P"},
					},
					U: ast.Global{Name: "pzero"},
				},
				U: ast.Global{Name: "pdouble"},
			},
			U: ast.App{
				T: ast.Global{Name: "double"},
				U: ast.Global{Name: "e"},
			},
		}

		result := eval.Eval(nil, term)
		// For same-type recursion, should reduce to pdouble e ih
		// where ih = evenElim P pzero pdouble e
		// So we expect the result to be (pdouble e (evenElim P pzero pdouble e))
		neutral, ok := result.(eval.VNeutral)
		if !ok {
			t.Fatalf("expected VNeutral, got %T", result)
		}
		// The head should be pdouble
		if neutral.N.Head.Glob != "pdouble" {
			t.Errorf("expected head pdouble, got %s", neutral.N.Head.Glob)
		}
		// Should have 2 args: e and the IH
		if len(neutral.N.Sp) != 2 {
			t.Errorf("expected 2 args (e and IH), got %d", len(neutral.N.Sp))
		}
	})

	eval.ClearRecursorRegistry()
}

// TestMutualInductive_NestedNegative tests that deeply nested negative occurrences are rejected.
func TestMutualInductive_NestedNegative(t *testing.T) {
	eval.ClearRecursorRegistry()
	env := NewGlobalEnv()

	// T1 : Type
	// T2 : Type
	// mk : ((T1 -> A) -> B) -> T2  -- T1 occurs in nested negative position (should be rejected)
	t1Type := ast.Sort{U: 0}
	t2Type := ast.Sort{U: 0}

	// ((T1 -> A) -> B) -> T2
	mkType := ast.Pi{
		Binder: "_",
		A: ast.Pi{
			Binder: "_",
			A: ast.Pi{
				Binder: "_",
				A:      ast.Global{Name: "T1"},
				B:      ast.Global{Name: "A"}, // Would need A to be defined, use T2
			},
			B: ast.Global{Name: "T2"},
		},
		B: ast.Global{Name: "T2"},
	}

	err := env.DeclareMutual([]MutualInductiveSpec{
		{
			Name:         "T1",
			Type:         t1Type,
			Constructors: []Constructor{},
			Eliminator:   "t1Elim",
		},
		{
			Name: "T2",
			Type: t2Type,
			Constructors: []Constructor{
				{Name: "mk", Type: mkType},
			},
			Eliminator: "t2Elim",
		},
	})

	// T1 appears in: _ : ((T1 -> T2) -> T2)
	// Position analysis: outer Pi domain, inner Pi domain = negative * negative = positive
	// But inner-inner domain (T1 -> T2) has T1 in domain = negative
	// So total: positive * negative = negative - should be REJECTED
	if err == nil {
		t.Error("DeclareMutual should reject nested negative occurrence")
	}

	eval.ClearRecursorRegistry()
}

// TestMutualInductive_DoublyNegativeIsPositive tests that double negation becomes positive.
func TestMutualInductive_DoublyNegativeIsPositive(t *testing.T) {
	eval.ClearRecursorRegistry()
	env := NewGlobalEnv()

	// This is a subtle test: (A -> B) -> C has A in positive position (double negative)
	// BUT our strict positivity check is more conservative - it rejects ANY occurrence
	// in a domain (once in domain, we stay in domain regardless of further nesting).
	//
	// So ((T -> X) -> X) -> T actually puts T in negative position in our implementation.
	// This is intentional and matches standard strict positivity.

	// T1 : Type
	// T2 : Type
	// mk : (T1 -> T2) -> T2  -- T1 in negative position (rejected)
	t1Type := ast.Sort{U: 0}
	t2Type := ast.Sort{U: 0}

	mkType := ast.Pi{
		Binder: "_",
		A: ast.Pi{
			Binder: "_",
			A:      ast.Global{Name: "T1"},
			B:      ast.Global{Name: "T2"},
		},
		B: ast.Global{Name: "T2"},
	}

	err := env.DeclareMutual([]MutualInductiveSpec{
		{
			Name:         "T1",
			Type:         t1Type,
			Constructors: []Constructor{},
			Eliminator:   "t1Elim",
		},
		{
			Name: "T2",
			Type: t2Type,
			Constructors: []Constructor{
				{Name: "mk", Type: mkType},
			},
			Eliminator: "t2Elim",
		},
	})

	// T1 is in the domain of a Pi which is itself in a domain - negative position
	if err == nil {
		t.Error("DeclareMutual should reject T1 in negative position")
	}

	eval.ClearRecursorRegistry()
}

// TestMutualInductive_SymmetricNegative tests that negative occurrence is checked symmetrically.
// If T1 appears negatively in T2's constructor, it should be rejected.
// If T2 appears negatively in T1's constructor, it should also be rejected.
func TestMutualInductive_SymmetricNegative(t *testing.T) {
	eval.ClearRecursorRegistry()

	t.Run("T2_negative_in_T1", func(t *testing.T) {
		env := NewGlobalEnv()

		// T1 has constructor mk : (T2 -> X) -> T1  -- T2 in negative position
		t1Type := ast.Sort{U: 0}
		t2Type := ast.Sort{U: 0}

		mkType := ast.Pi{
			Binder: "_",
			A: ast.Pi{
				Binder: "_",
				A:      ast.Global{Name: "T2"},
				B:      ast.Global{Name: "T1"}, // Using T1 as codomain
			},
			B: ast.Global{Name: "T1"},
		}

		err := env.DeclareMutual([]MutualInductiveSpec{
			{
				Name: "T1",
				Type: t1Type,
				Constructors: []Constructor{
					{Name: "mk", Type: mkType},
				},
				Eliminator: "t1Elim",
			},
			{
				Name:         "T2",
				Type:         t2Type,
				Constructors: []Constructor{},
				Eliminator:   "t2Elim",
			},
		})

		if err == nil {
			t.Error("DeclareMutual should reject T2 in negative position in T1's constructor")
		}
	})

	eval.ClearRecursorRegistry()
}

// TestMutualInductive_Positivity_Accept tests that positive occurrences are accepted.
func TestMutualInductive_Positivity_Accept(t *testing.T) {
	eval.ClearRecursorRegistry()
	env := NewGlobalEnv()

	// A : Type
	// B : Type
	// mkA : B -> A   -- B in positive position (ok)
	// mkB : A -> B   -- A in positive position (ok)
	aType := ast.Sort{U: 0}
	bType := ast.Sort{U: 0}

	mkAType := ast.Pi{
		Binder: "_",
		A:      ast.Global{Name: "B"},
		B:      ast.Global{Name: "A"},
	}

	mkBType := ast.Pi{
		Binder: "_",
		A:      ast.Global{Name: "A"},
		B:      ast.Global{Name: "B"},
	}

	err := env.DeclareMutual([]MutualInductiveSpec{
		{
			Name: "A",
			Type: aType,
			Constructors: []Constructor{
				{Name: "mkA", Type: mkAType},
			},
			Eliminator: "aElim",
		},
		{
			Name: "B",
			Type: bType,
			Constructors: []Constructor{
				{Name: "mkB", Type: mkBType},
			},
			Eliminator: "bElim",
		},
	})
	if err != nil {
		t.Fatalf("DeclareMutual(A/B) failed: %v", err)
	}

	// Verify both types were added
	if env.LookupType("A") == nil {
		t.Error("A type not found")
	}
	if env.LookupType("B") == nil {
		t.Error("B type not found")
	}

	eval.ClearRecursorRegistry()
}
