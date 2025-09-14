# Hypergraph-Go
[![CI (Windows)](https://github.com/watchthelight/HypergraphGo/actions/workflows/ci-windows.yml/badge.svg?branch=main)](https://github.com/watchthelight/HypergraphGo/actions/workflows/ci-windows.yml)
[![CI (Linux)](https://github.com/watchthelight/HypergraphGo/actions/workflows/ci-linux.yml/badge.svg?branch=main)](https://github.com/watchthelight/HypergraphGo/actions/workflows/ci-linux.yml)
[![release](https://img.shields.io/github/v/release/watchthelight/HypergraphGo?sort=semver)](https://github.com/watchthelight/HypergraphGo/releases)
[![AUR](https://img.shields.io/aur/version/hypergraphgo?label=AUR&logo=archlinux)](https://aur.archlinux.org/packages/hypergraphgo)
[![APT](https://img.shields.io/badge/apt-hypergraphgo-blue?logo=debian)](https://cloudsmith.io/~watchthelight/repos/hypergraphgo/packages/)
[![choco](https://img.shields.io/chocolatey/v/hypergraphgo?label=choco)](https://community.chocolatey.org/packages/hypergraphgo)

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

### Install via APT (Cloudsmith)
```bash
curl -1sLf 'https://dl.cloudsmith.io/public/watchthelight/hypergraphgo/setup.deb.sh' | sudo -E bash
sudo apt install hypergraphgo
```

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

## Releasing

Create and push a new tag based on the VERSION file.

Linux/macOS/WSL:

```bash
./scripts/release.sh patch   # or minor | major | 1.2.3
```

Windows PowerShell:

```powershell
./scripts/release.ps1 patch  # or minor | major | 1.2.3
```

This updates `VERSION`, creates tag `vX.Y.Z`, and pushes the tag to `origin`.

## License

MIT License © 2025 watchthelight


