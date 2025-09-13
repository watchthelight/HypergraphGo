package hypergraph

// BFS performs breadth-first search starting from a vertex, returning reachable vertices.
func (h *Hypergraph[V]) BFS(start V) []V {
	if !h.HasVertex(start) {
		return nil
	}
	// Preallocate to reduce allocations under heavy traversal.
	visited := make(map[V]struct{}, len(h.vertices))
	queue := make([]V, 0, len(h.vertices))
	queue = append(queue, start)
	visited[start] = struct{}{}
	result := make([]V, 0, len(h.vertices))
	result = append(result, start)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for edgeID := range h.vertexToEdges[current] {
			members := h.edges[edgeID].Set
			for v := range members {
				if _, exists := visited[v]; !exists {
					visited[v] = struct{}{}
					queue = append(queue, v)
					result = append(result, v)
				}
			}
		}
	}
	return result
}

// DFS performs depth-first search starting from a vertex, returning reachable vertices.
func (h *Hypergraph[V]) DFS(start V) []V {
	if !h.HasVertex(start) {
		return nil
	}
	visited := make(map[V]struct{}, len(h.vertices))
	stack := make([]V, 0, len(h.vertices))
	stack = append(stack, start)
	result := make([]V, 0, len(h.vertices))
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if _, exists := visited[current]; !exists {
			visited[current] = struct{}{}
			result = append(result, current)
			for edgeID := range h.vertexToEdges[current] {
				members := h.edges[edgeID].Set
				for v := range members {
					if _, exists := visited[v]; !exists {
						stack = append(stack, v)
					}
				}
			}
		}
	}
	return result
}

// ConnectedComponents returns the connected components as slices of vertices.
func (h *Hypergraph[V]) ConnectedComponents() [][]V {
	visited := make(map[V]struct{})
	var components [][]V
	for v := range h.vertices {
		if _, exists := visited[v]; !exists {
			component := h.BFS(v)
			for _, u := range component {
				visited[u] = struct{}{}
			}
			components = append(components, component)
		}
	}
	return components
}
