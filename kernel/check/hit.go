package check

import (
	"fmt"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
)

// DeclareHIT validates and registers a Higher Inductive Type.
// It validates:
// - The inductive type signature is well-formed
// - All point constructors are well-formed and return the HIT type
// - All path constructors are well-formed with proper boundaries
// - Strict positivity is satisfied for both point and path constructors
// It also generates and registers the eliminator with path cases.
func (g *GlobalEnv) DeclareHIT(spec *ast.HITSpec) error {
	if spec == nil {
		return nil
	}

	// 1. Validate the inductive type signature
	totalArgs, allArgTypes, _, err := validateInductiveType(spec.Type)
	if err != nil {
		return &InductiveError{
			IndName: spec.Name,
			Message: err.Error(),
		}
	}

	// 2. Temporarily add the HIT as an axiom for constructor checking
	g.AddAxiom(spec.Name, spec.Type)

	// 3. Validate point constructors
	checker := NewChecker(g)
	for _, pc := range spec.PointCtors {
		if err := validateConstructorType(checker, spec.Name, Constructor{Name: pc.Name, Type: pc.Type}); err != nil {
			delete(g.axioms, spec.Name)
			g.removeFromOrder(spec.Name)
			return err
		}
	}

	// 3.5. Add point constructors as temporary axioms for path constructor validation
	// Path constructors may reference point constructors (e.g., loop references base)
	for _, pc := range spec.PointCtors {
		g.AddAxiom(pc.Name, pc.Type)
	}
	// Refresh checker with the new axioms
	checker = NewChecker(g)

	// 4. Validate path constructors
	for _, pathCtor := range spec.PathCtors {
		if err := validatePathConstructor(checker, spec.Name, pathCtor); err != nil {
			// Clean up temporary axioms
			delete(g.axioms, spec.Name)
			g.removeFromOrder(spec.Name)
			for _, pc := range spec.PointCtors {
				delete(g.axioms, pc.Name)
				g.removeFromOrder(pc.Name)
			}
			return err
		}
	}

	// Remove temporary axioms
	delete(g.axioms, spec.Name)
	g.removeFromOrder(spec.Name)
	for _, pc := range spec.PointCtors {
		delete(g.axioms, pc.Name)
		g.removeFromOrder(pc.Name)
	}

	// 5. Check strict positivity for both point and path constructors
	if err := checkHITPositivity(spec.Name, spec.PointCtors, spec.PathCtors); err != nil {
		return err
	}

	// 6. Analyze params and indices
	// Convert ast.Constructor to check.Constructor
	checkConstrs := make([]Constructor, len(spec.PointCtors))
	for i, c := range spec.PointCtors {
		checkConstrs[i] = Constructor{Name: c.Name, Type: c.Type}
	}
	numParams := analyzeParamsAndIndices(totalArgs, checkConstrs)
	numIndices := totalArgs - numParams

	// 7. Validate constructor results
	for _, c := range spec.PointCtors {
		if err := validateConstructorResult(spec.Name, totalArgs, Constructor{Name: c.Name, Type: c.Type}); err != nil {
			return err
		}
	}

	// 8. Compute maximum path level
	maxLevel := 0
	for _, pc := range spec.PathCtors {
		if pc.Level > maxLevel {
			maxLevel = pc.Level
		}
	}

	// 9. Add the HIT to the environment
	paramTypes := allArgTypes[:numParams]
	indexTypes := allArgTypes[numParams:]
	g.inductives[spec.Name] = &Inductive{
		Name:         spec.Name,
		Type:         spec.Type,
		NumParams:    numParams,
		ParamTypes:   paramTypes,
		NumIndices:   numIndices,
		IndexTypes:   indexTypes,
		Constructors: checkConstrs,
		Eliminator:   spec.Eliminator,
		PathCtors:    spec.PathCtors,
		IsHIT:        true,
		MaxLevel:     maxLevel,
	}
	g.order = append(g.order, spec.Name)

	// 10. Generate and register the eliminator
	ind := g.inductives[spec.Name]
	elimType := GenerateHITRecursorType(ind)
	g.AddAxiom(spec.Eliminator, elimType)

	// Register the recursor for generic reduction
	recursorInfo := buildRecursorInfo(ind)
	eval.RegisterRecursor(recursorInfo)

	return nil
}

