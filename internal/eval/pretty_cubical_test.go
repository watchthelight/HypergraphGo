package eval

import (
	"bytes"
	"testing"
)

func TestSprintValue_Cubical(t *testing.T) {
	tests := []struct {
		name string
		val  Value
		want string
	}{
		{"VI0", VI0{}, "i0"},
		{"VI1", VI1{}, "i1"},
		{"VIVar", VIVar{Level: 3}, "i{3}"},
		{"VFaceTop", VFaceTop{}, "⊤"},
		{"VFaceBot", VFaceBot{}, "⊥"},
		{"VFaceEq_0", VFaceEq{ILevel: 0, IsOne: false}, "(i{0} = 0)"},
		{"VFaceEq_1", VFaceEq{ILevel: 1, IsOne: true}, "(i{1} = 1)"},
		{"VPath", VPath{A: VSort{Level: 0}, X: VI0{}, Y: VI1{}}, "(Path Type0 i0 i1)"},
		{"VPathP", VPathP{X: VI0{}, Y: VI1{}}, "(PathP <closure> i0 i1)"},
		{"VPathLam", VPathLam{}, "(<_> <closure>)"},
		{"VTransport", VTransport{E: VI0{}}, "(transport <closure> i0)"},
		{"VGlue", VGlue{A: VSort{Level: 0}}, "(Glue Type0 [...])"},
		{"VGlueElem", VGlueElem{Base: VI0{}}, "(glue [...] i0)"},
		{"VUnglue", VUnglue{G: VI0{}}, "(unglue i0)"},
		{"VUA", VUA{A: VSort{Level: 0}, B: VSort{Level: 0}, Equiv: VSort{Level: 1}}, "(ua Type0 Type0 Type1)"},
		{"VUABeta", VUABeta{Equiv: VSort{Level: 0}, Arg: VI0{}}, "(ua-β Type0 i0)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SprintValue(tt.val)
			if got != tt.want {
				t.Errorf("SprintValue(%v) = %q, want %q", tt.val, got, tt.want)
			}
		})
	}
}

func TestValueEqual_Cubical(t *testing.T) {
	tests := []struct {
		name string
		v1   Value
		v2   Value
		want bool
	}{
		{"VI0 equal", VI0{}, VI0{}, true},
		{"VI0 vs VI1", VI0{}, VI1{}, false},
		{"VI1 equal", VI1{}, VI1{}, true},
		{"VIVar equal", VIVar{Level: 5}, VIVar{Level: 5}, true},
		{"VIVar diff level", VIVar{Level: 5}, VIVar{Level: 3}, false},
		{"VFaceTop equal", VFaceTop{}, VFaceTop{}, true},
		{"VFaceBot equal", VFaceBot{}, VFaceBot{}, true},
		{"VFaceTop vs VFaceBot", VFaceTop{}, VFaceBot{}, false},
		{"VFaceEq equal", VFaceEq{ILevel: 1, IsOne: true}, VFaceEq{ILevel: 1, IsOne: true}, true},
		{"VFaceEq diff level", VFaceEq{ILevel: 1, IsOne: true}, VFaceEq{ILevel: 2, IsOne: true}, false},
		{"VFaceEq diff isone", VFaceEq{ILevel: 1, IsOne: true}, VFaceEq{ILevel: 1, IsOne: false}, false},
		{"VPath equal", VPath{A: VSort{Level: 0}, X: VI0{}, Y: VI1{}}, VPath{A: VSort{Level: 0}, X: VI0{}, Y: VI1{}}, true},
		{"VPath diff A", VPath{A: VSort{Level: 0}, X: VI0{}, Y: VI1{}}, VPath{A: VSort{Level: 1}, X: VI0{}, Y: VI1{}}, false},
		{"VUA equal", VUA{A: VSort{Level: 0}, B: VSort{Level: 0}, Equiv: VI0{}}, VUA{A: VSort{Level: 0}, B: VSort{Level: 0}, Equiv: VI0{}}, true},
		{"VUA diff", VUA{A: VSort{Level: 0}, B: VSort{Level: 0}, Equiv: VI0{}}, VUA{A: VSort{Level: 1}, B: VSort{Level: 0}, Equiv: VI0{}}, false},
		{"VUABeta equal", VUABeta{Equiv: VI0{}, Arg: VI1{}}, VUABeta{Equiv: VI0{}, Arg: VI1{}}, true},
		{"VUnglue equal", VUnglue{Ty: VSort{Level: 0}, G: VI0{}}, VUnglue{Ty: VSort{Level: 0}, G: VI0{}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValueEqual(tt.v1, tt.v2)
			if got != tt.want {
				t.Errorf("ValueEqual(%v, %v) = %v, want %v", tt.v1, tt.v2, got, tt.want)
			}
		})
	}
}

