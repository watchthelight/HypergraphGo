package check

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// Transparency controls whether a definition can be unfolded during conversion.
type Transparency int

const (
	Opaque      Transparency = iota // Never unfold
	Transparent                     // Always unfold during conversion
)

// Axiom represents a postulated constant with only a type.
type Axiom struct {
	Name string
	Type ast.Term
}

// Definition represents a defined constant with type and body.
type Definition struct {
	Name         string
	Type         ast.Term
	Body         ast.Term
	Transparency Transparency
}

// Constructor represents an inductive type constructor.
type Constructor struct {
	Name string
	Type ast.Term
}

// Inductive represents an inductive type definition.
type Inductive struct {
	Name         string
	Type         ast.Term
	Constructors []Constructor
	Eliminator   string // Name of the elimination principle
}

// Primitive represents a built-in type with special evaluation rules.
type Primitive struct {
	Name string
	Type ast.Term
}

// GlobalEnv holds all global declarations in dependency order.
type GlobalEnv struct {
	axioms     map[string]*Axiom
	defs       map[string]*Definition
	inductives map[string]*Inductive
	primitives map[string]*Primitive
	order      []string // Declaration order for dependency tracking
}

// NewGlobalEnv creates an empty global environment.
func NewGlobalEnv() *GlobalEnv {
	return &GlobalEnv{
		axioms:     make(map[string]*Axiom),
		defs:       make(map[string]*Definition),
		inductives: make(map[string]*Inductive),
		primitives: make(map[string]*Primitive),
		order:      nil,
	}
}

// NewGlobalEnvWithPrimitives creates a global environment with Nat and Bool.
func NewGlobalEnvWithPrimitives() *GlobalEnv {
	env := NewGlobalEnv()
	env.addPrimitives()
	return env
}

// addPrimitives adds built-in Nat and Bool types.
func (g *GlobalEnv) addPrimitives() {
	type0 := ast.Sort{U: 0}

	// Nat : Type₀
	natType := &Primitive{Name: "Nat", Type: type0}
	g.primitives["Nat"] = natType
	g.order = append(g.order, "Nat")

	// zero : Nat
	zeroType := &Primitive{Name: "zero", Type: ast.Global{Name: "Nat"}}
	g.primitives["zero"] = zeroType
	g.order = append(g.order, "zero")

	// succ : Nat → Nat
	succType := &Primitive{
		Name: "succ",
		Type: ast.Pi{
			Binder: "_",
			A:      ast.Global{Name: "Nat"},
			B:      ast.Global{Name: "Nat"},
		},
	}
	g.primitives["succ"] = succType
	g.order = append(g.order, "succ")

	// natElim : (P : Nat → Type) → P zero → ((n : Nat) → P n → P (succ n)) → (n : Nat) → P n
	// This is the dependent eliminator for Nat
	natElimType := &Primitive{
		Name: "natElim",
		Type: mkNatElimType(),
	}
	g.primitives["natElim"] = natElimType
	g.order = append(g.order, "natElim")

	// Bool : Type₀
	boolType := &Primitive{Name: "Bool", Type: type0}
	g.primitives["Bool"] = boolType
	g.order = append(g.order, "Bool")

	// true : Bool
	trueType := &Primitive{Name: "true", Type: ast.Global{Name: "Bool"}}
	g.primitives["true"] = trueType
	g.order = append(g.order, "true")

	// false : Bool
	falseType := &Primitive{Name: "false", Type: ast.Global{Name: "Bool"}}
	g.primitives["false"] = falseType
	g.order = append(g.order, "false")

	// boolElim : (P : Bool → Type) → P true → P false → (b : Bool) → P b
	boolElimType := &Primitive{
		Name: "boolElim",
		Type: mkBoolElimType(),
	}
	g.primitives["boolElim"] = boolElimType
	g.order = append(g.order, "boolElim")
}

