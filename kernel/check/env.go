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
	NumParams    int        // Number of parameters (uniform across constructors)
	ParamTypes   []ast.Term // Types of each parameter
	NumIndices   int        // Number of indices (vary per constructor)
	IndexTypes   []ast.Term // Types of each index (under param binders)
	Constructors []Constructor
	Eliminator   string   // Name of the elimination principle
	MutualGroup  []string // Names of other types in mutual block (nil for single inductives)
}

// MutualInductiveSpec specifies one inductive type in a mutual block.
type MutualInductiveSpec struct {
	Name         string
	Type         ast.Term
	Constructors []Constructor
	Eliminator   string
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
// For validated addition, use DeclareInductive or DeclareMutual.
func (g *GlobalEnv) AddInductive(name string, ty ast.Term, numParams int, paramTypes []ast.Term, numIndices int, indexTypes []ast.Term, constrs []Constructor, elim string, mutualGroup []string) {
	g.inductives[name] = &Inductive{
		Name:         name,
		Type:         ty,
		NumParams:    numParams,
		ParamTypes:   paramTypes,
		NumIndices:   numIndices,
		IndexTypes:   indexTypes,
		Constructors: constrs,
		Eliminator:   elim,
		MutualGroup:  mutualGroup,
	}
	g.order = append(g.order, name)
}

// DeclareInductive validates and adds a single inductive type to the environment.
// This is a convenience wrapper around DeclareMutual for non-mutual inductives.
func (g *GlobalEnv) DeclareInductive(name string, ty ast.Term, constrs []Constructor, elim string) error {
	return g.DeclareMutual([]MutualInductiveSpec{{
		Name:         name,
		Type:         ty,
		Constructors: constrs,
		Eliminator:   elim,
	}})
}

// DeclareMutual validates and adds mutually recursive inductive types.
// For a single inductive, use DeclareInductive for convenience.
//
// It checks:
// - Each inductive type is well-formed (Sort or Pi chain ending in Sort)
// - Each constructor type is well-formed (can reference any type in the mutual block)
// - Each constructor returns its inductive type applied to params/indices
// - The definition satisfies strict positivity across all mutual types
// It also generates and registers eliminators for each type.
func (g *GlobalEnv) DeclareMutual(specs []MutualInductiveSpec) error {
	if len(specs) == 0 {
		return nil
	}

	// Collect all type names for mutual reference
	indNames := make([]string, len(specs))
	for i, spec := range specs {
		indNames[i] = spec.Name
	}

	// Build mutual group (for non-single inductives)
	var mutualGroup []string
	if len(specs) > 1 {
		mutualGroup = indNames
	}

	// Validate and collect metadata for each type
	type indMetadata struct {
		totalArgs  int
		allArgTypes []ast.Term
		numParams  int
		numIndices int
	}
	metadata := make([]indMetadata, len(specs))

	// 1. Validate each inductive type signature
	for i, spec := range specs {
		totalArgs, allArgTypes, _, err := validateInductiveType(spec.Type)
		if err != nil {
			return &InductiveError{
				IndName: spec.Name,
				Message: err.Error(),
			}
		}
		metadata[i].totalArgs = totalArgs
		metadata[i].allArgTypes = allArgTypes
	}

	// 2. Temporarily add ALL types as axioms so constructors can reference each other
	for _, spec := range specs {
		g.AddAxiom(spec.Name, spec.Type)
	}

	// 3. Validate each constructor type is well-formed
	checker := NewChecker(g)
	for _, spec := range specs {
		for _, c := range spec.Constructors {
			if err := validateConstructorType(checker, spec.Name, c); err != nil {
				// Remove all temporary axioms on failure
				for _, s := range specs {
					delete(g.axioms, s.Name)
					g.removeFromOrder(s.Name)
				}
				return err
			}
		}
	}

	// Remove all temporary axioms - we'll add the real inductives
	for _, spec := range specs {
		delete(g.axioms, spec.Name)
		g.removeFromOrder(spec.Name)
	}

	// 4. Check strict positivity across all mutual types
	allConstrs := make(map[string][]Constructor)
	for _, spec := range specs {
		allConstrs[spec.Name] = spec.Constructors
	}
	if err := CheckMutualPositivity(indNames, allConstrs); err != nil {
		return err
	}

	// 5. Analyze params/indices and validate constructor results for each type
	for i, spec := range specs {
		numParams := analyzeParamsAndIndices(metadata[i].totalArgs, spec.Constructors)
		numIndices := metadata[i].totalArgs - numParams
		metadata[i].numParams = numParams
		metadata[i].numIndices = numIndices

		for _, c := range spec.Constructors {
			if err := validateConstructorResult(spec.Name, metadata[i].totalArgs, c); err != nil {
				return err
			}
		}
	}

	// 6. Add all inductives to the environment
	for i, spec := range specs {
		m := metadata[i]
		paramTypes := m.allArgTypes[:m.numParams]
		indexTypes := m.allArgTypes[m.numParams:]
		g.AddInductive(spec.Name, spec.Type, m.numParams, paramTypes, m.numIndices, indexTypes, spec.Constructors, spec.Eliminator, mutualGroup)
	}

	// 7. Generate and register eliminators for each type
	for _, spec := range specs {
		ind := g.inductives[spec.Name]
		elimType := GenerateRecursorType(ind)
		g.AddAxiom(spec.Eliminator, elimType)

		// Register the recursor for generic reduction
		recursorInfo := buildRecursorInfo(ind)
		eval.RegisterRecursor(recursorInfo)
	}

	return nil
}

// buildRecursorInfo builds RecursorInfo from an inductive definition.
// For parameterized inductives, NumParams is extracted from the inductive type,
// and constructor arg counts exclude parameters.
// For indexed inductives, NumIndices is also extracted and IndexArgPositions is computed.
//
// For mutual inductives with SEPARATE eliminators, only same-type recursive args
// generate IHs. Cross-type recursion must be handled explicitly in case functions.
// This matches the generated eliminator types where IHs are only for the current type.
func buildRecursorInfo(ind *Inductive) *eval.RecursorInfo {
	info := &eval.RecursorInfo{
		ElimName:   ind.Eliminator,
		IndName:    ind.Name,
		NumParams:  ind.NumParams,
		NumIndices: ind.NumIndices,
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

		// Find recursive arguments among data args and compute their index positions
		// For separate eliminators, only same-type args are considered recursive.
		recursiveIdx := []int{}
		indexArgPositions := make(map[int][]int)
		recursiveArgElims := make(map[int]string)

		for j, arg := range dataArgs {
			// Only check for same-type recursion (not cross-type for mutual inductives)
			if isRecursiveArgType(ind.Name, arg.Type) {
				recursiveIdx = append(recursiveIdx, j)
				recursiveArgElims[j] = ind.Eliminator

				// Compute index positions for this recursive arg
				if ind.NumIndices > 0 {
					idxPositions := computeIndexArgPositions(arg.Type, j, ind.NumParams, ind.NumIndices)
					if len(idxPositions) > 0 {
						indexArgPositions[j] = idxPositions
					}
				}
			}
		}

		info.Ctors[i] = eval.ConstructorInfo{
			Name:              c.Name,
			NumArgs:           len(dataArgs), // Only count non-param args
			RecursiveIdx:      recursiveIdx,
			IndexArgPositions: indexArgPositions,
			RecursiveArgElims: recursiveArgElims,
		}
	}

	return info
}

// computeIndexArgPositions computes the data-arg positions that serve as indices
// for a recursive argument's type.
//
// For a recursive arg at data position j with type (Ind params... indices...):
//   - Extract the index args from the type's application chain
//   - For each index that is a Var{V}, compute its data-arg position
//   - Return the list of data-arg positions
//
// DE BRUIJN INVARIANT:
// When checking arg at data position j, Var{V} refers to the (j-1-V)th data arg.
// This follows from the binder structure: params come first, then data args in order.
// Under binders [param1, ..., paramK, data0, ..., data(j-1)], at position j we have
// total binder count K+j. A Var{V} refers to binder at index (K+j-1-V), and after
// subtracting K params, we get data position (j-1-V).
//
// Example: vcons has data args [n, x, xs] where xs : Vec A n
//   - xs is at data position 2
//   - Its type Vec A n has index n = Var{1} (under binders A, n, x)
//   - data-arg position of n = 2 - 1 - 1 = 0
//   - Returns [0]
//
// COMPLETENESS: If any index is a computed expression (not a Var), it won't have
// a position entry. The evaluator checks len(positions) == NumIndices before using
// metadata and falls back to heuristics if incomplete.
func computeIndexArgPositions(argType ast.Term, dataArgPos int, numParams int, numIndices int) []int {
	// Extract the application args from the recursive arg's type
	typeArgs := extractAppArgs(argType)

	// Skip parameters, get index args
	if len(typeArgs) <= numParams {
		return nil // No indices
	}
	indexArgs := typeArgs[numParams:]

	// Limit to expected number of indices
	if len(indexArgs) > numIndices {
		indexArgs = indexArgs[:numIndices]
	}

	// Map each index Var to its data-arg position
	var positions []int
	for _, idxArg := range indexArgs {
		if v, ok := idxArg.(ast.Var); ok {
			// Under the context where this type is checked:
			// - We're at data-arg position dataArgPos
			// - All-arg position is dataArgPos + numParams
			// - Var{V} refers to all-arg position (dataArgPos + numParams - 1 - V)
			// - Data-arg position = dataArgPos - 1 - V
			dataPos := dataArgPos - 1 - v.Ix
			if dataPos >= 0 && dataPos < dataArgPos {
				positions = append(positions, dataPos)
			}
		}
		// Non-variable indices (like computed expressions) are handled
		// by evaluating at runtime; we just won't have a position for them.
	}

	return positions
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

// extractPiChain extracts all Pi arguments from a type.
// For Type -> Nat -> Type, returns ([Type, Nat], Type)
// For Type, returns ([], Type)
func extractPiChain(ty ast.Term) (argTypes []ast.Term, resultSort ast.Term) {
	current := ty
	for {
		if pi, ok := current.(ast.Pi); ok {
			argTypes = append(argTypes, pi.A)
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
// For parameterized/indexed: must be Pi chain ending in Sort
//
// Note: This extracts ALL Pi args. The caller must determine which are
// parameters vs indices by analyzing constructor result types.
// Returns the total number of args, their types, and the result sort.
func validateInductiveType(ty ast.Term) (numArgs int, argTypes []ast.Term, resultSort ast.Sort, err error) {
	argTypes, result := extractPiChain(ty)
	numArgs = len(argTypes)
	if sort, ok := result.(ast.Sort); ok {
		return numArgs, argTypes, sort, nil
	}
	return 0, nil, ast.Sort{}, &ValidationError{
		Msg: "inductive type must end in a Sort, got " + ast.Sprint(result),
	}
}

// analyzeParamsAndIndices determines how many of the inductive type's arguments
// are parameters (uniform across constructors) vs indices (can vary).
//
// For Vec : Type -> Nat -> Type with constructors:
//
//	vnil  : (A : Type) -> Vec A zero
//	vcons : (A : Type) -> A -> (n : Nat) -> Vec A n -> Vec A (succ n)
//
// The first arg (Type) is a parameter (always Var referring to the bound A).
// The second arg (Nat) is an index (varies: zero, succ n).
//
// Algorithm: For each position in the result type's application chain,
// check if ALL constructors use the same variable reference. If so, it's a param.
func analyzeParamsAndIndices(totalArgs int, constrs []Constructor) (numParams int) {
	if len(constrs) == 0 || totalArgs == 0 {
		return 0
	}

	// For each position, check if all constructors agree on using a variable
	for pos := 0; pos < totalArgs; pos++ {
		isParam := true
		for _, c := range constrs {
			resultTy := constructorResultType(c.Type)
			args := extractAppArgs(resultTy)
			if pos >= len(args) {
				isParam = false
				break
			}
			// Check if this arg is a variable (parameter reference)
			// For a param at position pos, under the constructor's binders,
			// it should reference one of the parameter variables
			if !isParamReference(args[pos]) {
				isParam = false
				break
			}
		}
		if isParam {
			numParams++
		} else {
			// Once we hit a non-param, all remaining are indices
			break
		}
	}
	return numParams
}

// isParamReference checks if a term is a simple variable reference (parameter).
// Parameters are always passed through as variables, while indices can be
// arbitrary expressions.
func isParamReference(t ast.Term) bool {
	_, ok := t.(ast.Var)
	return ok
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
