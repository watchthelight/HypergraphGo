// unify.go implements the core unification algorithm.
//
// See doc.go for package overview.

package unify

import (
	"fmt"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
	"github.com/watchthelight/HypergraphGo/kernel/subst"
)

// Constraint represents an equality constraint between two terms.
type Constraint struct {
	LHS ast.Term // Left-hand side
	RHS ast.Term // Right-hand side
}

// UnifyError represents an error during unification.
type UnifyError struct {
	Message     string
	Constraints []Constraint // Remaining constraints when error occurred
}

func (e *UnifyError) Error() string {
	return e.Message
}

// UnifyResult represents the result of unification.
type UnifyResult struct {
	// Solved maps metavariable IDs to their solutions.
	Solved map[int]ast.Term

	// Unsolved contains constraints that couldn't be solved.
	Unsolved []Constraint

	// Errors contains any errors encountered.
	Errors []UnifyError
}

// Unifier performs unification of type constraints.
type Unifier struct {
	// Solutions maps metavariable IDs to their solutions.
	solutions map[int]ast.Term

	// worklist contains constraints still to be processed.
	worklist []Constraint

	// unsolved contains constraints that couldn't be solved yet.
	unsolved []Constraint

	// errors contains any errors encountered.
	errors []UnifyError
}

// NewUnifier creates a new unifier.
func NewUnifier() *Unifier {
	return &Unifier{
		solutions: make(map[int]ast.Term),
	}
}

// AddConstraint adds a constraint to the worklist.
func (u *Unifier) AddConstraint(lhs, rhs ast.Term) {
	u.worklist = append(u.worklist, Constraint{LHS: lhs, RHS: rhs})
}

// Solve processes all constraints and returns the result.
func (u *Unifier) Solve() UnifyResult {
	// Process worklist until empty
	for len(u.worklist) > 0 {
		c := u.worklist[0]
		u.worklist = u.worklist[1:]

		if err := u.processConstraint(c); err != nil {
			u.errors = append(u.errors, *err)
		}
	}

	return UnifyResult{
		Solved:   u.solutions,
		Unsolved: u.unsolved,
		Errors:   u.errors,
	}
}

// GetSolution returns the solution for a metavariable if it exists.
func (u *Unifier) GetSolution(id int) (ast.Term, bool) {
	sol, ok := u.solutions[id]
	return sol, ok
}

// processConstraint processes a single constraint.
func (u *Unifier) processConstraint(c Constraint) *UnifyError {
	// Normalize both sides
	lhs := normalize(c.LHS)
	rhs := normalize(c.RHS)

	// Apply known solutions
	lhs = u.applySolutions(lhs)
	rhs = u.applySolutions(rhs)

	// Try to unify
	return u.unify(lhs, rhs)
}

// normalize normalizes a term using NbE.
// Meta terms are left unchanged since they can't be evaluated.
func normalize(t ast.Term) ast.Term {
	// Check if term contains metavariables - if so, skip NbE
	// since the evaluator doesn't handle metas
	if hasMeta(t) {
		return t
	}
	return eval.EvalNBE(t)
}

// applySolutions substitutes known solutions into a term.
func (u *Unifier) applySolutions(t ast.Term) ast.Term {
	if len(u.solutions) == 0 {
		return t
	}
	return u.zonkTerm(t)
}

