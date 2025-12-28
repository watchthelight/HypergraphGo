package eval

import (
	"strings"
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
)

// ============================================================================
// SprintValue Tests
// ============================================================================

func TestSprintValue_VSort(t *testing.T) {
	t.Parallel()
	tests := []struct {
		level int
		want  string
	}{
		{0, "Type0"},
		{1, "Type1"},
		{10, "Type10"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			v := VSort{Level: tt.level}
			got := SprintValue(v)
			if got != tt.want {
				t.Errorf("SprintValue(VSort{%d}) = %q, want %q", tt.level, got, tt.want)
			}
		})
	}
}

func TestSprintValue_VGlobal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want string
	}{
		{"foo", "foo"},
		{"Bar", "Bar"},
		{"error:test", "error:test"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := VGlobal{Name: tt.name}
			got := SprintValue(v)
			if got != tt.want {
				t.Errorf("SprintValue(VGlobal{%q}) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestSprintValue_VPair(t *testing.T) {
	t.Parallel()
	// Simple pair
	v := VPair{
		Fst: VSort{Level: 0},
		Snd: VSort{Level: 1},
	}
	got := SprintValue(v)
	if got != "(Type0 , Type1)" {
		t.Errorf("SprintValue(VPair) = %q, want %q", got, "(Type0 , Type1)")
	}
}

func TestSprintValue_VPair_Nested(t *testing.T) {
	t.Parallel()
	v := VPair{
		Fst: VPair{
			Fst: VSort{Level: 0},
			Snd: VSort{Level: 1},
		},
		Snd: VGlobal{Name: "x"},
	}
	got := SprintValue(v)
	if got != "((Type0 , Type1) , x)" {
		t.Errorf("SprintValue(nested VPair) = %q, want %q", got, "((Type0 , Type1) , x)")
	}
}

func TestSprintValue_VLam(t *testing.T) {
	t.Parallel()
	v := VLam{Body: &Closure{
		Env:  &Env{Bindings: nil},
		Term: ast.Var{Ix: 0},
	}}
	got := SprintValue(v)
	if !strings.Contains(got, "<closure>") {
		t.Errorf("SprintValue(VLam) = %q, want to contain <closure>", got)
	}
}

func TestSprintValue_VPi(t *testing.T) {
	t.Parallel()
	v := VPi{
		A: VSort{Level: 0},
		B: &Closure{
			Env:  &Env{Bindings: nil},
			Term: ast.Sort{U: 1},
		},
	}
	got := SprintValue(v)
	if !strings.Contains(got, "Pi") && !strings.Contains(got, "Type0") {
		t.Errorf("SprintValue(VPi) = %q, want to contain Pi and Type0", got)
	}
}

func TestSprintValue_VSigma(t *testing.T) {
	t.Parallel()
	v := VSigma{
		A: VSort{Level: 0},
		B: &Closure{
			Env:  &Env{Bindings: nil},
			Term: ast.Sort{U: 1},
		},
	}
	got := SprintValue(v)
	if !strings.Contains(got, "Sigma") && !strings.Contains(got, "Type0") {
		t.Errorf("SprintValue(VSigma) = %q, want to contain Sigma and Type0", got)
	}
}

func TestSprintValue_VId(t *testing.T) {
	t.Parallel()
	v := VId{
		A: VSort{Level: 0},
		X: VGlobal{Name: "a"},
		Y: VGlobal{Name: "b"},
	}
	got := SprintValue(v)
	if !strings.Contains(got, "Id") {
		t.Errorf("SprintValue(VId) = %q, want to contain Id", got)
	}
	if !strings.Contains(got, "a") || !strings.Contains(got, "b") {
		t.Errorf("SprintValue(VId) = %q, want to contain a and b", got)
	}
}

func TestSprintValue_VRefl(t *testing.T) {
	t.Parallel()
	v := VRefl{
		A: VSort{Level: 0},
		X: VGlobal{Name: "a"},
	}
	got := SprintValue(v)
	if !strings.Contains(got, "refl") {
		t.Errorf("SprintValue(VRefl) = %q, want to contain refl", got)
	}
}

func TestSprintValue_VNeutral_VarHead(t *testing.T) {
	t.Parallel()
	v := VNeutral{N: Neutral{
		Head: Head{Var: 0},
		Sp:   nil,
	}}
	got := SprintValue(v)
	if got != "{0}" {
		t.Errorf("SprintValue(VNeutral var) = %q, want {0}", got)
	}
}

func TestSprintValue_VNeutral_GlobalHead(t *testing.T) {
	t.Parallel()
	v := VNeutral{N: Neutral{
		Head: Head{Glob: "f"},
		Sp:   nil,
	}}
	got := SprintValue(v)
	if got != "f" {
		t.Errorf("SprintValue(VNeutral global) = %q, want f", got)
	}
}

func TestSprintValue_VNeutral_WithSpine(t *testing.T) {
	t.Parallel()
	v := VNeutral{N: Neutral{
		Head: Head{Glob: "f"},
		Sp:   []Value{VSort{Level: 0}, VGlobal{Name: "x"}},
	}}
	got := SprintValue(v)
	if got != "(f Type0 x)" {
		t.Errorf("SprintValue(VNeutral with spine) = %q, want (f Type0 x)", got)
	}
}

// ============================================================================
// SprintNeutral Tests
// ============================================================================

func TestSprintNeutral_VarHead_EmptySpine(t *testing.T) {
	t.Parallel()
	n := Neutral{Head: Head{Var: 5}, Sp: nil}
	got := SprintNeutral(n)
	if got != "{5}" {
		t.Errorf("SprintNeutral(var empty) = %q, want {5}", got)
	}
}

func TestSprintNeutral_GlobalHead_EmptySpine(t *testing.T) {
	t.Parallel()
	n := Neutral{Head: Head{Glob: "myFunc"}, Sp: nil}
	got := SprintNeutral(n)
	if got != "myFunc" {
		t.Errorf("SprintNeutral(global empty) = %q, want myFunc", got)
	}
}

func TestSprintNeutral_WithSpine(t *testing.T) {
	t.Parallel()
	n := Neutral{
		Head: Head{Glob: "app"},
		Sp:   []Value{VSort{Level: 1}},
	}
	got := SprintNeutral(n)
	if got != "(app Type1)" {
		t.Errorf("SprintNeutral(with spine) = %q, want (app Type1)", got)
	}
}

// ============================================================================
// PrettyValue Tests
// ============================================================================

func TestPrettyValue_VSort(t *testing.T) {
	t.Parallel()
	v := VSort{Level: 2}
	got := PrettyValue(v)
	// PrettyValue reifies then uses ast.Sprint
	if !strings.Contains(got, "2") && !strings.Contains(got, "Type") {
		t.Errorf("PrettyValue(VSort{2}) = %q, expected to contain Type/2", got)
	}
}

func TestPrettyValue_VGlobal(t *testing.T) {
	t.Parallel()
	v := VGlobal{Name: "myGlobal"}
	got := PrettyValue(v)
	if !strings.Contains(got, "myGlobal") {
		t.Errorf("PrettyValue(VGlobal) = %q, expected to contain myGlobal", got)
	}
}

// ============================================================================
// PrettyNeutral Tests
// ============================================================================

func TestPrettyNeutral_VarHead(t *testing.T) {
	t.Parallel()
	n := Neutral{Head: Head{Var: 0}, Sp: nil}
	got := PrettyNeutral(n)
	// Should produce a variable representation
	if got == "" {
		t.Error("PrettyNeutral returned empty string")
	}
}

func TestPrettyNeutral_GlobalHead(t *testing.T) {
	t.Parallel()
	n := Neutral{Head: Head{Glob: "test"}, Sp: nil}
	got := PrettyNeutral(n)
	if !strings.Contains(got, "test") {
		t.Errorf("PrettyNeutral(global) = %q, expected to contain test", got)
	}
}

// ============================================================================
// DebugValue Tests
// ============================================================================

func TestDebugValue_VSort(t *testing.T) {
	t.Parallel()
	v := VSort{Level: 0}
	got := DebugValue(v)

	if !strings.Contains(got, "Value:") {
		t.Error("DebugValue missing Value: prefix")
	}
	if !strings.Contains(got, "Type:") {
		t.Error("DebugValue missing Type: prefix")
	}
	if !strings.Contains(got, "VSort") {
		t.Error("DebugValue missing VSort type name")
	}
}

func TestDebugValue_VNeutral(t *testing.T) {
	t.Parallel()
	v := VNeutral{N: Neutral{Head: Head{Glob: "x"}, Sp: nil}}
	got := DebugValue(v)

	if !strings.Contains(got, "VNeutral") {
		t.Error("DebugValue missing VNeutral type name")
	}
}
