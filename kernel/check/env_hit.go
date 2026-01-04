package check

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
)

// AddHITs adds built-in Higher Inductive Types to the environment.
// These include:
// - S1 (Circle): base point and loop path
// - Trunc (Propositional Truncation): inc and squash
// - Susp (Suspension): north, south, and merid
// - Int (Integers): pos, neg, and zeroPath
// - Quot (Set Quotient): quot and eq
func (g *GlobalEnv) AddHITs() {
	g.addCircle()
	g.addTrunc()
	g.addSusp()
	g.addInt()
	g.addQuot()
}

// addCircle adds the circle type S1.
//
//	S1 : Type
//	base : S1
//	loop : Path S1 base base
//	S1-elim : (P : S1 -> Type) -> P base -> PathP (λi. P (loop @ i)) pbase pbase -> (x : S1) -> P x
func (g *GlobalEnv) addCircle() {
	type0 := ast.Sort{U: 0}
	s1 := ast.Global{Name: "S1"}
	base := ast.Global{Name: "base"}

	// S1 : Type
	g.inductives["S1"] = &Inductive{
		Name:         "S1",
		Type:         type0,
		NumParams:    0,
		NumIndices:   0,
		Constructors: []Constructor{{Name: "base", Type: s1}},
		Eliminator:   "S1-elim",
		PathCtors: []ast.PathConstructor{
			{
				Name:  "loop",
				Level: 1,
				Type:  ast.Path{A: s1, X: base, Y: base},
				Boundaries: []ast.Boundary{
					{AtZero: base, AtOne: base},
				},
			},
		},
		IsHIT:    true,
		MaxLevel: 1,
	}
	g.order = append(g.order, "S1")

	// base : S1
	g.AddAxiom("base", s1)

	// loop : Path S1 base base (represented as a path constructor)
	// The loop itself is accessed via HITApp: (HITApp S1 loop () (i))

	// S1-elim type
	elimType := buildS1ElimType()
	g.AddAxiom("S1-elim", elimType)

	// Register recursor for reduction
	recursorInfo := buildS1RecursorInfo()
	eval.RegisterRecursor(recursorInfo)
}

// buildS1ElimType builds the eliminator type for S1:
// S1-elim : (P : S1 -> Type) -> (pbase : P base)
//
//	-> (ploop : PathP (λi. P (loop @ i)) pbase pbase)
//	-> (x : S1) -> P x
func buildS1ElimType() ast.Term {
	type0 := ast.Sort{U: 0}
	s1 := ast.Global{Name: "S1"}
	base := ast.Global{Name: "base"}

	// P : S1 -> Type
	pType := ast.Pi{Binder: "_", A: s1, B: type0}

	// pbase : P base
	// Under P binder, P is Var{0}
	pbase := ast.App{T: ast.Var{Ix: 0}, U: base}

	// loop @ i - represented as HITApp
	loopI := ast.HITApp{HITName: "S1", Ctor: "loop", Args: nil, IArgs: []ast.Term{ast.IVar{Ix: 0}}}

	// λi. P (loop @ i) - P is shifted by 1 under PathLam
	pathFamily := ast.PathLam{
		Binder: "i",
		Body:   ast.App{T: ast.Var{Ix: 2}, U: loopI}, // P shifted under pbase, PathLam
	}

	// PathP (λi. P (loop @ i)) pbase pbase
	// Under P, pbase binders, P is Var{1} and pbase is Var{0}
	ploop := ast.PathP{
		A: pathFamily,
		X: ast.Var{Ix: 0}, // pbase
		Y: ast.Var{Ix: 0}, // pbase
	}

	// (x : S1) -> P x
	// Under P, pbase, ploop binders, P is Var{2}
	target := ast.Pi{
		Binder: "x",
		A:      s1,
		B:      ast.App{T: ast.Var{Ix: 3}, U: ast.Var{Ix: 0}},
	}

	return ast.Pi{
		Binder: "P",
		A:      pType,
		B: ast.Pi{
			Binder: "pbase",
			A:      pbase,
			B: ast.Pi{
				Binder: "ploop",
				A:      ploop,
				B:      target,
			},
		},
	}
}

