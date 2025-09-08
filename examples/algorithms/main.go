package main

import (
    "fmt"
    "time"

    "github.com/watchthelight/hypergraphgo/hypergraph"
)

// Tiny example showcasing algorithms on a small hypergraph.
func main() {
    h := hypergraph.NewHypergraph[string]()
    _ = h.AddEdge("E1", []string{"A", "B"})
    _ = h.AddEdge("E2", []string{"B", "C"})
    _ = h.AddEdge("E3", []string{"C", "D"})

    fmt.Println("Greedy hitting set:", h.GreedyHittingSet())

    coloring := h.GreedyColoring()
    fmt.Println("Greedy coloring:", coloring)

    transversals, err := h.EnumerateMinimalTransversals(5, 200*time.Millisecond)
    fmt.Println("Minimal transversals (cutoff may apply):", transversals, err)
}

