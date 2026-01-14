# Getting Started with HoTTGo

This guide walks you through installing HoTTGo, using the REPL, type checking terms, and doing a simple proof with tactics.

## Installation

### From Package Manager

**macOS (Homebrew):**
```bash
brew tap watchthelight/tap
brew install hypergraphgo
```

**Windows (Scoop):**
```bash
scoop bucket add hottgo https://github.com/watchthelight/scoop-bucket
scoop install hypergraphgo
```

**Windows (Chocolatey):**
```bash
choco install hypergraphgo
```

### From Binary

Download the latest release from [GitHub Releases](https://github.com/watchthelight/HypergraphGo/releases) and add it to your PATH.

### From Source

```bash
go install github.com/watchthelight/HypergraphGo/cmd/hottgo@latest
```

Verify installation:
```bash
hottgo --version
```

## Using the REPL

Start the interactive REPL:

```bash
hottgo
```

You'll see:
```
hottgo - HoTT Kernel REPL
Commands: :eval EXPR, :synth EXPR, :quit
```

### Synthesize Types

Use `:synth` to find the type of a term:

```
> :synth (Lam A (Lam x 0))
(Pi A Type (Pi _ 0 1))
```

This shows the identity function `λA.λx.x` has type `(A : Type) → A → A`.

### Evaluate Terms

Use `:eval` to normalize a term:

```
> :eval (App (Lam x 0) zero)
zero
```

Beta reduction: `(λx.x) zero` reduces to `zero`.

### Built-in Types

The REPL includes `Nat` and `Bool` primitives:

```
> :synth zero
Nat

> :synth (App succ zero)
Nat

> :synth true
Bool
```

### Exit

Type `:quit` to exit the REPL.

## Type Checking a File

Create a file `myterms.sexpr`:

```lisp
; Identity function
(Lam A (Lam x 0))

; K combinator: λA.λB.λa.λb.a
(Lam A (Lam B (Lam a (Lam b 1))))

; Natural number: succ(succ(zero))
(App succ (App succ zero))
```

Type check the file:

```bash
hottgo --check myterms.sexpr
```

Output:
```
Term 1: OK - (Pi A Type (Pi _ 0 1))
Term 2: OK - (Pi A Type (Pi B Type (Pi _ 1 (Pi _ 0 2))))
Term 3: OK - Nat
```

## S-Expression Syntax

HoTTGo uses S-expression syntax for terms:

| Construct | Syntax | Example |
|-----------|--------|---------|
| Universe | `Type` or `(Type N)` | `Type`, `(Type 1)` |
| Variable | De Bruijn index | `0`, `1`, `2` |
| Lambda | `(Lam BINDER BODY)` | `(Lam x 0)` |
| Application | `(App FN ARG)` | `(App f x)` |
| Pi type | `(Pi BINDER DOMAIN CODOMAIN)` | `(Pi x Nat Type)` |
| Sigma type | `(Sigma BINDER A B)` | `(Sigma x Nat Nat)` |
| Pair | `(Pair FST SND)` | `(Pair zero true)` |
| Projections | `(Fst P)`, `(Snd P)` | `(Fst p)` |
| Identity | `(Id A X Y)` | `(Id Nat zero zero)` |
| Reflexivity | `(Refl A X)` | `(Refl Nat zero)` |
| J eliminator | `(J A C D X Y P)` | `(J Nat ...)` |

**Note:** Variables use de Bruijn indices. `0` refers to the innermost binder, `1` to the next, etc.

## Using the Go Library

### Type Checking

```go
package main

import (
    "fmt"
    "github.com/watchthelight/HypergraphGo/internal/parser"
    "github.com/watchthelight/HypergraphGo/kernel/check"
)

func main() {
    checker := check.NewCheckerWithPrimitives()

    // Parse and type check
    term := parser.MustParse("(Lam A (Lam x 0))")
    ty, err := checker.Synth(nil, check.NoSpan(), term)
    if err != nil {
        fmt.Printf("Type error: %v\n", err)
        return
    }
    fmt.Printf("Type: %s\n", parser.FormatTerm(ty))
    // Output: Type: (Pi A Type (Pi _ 0 1))
}
```

### Normalization

```go
package main

import (
    "fmt"
    "github.com/watchthelight/HypergraphGo/internal/ast"
    "github.com/watchthelight/HypergraphGo/internal/eval"
    "github.com/watchthelight/HypergraphGo/internal/parser"
)

func main() {
    // Build (λx.x) zero
    term := ast.App{
        T: ast.Lam{Binder: "x", Body: ast.Var{Ix: 0}},
        U: ast.Global{Name: "zero"},
    }

    // Normalize
    result := eval.EvalNBE(term)
    fmt.Printf("Normal form: %s\n", parser.FormatTerm(result))
    // Output: Normal form: zero
}
```

### Conversion Checking

```go
package main

import (
    "fmt"
    "github.com/watchthelight/HypergraphGo/internal/core"
    "github.com/watchthelight/HypergraphGo/internal/parser"
)

func main() {
    env := core.NewEnv()

    // (λx.x) zero ≡ zero ?
    term1 := parser.MustParse("(App (Lam x 0) zero)")
    term2 := parser.MustParse("zero")

    equal := core.Conv(env, term1, term2, core.ConvOptions{})
    fmt.Printf("Definitionally equal: %v\n", equal)
    // Output: Definitionally equal: true
}
```

## Proof Tactics

The tactics package provides Ltac-style proof construction.

### Simple Proof

Prove that `(A : Type) → A → A` is inhabited:

```go
package main

import (
    "fmt"
    "github.com/watchthelight/HypergraphGo/internal/ast"
    "github.com/watchthelight/HypergraphGo/internal/parser"
    "github.com/watchthelight/HypergraphGo/tactics"
)

func main() {
    // Goal: (A : Type) → A → A
    goalType := ast.Pi{
        Binder: "A",
        A:      ast.Sort{U: 0},
        B: ast.Pi{
            Binder: "x",
            A:      ast.Var{Ix: 0}, // A
            B:      ast.Var{Ix: 1}, // A (shifted)
        },
    }

    prover := tactics.NewProver(goalType)

    // intro A
    if err := prover.Apply(tactics.Intro("A")); err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    // intro x
    if err := prover.Apply(tactics.Intro("x")); err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    // exact x (use hypothesis x)
    if err := prover.Apply(tactics.Assumption()); err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }

    // Check proof is complete
    if !prover.Done() {
        fmt.Println("Proof incomplete!")
        return
    }

    // Extract proof term
    term, err := prover.Extract()
    if err != nil {
        fmt.Printf("Extraction error: %v\n", err)
        return
    }

    fmt.Printf("Proof term: %s\n", parser.FormatTerm(term))
    // Output: Proof term: (Lam A (Lam x 0))
}
```

### Available Tactics

| Tactic | Description |
|--------|-------------|
| `Intro(name)` | Introduce variable from Pi type |
| `IntroN(names...)` | Introduce multiple variables |
| `Exact(term)` | Provide exact proof term |
| `Assumption()` | Use matching hypothesis |
| `Apply(fn)` | Apply function to goal |
| `Split()` | Split Sigma goal into components |
| `Reflexivity()` | Prove `Id A x x` with `refl` |
| `Rewrite(eq)` | Rewrite using equality |

### Tactic Combinators

```go
// Sequence tactics
tactics.Seq(tactics.Intro("A"), tactics.Intro("x"), tactics.Assumption())

// Try first, else second
tactics.OrElse(tactics.Assumption(), tactics.Reflexivity())

// Try tactic, succeed either way
tactics.Try(tactics.Assumption())

// Repeat until failure
tactics.Repeat(tactics.Intro(""))
```

## Interpreting Errors

### Type Errors

```
error: expected function type, got Nat
```
You tried to apply something that isn't a function.

```
error: type mismatch: expected Bool, got Nat
```
Term has wrong type for context.

### Parse Errors

```
error: unexpected token: expected ')', got EOF
```
Unbalanced parentheses in S-expression.

```
error: unknown term constructor: Lambda
```
Use `Lam` not `Lambda` (see syntax table above).

### Tactic Errors

```
error: goal is not a Pi type
```
`Intro` requires a function type goal.

```
error: no matching hypothesis found
```
`Assumption` found no hypothesis matching the goal.

## Next Steps

- **Examples:** See `examples/` directory for more code samples
- **API Reference:** See `docs/API.md` for detailed API documentation
- **Architecture:** See `docs/architecture.md` to understand the codebase
- **Cubical Types:** See `examples/cubical/` for path types and transport
- **Inductive Types:** See `examples/inductive/` for Nat and Bool eliminators

## Running the Examples

```bash
# Clone the repository
git clone https://github.com/watchthelight/HypergraphGo.git
cd HypergraphGo

# Run examples
go run examples/typecheck/main.go
go run examples/inductive/main.go
go run examples/cubical/main.go
```