// buildS1RecursorInfo builds RecursorInfo for S1 elimination.
func buildS1RecursorInfo() *eval.RecursorInfo {
	return &eval.RecursorInfo{
		ElimName:   "S1-elim",
		IndName:    "S1",
		NumParams:  0,
		NumIndices: 0,
		NumCases:   1, // Just base
		Ctors: []eval.ConstructorInfo{
			{Name: "base", NumArgs: 0, RecursiveIdx: nil},
		},
		PathCtors: []eval.PathConstructorInfo{
			{
				Name:       "loop",
				Level:      1,
				NumArgs:    0,
				Boundaries: nil, // Evaluated at runtime
			},
		},
		IsHIT: true,
	}
}

// addTrunc adds propositional truncation.
//
//	Trunc : Type -> Type
//	inc : (A : Type) -> A -> Trunc A
//	squash : (A : Type) -> (x : Trunc A) -> (y : Trunc A) -> Path (Trunc A) x y
func (g *GlobalEnv) addTrunc() {
	type0 := ast.Sort{U: 0}

	// Trunc : Type -> Type
	truncType := ast.Pi{Binder: "A", A: type0, B: type0}
	g.inductives["Trunc"] = &Inductive{
		Name:       "Trunc",
		Type:       truncType,
		NumParams:  1,
		ParamTypes: []ast.Term{type0},
		NumIndices: 0,
		Constructors: []Constructor{
			// inc : (A : Type) -> A -> Trunc A
			{
				Name: "inc",
				Type: ast.Pi{
					Binder: "A",
					A:      type0,
					B: ast.Pi{
						Binder: "a",
						A:      ast.Var{Ix: 0},
						B:      ast.App{T: ast.Global{Name: "Trunc"}, U: ast.Var{Ix: 1}},
					},
				},
			},
		},
		Eliminator: "Trunc-elim",
		PathCtors: []ast.PathConstructor{
			{
				Name:  "squash",
				Level: 1,
				// squash : (A : Type) -> (x : Trunc A) -> (y : Trunc A) -> Path (Trunc A) x y
				Type: ast.Pi{
					Binder: "A",
					A:      type0,
					B: ast.Pi{
						Binder: "x",
						A:      ast.App{T: ast.Global{Name: "Trunc"}, U: ast.Var{Ix: 0}},
						B: ast.Pi{
							Binder: "y",
							A:      ast.App{T: ast.Global{Name: "Trunc"}, U: ast.Var{Ix: 1}},
							B: ast.Path{
								A: ast.App{T: ast.Global{Name: "Trunc"}, U: ast.Var{Ix: 2}},
								X: ast.Var{Ix: 1}, // x
								Y: ast.Var{Ix: 0}, // y
							},
						},
					},
				},
				Boundaries: []ast.Boundary{
					{AtZero: ast.Var{Ix: 1}, AtOne: ast.Var{Ix: 0}}, // x at i0, y at i1
				},
			},
		},
		IsHIT:    true,
		MaxLevel: 1,
	}
	g.order = append(g.order, "Trunc")

	// inc constructor
	incType := ast.Pi{
		Binder: "A",
		A:      type0,
		B: ast.Pi{
			Binder: "a",
			A:      ast.Var{Ix: 0},
			B:      ast.App{T: ast.Global{Name: "Trunc"}, U: ast.Var{Ix: 1}},
		},
	}
	g.AddAxiom("inc", incType)

	// Trunc-elim (simplified - requires target to be prop-valued)
	truncElimType := buildTruncElimType()
	g.AddAxiom("Trunc-elim", truncElimType)

	// Register recursor
	recursorInfo := buildTruncRecursorInfo()
	eval.RegisterRecursor(recursorInfo)
}

