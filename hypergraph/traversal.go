package hypergraph

// BFS performs breadth-first search starting from a vertex, returning reachable vertices.
func (h *Hypergraph[V]) BFS(start V) []V {
	if !h.HasVertex(start) {
		return nil
	}
	visited := make(map[V]struct{})
	queue := []V{start}
	visited[start] = struct{}{}
	result := []V{start}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for edgeID := range h.vertexToEdges[current] {
			for v := range h.edges[edgeID].Set {
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
	visited := make(map[V]struct{})
	stack := []V{start}
	result := []V{}
	for len(stack) > 0 {
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if _, exists := visited[current]; !exists {
			visited[current] = struct{}{}
			result = append(result, current)
			for edgeID := range h.vertexToEdges[current] {
				for v := range h.edges[edgeID].Set {
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
