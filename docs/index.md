# HoTTGo

<p align="center">
  <img src="../assets/hottgo-banner.png" alt="HoTTGo Banner" width="100%">
</p>

**Homotopy Type Theory in Go**

---

## What Is This?

A production-quality Go library combining:

- **Hypergraph theory** — generic vertex types, transforms, algorithms
- **HoTT kernel** — normalization by evaluation, bidirectional typing, identity types, cubical path types
- **CLI tools** — `hg` for hypergraphs, `hottgo` for type checking

---

## Quick Links

| Resource | Description |
|----------|-------------|
| [GitHub](https://github.com/watchthelight/HypergraphGo) | Source code |
| [Architecture](architecture.md) | Kernel design overview |
| [DESIGN.md](https://github.com/watchthelight/HypergraphGo/blob/main/DESIGN.md) | Design decisions |
| [DIAGRAMS.md](https://github.com/watchthelight/HypergraphGo/blob/main/DIAGRAMS.md) | Mermaid architecture diagrams |
| [CHANGELOG.md](https://github.com/watchthelight/HypergraphGo/blob/main/CHANGELOG.md) | Version history |

---

## Installation

```bash
go get github.com/watchthelight/hypergraphgo
```

---

## Current Status

**Phase 4 Complete**: Identity Types + Cubical Path Types

| Phase | Status | Description |
|-------|--------|-------------|
| Phase 0–4 | ✅ | Syntax, NbE, type checking, Id types, paths |
| Phase 5 | ⏳ | Inductives, recursors, positivity |
| Phase 6–10 | ⏳ | Univalence, HITs, tactics, stdlib |

---

## License

MIT License © 2025 [watchthelight](https://github.com/watchthelight)