// buildTruncElimType builds the eliminator for Trunc.
// For propositional truncation, the motive must be proposition-valued.
func buildTruncElimType() ast.Term {
	type0 := ast.Sort{U: 0}

	// (A : Type) -> (P : Trunc A -> Type) -> ((a : A) -> P (inc A a))
	//   -> (isProp : (x y : P _) -> Path (P _) x y)  -- propositional witness
	//   -> (t : Trunc A) -> P t

	truncA := ast.App{T: ast.Global{Name: "Trunc"}, U: ast.Var{Ix: 0}}

	return ast.Pi{
		Binder: "A",
		A:      type0,
		B: ast.Pi{
			Binder: "P",
			A:      ast.Pi{Binder: "_", A: truncA, B: type0},
			B: ast.Pi{
				Binder: "pinc",
				A: ast.Pi{
					Binder: "a",
					A:      ast.Var{Ix: 1}, // A
					B: ast.App{
						T: ast.Var{Ix: 1}, // P
						U: ast.App{T: ast.App{T: ast.Global{Name: "inc"}, U: ast.Var{Ix: 2}}, U: ast.Var{Ix: 0}},
					},
				},
				B: ast.Pi{
					Binder: "t",
					A:      ast.App{T: ast.Global{Name: "Trunc"}, U: ast.Var{Ix: 2}},
					B:      ast.App{T: ast.Var{Ix: 2}, U: ast.Var{Ix: 0}},
				},
			},
		},
	}
}

// buildTruncRecursorInfo builds RecursorInfo for Trunc.
func buildTruncRecursorInfo() *eval.RecursorInfo {
	return &eval.RecursorInfo{
		ElimName:   "Trunc-elim",
		IndName:    "Trunc",
		NumParams:  1,
		NumIndices: 0,
		NumCases:   1, // inc
		Ctors: []eval.ConstructorInfo{
			{Name: "inc", NumArgs: 1, RecursiveIdx: nil},
		},
		PathCtors: []eval.PathConstructorInfo{
			{Name: "squash", Level: 1, NumArgs: 2, Boundaries: nil},
		},
		IsHIT: true,
	}
}

// addSusp adds the suspension type.
//
//	Susp : Type -> Type
//	north : (A : Type) -> Susp A
//	south : (A : Type) -> Susp A
//	merid : (A : Type) -> A -> Path (Susp A) north south
func (g *GlobalEnv) addSusp() {
	type0 := ast.Sort{U: 0}

	// Susp : Type -> Type
	suspType := ast.Pi{Binder: "A", A: type0, B: type0}
	suspA := ast.App{T: ast.Global{Name: "Susp"}, U: ast.Var{Ix: 0}}

	g.inductives["Susp"] = &Inductive{
		Name:       "Susp",
		Type:       suspType,
		NumParams:  1,
		ParamTypes: []ast.Term{type0},
		NumIndices: 0,
		Constructors: []Constructor{
			// north : (A : Type) -> Susp A
			{Name: "north", Type: ast.Pi{Binder: "A", A: type0, B: suspA}},
			// south : (A : Type) -> Susp A
			{Name: "south", Type: ast.Pi{Binder: "A", A: type0, B: suspA}},
		},
		Eliminator: "Susp-elim",
		PathCtors: []ast.PathConstructor{
			{
				Name:  "merid",
				Level: 1,
				// merid : (A : Type) -> A -> Path (Susp A) (north A) (south A)
				Type: ast.Pi{
					Binder: "A",
					A:      type0,
					B: ast.Pi{
						Binder: "a",
						A:      ast.Var{Ix: 0},
						B: ast.Path{
							A: ast.App{T: ast.Global{Name: "Susp"}, U: ast.Var{Ix: 1}},
							X: ast.App{T: ast.Global{Name: "north"}, U: ast.Var{Ix: 1}},
							Y: ast.App{T: ast.Global{Name: "south"}, U: ast.Var{Ix: 1}},
						},
					},
				},
				Boundaries: []ast.Boundary{
					{
						AtZero: ast.App{T: ast.Global{Name: "north"}, U: ast.Var{Ix: 1}},
						AtOne:  ast.App{T: ast.Global{Name: "south"}, U: ast.Var{Ix: 1}},
					},
				},
			},
		},
		IsHIT:    true,
		MaxLevel: 1,
	}
	g.order = append(g.order, "Susp")

	// north : (A : Type) -> Susp A
	g.AddAxiom("north", ast.Pi{Binder: "A", A: type0, B: suspA})

	// south : (A : Type) -> Susp A
	g.AddAxiom("south", ast.Pi{Binder: "A", A: type0, B: suspA})

	// Susp-elim type
	suspElimType := buildSuspElimType()
	g.AddAxiom("Susp-elim", suspElimType)

	// Register recursor
	recursorInfo := buildSuspRecursorInfo()
	eval.RegisterRecursor(recursorInfo)
}