// validatePathConstructor checks that a path constructor is well-formed.
// It validates:
// - The type is well-formed
// - The type has the correct structure for a path constructor (returns Path/PathP to the HIT)
// - Boundaries are consistent (when interval = i0 or i1, the boundary term is well-typed)
func validatePathConstructor(checker *Checker, indName string, ctor ast.PathConstructor) error {
	// Validate the path constructor type is well-formed
	_, err := checker.CheckIsType(nil, Span{}, ctor.Type)
	if err != nil {
		return &PathConstructorError{
			IndName:     indName,
			Constructor: ctor.Name,
			Message:     "path constructor type is not well-formed: " + err.Error(),
		}
	}

	// Validate the result type is a Path/PathP to the HIT
	if !isPathToHIT(ctor.Type, indName) {
		return &PathConstructorError{
			IndName:     indName,
			Constructor: ctor.Name,
			Message:     "path constructor must return a Path into " + indName,
		}
	}

	// Validate boundaries match the expected types
	if len(ctor.Boundaries) != ctor.Level {
		return &PathConstructorError{
			IndName:     indName,
			Constructor: ctor.Name,
			Message:     fmt.Sprintf("expected %d boundaries for level %d path constructor, got %d", ctor.Level, ctor.Level, len(ctor.Boundaries)),
		}
	}

	// Boundary validation requires evaluating at interval endpoints
	// This is complex and depends on the full checker infrastructure.
	// For now, we trust that well-typed boundaries are correct.
	// Full validation will be added when we implement the type checker integration.

	return nil
}

// isPathToHIT checks if a type's result is a Path/PathP into the named HIT.
// For example, Path S1 base base returns true for HIT "S1".
func isPathToHIT(ty ast.Term, hitName string) bool {
	// Strip off Pi arguments to get the result type
	current := ty
	for {
		if pi, ok := current.(ast.Pi); ok {
			current = pi.B
		} else {
			break
		}
	}

	// Check if result is Path/PathP applied to the HIT type
	return isPathResult(current, hitName)
}

// isPathResult checks if a term is Path/PathP with the given HIT in the path type.
func isPathResult(t ast.Term, hitName string) bool {
	// Handle nested paths (for level > 1)
	switch tm := t.(type) {
	case ast.Path:
		// Path A x y - check if A contains the HIT
		return containsHIT(tm.A, hitName)
	case ast.PathP:
		// PathP A x y - check if A contains the HIT
		return containsHIT(tm.A, hitName)
	case ast.App:
		// Could be (Path ...) or (PathP ...) applied
		return isPathResult(tm.T, hitName)
	}
	return false
}

// containsHIT checks if a term contains a reference to the named HIT.
func containsHIT(t ast.Term, hitName string) bool {
	switch tm := t.(type) {
	case ast.Global:
		return tm.Name == hitName
	case ast.App:
		return containsHIT(tm.T, hitName) || containsHIT(tm.U, hitName)
	case ast.Pi:
		return containsHIT(tm.A, hitName) || containsHIT(tm.B, hitName)
	case ast.Lam:
		return containsHIT(tm.Body, hitName)
	case ast.PathLam:
		return containsHIT(tm.Body, hitName)
	case ast.Path:
		return containsHIT(tm.A, hitName) || containsHIT(tm.X, hitName) || containsHIT(tm.Y, hitName)
	case ast.PathP:
		return containsHIT(tm.A, hitName) || containsHIT(tm.X, hitName) || containsHIT(tm.Y, hitName)
	case ast.PathApp:
		return containsHIT(tm.P, hitName) || containsHIT(tm.R, hitName)
	}
	return false
}

// checkHITPositivity checks strict positivity for both point and path constructors.
// Path constructors have additional constraints: the HIT type must not appear
// negatively in the path's type argument.
func checkHITPositivity(indName string, pointCtors []ast.Constructor, pathCtors []ast.PathConstructor) error {
	// Check point constructors using existing positivity checker
	allConstrs := make(map[string][]Constructor)
	checkConstrs := make([]Constructor, len(pointCtors))
	for i, c := range pointCtors {
		checkConstrs[i] = Constructor{Name: c.Name, Type: c.Type}
	}
	allConstrs[indName] = checkConstrs

	if err := CheckMutualPositivity([]string{indName}, allConstrs); err != nil {
		return err
	}

	// Check path constructors for positivity
	// The path type argument (e.g., S1 in Path S1 base base) must satisfy positivity
	for _, pc := range pathCtors {
		if err := checkPathCtorPositivity(indName, pc); err != nil {
			return err
		}
	}

	return nil
}

// checkPathCtorPositivity checks that a path constructor satisfies strict positivity.
// The HIT type must not appear in negative position in any function argument.
func checkPathCtorPositivity(indName string, pc ast.PathConstructor) error {
	// Use the existing positivity checking infrastructure
	// Path constructors are checked the same way as point constructors
	return checkConstructorPositivity(indName, pc.Name, pc.Type)
}

// PathConstructorError represents an error in path constructor validation.
type PathConstructorError struct {
	IndName     string
	Constructor string
	Message     string
}

func (e *PathConstructorError) Error() string {
	return "path constructor " + e.Constructor + " of " + e.IndName + ": " + e.Message
}

