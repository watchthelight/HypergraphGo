package subst

import (
	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// IShift shifts all free interval variables >= cutoff by d positions.
// This operates in the interval variable namespace, separate from term variables.
func IShift(d, cutoff int, t ast.Term) ast.Term {
	if t == nil {
		return nil
	}
	switch tm := t.(type) {
	// Interval terms
	case ast.IVar:
		if tm.Ix >= cutoff {
			return ast.IVar{Ix: tm.Ix + d}
		}
		return tm
	case ast.I0, ast.I1, ast.Interval:
		return tm // Endpoints and interval type are constants

	// Path types - PathP and Transport bind interval variables
	case ast.Path:
		return ast.Path{
			A: IShift(d, cutoff, tm.A),
			X: IShift(d, cutoff, tm.X),
			Y: IShift(d, cutoff, tm.Y),
		}
	case ast.PathP:
		// A binds an interval variable
		return ast.PathP{
			A: IShift(d, cutoff+1, tm.A),
			X: IShift(d, cutoff, tm.X),
			Y: IShift(d, cutoff, tm.Y),
		}
	case ast.PathLam:
		// PathLam binds an interval variable
		return ast.PathLam{
			Binder: tm.Binder,
			Body:   IShift(d, cutoff+1, tm.Body),
		}
	case ast.PathApp:
		return ast.PathApp{
			P: IShift(d, cutoff, tm.P),
			R: IShift(d, cutoff, tm.R),
		}
	case ast.Transport:
		// A binds an interval variable
		return ast.Transport{
			A: IShift(d, cutoff+1, tm.A),
			E: IShift(d, cutoff, tm.E),
		}

	// Face formulas
	case ast.FaceTop, ast.FaceBot:
		return tm
	case ast.FaceEq:
		if tm.IVar >= cutoff {
			return ast.FaceEq{IVar: tm.IVar + d, IsOne: tm.IsOne}
		}
		return tm
	case ast.FaceAnd:
		return ast.FaceAnd{
			Left:  IShiftFace(d, cutoff, tm.Left),
			Right: IShiftFace(d, cutoff, tm.Right),
		}
	case ast.FaceOr:
		return ast.FaceOr{
			Left:  IShiftFace(d, cutoff, tm.Left),
			Right: IShiftFace(d, cutoff, tm.Right),
		}

	// Partial types
	case ast.Partial:
		return ast.Partial{
			Phi: IShiftFace(d, cutoff, tm.Phi),
			A:   IShift(d, cutoff, tm.A),
		}
	case ast.System:
		branches := make([]ast.SystemBranch, len(tm.Branches))
		for i, br := range tm.Branches {
			branches[i] = ast.SystemBranch{
				Phi:  IShiftFace(d, cutoff, br.Phi),
				Term: IShift(d, cutoff, br.Term),
			}
		}
		return ast.System{Branches: branches}

	// Composition operations
	case ast.Comp:
		// A and Tube bind an interval variable
		return ast.Comp{
			IBinder: tm.IBinder,
			A:       IShift(d, cutoff+1, tm.A),
			Phi:     IShiftFace(d, cutoff+1, tm.Phi),
			Tube:    IShift(d, cutoff+1, tm.Tube),
			Base:    IShift(d, cutoff, tm.Base),
		}
	case ast.HComp:
		// Tube binds an interval variable, A does not
		return ast.HComp{
			A:    IShift(d, cutoff, tm.A),
			Phi:  IShiftFace(d, cutoff+1, tm.Phi),
			Tube: IShift(d, cutoff+1, tm.Tube),
			Base: IShift(d, cutoff, tm.Base),
		}
	case ast.Fill:
		// A and Tube bind an interval variable
		return ast.Fill{
			IBinder: tm.IBinder,
			A:       IShift(d, cutoff+1, tm.A),
			Phi:     IShiftFace(d, cutoff+1, tm.Phi),
			Tube:    IShift(d, cutoff+1, tm.Tube),
			Base:    IShift(d, cutoff, tm.Base),
		}

	// Glue types - no interval binders
	case ast.Glue:
		branches := make([]ast.GlueBranch, len(tm.System))
		for i, br := range tm.System {
			branches[i] = ast.GlueBranch{
				Phi:   IShiftFace(d, cutoff, br.Phi),
				T:     IShift(d, cutoff, br.T),
				Equiv: IShift(d, cutoff, br.Equiv),
			}
		}
		return ast.Glue{
			A:      IShift(d, cutoff, tm.A),
			System: branches,
		}
	case ast.GlueElem:
		branches := make([]ast.GlueElemBranch, len(tm.System))
		for i, br := range tm.System {
			branches[i] = ast.GlueElemBranch{
				Phi:  IShiftFace(d, cutoff, br.Phi),
				Term: IShift(d, cutoff, br.Term),
			}
		}
		return ast.GlueElem{
			System: branches,
			Base:   IShift(d, cutoff, tm.Base),
		}
	case ast.Unglue:
		return ast.Unglue{
			Ty: IShift(d, cutoff, tm.Ty),
			G:  IShift(d, cutoff, tm.G),
		}

	// Univalence - no interval binders
	case ast.UA:
		return ast.UA{
			A:     IShift(d, cutoff, tm.A),
			B:     IShift(d, cutoff, tm.B),
			Equiv: IShift(d, cutoff, tm.Equiv),
		}
	case ast.UABeta:
		return ast.UABeta{
			Equiv: IShift(d, cutoff, tm.Equiv),
			Arg:   IShift(d, cutoff, tm.Arg),
		}
	// Higher Inductive Types - no interval binders in HITApp itself
	case ast.HITApp:
		args := make([]ast.Term, len(tm.Args))
		for i, arg := range tm.Args {
			args[i] = IShift(d, cutoff, arg)
		}
		iargs := make([]ast.Term, len(tm.IArgs))
		for i, iarg := range tm.IArgs {
			iargs[i] = IShift(d, cutoff, iarg)
		}
		return ast.HITApp{
			HITName: tm.HITName,
			Ctor:    tm.Ctor,
			Args:    args,
			IArgs:   iargs,
		}

	// Standard terms - recurse without changing cutoff (no interval binders)
	case ast.Var, ast.Sort, ast.Global:
		return tm
	case ast.Pi:
		return ast.Pi{
			Binder: tm.Binder,
			A:      IShift(d, cutoff, tm.A),
			B:      IShift(d, cutoff, tm.B), // Pi doesn't bind interval vars
		}
	case ast.Lam:
		return ast.Lam{
			Binder: tm.Binder,
			Ann:    IShift(d, cutoff, tm.Ann),
			Body:   IShift(d, cutoff, tm.Body),
		}
	case ast.App:
		return ast.App{
			T: IShift(d, cutoff, tm.T),
			U: IShift(d, cutoff, tm.U),
		}
	case ast.Sigma:
		return ast.Sigma{
			Binder: tm.Binder,
			A:      IShift(d, cutoff, tm.A),
			B:      IShift(d, cutoff, tm.B),
		}
	case ast.Pair:
		return ast.Pair{
			Fst: IShift(d, cutoff, tm.Fst),
			Snd: IShift(d, cutoff, tm.Snd),
		}
	case ast.Fst:
		return ast.Fst{P: IShift(d, cutoff, tm.P)}
	case ast.Snd:
		return ast.Snd{P: IShift(d, cutoff, tm.P)}
	case ast.Let:
		return ast.Let{
			Binder: tm.Binder,
			Ann:    IShift(d, cutoff, tm.Ann),
			Val:    IShift(d, cutoff, tm.Val),
			Body:   IShift(d, cutoff, tm.Body),
		}
	case ast.Id:
		return ast.Id{
			A: IShift(d, cutoff, tm.A),
			X: IShift(d, cutoff, tm.X),
			Y: IShift(d, cutoff, tm.Y),
		}
	case ast.Refl:
		return ast.Refl{
			A: IShift(d, cutoff, tm.A),
			X: IShift(d, cutoff, tm.X),
		}
	case ast.J:
		return ast.J{
			A: IShift(d, cutoff, tm.A),
			C: IShift(d, cutoff, tm.C),
			D: IShift(d, cutoff, tm.D),
			X: IShift(d, cutoff, tm.X),
			Y: IShift(d, cutoff, tm.Y),
			P: IShift(d, cutoff, tm.P),
		}
	default:
		return t
	}
}

// ISubst substitutes interval term s for interval variable j in t.
// This operates in the interval variable namespace.
func ISubst(j int, s ast.Term, t ast.Term) ast.Term {
	if t == nil {
		return nil
	}
	switch tm := t.(type) {
	// Interval terms
	case ast.IVar:
		if tm.Ix == j {
			return s
		} else if tm.Ix > j {
			return ast.IVar{Ix: tm.Ix - 1}
		}
		return tm
	case ast.I0, ast.I1, ast.Interval:
		return tm

	// Path types - PathP, PathLam, and Transport bind interval variables
	case ast.Path:
		return ast.Path{
			A: ISubst(j, s, tm.A),
			X: ISubst(j, s, tm.X),
			Y: ISubst(j, s, tm.Y),
		}
	case ast.PathP:
		return ast.PathP{
			A: ISubst(j+1, IShift(1, 0, s), tm.A),
			X: ISubst(j, s, tm.X),
			Y: ISubst(j, s, tm.Y),
		}
	case ast.PathLam:
		return ast.PathLam{
			Binder: tm.Binder,
			Body:   ISubst(j+1, IShift(1, 0, s), tm.Body),
		}
	case ast.PathApp:
		return ast.PathApp{
			P: ISubst(j, s, tm.P),
			R: ISubst(j, s, tm.R),
		}
	case ast.Transport:
		return ast.Transport{
			A: ISubst(j+1, IShift(1, 0, s), tm.A),
			E: ISubst(j, s, tm.E),
		}

	// Face formulas
	case ast.FaceTop, ast.FaceBot:
		return tm
	case ast.FaceEq:
		if tm.IVar == j {
			// Substituting this variable - need to check endpoint
			switch iv := s.(type) {
			case ast.I0:
				if tm.IsOne {
					return ast.FaceBot{} // (i0 = 1) is false
				}
				return ast.FaceTop{} // (i0 = 0) is true
			case ast.I1:
				if tm.IsOne {
					return ast.FaceTop{} // (i1 = 1) is true
				}
				return ast.FaceBot{} // (i1 = 0) is false
			case ast.IVar:
				return ast.FaceEq{IVar: iv.Ix, IsOne: tm.IsOne}
			default:
				return tm
			}
		} else if tm.IVar > j {
			return ast.FaceEq{IVar: tm.IVar - 1, IsOne: tm.IsOne}
		}
		return tm
	case ast.FaceAnd:
		result := simplifyFaceAndAST(
			ISubstFace(j, s, tm.Left),
			ISubstFace(j, s, tm.Right),
		)
		return faceToTerm(result)
	case ast.FaceOr:
		result := simplifyFaceOrAST(
			ISubstFace(j, s, tm.Left),
			ISubstFace(j, s, tm.Right),
		)
		return faceToTerm(result)

	// Partial types
	case ast.Partial:
		return ast.Partial{
			Phi: ISubstFace(j, s, tm.Phi),
			A:   ISubst(j, s, tm.A),
		}
	case ast.System:
		branches := make([]ast.SystemBranch, len(tm.Branches))
		for i, br := range tm.Branches {
			branches[i] = ast.SystemBranch{
				Phi:  ISubstFace(j, s, br.Phi),
				Term: ISubst(j, s, br.Term),
			}
		}
		return ast.System{Branches: branches}

	// Composition operations
	case ast.Comp:
		// A and Tube bind an interval variable
		return ast.Comp{
			IBinder: tm.IBinder,
			A:       ISubst(j+1, IShift(1, 0, s), tm.A),
			Phi:     ISubstFace(j+1, IShift(1, 0, s), tm.Phi),
			Tube:    ISubst(j+1, IShift(1, 0, s), tm.Tube),
			Base:    ISubst(j, s, tm.Base),
		}
	case ast.HComp:
		// Tube binds an interval variable, A does not
		return ast.HComp{
			A:    ISubst(j, s, tm.A),
			Phi:  ISubstFace(j+1, IShift(1, 0, s), tm.Phi),
			Tube: ISubst(j+1, IShift(1, 0, s), tm.Tube),
			Base: ISubst(j, s, tm.Base),
		}
	case ast.Fill:
		// A and Tube bind an interval variable
		return ast.Fill{
			IBinder: tm.IBinder,
			A:       ISubst(j+1, IShift(1, 0, s), tm.A),
			Phi:     ISubstFace(j+1, IShift(1, 0, s), tm.Phi),
			Tube:    ISubst(j+1, IShift(1, 0, s), tm.Tube),
			Base:    ISubst(j, s, tm.Base),
		}

	// Glue types - no interval binders
	case ast.Glue:
		branches := make([]ast.GlueBranch, len(tm.System))
		for i, br := range tm.System {
			branches[i] = ast.GlueBranch{
				Phi:   ISubstFace(j, s, br.Phi),
				T:     ISubst(j, s, br.T),
				Equiv: ISubst(j, s, br.Equiv),
			}
		}
		return ast.Glue{
			A:      ISubst(j, s, tm.A),
			System: branches,
		}
	case ast.GlueElem:
		branches := make([]ast.GlueElemBranch, len(tm.System))
		for i, br := range tm.System {
			branches[i] = ast.GlueElemBranch{
				Phi:  ISubstFace(j, s, br.Phi),
				Term: ISubst(j, s, br.Term),
			}
		}
		return ast.GlueElem{
			System: branches,
			Base:   ISubst(j, s, tm.Base),
		}
	case ast.Unglue:
		return ast.Unglue{
			Ty: ISubst(j, s, tm.Ty),
			G:  ISubst(j, s, tm.G),
		}

	// Univalence - no interval binders
	case ast.UA:
		return ast.UA{
			A:     ISubst(j, s, tm.A),
			B:     ISubst(j, s, tm.B),
			Equiv: ISubst(j, s, tm.Equiv),
		}
	case ast.UABeta:
		return ast.UABeta{
			Equiv: ISubst(j, s, tm.Equiv),
			Arg:   ISubst(j, s, tm.Arg),
		}
	// Higher Inductive Types - substitute in args and interval args
	case ast.HITApp:
		args := make([]ast.Term, len(tm.Args))
		for i, arg := range tm.Args {
			args[i] = ISubst(j, s, arg)
		}
		iargs := make([]ast.Term, len(tm.IArgs))
		for i, iarg := range tm.IArgs {
			iargs[i] = ISubst(j, s, iarg)
		}
		return ast.HITApp{
			HITName: tm.HITName,
			Ctor:    tm.Ctor,
			Args:    args,
			IArgs:   iargs,
		}

	// Standard terms - recurse without changing j (no interval binders)
	case ast.Var, ast.Sort, ast.Global:
		return tm
	case ast.Pi:
		return ast.Pi{
			Binder: tm.Binder,
			A:      ISubst(j, s, tm.A),
			B:      ISubst(j, s, tm.B),
		}
	case ast.Lam:
		return ast.Lam{
			Binder: tm.Binder,
			Ann:    ISubst(j, s, tm.Ann),
			Body:   ISubst(j, s, tm.Body),
		}
	case ast.App:
		return ast.App{
			T: ISubst(j, s, tm.T),
			U: ISubst(j, s, tm.U),
		}
	case ast.Sigma:
		return ast.Sigma{
			Binder: tm.Binder,
			A:      ISubst(j, s, tm.A),
			B:      ISubst(j, s, tm.B),
		}
	case ast.Pair:
		return ast.Pair{
			Fst: ISubst(j, s, tm.Fst),
			Snd: ISubst(j, s, tm.Snd),
		}
	case ast.Fst:
		return ast.Fst{P: ISubst(j, s, tm.P)}
	case ast.Snd:
		return ast.Snd{P: ISubst(j, s, tm.P)}
	case ast.Let:
		return ast.Let{
			Binder: tm.Binder,
			Ann:    ISubst(j, s, tm.Ann),
			Val:    ISubst(j, s, tm.Val),
			Body:   ISubst(j, s, tm.Body),
		}
	case ast.Id:
		return ast.Id{
			A: ISubst(j, s, tm.A),
			X: ISubst(j, s, tm.X),
			Y: ISubst(j, s, tm.Y),
		}
	case ast.Refl:
		return ast.Refl{
			A: ISubst(j, s, tm.A),
			X: ISubst(j, s, tm.X),
		}
	case ast.J:
		return ast.J{
			A: ISubst(j, s, tm.A),
			C: ISubst(j, s, tm.C),
			D: ISubst(j, s, tm.D),
			X: ISubst(j, s, tm.X),
			Y: ISubst(j, s, tm.Y),
			P: ISubst(j, s, tm.P),
		}
	default:
		return t
	}
}

// shiftExtension extends Shift to handle cubical term types.
// Called from the Shift function when cubical is enabled.
func shiftExtension(d, cutoff int, t ast.Term) (ast.Term, bool) {
	switch tm := t.(type) {
	case ast.Interval, ast.I0, ast.I1, ast.IVar:
		return tm, true // Interval terms have no term variables
	case ast.Path:
		return ast.Path{
			A: Shift(d, cutoff, tm.A),
			X: Shift(d, cutoff, tm.X),
			Y: Shift(d, cutoff, tm.Y),
		}, true
	case ast.PathP:
		// PathP doesn't bind term variables, only interval variables
		return ast.PathP{
			A: Shift(d, cutoff, tm.A),
			X: Shift(d, cutoff, tm.X),
			Y: Shift(d, cutoff, tm.Y),
		}, true
	case ast.PathLam:
		// PathLam doesn't bind term variables
		return ast.PathLam{
			Binder: tm.Binder,
			Body:   Shift(d, cutoff, tm.Body),
		}, true
	case ast.PathApp:
		return ast.PathApp{
			P: Shift(d, cutoff, tm.P),
			R: Shift(d, cutoff, tm.R),
		}, true
	case ast.Transport:
		return ast.Transport{
			A: Shift(d, cutoff, tm.A),
			E: Shift(d, cutoff, tm.E),
		}, true
	// Face formulas - no term variables
	case ast.FaceTop, ast.FaceBot, ast.FaceEq:
		return tm, true
	case ast.FaceAnd:
		return ast.FaceAnd{
			Left:  ShiftFace(d, cutoff, tm.Left),
			Right: ShiftFace(d, cutoff, tm.Right),
		}, true
	case ast.FaceOr:
		return ast.FaceOr{
			Left:  ShiftFace(d, cutoff, tm.Left),
			Right: ShiftFace(d, cutoff, tm.Right),
		}, true
	// Partial types
	case ast.Partial:
		return ast.Partial{
			Phi: ShiftFace(d, cutoff, tm.Phi),
			A:   Shift(d, cutoff, tm.A),
		}, true
	case ast.System:
		branches := make([]ast.SystemBranch, len(tm.Branches))
		for i, br := range tm.Branches {
			branches[i] = ast.SystemBranch{
				Phi:  ShiftFace(d, cutoff, br.Phi),
				Term: Shift(d, cutoff, br.Term),
			}
		}
		return ast.System{Branches: branches}, true
	// Composition operations - no term variable binders
	case ast.Comp:
		return ast.Comp{
			IBinder: tm.IBinder,
			A:       Shift(d, cutoff, tm.A),
			Phi:     ShiftFace(d, cutoff, tm.Phi),
			Tube:    Shift(d, cutoff, tm.Tube),
			Base:    Shift(d, cutoff, tm.Base),
		}, true
	case ast.HComp:
		return ast.HComp{
			A:    Shift(d, cutoff, tm.A),
			Phi:  ShiftFace(d, cutoff, tm.Phi),
			Tube: Shift(d, cutoff, tm.Tube),
			Base: Shift(d, cutoff, tm.Base),
		}, true
	case ast.Fill:
		return ast.Fill{
			IBinder: tm.IBinder,
			A:       Shift(d, cutoff, tm.A),
			Phi:     ShiftFace(d, cutoff, tm.Phi),
			Tube:    Shift(d, cutoff, tm.Tube),
			Base:    Shift(d, cutoff, tm.Base),
		}, true
	// Glue types - no term variable binders
	case ast.Glue:
		branches := make([]ast.GlueBranch, len(tm.System))
		for i, br := range tm.System {
			branches[i] = ast.GlueBranch{
				Phi:   ShiftFace(d, cutoff, br.Phi),
				T:     Shift(d, cutoff, br.T),
				Equiv: Shift(d, cutoff, br.Equiv),
			}
		}
		return ast.Glue{
			A:      Shift(d, cutoff, tm.A),
			System: branches,
		}, true
	case ast.GlueElem:
		branches := make([]ast.GlueElemBranch, len(tm.System))
		for i, br := range tm.System {
			branches[i] = ast.GlueElemBranch{
				Phi:  ShiftFace(d, cutoff, br.Phi),
				Term: Shift(d, cutoff, br.Term),
			}
		}
		return ast.GlueElem{
			System: branches,
			Base:   Shift(d, cutoff, tm.Base),
		}, true
	case ast.Unglue:
		return ast.Unglue{
			Ty: Shift(d, cutoff, tm.Ty),
			G:  Shift(d, cutoff, tm.G),
		}, true
	// Univalence - no term variable binders
	case ast.UA:
		return ast.UA{
			A:     Shift(d, cutoff, tm.A),
			B:     Shift(d, cutoff, tm.B),
			Equiv: Shift(d, cutoff, tm.Equiv),
		}, true
	case ast.UABeta:
		return ast.UABeta{
			Equiv: Shift(d, cutoff, tm.Equiv),
			Arg:   Shift(d, cutoff, tm.Arg),
		}, true
	// Higher Inductive Types - no term variable binders in HITApp
	case ast.HITApp:
		args := make([]ast.Term, len(tm.Args))
		for i, arg := range tm.Args {
			args[i] = Shift(d, cutoff, arg)
		}
		iargs := make([]ast.Term, len(tm.IArgs))
		for i, iarg := range tm.IArgs {
			iargs[i] = Shift(d, cutoff, iarg)
		}
		return ast.HITApp{
			HITName: tm.HITName,
			Ctor:    tm.Ctor,
			Args:    args,
			IArgs:   iargs,
		}, true
	default:
		return nil, false
	}
}

// substExtension extends Subst to handle cubical term types.
// Called from the Subst function when cubical is enabled.
func substExtension(j int, s ast.Term, t ast.Term) (ast.Term, bool) {
	switch tm := t.(type) {
	case ast.Interval, ast.I0, ast.I1, ast.IVar:
		return tm, true // Interval terms have no term variables
	case ast.Path:
		return ast.Path{
			A: Subst(j, s, tm.A),
			X: Subst(j, s, tm.X),
			Y: Subst(j, s, tm.Y),
		}, true
	case ast.PathP:
		return ast.PathP{
			A: Subst(j, s, tm.A),
			X: Subst(j, s, tm.X),
			Y: Subst(j, s, tm.Y),
		}, true
	case ast.PathLam:
		return ast.PathLam{
			Binder: tm.Binder,
			Body:   Subst(j, s, tm.Body),
		}, true
	case ast.PathApp:
		return ast.PathApp{
			P: Subst(j, s, tm.P),
			R: Subst(j, s, tm.R),
		}, true
	case ast.Transport:
		return ast.Transport{
			A: Subst(j, s, tm.A),
			E: Subst(j, s, tm.E),
		}, true
	// Face formulas - no term variables
	case ast.FaceTop, ast.FaceBot, ast.FaceEq:
		return tm, true
	case ast.FaceAnd:
		return ast.FaceAnd{
			Left:  SubstFace(j, s, tm.Left),
			Right: SubstFace(j, s, tm.Right),
		}, true
	case ast.FaceOr:
		return ast.FaceOr{
			Left:  SubstFace(j, s, tm.Left),
			Right: SubstFace(j, s, tm.Right),
		}, true
	// Partial types
	case ast.Partial:
		return ast.Partial{
			Phi: SubstFace(j, s, tm.Phi),
			A:   Subst(j, s, tm.A),
		}, true
	case ast.System:
		branches := make([]ast.SystemBranch, len(tm.Branches))
		for i, br := range tm.Branches {
			branches[i] = ast.SystemBranch{
				Phi:  SubstFace(j, s, br.Phi),
				Term: Subst(j, s, br.Term),
			}
		}
		return ast.System{Branches: branches}, true
	// Composition operations - no term variable binders
	case ast.Comp:
		return ast.Comp{
			IBinder: tm.IBinder,
			A:       Subst(j, s, tm.A),
			Phi:     SubstFace(j, s, tm.Phi),
			Tube:    Subst(j, s, tm.Tube),
			Base:    Subst(j, s, tm.Base),
		}, true
	case ast.HComp:
		return ast.HComp{
			A:    Subst(j, s, tm.A),
			Phi:  SubstFace(j, s, tm.Phi),
			Tube: Subst(j, s, tm.Tube),
			Base: Subst(j, s, tm.Base),
		}, true
	case ast.Fill:
		return ast.Fill{
			IBinder: tm.IBinder,
			A:       Subst(j, s, tm.A),
			Phi:     SubstFace(j, s, tm.Phi),
			Tube:    Subst(j, s, tm.Tube),
			Base:    Subst(j, s, tm.Base),
		}, true
	// Glue types - no term variable binders
	case ast.Glue:
		branches := make([]ast.GlueBranch, len(tm.System))
		for i, br := range tm.System {
			branches[i] = ast.GlueBranch{
				Phi:   SubstFace(j, s, br.Phi),
				T:     Subst(j, s, br.T),
				Equiv: Subst(j, s, br.Equiv),
			}
		}
		return ast.Glue{
			A:      Subst(j, s, tm.A),
			System: branches,
		}, true
	case ast.GlueElem:
		branches := make([]ast.GlueElemBranch, len(tm.System))
		for i, br := range tm.System {
			branches[i] = ast.GlueElemBranch{
				Phi:  SubstFace(j, s, br.Phi),
				Term: Subst(j, s, br.Term),
			}
		}
		return ast.GlueElem{
			System: branches,
			Base:   Subst(j, s, tm.Base),
		}, true
	case ast.Unglue:
		return ast.Unglue{
			Ty: Subst(j, s, tm.Ty),
			G:  Subst(j, s, tm.G),
		}, true
	// Univalence - no term variable binders
	case ast.UA:
		return ast.UA{
			A:     Subst(j, s, tm.A),
			B:     Subst(j, s, tm.B),
			Equiv: Subst(j, s, tm.Equiv),
		}, true
	case ast.UABeta:
		return ast.UABeta{
			Equiv: Subst(j, s, tm.Equiv),
			Arg:   Subst(j, s, tm.Arg),
		}, true
	// Higher Inductive Types - no term variable binders in HITApp
	case ast.HITApp:
		args := make([]ast.Term, len(tm.Args))
		for i, arg := range tm.Args {
			args[i] = Subst(j, s, arg)
		}
		iargs := make([]ast.Term, len(tm.IArgs))
		for i, iarg := range tm.IArgs {
			iargs[i] = Subst(j, s, iarg)
		}
		return ast.HITApp{
			HITName: tm.HITName,
			Ctor:    tm.Ctor,
			Args:    args,
			IArgs:   iargs,
		}, true
	default:
		return nil, false
	}
}

// --- Face Formula Helpers ---

// faceToTerm converts a Face to a Term (all Face types implement Term).
func faceToTerm(f ast.Face) ast.Term {
	switch ft := f.(type) {
	case ast.FaceTop:
		return ft
	case ast.FaceBot:
		return ft
	case ast.FaceEq:
		return ft
	case ast.FaceAnd:
		return ft
	case ast.FaceOr:
		return ft
	default:
		// Should not happen for valid faces
		return ast.FaceBot{}
	}
}

// IShiftFace shifts interval variables in a face formula.
func IShiftFace(d, cutoff int, f ast.Face) ast.Face {
	if f == nil {
		return nil
	}
	switch fm := f.(type) {
	case ast.FaceTop, ast.FaceBot:
		return fm
	case ast.FaceEq:
		if fm.IVar >= cutoff {
			return ast.FaceEq{IVar: fm.IVar + d, IsOne: fm.IsOne}
		}
		return fm
	case ast.FaceAnd:
		return ast.FaceAnd{
			Left:  IShiftFace(d, cutoff, fm.Left),
			Right: IShiftFace(d, cutoff, fm.Right),
		}
	case ast.FaceOr:
		return ast.FaceOr{
			Left:  IShiftFace(d, cutoff, fm.Left),
			Right: IShiftFace(d, cutoff, fm.Right),
		}
	default:
		return f
	}
}

// ISubstFace substitutes an interval term for an interval variable in a face formula.
func ISubstFace(j int, s ast.Term, f ast.Face) ast.Face {
	if f == nil {
		return nil
	}
	switch fm := f.(type) {
	case ast.FaceTop, ast.FaceBot:
		return fm
	case ast.FaceEq:
		if fm.IVar == j {
			// Substituting this variable
			switch iv := s.(type) {
			case ast.I0:
				if fm.IsOne {
					return ast.FaceBot{} // (i0 = 1) is false
				}
				return ast.FaceTop{} // (i0 = 0) is true
			case ast.I1:
				if fm.IsOne {
					return ast.FaceTop{} // (i1 = 1) is true
				}
				return ast.FaceBot{} // (i1 = 0) is false
			case ast.IVar:
				return ast.FaceEq{IVar: iv.Ix, IsOne: fm.IsOne}
			default:
				return fm
			}
		} else if fm.IVar > j {
			return ast.FaceEq{IVar: fm.IVar - 1, IsOne: fm.IsOne}
		}
		return fm
	case ast.FaceAnd:
		return simplifyFaceAndAST(
			ISubstFace(j, s, fm.Left),
			ISubstFace(j, s, fm.Right),
		)
	case ast.FaceOr:
		return simplifyFaceOrAST(
			ISubstFace(j, s, fm.Left),
			ISubstFace(j, s, fm.Right),
		)
	default:
		return f
	}
}

// ShiftFace shifts term variables in a face formula (faces have no term vars).
func ShiftFace(d, cutoff int, f ast.Face) ast.Face {
	// Faces contain no term variables, so return unchanged
	return f
}

// SubstFace substitutes a term for a term variable in a face formula.
func SubstFace(j int, s ast.Term, f ast.Face) ast.Face {
	// Faces contain no term variables, so return unchanged
	return f
}

// simplifyFaceAndAST simplifies φ ∧ ψ in AST representation.
func simplifyFaceAndAST(left, right ast.Face) ast.Face {
	// ⊥ ∧ ψ = ⊥
	if _, ok := left.(ast.FaceBot); ok {
		return ast.FaceBot{}
	}
	// φ ∧ ⊥ = ⊥
	if _, ok := right.(ast.FaceBot); ok {
		return ast.FaceBot{}
	}
	// ⊤ ∧ ψ = ψ
	if _, ok := left.(ast.FaceTop); ok {
		return right
	}
	// φ ∧ ⊤ = φ
	if _, ok := right.(ast.FaceTop); ok {
		return left
	}
	// Check for (i=0) ∧ (i=1) = ⊥
	if leq, lok := left.(ast.FaceEq); lok {
		if req, rok := right.(ast.FaceEq); rok {
			if leq.IVar == req.IVar && leq.IsOne != req.IsOne {
				return ast.FaceBot{}
			}
		}
	}
	return ast.FaceAnd{Left: left, Right: right}
}

// simplifyFaceOrAST simplifies φ ∨ ψ in AST representation.
func simplifyFaceOrAST(left, right ast.Face) ast.Face {
	// ⊤ ∨ ψ = ⊤
	if _, ok := left.(ast.FaceTop); ok {
		return ast.FaceTop{}
	}
	// φ ∨ ⊤ = ⊤
	if _, ok := right.(ast.FaceTop); ok {
		return ast.FaceTop{}
	}
	// ⊥ ∨ ψ = ψ
	if _, ok := left.(ast.FaceBot); ok {
		return right
	}
	// φ ∨ ⊥ = φ
	if _, ok := right.(ast.FaceBot); ok {
		return left
	}
	// Check for (i=0) ∨ (i=1) = ⊤
	if leq, lok := left.(ast.FaceEq); lok {
		if req, rok := right.(ast.FaceEq); rok {
			if leq.IVar == req.IVar && leq.IsOne != req.IsOne {
				return ast.FaceTop{}
			}
		}
	}
	return ast.FaceOr{Left: left, Right: right}
}

// substClosedExtension extends substClosed to handle cubical term types.
// This is an optimization when s is a closed term (no shifting needed).
func substClosedExtension(j int, s ast.Term, t ast.Term) (ast.Term, bool) {
	switch tm := t.(type) {
	case ast.Interval, ast.I0, ast.I1, ast.IVar:
		return tm, true // Interval terms have no term variables
	case ast.Path:
		return ast.Path{
			A: substClosed(j, s, tm.A),
			X: substClosed(j, s, tm.X),
			Y: substClosed(j, s, tm.Y),
		}, true
	case ast.PathP:
		return ast.PathP{
			A: substClosed(j, s, tm.A),
			X: substClosed(j, s, tm.X),
			Y: substClosed(j, s, tm.Y),
		}, true
	case ast.PathLam:
		return ast.PathLam{
			Binder: tm.Binder,
			Body:   substClosed(j, s, tm.Body),
		}, true
	case ast.PathApp:
		return ast.PathApp{
			P: substClosed(j, s, tm.P),
			R: substClosed(j, s, tm.R),
		}, true
	case ast.Transport:
		return ast.Transport{
			A: substClosed(j, s, tm.A),
			E: substClosed(j, s, tm.E),
		}, true
	// Face formulas - no term variables
	case ast.FaceTop, ast.FaceBot, ast.FaceEq:
		return tm, true
	case ast.FaceAnd:
		return ast.FaceAnd{
			Left:  SubstFace(j, s, tm.Left),
			Right: SubstFace(j, s, tm.Right),
		}, true
	case ast.FaceOr:
		return ast.FaceOr{
			Left:  SubstFace(j, s, tm.Left),
			Right: SubstFace(j, s, tm.Right),
		}, true
	// Partial types
	case ast.Partial:
		return ast.Partial{
			Phi: SubstFace(j, s, tm.Phi),
			A:   substClosed(j, s, tm.A),
		}, true
	case ast.System:
		branches := make([]ast.SystemBranch, len(tm.Branches))
		for i, br := range tm.Branches {
			branches[i] = ast.SystemBranch{
				Phi:  SubstFace(j, s, br.Phi),
				Term: substClosed(j, s, br.Term),
			}
		}
		return ast.System{Branches: branches}, true
	// Composition operations - no term variable binders
	case ast.Comp:
		return ast.Comp{
			IBinder: tm.IBinder,
			A:       substClosed(j, s, tm.A),
			Phi:     SubstFace(j, s, tm.Phi),
			Tube:    substClosed(j, s, tm.Tube),
			Base:    substClosed(j, s, tm.Base),
		}, true
	case ast.HComp:
		return ast.HComp{
			A:    substClosed(j, s, tm.A),
			Phi:  SubstFace(j, s, tm.Phi),
			Tube: substClosed(j, s, tm.Tube),
			Base: substClosed(j, s, tm.Base),
		}, true
	case ast.Fill:
		return ast.Fill{
			IBinder: tm.IBinder,
			A:       substClosed(j, s, tm.A),
			Phi:     SubstFace(j, s, tm.Phi),
			Tube:    substClosed(j, s, tm.Tube),
			Base:    substClosed(j, s, tm.Base),
		}, true
	// Glue types - no term variable binders
	case ast.Glue:
		branches := make([]ast.GlueBranch, len(tm.System))
		for i, br := range tm.System {
			branches[i] = ast.GlueBranch{
				Phi:   SubstFace(j, s, br.Phi),
				T:     substClosed(j, s, br.T),
				Equiv: substClosed(j, s, br.Equiv),
			}
		}
		return ast.Glue{
			A:      substClosed(j, s, tm.A),
			System: branches,
		}, true
	case ast.GlueElem:
		branches := make([]ast.GlueElemBranch, len(tm.System))
		for i, br := range tm.System {
			branches[i] = ast.GlueElemBranch{
				Phi:  SubstFace(j, s, br.Phi),
				Term: substClosed(j, s, br.Term),
			}
		}
		return ast.GlueElem{
			System: branches,
			Base:   substClosed(j, s, tm.Base),
		}, true
	case ast.Unglue:
		return ast.Unglue{
			Ty: substClosed(j, s, tm.Ty),
			G:  substClosed(j, s, tm.G),
		}, true
	// Univalence - no term variable binders
	case ast.UA:
		return ast.UA{
			A:     substClosed(j, s, tm.A),
			B:     substClosed(j, s, tm.B),
			Equiv: substClosed(j, s, tm.Equiv),
		}, true
	case ast.UABeta:
		return ast.UABeta{
			Equiv: substClosed(j, s, tm.Equiv),
			Arg:   substClosed(j, s, tm.Arg),
		}, true
	// Higher Inductive Types - no term variable binders in HITApp
	case ast.HITApp:
		args := make([]ast.Term, len(tm.Args))
		for i, arg := range tm.Args {
			args[i] = substClosed(j, s, arg)
		}
		iargs := make([]ast.Term, len(tm.IArgs))
		for i, iarg := range tm.IArgs {
			iargs[i] = substClosed(j, s, iarg)
		}
		return ast.HITApp{
			HITName: tm.HITName,
			Ctor:    tm.Ctor,
			Args:    args,
			IArgs:   iargs,
		}, true
	default:
		return nil, false
	}
}
