package main

import (
	"flag"
	"fmt"
	"slices"
	"strings"

	"github.com/watchthelight/HypergraphGo/hypergraph"
)

func cmdInfo(args []string) error {
	fs := flag.NewFlagSet("info", flag.ExitOnError)
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

	fmt.Printf("Vertices: %d\n", hg.NumVertices())
	fmt.Printf("Edges:    %d\n", hg.NumEdges())
	fmt.Printf("Empty:    %v\n", hg.IsEmpty())
	return nil
}

func cmdNew(args []string) error {
	fs := flag.NewFlagSet("new", flag.ExitOnError)
	output := fs.String("o", "", "output file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *output == "" {
		return fmt.Errorf("missing required flag: -o FILE")
	}

	hg := hypergraph.NewHypergraph[string]()
	return saveGraph(hg, *output)
}

func cmdValidate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" {
		return fmt.Errorf("missing required flag: -f FILE")
	}

	_, err := loadGraph(*file)
	if err != nil {
		return fmt.Errorf("invalid: %w", err)
	}

	fmt.Println("valid")
	return nil
}

func cmdAddVertex(args []string) error {
	fs := flag.NewFlagSet("add-vertex", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	vertex := fs.String("v", "", "vertex to add")
	output := fs.String("o", "", "output file (default: modify in-place)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" || *vertex == "" {
		return fmt.Errorf("missing required flags: -f FILE -v VERTEX")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	hg.AddVertex(*vertex)

	outFile := *output
	if outFile == "" {
		outFile = *file
	}
	return saveGraph(hg, outFile)
}

func cmdRemoveVertex(args []string) error {
	fs := flag.NewFlagSet("remove-vertex", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	vertex := fs.String("v", "", "vertex to remove")
	output := fs.String("o", "", "output file (default: modify in-place)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" || *vertex == "" {
		return fmt.Errorf("missing required flags: -f FILE -v VERTEX")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	if !hg.HasVertex(*vertex) {
		return fmt.Errorf("vertex not found: %s", *vertex)
	}
	hg.RemoveVertex(*vertex)

	outFile := *output
	if outFile == "" {
		outFile = *file
	}
	return saveGraph(hg, outFile)
}

func cmdHasVertex(args []string) error {
	fs := flag.NewFlagSet("has-vertex", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	vertex := fs.String("v", "", "vertex to check")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" || *vertex == "" {
		return fmt.Errorf("missing required flags: -f FILE -v VERTEX")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	if hg.HasVertex(*vertex) {
		fmt.Println("true")
		return nil
	}
	fmt.Println("false")
	return nil
}

func cmdAddEdge(args []string) error {
	fs := flag.NewFlagSet("add-edge", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	edgeID := fs.String("id", "", "edge ID")
	members := fs.String("m", "", "comma-separated member vertices")
	output := fs.String("o", "", "output file (default: modify in-place)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" || *edgeID == "" || *members == "" {
		return fmt.Errorf("missing required flags: -f FILE -id ID -m MEMBERS")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	memberList := strings.Split(*members, ",")
	for i := range memberList {
		memberList[i] = strings.TrimSpace(memberList[i])
	}

	if err := hg.AddEdge(*edgeID, memberList); err != nil {
		return err
	}

	outFile := *output
	if outFile == "" {
		outFile = *file
	}
	return saveGraph(hg, outFile)
}

func cmdRemoveEdge(args []string) error {
	fs := flag.NewFlagSet("remove-edge", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	edgeID := fs.String("id", "", "edge ID to remove")
	output := fs.String("o", "", "output file (default: modify in-place)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" || *edgeID == "" {
		return fmt.Errorf("missing required flags: -f FILE -id ID")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	if !hg.HasEdge(*edgeID) {
		return fmt.Errorf("edge not found: %s", *edgeID)
	}
	hg.RemoveEdge(*edgeID)

	outFile := *output
	if outFile == "" {
		outFile = *file
	}
	return saveGraph(hg, outFile)
}

func cmdHasEdge(args []string) error {
	fs := flag.NewFlagSet("has-edge", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	edgeID := fs.String("id", "", "edge ID to check")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" || *edgeID == "" {
		return fmt.Errorf("missing required flags: -f FILE -id ID")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	if hg.HasEdge(*edgeID) {
		fmt.Println("true")
		return nil
	}
	fmt.Println("false")
	return nil
}

func cmdVertices(args []string) error {
	fs := flag.NewFlagSet("vertices", flag.ExitOnError)
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

	verts := hg.Vertices()
	slices.Sort(verts)
	for _, v := range verts {
		fmt.Println(v)
	}
	return nil
}

func cmdEdges(args []string) error {
	fs := flag.NewFlagSet("edges", flag.ExitOnError)
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

	edges := hg.Edges()
	slices.Sort(edges)
	for _, id := range edges {
		members := hg.EdgeMembers(id)
		slices.Sort(members)
		fmt.Printf("%s: %s\n", id, strings.Join(members, ", "))
	}
	return nil
}

func cmdDegree(args []string) error {
	fs := flag.NewFlagSet("degree", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	vertex := fs.String("v", "", "vertex")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" || *vertex == "" {
		return fmt.Errorf("missing required flags: -f FILE -v VERTEX")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	if !hg.HasVertex(*vertex) {
		return fmt.Errorf("vertex not found: %s", *vertex)
	}
	fmt.Println(hg.VertexDegree(*vertex))
	return nil
}

func cmdEdgeSize(args []string) error {
	fs := flag.NewFlagSet("edge-size", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	edgeID := fs.String("id", "", "edge ID")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" || *edgeID == "" {
		return fmt.Errorf("missing required flags: -f FILE -id ID")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	size, ok := hg.EdgeSize(*edgeID)
	if !ok {
		return fmt.Errorf("edge not found: %s", *edgeID)
	}
	fmt.Println(size)
	return nil
}

func cmdCopy(args []string) error {
	fs := flag.NewFlagSet("copy", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	output := fs.String("o", "", "output file")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *file == "" || *output == "" {
		return fmt.Errorf("missing required flags: -f FILE -o OUTPUT")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	return saveGraph(hg.Copy(), *output)
}
