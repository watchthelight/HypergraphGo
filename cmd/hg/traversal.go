package main

import (
	"flag"
	"fmt"
	"slices"
	"strings"
)

func cmdBFS(args []string) error {
	fs := flag.NewFlagSet("bfs", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	start := fs.String("start", "", "starting vertex")
	fs.Parse(args)

	if *file == "" || *start == "" {
		return fmt.Errorf("missing required flags: -f FILE -start VERTEX")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	if !hg.HasVertex(*start) {
		return fmt.Errorf("vertex not found: %s", *start)
	}

	result := hg.BFS(*start)
	fmt.Println(strings.Join(result, " "))
	return nil
}

func cmdDFS(args []string) error {
	fs := flag.NewFlagSet("dfs", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	start := fs.String("start", "", "starting vertex")
	fs.Parse(args)

	if *file == "" || *start == "" {
		return fmt.Errorf("missing required flags: -f FILE -start VERTEX")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	if !hg.HasVertex(*start) {
		return fmt.Errorf("vertex not found: %s", *start)
	}

	result := hg.DFS(*start)
	fmt.Println(strings.Join(result, " "))
	return nil
}

func cmdComponents(args []string) error {
	fs := flag.NewFlagSet("components", flag.ExitOnError)
	file := fs.String("f", "", "input hypergraph JSON file")
	fs.Parse(args)

	if *file == "" {
		return fmt.Errorf("missing required flag: -f FILE")
	}

	hg, err := loadGraph(*file)
	if err != nil {
		return err
	}

	components := hg.ConnectedComponents()
	for i, comp := range components {
		slices.Sort(comp)
		fmt.Printf("Component %d: %s\n", i+1, strings.Join(comp, ", "))
	}
	return nil
}
