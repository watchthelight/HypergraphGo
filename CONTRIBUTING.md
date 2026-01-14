# Contributing to HoTTGo

We accept contributions. We have standards.

## Ground Rules

1. **Small PRs.** One logical change per PR. If you're touching 30 files, you're probably doing it wrong.

2. **Tests required.** If it's not tested, it doesn't exist. No exceptions.

3. **CHANGELOG entry required.** Update `CHANGELOG.md` under `[Unreleased]`. Use [Keep a Changelog](https://keepachangelog.com/) format.

4. **Kernel boundaries are sacred.** The kernel (`kernel/`) is minimal, total, and panic-free. Don't blur the lines between trusted and untrusted code.

5. **If your PR is a hot take, open an issue first.** It saves everyone grief.

## Process

1. Fork the repo
2. Create a branch (`git checkout -b feat/my-feature`)
3. Make your changes
4. Run tests: `go test ./...`
5. Run lints: `golangci-lint run`
6. Update CHANGELOG
7. Submit PR

## Code Standards

- **Go 1.25+** required
- **`go fmt`** your code
- **No panics** in kernel code — return typed errors
- **De Bruijn indices** in core AST — no raw names
- **Deterministic output** — sort maps before iteration

## What We Won't Accept

- PRs without tests
- PRs that break existing tests
- PRs that blur kernel boundaries
- Cosmetic-only changes without justification
- "Improvements" that add complexity without clear benefit

## Architecture & Boundaries

See [docs/architecture.md](docs/architecture.md) for full details. Key rules for contributors:

### Kernel Boundary Rules

The **trusted kernel** must remain minimal and correct:

| Package | Boundary | Can Import |
|---------|----------|------------|
| `kernel/check` | Trusted | `kernel/*`, `internal/ast`, `internal/eval`, `internal/core` |
| `kernel/ctx` | Trusted | `internal/ast` only |
| `kernel/subst` | Trusted | `internal/ast` only |
| `internal/eval` | Trusted | `internal/ast`, `internal/core` |
| `internal/core` | Trusted | `internal/ast`, `internal/eval` |

**Never:**
- Import `internal/parser` from kernel packages
- Import `cmd/*` from any library package
- Import `tactics` from kernel or internal packages
- Add panics to trusted kernel code

### Where to Add Things

| Adding... | Location | Notes |
|-----------|----------|-------|
| New AST term type | `internal/ast/term.go` | Implement `isCoreTerm()` |
| Cubical term type | `internal/ast/term_cubical.go` | Add cubical marker method |
| Type checking rule | `kernel/check/bidir.go` | Or `bidir_cubical.go` for cubical |
| Evaluation rule | `internal/eval/nbe.go` | Or `nbe_cubical.go` |
| New tactic | `tactics/core.go` | Follow existing patterns |
| Tactic combinator | `tactics/combinators.go` | Must be composable |
| Parser extension | `internal/parser/sexpr.go` | Or `sexpr_cubical.go` |
| New primitive type | `kernel/check/primitives.go` | Via `GlobalEnv.AddPrimitive` |
| Inductive type | Use `GlobalEnv.AddInductive` | Positivity is checked |

### Tactics System Isolation

The `tactics/` package is **not** imported by kernel or internal packages. This is intentional:
- Tactics build terms; kernel checks them
- Bugs in tactics can't compromise soundness
- New tactics don't require kernel changes

## Testing Standards

### Running Tests

```bash
# All tests
go test ./...

# With race detector (slower, catches data races)
go test -race ./...

# Specific package
go test ./internal/eval/...

# Verbose output
go test -v ./...

# Run specific test
go test -run TestConv ./internal/core/...
```

### Running Fuzz Tests

```bash
# Start fuzzer (runs indefinitely until stopped)
go test -fuzz=FuzzParseTerm ./internal/parser/...

# Run for specific duration
go test -fuzz=FuzzParseTerm -fuzztime=30s ./internal/parser/...

# Run all fuzz tests briefly
go test -fuzz=. -fuzztime=10s ./...
```

### Running Benchmarks

```bash
# All benchmarks
go test -bench=. ./...

# With memory stats
go test -bench=. -benchmem ./internal/core/...

# Specific benchmark
go test -bench=BenchmarkConv ./internal/core/...
```

### Writing Tests

1. **Unit tests** go in `*_test.go` next to the code
2. **Test function names**: `TestFunctionName_Scenario`
3. **Table-driven tests** preferred for multiple cases
4. **Use subtests** for related scenarios: `t.Run("case", ...)`

```go
func TestConv_BetaReduction(t *testing.T) {
    tests := []struct {
        name string
        t1, t2 ast.Term
        want bool
    }{
        {"id applied", app(id, x), x, true},
        {"different terms", x, y, false},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Conv(env, tt.t1, tt.t2, opts)
            if got != tt.want {
                t.Errorf("Conv() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Coverage

Check coverage locally:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Target: maintain or improve package coverage. Current baseline is ~68%.

## Commit Messages

Be clear. Be concise. Start with a verb.

```
feat(kernel): add J eliminator for identity types
fix(nbe): correct reification for nested Pi types
docs: update roadmap with Phase 5 status
```

## Questions?

Open an issue. Don't email.