// zonkTerm recursively substitutes metavariables with their solutions.
func (u *Unifier) zonkTerm(t ast.Term) ast.Term {
	if t == nil {
		return nil
	}

	switch tt := t.(type) {
	case ast.Meta:
		if sol, ok := u.solutions[tt.ID]; ok {
			// Apply arguments to solution
			result := u.zonkTerm(sol)
			for _, arg := range tt.Args {
				result = ast.App{T: result, U: u.zonkTerm(arg)}
			}
			return result
		}
		// Keep meta but zonk arguments
		newArgs := make([]ast.Term, len(tt.Args))
		for i, arg := range tt.Args {
			newArgs[i] = u.zonkTerm(arg)
		}
		return ast.Meta{ID: tt.ID, Args: newArgs}

	case ast.Var, ast.Global, ast.Sort:
		return t

	case ast.Pi:
		return ast.Pi{
			Binder: tt.Binder,
			A:      u.zonkTerm(tt.A),
			B:      u.zonkTerm(tt.B),
		}

	case ast.Lam:
		return ast.Lam{
			Binder: tt.Binder,
			Ann:    u.zonkTerm(tt.Ann),
			Body:   u.zonkTerm(tt.Body),
		}

	case ast.App:
		return ast.App{
			T: u.zonkTerm(tt.T),
			U: u.zonkTerm(tt.U),
		}

	case ast.Sigma:
		return ast.Sigma{
			Binder: tt.Binder,
			A:      u.zonkTerm(tt.A),
			B:      u.zonkTerm(tt.B),
		}

	case ast.Pair:
		return ast.Pair{
			Fst: u.zonkTerm(tt.Fst),
			Snd: u.zonkTerm(tt.Snd),
		}

	case ast.Fst:
		return ast.Fst{P: u.zonkTerm(tt.P)}

	case ast.Snd:
		return ast.Snd{P: u.zonkTerm(tt.P)}

	case ast.Let:
		return ast.Let{
			Binder: tt.Binder,
			Ann:    u.zonkTerm(tt.Ann),
			Val:    u.zonkTerm(tt.Val),
			Body:   u.zonkTerm(tt.Body),
		}

	case ast.Id:
		return ast.Id{
			A: u.zonkTerm(tt.A),
			X: u.zonkTerm(tt.X),
			Y: u.zonkTerm(tt.Y),
		}

	case ast.Refl:
		return ast.Refl{
			A: u.zonkTerm(tt.A),
			X: u.zonkTerm(tt.X),
		}

	case ast.J:
		return ast.J{
			A: u.zonkTerm(tt.A),
			C: u.zonkTerm(tt.C),
			D: u.zonkTerm(tt.D),
			X: u.zonkTerm(tt.X),
			Y: u.zonkTerm(tt.Y),
			P: u.zonkTerm(tt.P),
		}

	// Path types
	case ast.Path:
		return ast.Path{
			A: u.zonkTerm(tt.A),
			X: u.zonkTerm(tt.X),
			Y: u.zonkTerm(tt.Y),
		}

	case ast.PathP:
		return ast.PathP{
			A: u.zonkTerm(tt.A),
			X: u.zonkTerm(tt.X),
			Y: u.zonkTerm(tt.Y),
		}

	case ast.PathLam:
		return ast.PathLam{
			Binder: tt.Binder,
			Body:   u.zonkTerm(tt.Body),
		}

	case ast.PathApp:
		return ast.PathApp{
			P: u.zonkTerm(tt.P),
			R: u.zonkTerm(tt.R),
		}

	case ast.Transport:
		return ast.Transport{
			A: u.zonkTerm(tt.A),
			E: u.zonkTerm(tt.E),
		}

	// Interval and faces - no subterms contain metas
	case ast.Interval, ast.I0, ast.I1, ast.IVar:
		return t

	case ast.FaceTop, ast.FaceBot, ast.FaceEq, ast.FaceAnd, ast.FaceOr:
		return t // Face types don't contain Terms

	case ast.Partial:
		return ast.Partial{
			Phi: tt.Phi, // Face, not Term
			A:   u.zonkTerm(tt.A),
		}

	case ast.System:
		branches := make([]ast.SystemBranch, len(tt.Branches))
		for i, b := range tt.Branches {
			branches[i] = ast.SystemBranch{
				Phi:  b.Phi, // Face, not Term
				Term: u.zonkTerm(b.Term),
			}
		}
		return ast.System{Branches: branches}

	case ast.Comp:
		return ast.Comp{
			IBinder: tt.IBinder,
			A:       u.zonkTerm(tt.A),
			Phi:     tt.Phi, // Face, not Term
			Tube:    u.zonkTerm(tt.Tube),
			Base:    u.zonkTerm(tt.Base),
		}

	case ast.HComp:
		return ast.HComp{
			A:    u.zonkTerm(tt.A),
			Phi:  tt.Phi, // Face, not Term
			Tube: u.zonkTerm(tt.Tube),
			Base: u.zonkTerm(tt.Base),
		}

	case ast.Fill:
		return ast.Fill{
			IBinder: tt.IBinder,
			A:       u.zonkTerm(tt.A),
			Phi:     tt.Phi, // Face, not Term
			Tube:    u.zonkTerm(tt.Tube),
			Base:    u.zonkTerm(tt.Base),
		}

	case ast.Glue:
		branches := make([]ast.GlueBranch, len(tt.System))
		for i, b := range tt.System {
			branches[i] = ast.GlueBranch{
				Phi:   b.Phi, // Face, not Term
				T:     u.zonkTerm(b.T),
				Equiv: u.zonkTerm(b.Equiv),
			}
		}
		return ast.Glue{
			A:      u.zonkTerm(tt.A),
			System: branches,
		}

	case ast.GlueElem:
		branches := make([]ast.GlueElemBranch, len(tt.System))
		for i, b := range tt.System {
			branches[i] = ast.GlueElemBranch{
				Phi:  b.Phi, // Face, not Term
				Term: u.zonkTerm(b.Term),
			}
		}
		return ast.GlueElem{
			Base:   u.zonkTerm(tt.Base),
			System: branches,
		}

	case ast.Unglue:
		return ast.Unglue{
			Ty: u.zonkTerm(tt.Ty),
			G:  u.zonkTerm(tt.G),
		}

	case ast.UA:
		return ast.UA{
			A:     u.zonkTerm(tt.A),
			B:     u.zonkTerm(tt.B),
			Equiv: u.zonkTerm(tt.Equiv),
		}

	case ast.UABeta:
		return ast.UABeta{
			Equiv: u.zonkTerm(tt.Equiv),
			Arg:   u.zonkTerm(tt.Arg),
		}

	// HIT terms - only HITApp implements isCoreTerm()
	case ast.HITApp:
		args := make([]ast.Term, len(tt.Args))
		for i, arg := range tt.Args {
			args[i] = u.zonkTerm(arg)
		}
		iArgs := make([]ast.Term, len(tt.IArgs))
		for i, arg := range tt.IArgs {
			iArgs[i] = u.zonkTerm(arg)
		}
		return ast.HITApp{
			HITName: tt.HITName,
			Ctor:    tt.Ctor,
			Args:    args,
			IArgs:   iArgs,
		}

	default:
		// Unknown term type, return as-is
		return t
	}
}

