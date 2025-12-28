package main

import (
	"os"

	"github.com/watchthelight/HypergraphGo/hypergraph"
)

// loadGraph loads a hypergraph from a JSON file.
func loadGraph(filename string) (*hypergraph.Hypergraph[string], error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return hypergraph.LoadJSON[string](f)
}

// saveGraph saves a hypergraph to a JSON file.
func saveGraph(hg *hypergraph.Hypergraph[string], filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return hg.SaveJSON(f)
}
