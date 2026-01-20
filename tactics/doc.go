// Package tactics provides Ltac-style proof tactics for HoTT.
//
// Tactics allow interactive proof construction by transforming proof goals
// step by step. Each tactic takes a proof state and produces a new state
// with (usually) fewer or simpler goals.
//
// # Overview
//
// A typical proof session:
//
//  1. Create a [Prover] with the goal type
//  2. Apply tactics to transform goals
//  3. When all goals are solved, extract the proof term
//
// Example:
//
//	// Prove: (A -> B -> A)
//	prover := tactics.NewProver(piType) // goal: (A -> B -> A)
//	prover.Apply(tactics.Intro("a"))    // goal: (B -> A), hyp: a:A
//	prover.Apply(tactics.Intro("b"))    // goal: A, hyps: a:A, b:B
//	prover.Apply(tactics.Exact(varA))   // solved!
//	term, _ := prover.Extract()         // λa.λb.a
//
// # Core Tactics
//
// Basic proof construction:
//
//   - [Intro] - introduce a variable from a Pi type goal
//   - [IntroN] - introduce multiple variables
//   - [Intros] - introduce all variables
//   - [Exact] - provide the exact proof term
//   - [Apply] - apply a function to the goal
//   - [Split] - split a Sigma type goal into components
//   - [Rewrite], [RewriteRev] - rewrite using an equality
//   - [Reflexivity] - prove Id with refl
//   - [Assumption] - use a hypothesis matching the goal
//
// # Inductive Tactics
//
// Working with inductive types (Unit, Empty, Sum, List, Nat, Bool):
//
//   - [Contradiction] - prove any goal from Empty hypothesis
//   - [Left], [Right] - prove Sum goal with injection
//   - [Destruct] - case analysis on Sum or Bool hypothesis
//   - [Induction] - structural induction on Nat or List
//   - [Cases] - non-recursive case analysis
//   - [Constructor] - apply first applicable constructor
//   - [Exists] - provide witness for Sigma goal
//
// # Automation
//
//   - [Trivial] - try reflexivity then assumption
//   - [Auto] - automatic proof search
//   - [Simpl] - simplify the goal
//
// # Tactic Combinators
//
// Compose tactics for complex proofs:
//
//   - [Seq] - sequence tactics (t1; t2; t3)
//   - [OrElse] - try first, else second (t1 <|> t2)
//   - [Try] - try tactic, succeed either way
//   - [Repeat] - repeat until failure
//   - [RepeatN] - repeat at most N times
//   - [First] - try each until one succeeds
//   - [All] - apply to all goals
//   - [Progress] - fail if no progress made
//
// # Prover API
//
// [Prover] provides a high-level interface for proof sessions:
//
//   - [NewProver] - start a proof with a goal type
//   - [Prover.Apply] - apply a tactic
//   - [Prover.Goals] - get current goals
//   - [Prover.Done] - check if proof is complete
//   - [Prover.Extract] - get the proof term
//   - [Prover.Undo] - undo last tactic
//
// # Fluent API
//
// For concise proof scripts, use the fluent interface:
//
//	term := tactics.Prove(goalType).
//	    Intro("x").
//	    Intro("y").
//	    Exact(varX).
//	    QED()
//
// # Tactic Results
//
// Tactics return [TacticResult] containing:
//
//   - State: the new proof state (on success)
//   - Err: error description (on failure)
//   - Message: optional status message
//
// Use [Success], [SuccessMsg], [Fail], [Failf] to create results.
//
// # Proof State
//
// See the [proofstate] subpackage for:
//
//   - [proofstate.ProofState] - proof obligations and metavariable store
//   - [proofstate.Goal] - individual proof obligations
//   - [proofstate.Hypothesis] - local context bindings
package tactics
