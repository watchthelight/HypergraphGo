// stdlib.go provides standard library types for HoTTGo.
//
// These types are defined via AddInductive, not as primitives, so they
// use the standard inductive type machinery for type checking and computation.
//
// Standard library types:
//   - Unit: single-element type with tt constructor
//   - Empty: uninhabited type with no constructors
//   - Sum: disjoint union (coproduct) with inl/inr constructors
//   - List: polymorphic lists with nil/cons constructors

package check

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// AddStdlib adds standard library types to a GlobalEnv.
// Call this after NewGlobalEnvWithPrimitives() to include stdlib.
func AddStdlib(env *GlobalEnv) {
	addUnit(env)
	addEmpty(env)
	addSum(env)
	addList(env)
}

// NewGlobalEnvWithStdlib creates a global environment with primitives and stdlib.
func NewGlobalEnvWithStdlib() *GlobalEnv {
	env := NewGlobalEnvWithPrimitives()
	AddStdlib(env)
	return env
}

// NewCheckerWithStdlib creates a checker with primitives and stdlib.
func NewCheckerWithStdlib() *Checker {
	return NewChecker(NewGlobalEnvWithStdlib())
}

// addUnit adds the Unit type:
//
//	Unit : Type₀
//	tt : Unit
//	unitElim : (P : Unit → Type) → P tt → (u : Unit) → P u
//
// Computation rule: unitElim P p tt → p
func addUnit(env *GlobalEnv) {
	type0 := ast.Sort{U: 0}
	unitGlobal := ast.Global{Name: "Unit"}

	// Unit : Type₀
	// Use DeclareInductive to properly validate and generate the eliminator
	err := env.DeclareInductive(
		"Unit",          // name
		type0,           // type (Type₀)
		[]Constructor{
			{
				Name: "tt",
				Type: unitGlobal, // tt : Unit
			},
		},
		"unitElim",      // eliminator name
	)
	if err != nil {
		panic("failed to declare Unit type: " + err.Error())
	}
}

// addEmpty adds the Empty type (also known as Void or False):
//
//	Empty : Type₀
//	(no constructors)
//	emptyElim : (P : Empty → Type) → (e : Empty) → P e
//
// The eliminator is always stuck since there are no constructors to match.
func addEmpty(env *GlobalEnv) {
	type0 := ast.Sort{U: 0}

	// Empty : Type₀ with no constructors
	// Use DeclareInductive to properly validate and generate the eliminator
	err := env.DeclareInductive(
		"Empty",         // name
		type0,           // type (Type₀)
		[]Constructor{}, // no constructors
		"emptyElim",     // eliminator name
	)
	if err != nil {
		panic("failed to declare Empty type: " + err.Error())
	}
}

// addSum adds the Sum (coproduct/disjoint union) type:
//
//	Sum : Type → Type → Type
//	inl : (A : Type) → (B : Type) → A → Sum A B
//	inr : (A : Type) → (B : Type) → B → Sum A B
//	sumElim : (A : Type) → (B : Type) → (P : Sum A B → Type)
//	          → ((a : A) → P (inl A B a))
//	          → ((b : B) → P (inr A B b))
//	          → (s : Sum A B) → P s
//
// Computation rules:
//   - sumElim A B P f g (inl A B a) → f a
//   - sumElim A B P f g (inr A B b) → g b
func addSum(env *GlobalEnv) {
	type0 := ast.Sort{U: 0}
	sumGlobal := ast.Global{Name: "Sum"}

	// Sum : Type → Type → Type
	sumType := ast.Pi{
		Binder: "A",
		A:      type0,
		B: ast.Pi{
			Binder: "B",
			A:      type0,
			B:      type0,
		},
	}

	// inl : (A : Type) → (B : Type) → A → Sum A B
	// Under [A, B, _], A is Var{2}, B is Var{1}
	// Under [A, B], A is Var{1}
	inlType := ast.Pi{
		Binder: "A",
		A:      type0,
		B: ast.Pi{
			Binder: "B",
			A:      type0,
			B: ast.Pi{
				Binder: "_",
				A:      ast.Var{Ix: 1}, // A under [A, B]
				B: ast.App{
					T: ast.App{T: sumGlobal, U: ast.Var{Ix: 2}}, // Sum A
					U: ast.Var{Ix: 1},                           // Sum A B
				},
			},
		},
	}

	// inr : (A : Type) → (B : Type) → B → Sum A B
	// Under [A, B, _], A is Var{2}, B is Var{1}
	// Under [A, B], B is Var{0}
	inrType := ast.Pi{
		Binder: "A",
		A:      type0,
		B: ast.Pi{
			Binder: "B",
			A:      type0,
			B: ast.Pi{
				Binder: "_",
				A:      ast.Var{Ix: 0}, // B under [A, B]
				B: ast.App{
					T: ast.App{T: sumGlobal, U: ast.Var{Ix: 2}}, // Sum A
					U: ast.Var{Ix: 1},                           // Sum A B
				},
			},
		},
	}

	err := env.DeclareInductive(
		"Sum",
		sumType,
		[]Constructor{
			{Name: "inl", Type: inlType},
			{Name: "inr", Type: inrType},
		},
		"sumElim",
	)
	if err != nil {
		panic("failed to declare Sum type: " + err.Error())
	}
}

// addList adds the List (polymorphic list) type:
//
//	List : Type → Type
//	nil : (A : Type) → List A
//	cons : (A : Type) → A → List A → List A
//	listElim : (A : Type) → (P : List A → Type)
//	           → P nil
//	           → ((x : A) → (xs : List A) → P xs → P (cons A x xs))
//	           → (l : List A) → P l
//
// Computation rules:
//   - listElim A P pn pc nil → pn
//   - listElim A P pn pc (cons A x xs) → pc x xs (listElim A P pn pc xs)
func addList(env *GlobalEnv) {
	type0 := ast.Sort{U: 0}
	listGlobal := ast.Global{Name: "List"}

	// List : Type → Type
	listType := ast.Pi{
		Binder: "A",
		A:      type0,
		B:      type0,
	}

	// nil : (A : Type) → List A
	// Under [A], List A = App{List, Var{0}}
	nilType := ast.Pi{
		Binder: "A",
		A:      type0,
		B:      ast.App{T: listGlobal, U: ast.Var{Ix: 0}},
	}

	// cons : (A : Type) → A → List A → List A
	// Under [A], A is Var{0}
	// Under [A, x], A is Var{1}
	// Under [A, x, xs], A is Var{2}
	consType := ast.Pi{
		Binder: "A",
		A:      type0,
		B: ast.Pi{
			Binder: "_",
			A:      ast.Var{Ix: 0}, // A under [A]
			B: ast.Pi{
				Binder: "_",
				A:      ast.App{T: listGlobal, U: ast.Var{Ix: 1}}, // List A under [A, x]
				B:      ast.App{T: listGlobal, U: ast.Var{Ix: 2}}, // List A under [A, x, xs]
			},
		},
	}

	err := env.DeclareInductive(
		"List",
		listType,
		[]Constructor{
			{Name: "nil", Type: nilType},
			{Name: "cons", Type: consType},
		},
		"listElim",
	)
	if err != nil {
		panic("failed to declare List type: " + err.Error())
	}
}