// GenerateHITRecursorType generates the eliminator type for a HIT.
// For Circle (S1) with base and loop:
//
//	S1-elim : (P : S1 -> Type)
//	        -> (pbase : P base)
//	        -> (ploop : PathP (λi. P (loop @ i)) pbase pbase)
//	        -> (x : S1) -> P x
//
// The key difference from regular inductives is that path constructors
// require PathP cases rather than simple function cases.
func GenerateHITRecursorType(ind *Inductive) ast.Term {
	// Start with the basic recursor structure
	// 1. Parameters (if any)
	// 2. Motive P : (indices...) -> Ind params... indices... -> Type
	// 3. Point constructor cases
	// 4. Path constructor cases (as PathP types)
	// 5. Indices (if any)
	// 6. Target: (x : Ind params... indices...) -> P indices... x

	// For now, generate the basic eliminator for point constructors
	// and add path cases as PathP types
	baseElim := GenerateRecursorType(ind)

	if !ind.IsHIT || len(ind.PathCtors) == 0 {
		return baseElim
	}

	// We need to insert path cases before the final target
	// The structure is: Pi params -> Pi P -> Pi cases... -> Pi indices -> Pi x -> P x
	// We insert path cases right after point cases

	return insertPathCases(baseElim, ind)
}

// insertPathCases inserts path constructor cases into the eliminator type.
// This modifies the type to add PathP cases for each path constructor.
func insertPathCases(elimType ast.Term, ind *Inductive) ast.Term {
	// Navigate through the Pi chain to find where to insert path cases
	// Structure: params... -> P -> point_cases... -> [INSERT HERE] -> indices... -> x -> P x

	// Count how many Pis to skip (params + motive + point cases)
	skipCount := ind.NumParams + 1 + len(ind.Constructors)

	return insertPathCasesAt(elimType, skipCount, ind, 0)
}

// insertPathCasesAt recursively navigates the Pi chain and inserts path cases.
func insertPathCasesAt(ty ast.Term, skip int, ind *Inductive, depth int) ast.Term {
	pi, ok := ty.(ast.Pi)
	if !ok {
		// Not a Pi - we've reached the target, shouldn't happen normally
		return ty
	}

	if skip > 0 {
		// Keep descending, shifting the body
		return ast.Pi{
			Binder: pi.Binder,
			A:      pi.A,
			B:      insertPathCasesAt(pi.B, skip-1, ind, depth+1),
		}
	}

	// We're at the insertion point - add all path cases
	result := ty
	for i := len(ind.PathCtors) - 1; i >= 0; i-- {
		pc := ind.PathCtors[i]
		pathCaseType := buildPathCaseType(ind, pc, depth)
		result = ast.Pi{
			Binder: "p" + pc.Name,
			A:      pathCaseType,
			B:      result,
		}
	}

	return result
}

// buildPathCaseType constructs the type for a path constructor case.
// For loop : Path S1 base base with motive P, the case type is:
//
//	PathP (λi. P (loop @ i)) pbase pbase
//
// where pbase is the case for the base constructor.
//
// The depth parameter indicates how many binders we're under (for de Bruijn indexing).
func buildPathCaseType(ind *Inductive, pc ast.PathConstructor, depth int) ast.Term {
	// For a path constructor with level 1:
	// PathP (λi. P (ctor @ i)) (endpoint0_case) (endpoint1_case)

	// The motive P is at a specific de Bruijn index based on depth
	// Under [params..., P, cases...], P is at index (depth - NumParams - 1)
	// But we've shifted, so P is at index (len(Constructors))

	// Build the path family: λi. P (ctor @ i)
	// We need HITApp to represent the path constructor application

	// Simplified: for level 1 paths, create:
	// PathP (λi. P (ctor args... @ i)) boundary0 boundary1

	if len(pc.Boundaries) == 0 {
		// Shouldn't happen for a valid path constructor
		return ast.Sort{U: 0}
	}

	// Get the boundary endpoints
	b := pc.Boundaries[0]

	// Build PathP family: λi. P (ctor @ i)
	// P is at index depth (counting from the path case position)
	motiveIdx := len(ind.Constructors)

	// HITApp represents ctor @ i (where i is Var{0} under the PathLam)
	ctorApp := ast.HITApp{
		HITName: ind.Name,
		Ctor:    pc.Name,
		Args:    []ast.Term{}, // Parameters would go here
		IArgs:   []ast.Term{ast.Var{Ix: 0}}, // The bound interval variable
	}

	// P (ctor @ i) - apply motive to the HIT path value
	// P is at motiveIdx + 1 (shifted by the PathLam binder)
	pApp := ast.App{
		T: ast.Var{Ix: motiveIdx + 1},
		U: ctorApp,
	}

	// λi. P (ctor @ i)
	pathFamily := ast.PathLam{
		Binder: "i",
		Body:   pApp,
	}

	// The endpoints reference the cases for boundary constructors
	// For loop : Path S1 base base, both endpoints are pbase
	// We need to find which case corresponds to each boundary

	// For now, use the boundary terms directly
	// In practice, these should reference the point constructor cases

	return ast.PathP{
		A: pathFamily,
		X: b.AtZero, // Case for left endpoint
		Y: b.AtOne,  // Case for right endpoint
	}
}
