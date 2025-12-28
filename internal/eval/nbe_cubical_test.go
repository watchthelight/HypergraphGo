package eval

import (
	"testing"
)

// =============================================================================
// Phase 2: Interval Environment Tests
// =============================================================================

func TestEmptyIEnv(t *testing.T) {
	ienv := EmptyIEnv()
	if ienv == nil {
		t.Fatal("EmptyIEnv() returned nil")
	}
	if len(ienv.Bindings) != 0 {
		t.Errorf("EmptyIEnv() has %d bindings, want 0", len(ienv.Bindings))
	}
}

func TestIEnv_ILen(t *testing.T) {
	tests := []struct {
		name string
		ienv *IEnv
		want int
	}{
		{"nil", nil, 0},
		{"empty", EmptyIEnv(), 0},
		{"one binding", EmptyIEnv().Extend(VI0{}), 1},
		{"two bindings", EmptyIEnv().Extend(VI0{}).Extend(VI1{}), 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ienv.ILen()
			if got != tt.want {
				t.Errorf("ILen() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestIEnv_Extend(t *testing.T) {
	// Extend nil environment
	t.Run("nil receiver", func(t *testing.T) {
		var ienv *IEnv = nil
		extended := ienv.Extend(VI0{})
		if extended == nil {
			t.Fatal("Extend on nil returned nil")
		}
		if extended.ILen() != 1 {
			t.Errorf("ILen() = %d, want 1", extended.ILen())
		}
	})

	// Extend empty environment
	t.Run("empty", func(t *testing.T) {
		ienv := EmptyIEnv()
		extended := ienv.Extend(VI1{})
		if extended.ILen() != 1 {
			t.Errorf("ILen() = %d, want 1", extended.ILen())
		}
		// Original should be unchanged
		if ienv.ILen() != 0 {
			t.Errorf("original ILen() = %d, want 0", ienv.ILen())
		}
	})

	// Extend with multiple values
	t.Run("multiple", func(t *testing.T) {
		ienv := EmptyIEnv().Extend(VI0{}).Extend(VI1{}).Extend(VIVar{Level: 5})
		if ienv.ILen() != 3 {
			t.Errorf("ILen() = %d, want 3", ienv.ILen())
		}
	})
}

func TestIEnv_Lookup(t *testing.T) {
	// Build environment: [VIVar{5}, VI1, VI0] (most recent first)
	ienv := EmptyIEnv().Extend(VI0{}).Extend(VI1{}).Extend(VIVar{Level: 5})

	tests := []struct {
		name    string
		ix      int
		wantTyp string
	}{
		{"index 0 (most recent)", 0, "VIVar"},
		{"index 1", 1, "VI1"},
		{"index 2 (oldest)", 2, "VI0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ienv.Lookup(tt.ix)
			switch tt.wantTyp {
			case "VI0":
				if _, ok := got.(VI0); !ok {
					t.Errorf("Lookup(%d) = %T, want VI0", tt.ix, got)
				}
			case "VI1":
				if _, ok := got.(VI1); !ok {
					t.Errorf("Lookup(%d) = %T, want VI1", tt.ix, got)
				}
			case "VIVar":
				if _, ok := got.(VIVar); !ok {
					t.Errorf("Lookup(%d) = %T, want VIVar", tt.ix, got)
				}
			}
		})
	}
}

func TestIEnv_Lookup_OutOfBounds(t *testing.T) {
	ienv := EmptyIEnv().Extend(VI0{})

	tests := []struct {
		name string
		ix   int
	}{
		{"negative index", -1},
		{"out of bounds", 5},
		{"just past end", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ienv.Lookup(tt.ix)
			// Out of bounds should return VIVar with the index as level
			v, ok := got.(VIVar)
			if !ok {
				t.Errorf("Lookup(%d) = %T, want VIVar", tt.ix, got)
				return
			}
			if v.Level != tt.ix {
				t.Errorf("VIVar.Level = %d, want %d", v.Level, tt.ix)
			}
		})
	}
}

func TestIEnv_Lookup_NilReceiver(t *testing.T) {
	var ienv *IEnv = nil
	got := ienv.Lookup(0)
	v, ok := got.(VIVar)
	if !ok {
		t.Errorf("Lookup on nil = %T, want VIVar", got)
		return
	}
	if v.Level != 0 {
		t.Errorf("VIVar.Level = %d, want 0", v.Level)
	}
}
