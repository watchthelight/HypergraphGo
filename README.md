# Hypergraph-Go

A production-quality Go library for hypergraph theory, supporting generic vertex types, advanced algorithms, and CLI tools.

## Overview

Hypergraphs generalize graphs by allowing edges (called hyperedges) to connect any number of vertices. This library provides a flexible, efficient, and idiomatic Go implementation of hypergraphs with rich operations, transforms, and algorithms.

## Installation

```bash
go get github.com/watchthelight/hypergraphgo
```

Replace `watchthelight` and `hypergraphgo` with your GitHub username and repository name.

## Usage

### Basic usage

```go
package main

import (
    "fmt"
    "github.com/watchthelight/hypergraphgo/hypergraph"
)

func main() {
    hg := hypergraph.NewHypergraph[string]()
    hg.AddVertex("A")
    hg.AddVertex("B")
    hg.AddEdge("E1", []string{"A", "B"})
    fmt.Println("Vertices:", hg.Vertices())
    fmt.Println("Edges:", hg.Edges())
}
```

### CLI examples

```bash
hg info -file example.json
hg add-vertex -file example.json -v "A"
hg add-edge -file example.json -id E1 -members "A,B,C"
hg components -file example.json
hg dual -in example.json -out dual.json
hg section -in example.json -out section.json
hg save -out example.json
hg load -in example.json
```

## Algorithms and Complexity

- Greedy hitting set: polynomial time heuristic.
- Minimal transversals enumeration: exponential, supports cutoffs.
- Greedy coloring: heuristic.
- Dual, 2-section, and line graph transforms.
- Connectivity and traversal via BFS/DFS.

## Example

Build a small hypergraph, print degrees, components, dual, and 2-section graphs.

## Versioning

Starting at v0.1.0 with semantic versioning.

## License

MIT License Â© 2025 watchthelight