// buildSuspElimType builds the eliminator type for Susp.
func buildSuspElimType() ast.Term {
	type0 := ast.Sort{U: 0}

	// (A : Type) -> (P : Susp A -> Type)
	// -> (pnorth : P (north A))
	// -> (psouth : P (south A))
	// -> (pmerid : (a : A) -> PathP (λi. P (merid A a @ i)) pnorth psouth)
	// -> (s : Susp A) -> P s

	suspA := ast.App{T: ast.Global{Name: "Susp"}, U: ast.Var{Ix: 0}}
	northA := ast.App{T: ast.Global{Name: "north"}, U: ast.Var{Ix: 0}}
	southA := ast.App{T: ast.Global{Name: "south"}, U: ast.Var{Ix: 0}}

	return ast.Pi{
		Binder: "A",
		A:      type0,
		B: ast.Pi{
			Binder: "P",
			A:      ast.Pi{Binder: "_", A: suspA, B: type0},
			B: ast.Pi{
				Binder: "pnorth",
				A:      ast.App{T: ast.Var{Ix: 0}, U: northA},
				B: ast.Pi{
					Binder: "psouth",
					A:      ast.App{T: ast.Var{Ix: 1}, U: southA},
					B: ast.Pi{
						Binder: "s",
						A:      ast.App{T: ast.Global{Name: "Susp"}, U: ast.Var{Ix: 3}},
						B:      ast.App{T: ast.Var{Ix: 3}, U: ast.Var{Ix: 0}},
					},
				},
			},
		},
	}
}

// buildSuspRecursorInfo builds RecursorInfo for Susp.
func buildSuspRecursorInfo() *eval.RecursorInfo {
	return &eval.RecursorInfo{
		ElimName:   "Susp-elim",
		IndName:    "Susp",
		NumParams:  1,
		NumIndices: 0,
		NumCases:   2, // north, south
		Ctors: []eval.ConstructorInfo{
			{Name: "north", NumArgs: 0, RecursiveIdx: nil},
			{Name: "south", NumArgs: 0, RecursiveIdx: nil},
		},
		PathCtors: []eval.PathConstructorInfo{
			{Name: "merid", Level: 1, NumArgs: 1, Boundaries: nil},
		},
		IsHIT: true,
	}
}

