# HypergraphGo

[![CI](https://github.com/watchthelight/hypergraphgo/actions/workflows/ci.yml/badge.svg)](https://github.com/watchthelight/hypergraphgo/actions/workflows/ci.yml)

A work-in-progress implementation of a **native HoTT (Homotopy Type Theory) kernel in Go**, evolving from an earlier hypergraph theory library.

The end goal is a lightweight, sound, and hackable proof kernel with dependent types, inductives, higher inductives, univalence, and cubical features. The hypergraph library components will remain as utilities and examples.

---

## Project Status

- ✅ **Phase 0:** Repository bootstrap, design decisions, package scaffolding, CI checks  
- ✅ **Phase 1:** Core AST, raw AST, resolver, pretty-printer  
- 🔜 **Phase 2:** Normalization-by-Evaluation (NbE) skeleton  
- Future: definitional equality, type checker, inductives, paths, univalence, HITs, elaborator, stdlib.

See [`TODO.txt`](TODO.txt) for the detailed roadmap.

---

## Repo Layout

```
internal/ast/      Core & Raw AST, printer, resolver
internal/core/     Definitional equality (coming soon)
internal/eval/     NbE evaluator (coming soon)
internal/check/    Type checker (coming soon)
internal/kernel/   Trusted kernel boundary (Axiom, Def, Inductive)
pkg/env/           Front-end conveniences, untrusted
cmd/hottgo/        CLI entry point
docs/              DESIGN.md and contributing guidelines
examples/          Hypergraph algorithms and demo programs
```

---

## Quickstart

Clone the repository:

```bash
git clone https://github.com/watchthelight/HypergraphGo.git
cd HypergraphGo
```

Run all tests:

```bash
go test ./...
```

Run the CLI:

```bash
go run ./cmd/hottgo -version
```

---

## Hypergraph Examples

Although the project’s focus is now HoTT, you can still explore the original hypergraph algorithms:

```bash
go run ./examples/basic
go run ./examples/algorithms
```

### CLI usage

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

---

## Algorithms and Complexity

- Greedy hitting set: polynomial-time heuristic  
- Minimal transversals enumeration: exponential, supports cutoffs  
- Greedy coloring: heuristic  
- Dual, 2-section, and line graph transforms  
- Connectivity and traversal via BFS/DFS  

---

## Contributing

- The kernel (`internal/kernel`) must remain panic-free and minimal.  
- All sugar, tactics, or inference layers must live outside the kernel and re-check through it.  
- See [`docs/CONTRIBUTING.md`](docs/CONTRIBUTING.md) for guidelines.

---

## License

MIT License © 2025 [watchthelight](https://github.com/watchthelight)
