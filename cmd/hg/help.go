package main

import "fmt"

var commandHelp = map[string]string{
	"info": `hg info - Display hypergraph statistics

Usage: hg info -f FILE

Flags:
  -f FILE    Input hypergraph JSON file (required)`,

	"new": `hg new - Create empty hypergraph

Usage: hg new -o FILE

Flags:
  -o FILE    Output file (required)`,

	"validate": `hg validate - Validate JSON file format

Usage: hg validate -f FILE

Flags:
  -f FILE    Input hypergraph JSON file (required)`,

	"add-vertex": `hg add-vertex - Add a vertex

Usage: hg add-vertex -f FILE -v VERTEX [-o OUTPUT]

Flags:
  -f FILE      Input hypergraph JSON file (required)
  -v VERTEX    Vertex to add (required)
  -o OUTPUT    Output file (default: modify in-place)`,

	"remove-vertex": `hg remove-vertex - Remove a vertex

Usage: hg remove-vertex -f FILE -v VERTEX [-o OUTPUT]

Flags:
  -f FILE      Input hypergraph JSON file (required)
  -v VERTEX    Vertex to remove (required)
  -o OUTPUT    Output file (default: modify in-place)`,

	"has-vertex": `hg has-vertex - Check vertex existence

Usage: hg has-vertex -f FILE -v VERTEX

Flags:
  -f FILE      Input hypergraph JSON file (required)
  -v VERTEX    Vertex to check (required)

Prints "true" or "false".`,

	"add-edge": `hg add-edge - Add a hyperedge

Usage: hg add-edge -f FILE -id ID -m MEMBERS [-o OUTPUT]

Flags:
  -f FILE      Input hypergraph JSON file (required)
  -id ID       Edge ID (required)
  -m MEMBERS   Comma-separated member vertices (required)
  -o OUTPUT    Output file (default: modify in-place)`,

	"remove-edge": `hg remove-edge - Remove a hyperedge

Usage: hg remove-edge -f FILE -id ID [-o OUTPUT]

Flags:
  -f FILE      Input hypergraph JSON file (required)
  -id ID       Edge ID to remove (required)
  -o OUTPUT    Output file (default: modify in-place)`,

	"has-edge": `hg has-edge - Check edge existence

Usage: hg has-edge -f FILE -id ID

Flags:
  -f FILE      Input hypergraph JSON file (required)
  -id ID       Edge ID to check (required)

Prints "true" or "false".`,

	"vertices": `hg vertices - List all vertices

Usage: hg vertices -f FILE

Flags:
  -f FILE    Input hypergraph JSON file (required)`,

	"edges": `hg edges - List all edges

Usage: hg edges -f FILE

Flags:
  -f FILE    Input hypergraph JSON file (required)`,

	"degree": `hg degree - Get vertex degree

Usage: hg degree -f FILE -v VERTEX

Flags:
  -f FILE      Input hypergraph JSON file (required)
  -v VERTEX    Vertex (required)`,

	"edge-size": `hg edge-size - Get edge size

Usage: hg edge-size -f FILE -id ID

Flags:
  -f FILE      Input hypergraph JSON file (required)
  -id ID       Edge ID (required)`,

	"copy": `hg copy - Copy hypergraph

Usage: hg copy -f FILE -o OUTPUT

Flags:
  -f FILE      Input hypergraph JSON file (required)
  -o OUTPUT    Output file (required)`,

	"dual": `hg dual - Compute dual hypergraph

Usage: hg dual -f FILE -o OUTPUT

The dual swaps vertices and edges: each original edge becomes a vertex,
and each original vertex becomes an edge containing its incident edges.

Flags:
  -f FILE      Input hypergraph JSON file (required)
  -o OUTPUT    Output file (required)`,

	"two-section": `hg two-section - Compute 2-section graph

Usage: hg two-section -f FILE -o OUTPUT

The 2-section is a graph where vertices are connected if they share
a hyperedge in the original hypergraph.

Flags:
  -f FILE      Input hypergraph JSON file (required)
  -o OUTPUT    Output file (required)`,

	"line-graph": `hg line-graph - Compute line graph

Usage: hg line-graph -f FILE -o OUTPUT

The line graph has edges as vertices, connected if the original edges
shared a vertex.

Flags:
  -f FILE      Input hypergraph JSON file (required)
  -o OUTPUT    Output file (required)`,

	"bfs": `hg bfs - Breadth-first search

Usage: hg bfs -f FILE -start VERTEX

Flags:
  -f FILE        Input hypergraph JSON file (required)
  -start VERTEX  Starting vertex (required)`,

	"dfs": `hg dfs - Depth-first search

Usage: hg dfs -f FILE -start VERTEX

Flags:
  -f FILE        Input hypergraph JSON file (required)
  -start VERTEX  Starting vertex (required)`,

	"components": `hg components - Connected components

Usage: hg components -f FILE

Flags:
  -f FILE    Input hypergraph JSON file (required)`,

	"hitting-set": `hg hitting-set - Greedy hitting set

Usage: hg hitting-set -f FILE

Computes a hitting set using a greedy algorithm. A hitting set contains
at least one vertex from each edge.

Flags:
  -f FILE    Input hypergraph JSON file (required)`,

	"transversals": `hg transversals - Minimal transversals

Usage: hg transversals -f FILE [-max N] [-timeout DURATION]

Enumerates minimal transversals (hitting sets where no proper subset
is also a hitting set).

Flags:
  -f FILE            Input hypergraph JSON file (required)
  -max N             Maximum number of transversals (default: 100)
  -timeout DURATION  Maximum time (default: 10s)`,

	"coloring": `hg coloring - Greedy vertex coloring

Usage: hg coloring -f FILE

Computes a vertex coloring using a greedy algorithm. No two vertices
sharing a hyperedge receive the same color.

Flags:
  -f FILE    Input hypergraph JSON file (required)`,

	"incidence": `hg incidence - Print incidence matrix

Usage: hg incidence -f FILE

Prints the incidence matrix in COO (coordinate) format.

Flags:
  -f FILE    Input hypergraph JSON file (required)`,

	"repl": `hg repl - Interactive mode

Usage: hg repl [-f FILE]

Starts an interactive REPL for hypergraph manipulation.

Flags:
  -f FILE    Initial file to load (optional)

REPL Commands:
  :load FILE    Load hypergraph from file
  :save [FILE]  Save to file
  :new          Create new empty hypergraph
  :info         Show hypergraph info
  :help         Show available commands
  :quit         Exit REPL

Operations (without flags):
  add-vertex V         Add vertex
  remove-vertex V      Remove vertex
  add-edge ID V1,V2... Add edge
  remove-edge ID       Remove edge
  vertices             List vertices
  edges                List edges
  bfs V                BFS from vertex
  dfs V                DFS from vertex
  components           Show components
  hitting-set          Compute hitting set
  coloring             Compute coloring
  dual                 Compute dual (replaces current)`,
}

func cmdHelp(args []string) error {
	if len(args) == 0 {
		printUsage()
		return nil
	}

	cmd := args[0]
	if help, ok := commandHelp[cmd]; ok {
		fmt.Println(help)
		return nil
	}

	return fmt.Errorf("unknown command: %s", cmd)
}