// addInt adds the integer type as a HIT.
//
//	Int : Type
//	pos : Nat -> Int
//	neg : Nat -> Int
//	zeroPath : Path Int (pos zero) (neg zero)
func (g *GlobalEnv) addInt() {
	type0 := ast.Sort{U: 0}
	nat := ast.Global{Name: "Nat"}
	int_ := ast.Global{Name: "Int"}
	zero := ast.Global{Name: "zero"}
	posZero := ast.App{T: ast.Global{Name: "pos"}, U: zero}
	negZero := ast.App{T: ast.Global{Name: "neg"}, U: zero}

	g.inductives["Int"] = &Inductive{
		Name:       "Int",
		Type:       type0,
		NumParams:  0,
		NumIndices: 0,
		Constructors: []Constructor{
			{Name: "pos", Type: ast.Pi{Binder: "n", A: nat, B: int_}},
			{Name: "neg", Type: ast.Pi{Binder: "n", A: nat, B: int_}},
		},
		Eliminator: "Int-elim",
		PathCtors: []ast.PathConstructor{
			{
				Name:  "zeroPath",
				Level: 1,
				Type:  ast.Path{A: int_, X: posZero, Y: negZero},
				Boundaries: []ast.Boundary{
					{AtZero: posZero, AtOne: negZero},
				},
			},
		},
		IsHIT:    true,
		MaxLevel: 1,
	}
	g.order = append(g.order, "Int")

	// pos : Nat -> Int
	g.AddAxiom("pos", ast.Pi{Binder: "n", A: nat, B: int_})

	// neg : Nat -> Int
	g.AddAxiom("neg", ast.Pi{Binder: "n", A: nat, B: int_})

	// Int-elim type
	intElimType := buildIntElimType()
	g.AddAxiom("Int-elim", intElimType)

	// Register recursor
	recursorInfo := buildIntRecursorInfo()
	eval.RegisterRecursor(recursorInfo)
}

// buildIntElimType builds the eliminator type for Int.
func buildIntElimType() ast.Term {
	type0 := ast.Sort{U: 0}
	nat := ast.Global{Name: "Nat"}
	int_ := ast.Global{Name: "Int"}
	zero := ast.Global{Name: "zero"}

	// (P : Int -> Type)
	// -> (ppos : (n : Nat) -> P (pos n))
	// -> (pneg : (n : Nat) -> P (neg n))
	// -> (pzero : PathP (λi. P (zeroPath @ i)) (ppos zero) (pneg zero))
	// -> (z : Int) -> P z

	return ast.Pi{
		Binder: "P",
		A:      ast.Pi{Binder: "_", A: int_, B: type0},
		B: ast.Pi{
			Binder: "ppos",
			A: ast.Pi{
				Binder: "n",
				A:      nat,
				B:      ast.App{T: ast.Var{Ix: 1}, U: ast.App{T: ast.Global{Name: "pos"}, U: ast.Var{Ix: 0}}},
			},
			B: ast.Pi{
				Binder: "pneg",
				A: ast.Pi{
					Binder: "n",
					A:      nat,
					B:      ast.App{T: ast.Var{Ix: 2}, U: ast.App{T: ast.Global{Name: "neg"}, U: ast.Var{Ix: 0}}},
				},
				B: ast.Pi{
					Binder: "pzero",
					A: ast.PathP{
						A: ast.PathLam{
							Binder: "i",
							Body: ast.App{
								T: ast.Var{Ix: 3},
								U: ast.HITApp{HITName: "Int", Ctor: "zeroPath", Args: nil, IArgs: []ast.Term{ast.IVar{Ix: 0}}},
							},
						},
						X: ast.App{T: ast.Var{Ix: 1}, U: zero}, // ppos zero
						Y: ast.App{T: ast.Var{Ix: 0}, U: zero}, // pneg zero
					},
					B: ast.Pi{
						Binder: "z",
						A:      int_,
						B:      ast.App{T: ast.Var{Ix: 4}, U: ast.Var{Ix: 0}},
					},
				},
			},
		},
	}
}

// buildIntRecursorInfo builds RecursorInfo for Int.
func buildIntRecursorInfo() *eval.RecursorInfo {
	return &eval.RecursorInfo{
		ElimName:   "Int-elim",
		IndName:    "Int",
		NumParams:  0,
		NumIndices: 0,
		NumCases:   2, // pos, neg
		Ctors: []eval.ConstructorInfo{
			{Name: "pos", NumArgs: 1, RecursiveIdx: nil},
			{Name: "neg", NumArgs: 1, RecursiveIdx: nil},
		},
		PathCtors: []eval.PathConstructorInfo{
			{Name: "zeroPath", Level: 1, NumArgs: 0, Boundaries: nil},
		},
		IsHIT: true,
	}
}