func TestValueTypeName_Cubical(t *testing.T) {
	tests := []struct {
		val  Value
		want string
	}{
		{VI0{}, "VI0"},
		{VI1{}, "VI1"},
		{VIVar{}, "VIVar"},
		{VPath{}, "VPath"},
		{VPathP{}, "VPathP"},
		{VPathLam{}, "VPathLam"},
		{VTransport{}, "VTransport"},
		{VFaceTop{}, "VFaceTop"},
		{VFaceBot{}, "VFaceBot"},
		{VFaceEq{}, "VFaceEq"},
		{VFaceAnd{}, "VFaceAnd"},
		{VFaceOr{}, "VFaceOr"},
		{VPartial{}, "VPartial"},
		{VSystem{}, "VSystem"},
		{VComp{}, "VComp"},
		{VHComp{}, "VHComp"},
		{VFill{}, "VFill"},
		{VGlue{}, "VGlue"},
		{VGlueElem{}, "VGlueElem"},
		{VUnglue{}, "VUnglue"},
		{VUA{}, "VUA"},
		{VUABeta{}, "VUABeta"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := valueTypeName(tt.val)
			if got != tt.want {
				t.Errorf("valueTypeName(%T) = %q, want %q", tt.val, got, tt.want)
			}
		})
	}
}

func TestFaceValueEqual(t *testing.T) {
	tests := []struct {
		name string
		f1   FaceValue
		f2   FaceValue
		want bool
	}{
		{"nil nil", nil, nil, true},
		{"nil top", nil, VFaceTop{}, false},
		{"top top", VFaceTop{}, VFaceTop{}, true},
		{"bot bot", VFaceBot{}, VFaceBot{}, true},
		{"top bot", VFaceTop{}, VFaceBot{}, false},
		{"eq eq", VFaceEq{ILevel: 1, IsOne: true}, VFaceEq{ILevel: 1, IsOne: true}, true},
		{"eq diff", VFaceEq{ILevel: 1, IsOne: true}, VFaceEq{ILevel: 2, IsOne: true}, false},
		{"and and", VFaceAnd{Left: VFaceTop{}, Right: VFaceBot{}}, VFaceAnd{Left: VFaceTop{}, Right: VFaceBot{}}, true},
		{"and diff", VFaceAnd{Left: VFaceTop{}, Right: VFaceBot{}}, VFaceAnd{Left: VFaceBot{}, Right: VFaceTop{}}, false},
		{"or or", VFaceOr{Left: VFaceTop{}, Right: VFaceBot{}}, VFaceOr{Left: VFaceTop{}, Right: VFaceBot{}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := faceValueEqual(tt.f1, tt.f2)
			if got != tt.want {
				t.Errorf("faceValueEqual(%v, %v) = %v, want %v", tt.f1, tt.f2, got, tt.want)
			}
		})
	}
}

func TestWriteFaceValue(t *testing.T) {
	tests := []struct {
		name string
		f    FaceValue
		want string
	}{
		{"top", VFaceTop{}, "⊤"},
		{"bot", VFaceBot{}, "⊥"},
		{"eq0", VFaceEq{ILevel: 0, IsOne: false}, "(i{0} = 0)"},
		{"eq1", VFaceEq{ILevel: 2, IsOne: true}, "(i{2} = 1)"},
		{"and", VFaceAnd{Left: VFaceTop{}, Right: VFaceBot{}}, "(⊤ ∧ ⊥)"},
		{"or", VFaceOr{Left: VFaceTop{}, Right: VFaceBot{}}, "(⊤ ∨ ⊥)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sprintFaceValue(tt.f)
			if got != tt.want {
				t.Errorf("sprintFaceValue(%v) = %q, want %q", tt.f, got, tt.want)
			}
		})
	}
}

// sprintFaceValue is a helper for testing writeFaceValue.
func sprintFaceValue(f FaceValue) string {
	var b bytes.Buffer
	writeFaceValue(&b, f)
	return b.String()
}