// unify attempts to unify two terms.
func (u *Unifier) unify(lhs, rhs ast.Term) *UnifyError {
	// If already equal, done
	if eval.AlphaEq(lhs, rhs) {
		return nil
	}

	// Check for metavariable on either side
	if meta, ok := lhs.(ast.Meta); ok {
		return u.solveMeta(meta, rhs)
	}
	if meta, ok := rhs.(ast.Meta); ok {
		return u.solveMeta(meta, lhs)
	}

	// Structural unification
	switch l := lhs.(type) {
	case ast.Var:
		r, ok := rhs.(ast.Var)
		if !ok || l.Ix != r.Ix {
			return u.fail("cannot unify %v with %v", lhs, rhs)
		}
		return nil

	case ast.Global:
		r, ok := rhs.(ast.Global)
		if !ok || l.Name != r.Name {
			return u.fail("cannot unify %v with %v", lhs, rhs)
		}
		return nil

	case ast.Sort:
		r, ok := rhs.(ast.Sort)
		if !ok || l.U != r.U {
			return u.fail("cannot unify %v with %v", lhs, rhs)
		}
		return nil

	case ast.Pi:
		r, ok := rhs.(ast.Pi)
		if !ok {
			return u.fail("cannot unify Pi with %T", rhs)
		}
		// Unify domains
		if err := u.unify(l.A, r.A); err != nil {
			return err
		}
		// Unify codomains
		return u.unify(l.B, r.B)

	case ast.Lam:
		r, ok := rhs.(ast.Lam)
		if !ok {
			return u.fail("cannot unify Lam with %T", rhs)
		}
		// Unify bodies (annotation is optional)
		return u.unify(l.Body, r.Body)

	case ast.App:
		r, ok := rhs.(ast.App)
		if !ok {
			return u.fail("cannot unify App with %T", rhs)
		}
		if err := u.unify(l.T, r.T); err != nil {
			return err
		}
		return u.unify(l.U, r.U)

	case ast.Sigma:
		r, ok := rhs.(ast.Sigma)
		if !ok {
			return u.fail("cannot unify Sigma with %T", rhs)
		}
		if err := u.unify(l.A, r.A); err != nil {
			return err
		}
		return u.unify(l.B, r.B)

	case ast.Pair:
		r, ok := rhs.(ast.Pair)
		if !ok {
			return u.fail("cannot unify Pair with %T", rhs)
		}
		if err := u.unify(l.Fst, r.Fst); err != nil {
			return err
		}
		return u.unify(l.Snd, r.Snd)

	case ast.Fst:
		r, ok := rhs.(ast.Fst)
		if !ok {
			return u.fail("cannot unify Fst with %T", rhs)
		}
		return u.unify(l.P, r.P)

	case ast.Snd:
		r, ok := rhs.(ast.Snd)
		if !ok {
			return u.fail("cannot unify Snd with %T", rhs)
		}
		return u.unify(l.P, r.P)

	case ast.Id:
		r, ok := rhs.(ast.Id)
		if !ok {
			return u.fail("cannot unify Id with %T", rhs)
		}
		if err := u.unify(l.A, r.A); err != nil {
			return err
		}
		if err := u.unify(l.X, r.X); err != nil {
			return err
		}
		return u.unify(l.Y, r.Y)

	case ast.Refl:
		r, ok := rhs.(ast.Refl)
		if !ok {
			return u.fail("cannot unify Refl with %T", rhs)
		}
		if err := u.unify(l.A, r.A); err != nil {
			return err
		}
		return u.unify(l.X, r.X)

	case ast.Path:
		r, ok := rhs.(ast.Path)
		if !ok {
			return u.fail("cannot unify Path with %T", rhs)
		}
		if err := u.unify(l.A, r.A); err != nil {
			return err
		}
		if err := u.unify(l.X, r.X); err != nil {
			return err
		}
		return u.unify(l.Y, r.Y)

	case ast.PathP:
		r, ok := rhs.(ast.PathP)
		if !ok {
			return u.fail("cannot unify PathP with %T", rhs)
		}
		if err := u.unify(l.A, r.A); err != nil {
			return err
		}
		if err := u.unify(l.X, r.X); err != nil {
			return err
		}
		return u.unify(l.Y, r.Y)

	default:
		// Can't unify, add to unsolved
		u.unsolved = append(u.unsolved, Constraint{LHS: lhs, RHS: rhs})
		return nil
	}
}