// mkNatElimType constructs the type of natElim:
// (P : Nat → Type) → P zero → ((n : Nat) → P n → P (succ n)) → (n : Nat) → P n
func mkNatElimType() ast.Term {
	nat := ast.Global{Name: "Nat"}
	type0 := ast.Sort{U: 0}

	// P : Nat → Type
	pType := ast.Pi{Binder: "_", A: nat, B: type0}

	// P zero
	pZero := ast.App{T: ast.Var{Ix: 0}, U: ast.Global{Name: "zero"}}

	// (n : Nat) → P n → P (succ n)
	// Under P binder, P is Var{0}
	// P n = App{Var{1}, Var{0}}  (P is shifted by 1 under n binder)
	// P (succ n) = App{Var{1}, App{Global{succ}, Var{0}}}
	pn := ast.App{T: ast.Var{Ix: 1}, U: ast.Var{Ix: 0}}
	pSuccN := ast.App{T: ast.Var{Ix: 2}, U: ast.App{T: ast.Global{Name: "succ"}, U: ast.Var{Ix: 1}}}
	succCase := ast.Pi{
		Binder: "n",
		A:      nat,
		B: ast.Pi{
			Binder: "_",
			A:      pn,
			B:      pSuccN,
		},
	}

	// (n : Nat) → P n
	// Under P, pZero, succCase binders, P is Var{3}
	pnResult := ast.App{T: ast.Var{Ix: 3}, U: ast.Var{Ix: 0}}
	target := ast.Pi{Binder: "n", A: nat, B: pnResult}

	return ast.Pi{
		Binder: "P",
		A:      pType,
		B: ast.Pi{
			Binder: "_",
			A:      pZero,
			B: ast.Pi{
				Binder: "_",
				A:      succCase,
				B:      target,
			},
		},
	}
}

// mkBoolElimType constructs the type of boolElim:
// (P : Bool → Type) → P true → P false → (b : Bool) → P b
func mkBoolElimType() ast.Term {
	bool_ := ast.Global{Name: "Bool"}
	type0 := ast.Sort{U: 0}

	// P : Bool → Type
	pType := ast.Pi{Binder: "_", A: bool_, B: type0}

	// P true
	pTrue := ast.App{T: ast.Var{Ix: 0}, U: ast.Global{Name: "true"}}

	// P false (under P, pTrue binders)
	pFalse := ast.App{T: ast.Var{Ix: 1}, U: ast.Global{Name: "false"}}

	// (b : Bool) → P b (under P, pTrue, pFalse binders)
	pb := ast.App{T: ast.Var{Ix: 3}, U: ast.Var{Ix: 0}}
	target := ast.Pi{Binder: "b", A: bool_, B: pb}

	return ast.Pi{
		Binder: "P",
		A:      pType,
		B: ast.Pi{
			Binder: "_",
			A:      pTrue,
			B: ast.Pi{
				Binder: "_",
				A:      pFalse,
				B:      target,
			},
		},
	}
}

// AddAxiom adds an axiom to the environment.
func (g *GlobalEnv) AddAxiom(name string, ty ast.Term) {
	g.axioms[name] = &Axiom{Name: name, Type: ty}
	g.order = append(g.order, name)
}

// AddDefinition adds a definition to the environment.
func (g *GlobalEnv) AddDefinition(name string, ty, body ast.Term, trans Transparency) {
	g.defs[name] = &Definition{Name: name, Type: ty, Body: body, Transparency: trans}
	g.order = append(g.order, name)
}

// AddInductive adds an inductive type to the environment without validation.
// For validated addition, use DeclareInductive.
func (g *GlobalEnv) AddInductive(name string, ty ast.Term, constrs []Constructor, elim string) {
	g.inductives[name] = &Inductive{Name: name, Type: ty, Constructors: constrs, Eliminator: elim}
	g.order = append(g.order, name)
}

// DeclareInductive validates and adds an inductive type to the environment.
// It checks:
// - The inductive type is well-formed (a Sort)
// - Each constructor type is well-formed
// - Each constructor returns the inductive type
// - The definition satisfies strict positivity
// It also generates and registers the eliminator.
func (g *GlobalEnv) DeclareInductive(name string, ty ast.Term, constrs []Constructor, elim string) error {
	// 1. Validate the inductive type is a Sort
	if err := validateIsSort(ty); err != nil {
		return &InductiveError{
			IndName: name,
			Message: "inductive type must be a Sort: " + err.Error(),
		}
	}

	// 2. Check strict positivity
	if err := CheckPositivity(name, constrs); err != nil {
		return err
	}

	// 3. Validate each constructor returns the inductive type
	for _, c := range constrs {
		if err := validateConstructorResult(name, c); err != nil {
			return err
		}
	}

	// 4. Add the inductive to the environment
	g.AddInductive(name, ty, constrs, elim)

	// 5. Generate and register the eliminator
	ind := g.inductives[name]
	elimType := GenerateRecursorType(ind)
	g.AddAxiom(elim, elimType)

	return nil
}

