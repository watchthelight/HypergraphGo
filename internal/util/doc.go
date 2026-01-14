// Package util provides generic utility types and functions.
//
// This package contains small, general-purpose helpers used across
// the HoTTGo codebase. Currently it provides a generic set type.
//
// # Set
//
// [Set] is a generic set type using Go's map-based implementation:
//
//	s := util.NewSet[string]()
//	s.Add("foo")
//	s.Add("bar")
//	s.Contains("foo") // true
//	s.Remove("foo")
//
// Set operations:
//
//   - [NewSet] - create an empty set
//   - [Set.Add] - add an element
//   - [Set.Remove] - remove an element
//   - [Set.Contains] - test membership
//   - [Union] - set union
//   - [Intersection] - set intersection
package util
