package check

import (
	"fmt"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
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
	NumParams    int        // Number of parameters extracted from Type
	ParamTypes   []ast.Term // Types of each parameter
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
func (g *GlobalEnv) AddInductive(name string, ty ast.Term, numParams int, paramTypes []ast.Term, constrs []Constructor, elim string) {
	g.inductives[name] = &Inductive{
		Name:         name,
		Type:         ty,
		NumParams:    numParams,
		ParamTypes:   paramTypes,
		Constructors: constrs,
		Eliminator:   elim,
	}
	g.order = append(g.order, name)
}

// DeclareInductive validates and adds an inductive type to the environment.
// It checks:
// - The inductive type is well-formed (Sort or Pi chain ending in Sort)
// - Each constructor type is well-formed (uses Checker API)
// - Each constructor returns the inductive type applied to parameters
// - The definition satisfies strict positivity
// It also generates and registers the eliminator.
func (g *GlobalEnv) DeclareInductive(name string, ty ast.Term, constrs []Constructor, elim string) error {
	// 1. Validate and extract parameters from inductive type
	numParams, paramTypes, _, err := validateInductiveType(ty)
	if err != nil {
		return &InductiveError{
			IndName: name,
			Message: err.Error(),
		}
	}

	// 2. Temporarily add the inductive type so constructor types can reference it.
	// This allows constructor types like (n : Nat) -> Nat to type-check.
	g.AddAxiom(name, ty)

	// 3. Validate each constructor type is well-formed using the Checker API.
	// Create a checker with our environment.
	checker := NewChecker(g)
	for _, c := range constrs {
		if err := validateConstructorType(checker, name, c); err != nil {
			// Remove the temporary axiom on failure
			delete(g.axioms, name)
			g.removeFromOrder(name)
			return err
		}
	}

	// Remove the temporary axiom - we'll add the real inductive
	delete(g.axioms, name)
	g.removeFromOrder(name)

	// 4. Check strict positivity
	if err := CheckPositivity(name, constrs); err != nil {
		return err
	}

	// 5. Validate each constructor returns the inductive type with correct params
	for _, c := range constrs {
		if err := validateConstructorResult(name, numParams, c); err != nil {
			return err
		}
	}

	// 6. Add the inductive to the environment
	g.AddInductive(name, ty, numParams, paramTypes, constrs, elim)

	// 7. Generate and register the eliminator
	ind := g.inductives[name]
	elimType := GenerateRecursorType(ind)
	g.AddAxiom(elim, elimType)

	// 8. Register the recursor for generic reduction
	recursorInfo := buildRecursorInfo(ind)
	eval.RegisterRecursor(recursorInfo)

	return nil
}

// buildRecursorInfo builds RecursorInfo from an inductive definition.
// For parameterized inductives, NumParams is extracted from the inductive type,
// and constructor arg counts exclude parameters.
func buildRecursorInfo(ind *Inductive) *eval.RecursorInfo {
	info := &eval.RecursorInfo{
		ElimName:   ind.Eliminator,
		IndName:    ind.Name,
		NumParams:  ind.NumParams,
		NumIndices: 0, // TODO: extract from inductive type when indexed support is added
		NumCases:   len(ind.Constructors),
		Ctors:      make([]eval.ConstructorInfo, len(ind.Constructors)),
	}

	for i, c := range ind.Constructors {
		// Extract all Pi args from constructor type
		allArgs := extractPiArgs(c.Type)

		// Skip parameter args (first NumParams args are parameters)
		dataArgs := allArgs
		if ind.NumParams > 0 && len(allArgs) >= ind.NumParams {
			dataArgs = allArgs[ind.NumParams:]
		}

		// Find recursive arguments among data args
		recursiveIdx := []int{}
		for j, arg := range dataArgs {
			if isRecursiveArgType(ind.Name, arg.Type) {
				recursiveIdx = append(recursiveIdx, j)
			}
		}

		info.Ctors[i] = eval.ConstructorInfo{
			Name:         c.Name,
			NumArgs:      len(dataArgs), // Only count non-param args
			RecursiveIdx: recursiveIdx,
		}
	}

	return info
}

// validateConstructorType checks that a constructor type is well-formed.
// It validates that the type is a valid type using the Checker's CheckIsType,
// which ensures all Pi domains are well-typed.
func validateConstructorType(checker *Checker, indName string, c Constructor) error {
	// Use CheckIsType to validate the constructor type is well-formed.
	// This traverses the Pi chain and validates each domain is a type.
	_, err := checker.CheckIsType(nil, Span{}, c.Type)
	if err != nil {
		return &ConstructorError{
			IndName:     indName,
			Constructor: c.Name,
			Message:     "constructor type is not well-formed: " + err.Error(),
		}
	}
	return nil
}

// removeFromOrder removes a name from the declaration order.
func (g *GlobalEnv) removeFromOrder(name string) {
	for i, n := range g.order {
		if n == name {
			g.order = append(g.order[:i], g.order[i+1:]...)
			return
		}
	}
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

// extractParams extracts parameters from an inductive type.
// For Type -> Type -> Type, returns (2, [Type, Type], Type)
// For Type, returns (0, [], Type)
func extractParams(ty ast.Term) (numParams int, paramTypes []ast.Term, resultSort ast.Term) {
	current := ty
	for {
		if pi, ok := current.(ast.Pi); ok {
			paramTypes = append(paramTypes, pi.A)
			numParams++
			current = pi.B
		} else {
			resultSort = current
			break
		}
	}
	return
}

// validateInductiveType validates that ty is a valid inductive type.
// For non-parameterized: must be a Sort
// For parameterized: must be Pi chain ending in Sort
// Returns the number of parameters, their types, and the result sort.
func validateInductiveType(ty ast.Term) (numParams int, paramTypes []ast.Term, resultSort ast.Sort, err error) {
	numParams, paramTypes, result := extractParams(ty)
	if sort, ok := result.(ast.Sort); ok {
		return numParams, paramTypes, sort, nil
	}
	return 0, nil, ast.Sort{}, &ValidationError{
		Msg: "inductive type must end in a Sort, got " + ast.Sprint(result),
	}
}

// ValidationError represents a validation error during inductive declaration.
type ValidationError struct {
	Msg string
}

func (e *ValidationError) Error() string {
	return e.Msg
}

// validateConstructorResult checks that a constructor's result type is the inductive
// applied to the correct number of parameters.
func validateConstructorResult(indName string, numParams int, c Constructor) error {
	resultTy := constructorResultType(c.Type)
	if resultTy == nil {
		return &ConstructorError{
			IndName:     indName,
			Constructor: c.Name,
			Message:     "could not determine result type",
		}
	}

	// For 0 params: result must be Global{indName}
	if numParams == 0 {
		if g, ok := resultTy.(ast.Global); ok && g.Name == indName {
			return nil
		}
		return &ConstructorError{
			IndName:     indName,
			Constructor: c.Name,
			Message:     "result type must be " + indName,
		}
	}

	// For n params: result must be App chain of indName applied to n args
	if !isAppOfGlobal(resultTy, indName) {
		return &ConstructorError{
			IndName:     indName,
			Constructor: c.Name,
			Message:     "result type must be " + indName + " applied to parameters",
		}
	}

	// Count applications must match numParams
	args := extractAppArgs(resultTy)
	if len(args) != numParams {
		return &ConstructorError{
			IndName:     indName,
			Constructor: c.Name,
			Message:     fmt.Sprintf("result type must have %d type arguments, got %d", numParams, len(args)),
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

// extractAppArgs collects arguments from an application chain.
// For ((f a) b) c, returns [a, b, c] in left-to-right order.
func extractAppArgs(t ast.Term) []ast.Term {
	var args []ast.Term
	for {
		if app, ok := t.(ast.App); ok {
			args = append([]ast.Term{app.U}, args...)
			t = app.T
		} else {
			break
		}
	}
	return args
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
