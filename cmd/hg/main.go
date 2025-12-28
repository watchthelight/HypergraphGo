// Command hg is the Hypergraph CLI.
//
// Usage:
//
//	hg --version           Print version info
//	hg <command> [flags]   Run a subcommand
//	hg repl                Start interactive REPL
//	hg help [command]      Show help
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/watchthelight/HypergraphGo/internal/version"
)

func main() {
	// Global flags (before subcommand)
	ver := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *ver {
		fmt.Printf("hg %s (%s, %s)\n", version.Version, version.Commit, version.Date)
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	subcommand := args[0]
	subArgs := args[1:]

	var err error
	switch subcommand {
	// Core operations
	case "info":
		err = cmdInfo(subArgs)
	case "add-vertex":
		err = cmdAddVertex(subArgs)
	case "remove-vertex":
		err = cmdRemoveVertex(subArgs)
	case "has-vertex":
		err = cmdHasVertex(subArgs)
	case "add-edge":
		err = cmdAddEdge(subArgs)
	case "remove-edge":
		err = cmdRemoveEdge(subArgs)
	case "has-edge":
		err = cmdHasEdge(subArgs)
	case "vertices":
		err = cmdVertices(subArgs)
	case "edges":
		err = cmdEdges(subArgs)
	case "degree":
		err = cmdDegree(subArgs)
	case "edge-size":
		err = cmdEdgeSize(subArgs)
	case "copy":
		err = cmdCopy(subArgs)

	// Transforms
	case "dual":
		err = cmdDual(subArgs)
	case "two-section":
		err = cmdTwoSection(subArgs)
	case "line-graph":
		err = cmdLineGraph(subArgs)

	// Traversal
	case "bfs":
		err = cmdBFS(subArgs)
	case "dfs":
		err = cmdDFS(subArgs)
	case "components":
		err = cmdComponents(subArgs)

	// Algorithms
	case "hitting-set":
		err = cmdHittingSet(subArgs)
	case "transversals":
		err = cmdTransversals(subArgs)
	case "coloring":
		err = cmdColoring(subArgs)

	// I/O
	case "new":
		err = cmdNew(subArgs)
	case "incidence":
		err = cmdIncidence(subArgs)
	case "validate":
		err = cmdValidate(subArgs)

	// Meta
	case "help":
		err = cmdHelp(subArgs)
	case "repl":
		err = cmdREPL(subArgs)

	default:
		fmt.Fprintf(os.Stderr, "hg: unknown command '%s'\n", subcommand)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`hg - Hypergraph CLI

Usage: hg <command> [flags]

Commands:
  Core:
    info          Display hypergraph info
    add-vertex    Add a vertex
    remove-vertex Remove a vertex
    has-vertex    Check vertex existence
    add-edge      Add a hyperedge
    remove-edge   Remove a hyperedge
    has-edge      Check edge existence
    vertices      List all vertices
    edges         List all edges
    degree        Get vertex degree
    edge-size     Get edge size
    copy          Copy hypergraph

  Transforms:
    dual          Compute dual hypergraph
    two-section   Compute 2-section graph
    line-graph    Compute line graph

  Traversal:
    bfs           Breadth-first search
    dfs           Depth-first search
    components    Connected components

  Algorithms:
    hitting-set   Greedy hitting set
    transversals  Minimal transversals
    coloring      Greedy coloring

  I/O:
    new           Create empty hypergraph
    incidence     Print incidence matrix
    validate      Validate JSON file

  Meta:
    help          Show command help
    repl          Interactive mode

Global Flags:
    --version     Print version and exit

Use "hg help <command>" for more information.`)
}