// solveMeta attempts to solve a metavariable.
func (u *Unifier) solveMeta(meta ast.Meta, solution ast.Term) *UnifyError {
	// Check if already solved
	if existing, ok := u.solutions[meta.ID]; ok {
		// Already solved, check consistency
		return u.unify(existing, solution)
	}

	// Check if solution is the same metavariable with same args
	if solMeta, ok := solution.(ast.Meta); ok && solMeta.ID == meta.ID {
		if len(meta.Args) == len(solMeta.Args) {
			allEqual := true
			for i := range meta.Args {
				if !eval.AlphaEq(meta.Args[i], solMeta.Args[i]) {
					allEqual = false
					break
				}
			}
			if allEqual {
				return nil // Same meta with same args - trivially equal
			}
		}
	}

	// Check occurs check
	if u.occurs(meta.ID, solution) {
		return u.fail("occurs check failed: ?%d occurs in %v", meta.ID, solution)
	}

	// If meta has no arguments, we can directly use the solution
	if len(meta.Args) == 0 {
		u.solutions[meta.ID] = solution
		u.recheckUnsolved()
		return nil
	}

	// Check if this is a Miller pattern (args are distinct variables)
	if !u.isPattern(meta) {
		// Not a pattern, defer
		u.unsolved = append(u.unsolved, Constraint{LHS: meta, RHS: solution})
		return nil
	}

	// Invert the pattern to get the solution
	inverted, err := u.invertPattern(meta, solution)
	if err != nil {
		// Can't invert, defer
		u.unsolved = append(u.unsolved, Constraint{LHS: meta, RHS: solution})
		return nil
	}

	// Record solution
	u.solutions[meta.ID] = inverted

	// Re-check any deferred constraints
	u.recheckUnsolved()

	return nil
}

// occurs performs the occurs check: does meta ID occur in term?
func (u *Unifier) occurs(id int, t ast.Term) bool {
	if t == nil {
		return false
	}

	switch tt := t.(type) {
	case ast.Meta:
		if tt.ID == id {
			return true
		}
		for _, arg := range tt.Args {
			if u.occurs(id, arg) {
				return true
			}
		}
		return false

	case ast.Var, ast.Global, ast.Sort:
		return false

	case ast.Pi:
		return u.occurs(id, tt.A) || u.occurs(id, tt.B)

	case ast.Lam:
		return u.occurs(id, tt.Ann) || u.occurs(id, tt.Body)

	case ast.App:
		return u.occurs(id, tt.T) || u.occurs(id, tt.U)

	case ast.Sigma:
		return u.occurs(id, tt.A) || u.occurs(id, tt.B)

	case ast.Pair:
		return u.occurs(id, tt.Fst) || u.occurs(id, tt.Snd)

	case ast.Fst:
		return u.occurs(id, tt.P)

	case ast.Snd:
		return u.occurs(id, tt.P)

	case ast.Let:
		return u.occurs(id, tt.Ann) || u.occurs(id, tt.Val) || u.occurs(id, tt.Body)

	case ast.Id:
		return u.occurs(id, tt.A) || u.occurs(id, tt.X) || u.occurs(id, tt.Y)

	case ast.Refl:
		return u.occurs(id, tt.A) || u.occurs(id, tt.X)

	case ast.J:
		return u.occurs(id, tt.A) || u.occurs(id, tt.C) || u.occurs(id, tt.D) ||
			u.occurs(id, tt.X) || u.occurs(id, tt.Y) || u.occurs(id, tt.P)

	case ast.Path:
		return u.occurs(id, tt.A) || u.occurs(id, tt.X) || u.occurs(id, tt.Y)

	case ast.PathP:
		return u.occurs(id, tt.A) || u.occurs(id, tt.X) || u.occurs(id, tt.Y)

	case ast.PathLam:
		return u.occurs(id, tt.Body)

	case ast.PathApp:
		return u.occurs(id, tt.P) || u.occurs(id, tt.R)

	case ast.Transport:
		return u.occurs(id, tt.A) || u.occurs(id, tt.E)

	// Interval and faces - no subterms contain metas
	case ast.Interval, ast.I0, ast.I1, ast.IVar:
		return false

	case ast.FaceTop, ast.FaceBot, ast.FaceEq, ast.FaceAnd, ast.FaceOr:
		return false // Face types don't contain Terms

	case ast.Partial:
		return u.occurs(id, tt.A) // Phi is Face, not Term

	case ast.System:
		for _, b := range tt.Branches {
			if u.occurs(id, b.Term) { // Phi is Face
				return true
			}
		}
		return false

	case ast.Comp:
		return u.occurs(id, tt.A) || u.occurs(id, tt.Tube) || u.occurs(id, tt.Base)

	case ast.HComp:
		return u.occurs(id, tt.A) || u.occurs(id, tt.Tube) || u.occurs(id, tt.Base)

	case ast.Fill:
		return u.occurs(id, tt.A) || u.occurs(id, tt.Tube) || u.occurs(id, tt.Base)

	case ast.Glue:
		if u.occurs(id, tt.A) {
			return true
		}
		for _, b := range tt.System {
			if u.occurs(id, b.T) || u.occurs(id, b.Equiv) {
				return true
			}
		}
		return false

	case ast.GlueElem:
		if u.occurs(id, tt.Base) {
			return true
		}
		for _, b := range tt.System {
			if u.occurs(id, b.Term) {
				return true
			}
		}
		return false

	case ast.Unglue:
		return u.occurs(id, tt.Ty) || u.occurs(id, tt.G)

	case ast.UA:
		return u.occurs(id, tt.A) || u.occurs(id, tt.B) || u.occurs(id, tt.Equiv)

	case ast.UABeta:
		return u.occurs(id, tt.Equiv) || u.occurs(id, tt.Arg)

	// HIT terms - only HITApp implements isCoreTerm()
	case ast.HITApp:
		for _, arg := range tt.Args {
			if u.occurs(id, arg) {
				return true
			}
		}
		for _, arg := range tt.IArgs {
			if u.occurs(id, arg) {
				return true
			}
		}
		return false

	default:
		return false
	}
}

