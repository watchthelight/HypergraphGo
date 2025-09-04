package util

// Set is a generic set type.
type Set[T comparable] map[T]struct{}

// NewSet creates a new set.
func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

// Add adds an element to the set.
func (s Set[T]) Add(v T) {
	s[v] = struct{}{}
}

// Remove removes an element from the set.
func (s Set[T]) Remove(v T) {
	delete(s, v)
}

// Contains checks if the set contains an element.
func (s Set[T]) Contains(v T) bool {
	_, exists := s[v]
	return exists
}

// Union returns the union of two sets.
func Union[T comparable](a, b Set[T]) Set[T] {
	result := NewSet[T]()
	for v := range a {
		result.Add(v)
	}
	for v := range b {
		result.Add(v)
	}
	return result
}

// Intersection returns the intersection of two sets.
func Intersection[T comparable](a, b Set[T]) Set[T] {
	result := NewSet[T]()
	for v := range a {
		if b.Contains(v) {
			result.Add(v)
		}
	}
	return result
}
