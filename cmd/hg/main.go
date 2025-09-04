package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/watchthelight/hypergraphgo/hypergraph"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: hg <subcommand> [options]")
		os.Exit(1)
	}
	subcommand := os.Args[1]
	switch subcommand {
	case "info":
		infoCmd()
	case "add-vertex":
		addVertexCmd()
	case "add-edge":
		addEdgeCmd()
	case "components":
		componentsCmd()
	case "dual":
		dualCmd()
	case "section":
		sectionCmd()
	case "save":
		saveCmd()
	case "load":
		loadCmd()
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		os.Exit(1)
	}
}

func loadHypergraph(file string) *hypergraph.Hypergraph[string] {
	f, err := os.Open(file)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	h, err := hypergraph.LoadJSON[string](f)
	if err != nil {
		fmt.Printf("Error loading JSON: %v\n", err)
		os.Exit(1)
	}
	return h
}

func saveHypergraph(h *hypergraph.Hypergraph[string], file string) {
	f, err := os.Create(file)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()
	if err := h.SaveJSON(f); err != nil {
		fmt.Printf("Error saving JSON: %v\n", err)
		os.Exit(1)
	}
}

func infoCmd() {
	fs := flag.NewFlagSet("info", flag.ExitOnError)
	file := fs.String("file", "", "JSON file")
	fs.Parse(os.Args[2:])
	if *file == "" {
		fmt.Println("Missing -file")
		os.Exit(1)
	}
	h := loadHypergraph(*file)
	fmt.Printf("Vertices: %d\n", h.NumVertices())
	fmt.Printf("Edges: %d\n", h.NumEdges())
}

func addVertexCmd() {
	fs := flag.NewFlagSet("add-vertex", flag.ExitOnError)
	file := fs.String("file", "", "JSON file")
	v := fs.String("v", "", "vertex")
	fs.Parse(os.Args[2:])
	if *file == "" || *v == "" {
		fmt.Println("Missing -file or -v")
		os.Exit(1)
	}
	h := loadHypergraph(*file)
	h.AddVertex(*v)
	saveHypergraph(h, *file)
}

func addEdgeCmd() {
	fs := flag.NewFlagSet("add-edge", flag.ExitOnError)
	file := fs.String("file", "", "JSON file")
	id := fs.String("id", "", "edge ID")
	members := fs.String("members", "", "comma-separated members")
	fs.Parse(os.Args[2:])
	if *file == "" || *id == "" || *members == "" {
		fmt.Println("Missing -file, -id, or -members")
		os.Exit(1)
	}
	h := loadHypergraph(*file)
	ms := strings.Split(*members, ",")
	if err := h.AddEdge(*id, ms); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	saveHypergraph(h, *file)
}

func componentsCmd() {
	fs := flag.NewFlagSet("components", flag.ExitOnError)
	file := fs.String("file", "", "JSON file")
	fs.Parse(os.Args[2:])
	if *file == "" {
		fmt.Println("Missing -file")
		os.Exit(1)
	}
	h := loadHypergraph(*file)
	comps := h.ConnectedComponents()
	data, _ := json.MarshalIndent(comps, "", "  ")
	fmt.Println(string(data))
}

func dualCmd() {
	fs := flag.NewFlagSet("dual", flag.ExitOnError)
	in := fs.String("in", "", "input JSON file")
	out := fs.String("out", "", "output JSON file")
	fs.Parse(os.Args[2:])
	if *in == "" || *out == "" {
		fmt.Println("Missing -in or -out")
		os.Exit(1)
	}
	h := loadHypergraph(*in)
	dual := h.Dual()
	saveHypergraph(dual, *out)
}

func sectionCmd() {
	fs := flag.NewFlagSet("section", flag.ExitOnError)
	in := fs.String("in", "", "input JSON file")
	out := fs.String("out", "", "output JSON file")
	fs.Parse(os.Args[2:])
	if *in == "" || *out == "" {
		fmt.Println("Missing -in or -out")
		os.Exit(1)
	}
	h := loadHypergraph(*in)
	section := h.TwoSection()
	// Save as JSON, but since Graph is not Hypergraph, need to marshal manually
	data := map[string]interface{}{
		"vertices": make([]string, 0),
		"edges":    make([]map[string]string, 0),
	}
	for v := range section.vertices {
		data["vertices"] = append(data["vertices"].([]string), v)
	}
	for _, e := range section.edges {
		data["edges"] = append(data["edges"].([]map[string]string), map[string]string{"from": e.From, "to": e.To})
	}
	f, _ := os.Create(*out)
	json.NewEncoder(f).Encode(data)
	f.Close()
}

func saveCmd() {
	fs := flag.NewFlagSet("save", flag.ExitOnError)
	out := fs.String("out", "", "output JSON file")
	fs.Parse(os.Args[2:])
	if *out == "" {
		fmt.Println("Missing -out")
		os.Exit(1)
	}
	h := hypergraph.NewHypergraph[string]()
	saveHypergraph(h, *out)
}

func loadCmd() {
	fs := flag.NewFlagSet("load", flag.ExitOnError)
	in := fs.String("in", "", "input JSON file")
	fs.Parse(os.Args[2:])
	if *in == "" {
		fmt.Println("Missing -in")
		os.Exit(1)
	}
	h := loadHypergraph(*in)
	fmt.Printf("Loaded hypergraph with %d vertices and %d edges\n", h.NumVertices(), h.NumEdges())
}
