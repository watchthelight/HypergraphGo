# Hypergraph-Go
[![CI](https://github.com/watchthelight/hypergraphgo/actions/workflows/ci.yml/badge.svg)](https://github.com/watchthelight/hypergraphgo/actions/workflows/ci.yml)

A production-quality Go library for hypergraph theory, supporting generic vertex types, advanced algorithms, and CLI tools.

## Overview

Hypergraphs generalize graphs by allowing edges (called hyperedges) to connect any number of vertices. This library provides a flexible, efficient, and idiomatic Go implementation of hypergraphs with rich operations, transforms, and algorithms.

## Quickstart

Install the module and try a minimal snippet:

```bash
go get github.com/watchthelight/hypergraphgo
```

```go
package main

import (
    "fmt"
    "github.com/watchthelight/hypergraphgo/hypergraph"
)

func main() {
    hg := hypergraph.NewHypergraph[string]()
    _ = hg.AddEdge("E1", []string{"A", "B", "C"})
    fmt.Println("Vertices:", hg.Vertices())
    fmt.Println("Edges:", hg.Edges())
}
```

Run the examples:

```bash
go run ./examples/basic
go run ./examples/algorithms
```

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

This updates `VERSION`, creates tag `vX.Y.Z`, and pushes the tag to `origin`.

## License

MIT License Â© 2025 watchthelight