// isPattern checks if a metavariable application is a Miller pattern.
// A pattern is a metavariable applied to distinct bound variables.
func (u *Unifier) isPattern(meta ast.Meta) bool {
	seen := make(map[int]bool)
	for _, arg := range meta.Args {
		v, ok := arg.(ast.Var)
		if !ok {
			return false
		}
		if seen[v.Ix] {
			return false // Duplicate variable
		}
		seen[v.Ix] = true
	}
	return true
}

// invertPattern inverts a pattern to create a solution.
// If ?X x y = t where x,y are distinct variables, then X = λx.λy.t[x↦0,y↦1]
// Variables not in the pattern are shifted to account for the new lambdas.
func (u *Unifier) invertPattern(meta ast.Meta, solution ast.Term) (ast.Term, error) {
	numArgs := len(meta.Args)

	// Build a substitution from variable indices to their position in the spine
	// This maps the original context index to the lambda variable index
	patternVars := make(map[int]int)
	for i, arg := range meta.Args {
		v := arg.(ast.Var) // Already checked in isPattern
		// Variables at position i in the spine map to lambda binding (numArgs - 1 - i)
		patternVars[v.Ix] = numArgs - 1 - i
	}

	// Rename and shift variables in solution
	renamed := u.renameAndShiftVars(solution, patternVars, numArgs)

	// Wrap in lambdas
	result := renamed
	for range meta.Args {
		result = ast.Lam{Binder: "_", Body: result}
	}

	return result, nil
}

// checkScope verifies that a term only uses variables in the given scope.
func (u *Unifier) checkScope(t ast.Term, scope map[int]int) bool {
	if t == nil {
		return true
	}

	switch tt := t.(type) {
	case ast.Var:
		_, ok := scope[tt.Ix]
		return ok

	case ast.Meta:
		for _, arg := range tt.Args {
			if !u.checkScope(arg, scope) {
				return false
			}
		}
		return true

	case ast.Global, ast.Sort:
		return true

	case ast.Pi:
		if !u.checkScope(tt.A, scope) {
			return false
		}
		// Under binder, shift scope
		shiftedScope := u.shiftScope(scope)
		shiftedScope[0] = -1 // Mark the new binding
		return u.checkScope(tt.B, shiftedScope)

	case ast.Lam:
		shiftedScope := u.shiftScope(scope)
		shiftedScope[0] = -1
		return u.checkScope(tt.Body, shiftedScope)

	case ast.App:
		return u.checkScope(tt.T, scope) && u.checkScope(tt.U, scope)

	case ast.Sigma:
		if !u.checkScope(tt.A, scope) {
			return false
		}
		shiftedScope := u.shiftScope(scope)
		shiftedScope[0] = -1
		return u.checkScope(tt.B, shiftedScope)

	case ast.Pair:
		return u.checkScope(tt.Fst, scope) && u.checkScope(tt.Snd, scope)

	case ast.Fst:
		return u.checkScope(tt.P, scope)

	case ast.Snd:
		return u.checkScope(tt.P, scope)

	case ast.Let:
		if !u.checkScope(tt.Val, scope) {
			return false
		}
		shiftedScope := u.shiftScope(scope)
		shiftedScope[0] = -1
		return u.checkScope(tt.Body, shiftedScope)

	case ast.Id:
		return u.checkScope(tt.A, scope) && u.checkScope(tt.X, scope) && u.checkScope(tt.Y, scope)

	case ast.Refl:
		return u.checkScope(tt.A, scope) && u.checkScope(tt.X, scope)

	default:
		// Conservatively allow
		return true
	}
}

