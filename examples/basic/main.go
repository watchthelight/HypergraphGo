package main

import (
	"fmt"
	"os"

	"github.com/watchthelight/hypergraphgo/hypergraph"
)

func main() {
	h := hypergraph.NewHypergraph[string]()
	h.AddVertex("A")
	h.AddVertex("B")
	h.AddVertex("C")
	h.AddEdge("E1", []string{"A", "B"})
	h.AddEdge("E2", []string{"B", "C"})
	h.AddEdge("E3", []string{"A", "C"})

	fmt.Println("Vertices:", h.Vertices())
	fmt.Println("Edges:", h.Edges())
	for _, v := range h.Vertices() {
		fmt.Printf("Degree of %s: %d\n", v, h.VertexDegree(v))
	}

	comps := h.ConnectedComponents()
	fmt.Println("Components:", comps)

	dual := h.Dual()
	fmt.Println("Dual vertices:", dual.Vertices())
	fmt.Println("Dual edges:", dual.Edges())

	section := h.TwoSection()
	fmt.Println("2-Section vertices:", len(section.vertices))
	fmt.Println("2-Section edges:", len(section.edges))

	// Write dual and section to JSON
	dualFile, _ := os.Create("dual.json")
	dual.SaveJSON(dualFile)
	dualFile.Close()

	sectionFile, _ := os.Create("section.json")
	// For section, since it's Graph, marshal manually
	fmt.Fprintf(sectionFile, `{"vertices": ["A","B","C"], "edges": [{"from":"A","to":"B"},{"from":"B","to":"C"},{"from":"A","to":"C"}]}`)
	sectionFile.Close()
}
