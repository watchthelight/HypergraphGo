package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"
)

func cmdHittingSet(args []string) error {
	fs := flag.NewFlagSet("hitting-set", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" {
		return fmt.Errorf("missing required flag: -f FILE")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	result := hg.GreedyHittingSet()
	slices.Sort(result)
	fmt.Println(strings.Join(result, " "))
	return nil
}

func cmdTransversals(args []string) error {
	fs := flag.NewFlagSet("transversals", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	max := fs.Int("max", 100, "maximum number of transversals")
	timeout := fs.Duration("timeout", 10*time.Second, "maximum time")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" {
		return fmt.Errorf("missing required flag: -f FILE")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	transversals, err := hg.EnumerateMinimalTransversals(*max, *timeout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warning: %v\n", err)
	}

	for i, t := range transversals {
		slices.Sort(t)
		fmt.Printf("%d: %s\n", i+1, strings.Join(t, ", "))
	}
	return nil
}

func cmdColoring(args []string) error {
	fs := flag.NewFlagSet("coloring", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" {
		return fmt.Errorf("missing required flag: -f FILE")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	coloring := hg.GreedyColoring()

	// Sort vertices for stable output
	vertices := hg.Vertices()
	slices.Sort(vertices)

	for _, v := range vertices {
		fmt.Printf("%s: %d\n", v, coloring[v])
	}
	return nil
}

func cmdIncidence(args []string) error {
	fs := flag.NewFlagSet("incidence", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" {
		return fmt.Errorf("missing required flag: -f FILE")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	vertexIndex, edgeIndex, coo := hg.IncidenceMatrix()

	// Get sorted vertices and edges
	vertices := make([]string, len(vertexIndex))
	for v, i := range vertexIndex {
		vertices[i] = v
	}
	edges := make([]string, len(edgeIndex))
	for e, i := range edgeIndex {
		edges[i] = e
	}

	fmt.Printf("Vertices: %s\n", strings.Join(vertices, ", "))
	fmt.Printf("Edges: %s\n", strings.Join(edges, ", "))
	fmt.Println("Incidence (row, col):")
	for i := range coo.Rows {
		fmt.Printf("  (%d, %d)\n", coo.Rows[i], coo.Cols[i])
	}
	return nil
}
