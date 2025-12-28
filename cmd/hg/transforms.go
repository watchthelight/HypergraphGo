package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"slices"
)

func cmdDual(args []string) error {
	fs := flag.NewFlagSet("dual", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	output := fs.String("o", "", "output file")
	fs.Parse(args)

	if *file == "" || *output == "" {
		return fmt.Errorf("missing required flags: -f FILE -o OUTPUT")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	dual := hg.Dual()
	return saveGraph(dual, *output)
}

// graphJSON represents a simple graph in JSON format.
type graphJSON struct {
	Vertices []string   `json:"vertices"`
	Edges    [][]string `json:"edges"`
}

func cmdTwoSection(args []string) error {
	fs := flag.NewFlagSet("two-section", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	output := fs.String("o", "", "output file")
	fs.Parse(args)

	if *file == "" || *output == "" {
		return fmt.Errorf("missing required flags: -f FILE -o OUTPUT")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	g := hg.TwoSection()
	return saveSimpleGraphString(g.Vertices(), g.Edges(), *output)
}

func cmdLineGraph(args []string) error {
	fs := flag.NewFlagSet("line-graph", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	output := fs.String("o", "", "output file")
	fs.Parse(args)

	if *file == "" || *output == "" {
		return fmt.Errorf("missing required flags: -f FILE -o OUTPUT")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	g := hg.LineGraph()
	return saveSimpleGraphString(g.Vertices(), g.Edges(), *output)
}

// saveSimpleGraphString saves a simple graph (with string vertices) to JSON.
func saveSimpleGraphString(vertices []string, edges []struct{ From, To string }, filename string) error {
	verts := make([]string, len(vertices))
	copy(verts, vertices)
	slices.Sort(verts)

	edgeList := make([][]string, 0, len(edges))
	for _, e := range edges {
		pair := []string{e.From, e.To}
		slices.Sort(pair)
		edgeList = append(edgeList, pair)
	}
	slices.SortFunc(edgeList, func(a, b []string) int {
		if a[0] < b[0] {
			return -1
		}
		if a[0] > b[0] {
			return 1
		}
		if a[1] < b[1] {
			return -1
		}
		if a[1] > b[1] {
			return 1
		}
		return 0
	})

	data := graphJSON{
		Vertices: verts,
		Edges:    edgeList,
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(data)
}