// shiftScope shifts all indices in a scope by 1.
func (u *Unifier) shiftScope(scope map[int]int) map[int]int {
	result := make(map[int]int)
	for k, v := range scope {
		result[k+1] = v + 1
	}
	return result
}

// renameAndShiftVars renames pattern variables and shifts non-pattern variables.
// patternVars maps pattern variable indices to their lambda binding indices.
// shift is the number of lambdas being added (to shift non-pattern variables).
func (u *Unifier) renameAndShiftVars(t ast.Term, patternVars map[int]int, shift int) ast.Term {
	return u.renameAndShiftVarsDepth(t, patternVars, shift, 0)
}

// renameAndShiftVarsDepth handles variable renaming with depth tracking for binders.
func (u *Unifier) renameAndShiftVarsDepth(t ast.Term, patternVars map[int]int, shift int, depth int) ast.Term {
	if t == nil {
		return nil
	}

	switch tt := t.(type) {
	case ast.Var:
		if tt.Ix < depth {
			// Variable bound within the solution term - keep as is
			return t
		}
		// Adjust index for outer context
		outerIx := tt.Ix - depth
		if newIx, ok := patternVars[outerIx]; ok {
			// This is a pattern variable - rename to lambda binding
			return ast.Var{Ix: newIx + depth}
		}
		// Not a pattern variable - shift by the number of lambdas
		return ast.Var{Ix: tt.Ix + shift}

	case ast.Meta:
		newArgs := make([]ast.Term, len(tt.Args))
		for i, arg := range tt.Args {
			newArgs[i] = u.renameAndShiftVarsDepth(arg, patternVars, shift, depth)
		}
		return ast.Meta{ID: tt.ID, Args: newArgs}

	case ast.Global, ast.Sort:
		return t

	case ast.Pi:
		newA := u.renameAndShiftVarsDepth(tt.A, patternVars, shift, depth)
		newB := u.renameAndShiftVarsDepth(tt.B, patternVars, shift, depth+1)
		return ast.Pi{Binder: tt.Binder, A: newA, B: newB}

	case ast.Lam:
		var newAnn ast.Term
		if tt.Ann != nil {
			newAnn = u.renameAndShiftVarsDepth(tt.Ann, patternVars, shift, depth)
		}
		newBody := u.renameAndShiftVarsDepth(tt.Body, patternVars, shift, depth+1)
		return ast.Lam{Binder: tt.Binder, Ann: newAnn, Body: newBody}

	case ast.App:
		return ast.App{
			T: u.renameAndShiftVarsDepth(tt.T, patternVars, shift, depth),
			U: u.renameAndShiftVarsDepth(tt.U, patternVars, shift, depth),
		}

	case ast.Sigma:
		newA := u.renameAndShiftVarsDepth(tt.A, patternVars, shift, depth)
		newB := u.renameAndShiftVarsDepth(tt.B, patternVars, shift, depth+1)
		return ast.Sigma{Binder: tt.Binder, A: newA, B: newB}

	case ast.Pair:
		return ast.Pair{
			Fst: u.renameAndShiftVarsDepth(tt.Fst, patternVars, shift, depth),
			Snd: u.renameAndShiftVarsDepth(tt.Snd, patternVars, shift, depth),
		}

	case ast.Fst:
		return ast.Fst{P: u.renameAndShiftVarsDepth(tt.P, patternVars, shift, depth)}

	case ast.Snd:
		return ast.Snd{P: u.renameAndShiftVarsDepth(tt.P, patternVars, shift, depth)}

	case ast.Let:
		var newAnn ast.Term
		if tt.Ann != nil {
			newAnn = u.renameAndShiftVarsDepth(tt.Ann, patternVars, shift, depth)
		}
		newVal := u.renameAndShiftVarsDepth(tt.Val, patternVars, shift, depth)
		newBody := u.renameAndShiftVarsDepth(tt.Body, patternVars, shift, depth+1)
		return ast.Let{Binder: tt.Binder, Ann: newAnn, Val: newVal, Body: newBody}

	case ast.Id:
		return ast.Id{
			A: u.renameAndShiftVarsDepth(tt.A, patternVars, shift, depth),
			X: u.renameAndShiftVarsDepth(tt.X, patternVars, shift, depth),
			Y: u.renameAndShiftVarsDepth(tt.Y, patternVars, shift, depth),
		}

	case ast.Refl:
		return ast.Refl{
			A: u.renameAndShiftVarsDepth(tt.A, patternVars, shift, depth),
			X: u.renameAndShiftVarsDepth(tt.X, patternVars, shift, depth),
		}

	default:
		return t
	}
}