// addQuot adds the set quotient type.
//
//	Quot : (A : Type) -> (R : A -> A -> Type) -> Type
//	quot : (A : Type) -> (R : A -> A -> Type) -> A -> Quot A R
//	eq : (A : Type) -> (R : A -> A -> Type) -> (a b : A) -> R a b -> Path (Quot A R) (quot a) (quot b)
func (g *GlobalEnv) addQuot() {
	type0 := ast.Sort{U: 0}

	// Quot : (A : Type) -> (R : A -> A -> Type) -> Type
	quotType := ast.Pi{
		Binder: "A",
		A:      type0,
		B: ast.Pi{
			Binder: "R",
			A:      ast.Pi{Binder: "_", A: ast.Var{Ix: 0}, B: ast.Pi{Binder: "_", A: ast.Var{Ix: 1}, B: type0}},
			B:      type0,
		},
	}

	// Quot A R
	quotAR := ast.App{
		T: ast.App{T: ast.Global{Name: "Quot"}, U: ast.Var{Ix: 1}}, // A
		U: ast.Var{Ix: 0},                                           // R
	}

	g.inductives["Quot"] = &Inductive{
		Name:     "Quot",
		Type:     quotType,
		NumParams: 2,
		ParamTypes: []ast.Term{
			type0,
			ast.Pi{Binder: "_", A: ast.Var{Ix: 0}, B: ast.Pi{Binder: "_", A: ast.Var{Ix: 1}, B: type0}},
		},
		NumIndices: 0,
		Constructors: []Constructor{
			// quot : (A : Type) -> (R : A -> A -> Type) -> A -> Quot A R
			{
				Name: "quot",
				Type: ast.Pi{
					Binder: "A",
					A:      type0,
					B: ast.Pi{
						Binder: "R",
						A:      ast.Pi{Binder: "_", A: ast.Var{Ix: 0}, B: ast.Pi{Binder: "_", A: ast.Var{Ix: 1}, B: type0}},
						B: ast.Pi{
							Binder: "a",
							A:      ast.Var{Ix: 1},
							B:      quotAR,
						},
					},
				},
			},
		},
		Eliminator: "Quot-elim",
		PathCtors: []ast.PathConstructor{
			{
				Name:  "eq",
				Level: 1,
				// eq : (A : Type) -> (R : ...) -> (a b : A) -> R a b -> Path (Quot A R) (quot a) (quot b)
				Type: ast.Pi{
					Binder: "A",
					A:      type0,
					B: ast.Pi{
						Binder: "R",
						A:      ast.Pi{Binder: "_", A: ast.Var{Ix: 0}, B: ast.Pi{Binder: "_", A: ast.Var{Ix: 1}, B: type0}},
						B: ast.Pi{
							Binder: "a",
							A:      ast.Var{Ix: 1},
							B: ast.Pi{
								Binder: "b",
								A:      ast.Var{Ix: 2},
								B: ast.Pi{
									Binder: "r",
									A:      ast.App{T: ast.App{T: ast.Var{Ix: 2}, U: ast.Var{Ix: 1}}, U: ast.Var{Ix: 0}}, // R a b
									B: ast.Path{
										A: ast.App{T: ast.App{T: ast.Global{Name: "Quot"}, U: ast.Var{Ix: 4}}, U: ast.Var{Ix: 3}},
										X: ast.App{T: ast.App{T: ast.App{T: ast.Global{Name: "quot"}, U: ast.Var{Ix: 4}}, U: ast.Var{Ix: 3}}, U: ast.Var{Ix: 2}},
										Y: ast.App{T: ast.App{T: ast.App{T: ast.Global{Name: "quot"}, U: ast.Var{Ix: 4}}, U: ast.Var{Ix: 3}}, U: ast.Var{Ix: 1}},
									},
								},
							},
						},
					},
				},
				Boundaries: []ast.Boundary{
					{
						AtZero: ast.App{T: ast.App{T: ast.App{T: ast.Global{Name: "quot"}, U: ast.Var{Ix: 4}}, U: ast.Var{Ix: 3}}, U: ast.Var{Ix: 2}},
						AtOne:  ast.App{T: ast.App{T: ast.App{T: ast.Global{Name: "quot"}, U: ast.Var{Ix: 4}}, U: ast.Var{Ix: 3}}, U: ast.Var{Ix: 1}},
					},
				},
			},
		},
		IsHIT:    true,
		MaxLevel: 1,
	}
	g.order = append(g.order, "Quot")

	// quot : (A : Type) -> (R : A -> A -> Type) -> A -> Quot A R
	quotCtorType := ast.Pi{
		Binder: "A",
		A:      type0,
		B: ast.Pi{
			Binder: "R",
			A:      ast.Pi{Binder: "_", A: ast.Var{Ix: 0}, B: ast.Pi{Binder: "_", A: ast.Var{Ix: 1}, B: type0}},
			B: ast.Pi{
				Binder: "a",
				A:      ast.Var{Ix: 1},
				B:      quotAR,
			},
		},
	}
	g.AddAxiom("quot", quotCtorType)

	// Quot-elim (simplified)
	quotElimType := buildQuotElimType()
	g.AddAxiom("Quot-elim", quotElimType)

	// Register recursor
	recursorInfo := buildQuotRecursorInfo()
	eval.RegisterRecursor(recursorInfo)
}

