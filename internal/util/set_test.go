package util

import "testing"

func TestSetBasic(t *testing.T) {
    t.Parallel()
    s := NewSet[int]()
    if s.Contains(1) {
        t.Fatalf("unexpected contains before add")
    }
    s.Add(1)
    if !s.Contains(1) {
        t.Fatalf("missing element after add")
    }
    s.Remove(1)
    if s.Contains(1) {
        t.Fatalf("still contains after remove")
    }
}

func TestUnionIntersectionEdgeCases(t *testing.T) {
    t.Parallel()
    a := NewSet[int]()
    b := NewSet[int]()
    a.Add(1); a.Add(2)
    // Union with empty set yields original elements
    u := Union(a, b)
    if len(u) != 2 || !u.Contains(1) || !u.Contains(2) {
        t.Fatalf("unexpected union: %#v", u)
    }
    // Intersection with disjoint set is empty
    b.Add(3)
    inter := Intersection(a, b)
    if len(inter) != 0 {
        t.Fatalf("expected empty intersection, got %#v", inter)
    }
}

