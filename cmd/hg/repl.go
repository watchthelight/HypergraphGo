package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/watchthelight/HypergraphGo/hypergraph"
)

type replState struct {
	hg       *hypergraph.Hypergraph[string]
	file     string
	modified bool
}

var errQuit = errors.New("quit")

func cmdREPL(args []string) error {
	fs := flag.NewFlagSet("repl", flag.ExitOnError)
	file := fs.String("f", "", "initial file to load")
	fs.Parse(args)

	state := &replState{
		hg: hypergraph.NewHypergraph[string](),
	}

	if *file != "" {
		hg, err := loadGraph(*file)
		if err != nil {
			return err
		}
		state.hg = hg
		state.file = *file
		fmt.Printf("Loaded %s\n", *file)
	}

	fmt.Println("hg - Hypergraph REPL")
	fmt.Println("Type :help for commands, :quit to exit")
	fmt.Println()

	return replLoop(state)
}

func replLoop(state *replState) error {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("hg> ")
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if err := executeReplCommand(state, line); err != nil {
			if errors.Is(err, errQuit) {
				break
			}
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
		}
	}

	return nil
}

func executeReplCommand(state *replState, line string) error {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return nil
	}
	cmd := parts[0]
	args := parts[1:]

	switch cmd {
	case ":quit", ":q":
		if state.modified {
			fmt.Println("Warning: unsaved changes. Use :save or :quit again to exit.")
			state.modified = false
			return nil
		}
		return errQuit

	case ":help", ":h":
		printReplHelp()
		return nil

	case ":load":
		if len(args) < 1 {
			return fmt.Errorf("usage: :load FILE")
		}
		hg, err := loadGraph(args[0])
		if err != nil {
			return err
		}
		state.hg = hg
		state.file = args[0]
		state.modified = false
		fmt.Printf("Loaded %s\n", args[0])
		return nil

	case ":save":
		file := state.file
		if len(args) > 0 {
			file = args[0]
		}
		if file == "" {
			return fmt.Errorf("no file specified (use :save FILE)")
		}
		if err := saveGraph(state.hg, file); err != nil {
			return err
		}
		state.file = file
		state.modified = false
		fmt.Printf("Saved to %s\n", file)
		return nil

	case ":new":
		if state.modified {
			fmt.Println("Warning: unsaved changes. Use :save first or :new again.")
			state.modified = false
			return nil
		}
		state.hg = hypergraph.NewHypergraph[string]()
		state.file = ""
		state.modified = false
		fmt.Println("Created empty hypergraph.")
		return nil

	case ":info":
		fmt.Printf("Vertices: %d\n", state.hg.NumVertices())
		fmt.Printf("Edges:    %d\n", state.hg.NumEdges())
		fmt.Printf("Empty:    %v\n", state.hg.IsEmpty())
		if state.file != "" {
			fmt.Printf("File:     %s\n", state.file)
		}
		if state.modified {
			fmt.Println("(unsaved changes)")
		}
		return nil

	case "add-vertex":
		if len(args) < 1 {
			return fmt.Errorf("usage: add-vertex VERTEX")
		}
		state.hg.AddVertex(args[0])
		state.modified = true
		return nil

	case "remove-vertex":
		if len(args) < 1 {
			return fmt.Errorf("usage: remove-vertex VERTEX")
		}
		if !state.hg.HasVertex(args[0]) {
			return fmt.Errorf("vertex not found: %s", args[0])
		}
		state.hg.RemoveVertex(args[0])
		state.modified = true
		return nil

	case "has-vertex":
		if len(args) < 1 {
			return fmt.Errorf("usage: has-vertex VERTEX")
		}
		fmt.Println(state.hg.HasVertex(args[0]))
		return nil

	case "add-edge":
		if len(args) < 2 {
			return fmt.Errorf("usage: add-edge ID V1,V2,...")
		}
		members := strings.Split(args[1], ",")
		for i := range members {
			members[i] = strings.TrimSpace(members[i])
		}
		if err := state.hg.AddEdge(args[0], members); err != nil {
			return err
		}
		state.modified = true
		return nil

	case "remove-edge":
		if len(args) < 1 {
			return fmt.Errorf("usage: remove-edge ID")
		}
		if !state.hg.HasEdge(args[0]) {
			return fmt.Errorf("edge not found: %s", args[0])
		}
		state.hg.RemoveEdge(args[0])
		state.modified = true
		return nil

	case "has-edge":
		if len(args) < 1 {
			return fmt.Errorf("usage: has-edge ID")
		}
		fmt.Println(state.hg.HasEdge(args[0]))
		return nil

	case "vertices":
		verts := state.hg.Vertices()
		slices.Sort(verts)
		for _, v := range verts {
			fmt.Println(v)
		}
		return nil

	case "edges":
		edges := state.hg.Edges()
		slices.Sort(edges)
		for _, id := range edges {
			members := state.hg.EdgeMembers(id)
			slices.Sort(members)
			fmt.Printf("%s: %s\n", id, strings.Join(members, ", "))
		}
		return nil

	case "degree":
		if len(args) < 1 {
			return fmt.Errorf("usage: degree VERTEX")
		}
		if !state.hg.HasVertex(args[0]) {
			return fmt.Errorf("vertex not found: %s", args[0])
		}
		fmt.Println(state.hg.VertexDegree(args[0]))
		return nil

	case "edge-size":
		if len(args) < 1 {
			return fmt.Errorf("usage: edge-size ID")
		}
		size, ok := state.hg.EdgeSize(args[0])
		if !ok {
			return fmt.Errorf("edge not found: %s", args[0])
		}
		fmt.Println(size)
		return nil

	case "bfs":
		if len(args) < 1 {
			return fmt.Errorf("usage: bfs VERTEX")
		}
		if !state.hg.HasVertex(args[0]) {
			return fmt.Errorf("vertex not found: %s", args[0])
		}
		result := state.hg.BFS(args[0])
		fmt.Println(strings.Join(result, " "))
		return nil

	case "dfs":
		if len(args) < 1 {
			return fmt.Errorf("usage: dfs VERTEX")
		}
		if !state.hg.HasVertex(args[0]) {
			return fmt.Errorf("vertex not found: %s", args[0])
		}
		result := state.hg.DFS(args[0])
		fmt.Println(strings.Join(result, " "))
		return nil

	case "components":
		components := state.hg.ConnectedComponents()
		for i, comp := range components {
			slices.Sort(comp)
			fmt.Printf("Component %d: %s\n", i+1, strings.Join(comp, ", "))
		}
		return nil

	case "hitting-set":
		result := state.hg.GreedyHittingSet()
		slices.Sort(result)
		fmt.Println(strings.Join(result, " "))
		return nil

	case "coloring":
		coloring := state.hg.GreedyColoring()
		vertices := state.hg.Vertices()
		slices.Sort(vertices)
		for _, v := range vertices {
			fmt.Printf("%s: %d\n", v, coloring[v])
		}
		return nil

	case "dual":
		dual := state.hg.Dual()
		state.hg = dual
		state.modified = true
		fmt.Println("Computed dual (current hypergraph replaced).")
		return nil

	default:
		return fmt.Errorf("unknown command: %s (type :help for commands)", cmd)
	}
}

func printReplHelp() {
	fmt.Println(`REPL Commands:
  :load FILE    Load hypergraph from file
  :save [FILE]  Save to file
  :new          Create new empty hypergraph
  :info         Show hypergraph info
  :help         Show this help
  :quit         Exit REPL

Operations:
  add-vertex V         Add vertex
  remove-vertex V      Remove vertex
  has-vertex V         Check vertex existence
  add-edge ID V1,V2... Add edge
  remove-edge ID       Remove edge
  has-edge ID          Check edge existence
  vertices             List all vertices
  edges                List all edges
  degree V             Get vertex degree
  edge-size ID         Get edge size
  bfs V                BFS from vertex
  dfs V                DFS from vertex
  components           Show connected components
  hitting-set          Compute greedy hitting set
  coloring             Compute greedy coloring
  dual                 Compute dual (replaces current)`)
}