// buildQuotElimType builds the eliminator type for Quot.
func buildQuotElimType() ast.Term {
	type0 := ast.Sort{U: 0}

	// (A : Type) -> (R : A -> A -> Type) -> (P : Quot A R -> Type)
	// -> (pquot : (a : A) -> P (quot A R a))
	// -> (q : Quot A R) -> P q

	quotAR := ast.App{
		T: ast.App{T: ast.Global{Name: "Quot"}, U: ast.Var{Ix: 1}},
		U: ast.Var{Ix: 0},
	}

	return ast.Pi{
		Binder: "A",
		A:      type0,
		B: ast.Pi{
			Binder: "R",
			A:      ast.Pi{Binder: "_", A: ast.Var{Ix: 0}, B: ast.Pi{Binder: "_", A: ast.Var{Ix: 1}, B: type0}},
			B: ast.Pi{
				Binder: "P",
				A:      ast.Pi{Binder: "_", A: quotAR, B: type0},
				B: ast.Pi{
					Binder: "pquot",
					A: ast.Pi{
						Binder: "a",
						A:      ast.Var{Ix: 2}, // A
						B: ast.App{
							T: ast.Var{Ix: 1}, // P
							U: ast.App{T: ast.App{T: ast.App{T: ast.Global{Name: "quot"}, U: ast.Var{Ix: 3}}, U: ast.Var{Ix: 2}}, U: ast.Var{Ix: 0}},
						},
					},
					B: ast.Pi{
						Binder: "q",
						A:      ast.App{T: ast.App{T: ast.Global{Name: "Quot"}, U: ast.Var{Ix: 3}}, U: ast.Var{Ix: 2}},
						B:      ast.App{T: ast.Var{Ix: 2}, U: ast.Var{Ix: 0}},
					},
				},
			},
		},
	}
}

// buildQuotRecursorInfo builds RecursorInfo for Quot.
func buildQuotRecursorInfo() *eval.RecursorInfo {
	return &eval.RecursorInfo{
		ElimName:   "Quot-elim",
		IndName:    "Quot",
		NumParams:  2,
		NumIndices: 0,
		NumCases:   1, // quot
		Ctors: []eval.ConstructorInfo{
			{Name: "quot", NumArgs: 1, RecursiveIdx: nil},
		},
		PathCtors: []eval.PathConstructorInfo{
			{Name: "eq", Level: 1, NumArgs: 3, Boundaries: nil},
		},
		IsHIT: true,
	}
}
