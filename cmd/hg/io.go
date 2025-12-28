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

// saveGraph saves a hypergraph to a JSON file atomically.
// It writes to a temp file first, then renames to the target.
func saveGraph(hg *hypergraph.Hypergraph[string], filename string) error {
	// Write to temp file first
	tmpFile := filename + ".tmp"
	f, err := os.Create(tmpFile)
	if err != nil {
		return err
	}

	if err := hg.SaveJSON(f); err != nil {
		f.Close()
		os.Remove(tmpFile)
		return err
	}

	if err := f.Close(); err != nil {
		os.Remove(tmpFile)
		return err
	}

	// Atomic rename
	return os.Rename(tmpFile, filename)
}