// InductiveError represents an error in inductive type validation.
type InductiveError struct {
	IndName string
	Message string
}

func (e *InductiveError) Error() string {
	return "inductive " + e.IndName + ": " + e.Message
}

// validateIsSort checks that ty is a Sort.
func validateIsSort(ty ast.Term) error {
	switch ty.(type) {
	case ast.Sort:
		return nil
	default:
		return &ValidationError{Msg: "expected Sort, got " + ast.Sprint(ty)}
	}
}

// ValidationError represents a validation error during inductive declaration.
type ValidationError struct {
	Msg string
}

func (e *ValidationError) Error() string {
	return e.Msg
}

// validateConstructorResult checks that a constructor's result type is the inductive.
func validateConstructorResult(indName string, c Constructor) error {
	resultTy := constructorResultType(c.Type)
	if resultTy == nil {
		return &ConstructorError{
			IndName:     indName,
			Constructor: c.Name,
			Message:     "could not determine result type",
		}
	}

	// Result should be the inductive itself or an application of it
	switch r := resultTy.(type) {
	case ast.Global:
		if r.Name != indName {
			return &ConstructorError{
				IndName:     indName,
				Constructor: c.Name,
				Message:     "result type must be " + indName + ", got " + r.Name,
			}
		}
	case ast.App:
		// Check if it's an application with the inductive as the function
		if !isAppOfGlobal(r, indName) {
			return &ConstructorError{
				IndName:     indName,
				Constructor: c.Name,
				Message:     "result type must be " + indName + " or its application",
			}
		}
	default:
		return &ConstructorError{
			IndName:     indName,
			Constructor: c.Name,
			Message:     "result type must be " + indName,
		}
	}
	return nil
}

// constructorResultType extracts the result type from a constructor type.
// For (x : A) -> B, it recursively finds the final codomain.
func constructorResultType(ty ast.Term) ast.Term {
	switch t := ty.(type) {
	case ast.Pi:
		return constructorResultType(t.B)
	default:
		return ty
	}
}

// isAppOfGlobal checks if a term is an application chain with the given global at the head.
func isAppOfGlobal(t ast.Term, name string) bool {
	switch app := t.(type) {
	case ast.App:
		return isAppOfGlobal(app.T, name)
	case ast.Global:
		return app.Name == name
	default:
		return false
	}
}

// ConstructorError represents an error in constructor validation.
type ConstructorError struct {
	IndName     string
	Constructor string
	Message     string
}

func (e *ConstructorError) Error() string {
	return "constructor " + e.Constructor + " of " + e.IndName + ": " + e.Message
}

// LookupType returns the type of a global name, or nil if not found.
func (g *GlobalEnv) LookupType(name string) ast.Term {
	if ax, ok := g.axioms[name]; ok {
		return ax.Type
	}
	if def, ok := g.defs[name]; ok {
		return def.Type
	}
	if ind, ok := g.inductives[name]; ok {
		return ind.Type
	}
	if prim, ok := g.primitives[name]; ok {
		return prim.Type
	}
	// Check constructors
	for _, ind := range g.inductives {
		for _, c := range ind.Constructors {
			if c.Name == name {
				return c.Type
			}
		}
	}
	return nil
}

// LookupDefinitionBody returns the body of a definition if transparent.
func (g *GlobalEnv) LookupDefinitionBody(name string) (ast.Term, bool) {
	if def, ok := g.defs[name]; ok && def.Transparency == Transparent {
		return def.Body, true
	}
	return nil, false
}

// Has returns true if the name is defined in the environment.
func (g *GlobalEnv) Has(name string) bool {
	return g.LookupType(name) != nil
}

// Order returns the declaration order.
func (g *GlobalEnv) Order() []string {
	return g.order
}
