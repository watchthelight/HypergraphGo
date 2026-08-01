package script

import (
	"strings"
	"testing"

	"github.com/watchthelight/HypergraphGo/kernel/check"
)

// Every failure a script author can hit should name a source line.
// Failures inside the tactic loop blame the tactic's line; failures after
// the loop (incomplete proof, extraction, final re-check) blame the
// theorem's line because no single tactic caused them.

func execSingle(t *testing.T, input string) TheoremResult {
	t.Helper()
	script, err := ParseString(input)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	result := Execute(script, check.NewCheckerWithStdlib())
	if len(result.Theorems) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Theorems))
	}
	return result.Theorems[0]
}

func TestExecute_TacticFailureReportsTacticLine(t *testing.T) {
	thm := execSingle(t, `
Theorem bad : Type
Proof
  unknown_tactic
Qed
`)
	if thm.Success {
		t.Fatal("theorem should fail (unknown tactic)")
	}
	if msg := thm.Error.Error(); !strings.Contains(msg, "(line 4)") {
		t.Errorf("tactic failure should cite line 4, got: %s", msg)
	}
}

func TestExecute_IncompleteProofReportsTheoremLine(t *testing.T) {
	thm := execSingle(t, `
Theorem incomplete : (Pi A Type (Pi B Type (Var 1)))
Proof
  intro A
Qed
`)
	if thm.Success {
		t.Fatal("theorem should fail (incomplete proof)")
	}
	if msg := thm.Error.Error(); !strings.Contains(msg, "(line 2)") {
		t.Errorf("incomplete proof should cite the theorem's line 2, got: %s", msg)
	}
}