// renameVars renames variables in a term according to the given mapping.
func (u *Unifier) renameVars(t ast.Term, renaming map[int]int) ast.Term {
	if t == nil {
		return nil
	}

	switch tt := t.(type) {
	case ast.Var:
		if newIx, ok := renaming[tt.Ix]; ok {
			return ast.Var{Ix: newIx}
		}
		return t

	case ast.Meta:
		newArgs := make([]ast.Term, len(tt.Args))
		for i, arg := range tt.Args {
			newArgs[i] = u.renameVars(arg, renaming)
		}
		return ast.Meta{ID: tt.ID, Args: newArgs}

	case ast.Global, ast.Sort:
		return t

	case ast.Pi:
		newA := u.renameVars(tt.A, renaming)
		shiftedRenaming := u.shiftRenaming(renaming)
		newB := u.renameVars(tt.B, shiftedRenaming)
		return ast.Pi{Binder: tt.Binder, A: newA, B: newB}

	case ast.Lam:
		var newAnn ast.Term
		if tt.Ann != nil {
			newAnn = u.renameVars(tt.Ann, renaming)
		}
		shiftedRenaming := u.shiftRenaming(renaming)
		newBody := u.renameVars(tt.Body, shiftedRenaming)
		return ast.Lam{Binder: tt.Binder, Ann: newAnn, Body: newBody}

	case ast.App:
		return ast.App{
			T: u.renameVars(tt.T, renaming),
			U: u.renameVars(tt.U, renaming),
		}

	case ast.Sigma:
		newA := u.renameVars(tt.A, renaming)
		shiftedRenaming := u.shiftRenaming(renaming)
		newB := u.renameVars(tt.B, shiftedRenaming)
		return ast.Sigma{Binder: tt.Binder, A: newA, B: newB}

	case ast.Pair:
		return ast.Pair{
			Fst: u.renameVars(tt.Fst, renaming),
			Snd: u.renameVars(tt.Snd, renaming),
		}

	case ast.Fst:
		return ast.Fst{P: u.renameVars(tt.P, renaming)}

	case ast.Snd:
		return ast.Snd{P: u.renameVars(tt.P, renaming)}

	case ast.Let:
		var newAnn ast.Term
		if tt.Ann != nil {
			newAnn = u.renameVars(tt.Ann, renaming)
		}
		newVal := u.renameVars(tt.Val, renaming)
		shiftedRenaming := u.shiftRenaming(renaming)
		newBody := u.renameVars(tt.Body, shiftedRenaming)
		return ast.Let{Binder: tt.Binder, Ann: newAnn, Val: newVal, Body: newBody}

	case ast.Id:
		return ast.Id{
			A: u.renameVars(tt.A, renaming),
			X: u.renameVars(tt.X, renaming),
			Y: u.renameVars(tt.Y, renaming),
		}

	case ast.Refl:
		return ast.Refl{
			A: u.renameVars(tt.A, renaming),
			X: u.renameVars(tt.X, renaming),
		}

	default:
		return t
	}
}

// shiftRenaming shifts all indices in a renaming by 1.
func (u *Unifier) shiftRenaming(renaming map[int]int) map[int]int {
	result := make(map[int]int)
	for k, v := range renaming {
		result[k+1] = v + 1
	}
	result[0] = 0 // The new binder maps to itself
	return result
}

// recheckUnsolved re-processes unsolved constraints after finding a solution.
func (u *Unifier) recheckUnsolved() {
	// Move unsolved back to worklist
	worklist := u.unsolved
	u.unsolved = nil
	u.worklist = append(u.worklist, worklist...)
}

// fail creates an error.
func (u *Unifier) fail(format string, args ...any) *UnifyError {
	return &UnifyError{Message: fmt.Sprintf(format, args...)}
}

// Unify is a convenience function that unifies two terms.
func Unify(lhs, rhs ast.Term) UnifyResult {
	u := NewUnifier()
	u.AddConstraint(lhs, rhs)
	return u.Solve()
}

// UnifyAll unifies multiple pairs of terms.
func UnifyAll(constraints []Constraint) UnifyResult {
	u := NewUnifier()
	for _, c := range constraints {
		u.AddConstraint(c.LHS, c.RHS)
	}
	return u.Solve()
}

// Zonk applies all solutions to a term, fully substituting metavariables.
func Zonk(solutions map[int]ast.Term, t ast.Term) ast.Term {
	u := &Unifier{solutions: solutions}
	return u.zonkTerm(t)
}

// ZonkFull applies all solutions and errors if any metavariables remain.
func ZonkFull(solutions map[int]ast.Term, t ast.Term) (ast.Term, error) {
	result := Zonk(solutions, t)
	if hasMeta(result) {
		return result, fmt.Errorf("unsolved metavariables remain")
	}
	return result, nil
}

// hasMeta checks if a term contains any metavariables.
func hasMeta(t ast.Term) bool {
	if t == nil {
		return false
	}

	switch tt := t.(type) {
	case ast.Meta:
		return true

	case ast.Var, ast.Global, ast.Sort:
		return false

	case ast.Pi:
		return hasMeta(tt.A) || hasMeta(tt.B)

	case ast.Lam:
		return hasMeta(tt.Ann) || hasMeta(tt.Body)

	case ast.App:
		return hasMeta(tt.T) || hasMeta(tt.U)

	case ast.Sigma:
		return hasMeta(tt.A) || hasMeta(tt.B)

	case ast.Pair:
		return hasMeta(tt.Fst) || hasMeta(tt.Snd)

	case ast.Fst:
		return hasMeta(tt.P)

	case ast.Snd:
		return hasMeta(tt.P)

	case ast.Let:
		return hasMeta(tt.Ann) || hasMeta(tt.Val) || hasMeta(tt.Body)

	case ast.Id:
		return hasMeta(tt.A) || hasMeta(tt.X) || hasMeta(tt.Y)

	case ast.Refl:
		return hasMeta(tt.A) || hasMeta(tt.X)

	case ast.J:
		return hasMeta(tt.A) || hasMeta(tt.C) || hasMeta(tt.D) ||
			hasMeta(tt.X) || hasMeta(tt.Y) || hasMeta(tt.P)

	// Path types
	case ast.Path:
		return hasMeta(tt.A) || hasMeta(tt.X) || hasMeta(tt.Y)

	case ast.PathP:
		return hasMeta(tt.A) || hasMeta(tt.X) || hasMeta(tt.Y)

	case ast.PathLam:
		return hasMeta(tt.Body)

	case ast.PathApp:
		return hasMeta(tt.P) || hasMeta(tt.R)

	case ast.Transport:
		return hasMeta(tt.A) || hasMeta(tt.E)

	// Interval and faces - no subterms contain metas
	case ast.Interval, ast.I0, ast.I1, ast.IVar:
		return false

	case ast.FaceTop, ast.FaceBot, ast.FaceEq, ast.FaceAnd, ast.FaceOr:
		return false // Face types don't contain Terms

	case ast.Partial:
		return hasMeta(tt.A) // Phi is Face, not Term

	case ast.System:
		for _, b := range tt.Branches {
			if hasMeta(b.Term) { // Phi is Face
				return true
			}
		}
		return false

	case ast.Comp:
		return hasMeta(tt.A) || hasMeta(tt.Tube) || hasMeta(tt.Base)

	case ast.HComp:
		return hasMeta(tt.A) || hasMeta(tt.Tube) || hasMeta(tt.Base)

	case ast.Fill:
		return hasMeta(tt.A) || hasMeta(tt.Tube) || hasMeta(tt.Base)

	case ast.Glue:
		if hasMeta(tt.A) {
			return true
		}
		for _, b := range tt.System {
			if hasMeta(b.T) || hasMeta(b.Equiv) {
				return true
			}
		}
		return false

	case ast.GlueElem:
		if hasMeta(tt.Base) {
			return true
		}
		for _, b := range tt.System {
			if hasMeta(b.Term) {
				return true
			}
		}
		return false

	case ast.Unglue:
		return hasMeta(tt.Ty) || hasMeta(tt.G)

	case ast.UA:
		return hasMeta(tt.A) || hasMeta(tt.B) || hasMeta(tt.Equiv)

	case ast.UABeta:
		return hasMeta(tt.Equiv) || hasMeta(tt.Arg)

	// HIT terms - only HITApp implements isCoreTerm()
	case ast.HITApp:
		for _, arg := range tt.Args {
			if hasMeta(arg) {
				return true
			}
		}
		for _, arg := range tt.IArgs {
			if hasMeta(arg) {
				return true
			}
		}
		return false

	default:
		return false
	}
}

// Ensure subst package is used (for future use in pattern inversion)
var _ = subst.Shift
