# HoTT Kernel API Reference

This document provides a high-level API reference for the HoTT (Homotopy Type Theory) kernel in HypergraphGo. The kernel implements a dependently typed lambda calculus with identity types, cubical features, and inductive types.

## Package Overview

The kernel is organized into several packages:

| Package | Description |
|---------|-------------|
| `internal/ast` | Abstract syntax tree definitions |
| `internal/eval` | Normalization by Evaluation (NbE) |
| `internal/parser` | S-expression parser |
| `kernel/check` | Bidirectional type checker |
| `kernel/ctx` | Typing context |
| `kernel/subst` | Variable substitution |

## Quick Start

```go
import (
    "github.com/watchthelight/HypergraphGo/internal/parser"
    "github.com/watchthelight/HypergraphGo/kernel/check"
)

// Create a type checker with built-in Nat and Bool
checker := check.NewCheckerWithPrimitives()

// Parse a term
term := parser.MustParse("(Lam x 0)")

// Synthesize (infer) its type
ty, err := checker.Synth(nil, check.NoSpan(), term)
if err != nil {
    // Handle type error
}

// Check a term against an expected type
expected := parser.MustParse("(Pi A Type A)")
err = checker.Check(nil, check.NoSpan(), term, expected)
```

## Core Types

### Terms (`internal/ast`)

The `ast.Term` interface represents all term types:

```go
// Base terms
ast.Sort{U: 0}              // Type₀ (universe)
ast.Var{Ix: 0}              // de Bruijn variable
ast.Global{Name: "foo"}     // global constant

// Function types
ast.Pi{Binder: "x", A: domain, B: codomain}
ast.Lam{Binder: "x", Body: body}
ast.App{T: fn, U: arg}

// Pair types
ast.Sigma{Binder: "x", A: fstType, B: sndType}
ast.Pair{Fst: a, Snd: b}
ast.Fst{P: pair}
ast.Snd{P: pair}

// Identity types
ast.Id{A: ty, X: left, Y: right}
ast.Refl{A: ty, X: point}
ast.J{A: ty, C: motive, D: base, X: left, Y: right, P: proof}
```

### Cubical Terms

```go
// Interval
ast.Interval{}    // I (interval type)
ast.I0{}          // i0 (left endpoint)
ast.I1{}          // i1 (right endpoint)
ast.IVar{Ix: 0}   // interval variable

// Paths
ast.Path{A: ty, X: left, Y: right}
ast.PathP{A: tyFamily, X: left, Y: right}
ast.PathLam{Binder: "i", Body: body}
ast.PathApp{P: path, R: interval}

// Transport
ast.Transport{A: tyFamily, E: element}
```

## Type Checking (`kernel/check`)

### Creating a Checker

```go
// Empty environment
checker := check.NewChecker(nil)

// With custom globals
env := check.NewGlobalEnv()
env.DefineConst("myConst", myType, myValue)
checker := check.NewChecker(env)

// With built-in Nat and Bool
checker := check.NewCheckerWithPrimitives()

// With eta-equality enabled
checker := check.NewCheckerWithEta(nil)
```

### Type Synthesis and Checking

```go
// Synthesize a type (inference)
inferredType, err := checker.Synth(ctx, span, term)

// Check against expected type
err := checker.Check(ctx, span, term, expectedType)

// Check that a term is a valid type
level, err := checker.CheckIsType(ctx, span, term)

// Infer type and check against expected
err := checker.InferAndCheck(ctx, span, term, expectedType)
```

### Contexts

```go
import tyctx "github.com/watchthelight/HypergraphGo/kernel/ctx"

// Create empty context
ctx := &tyctx.Ctx{}

// Add binding
ctx.Extend("x", xType)

// Look up variable (de Bruijn index)
ty, ok := ctx.LookupVar(0)  // most recent binding

// Remove most recent binding
ctx = ctx.Drop()
```

### Span (Source Location)

```go
// No source location (for generated terms)
span := check.NoSpan()

// With source location
span := check.NewSpan("file.hott", startLine, startCol, endLine, endCol)
```

### Type Errors

```go
err := checker.Synth(ctx, span, term)
if err != nil {
    fmt.Printf("Error at %s: %s\n", err.Span, err.Message)
}
```

## Evaluation (`internal/eval`)

### Normalization

```go
import "github.com/watchthelight/HypergraphGo/internal/eval"

// Full normalization (eval + reify)
normalForm := eval.EvalNBE(term)

// Step by step
env := &eval.Env{}
value := eval.Eval(env, term)
normalForm := eval.Reify(value)
```

### Debug Mode

```go
// Enable strict error handling (panics on internal errors)
eval.DebugMode = true
// Or set environment variable: HOTTGO_DEBUG=1
```

## Parsing (`internal/parser`)

### Basic Parsing

```go
import "github.com/watchthelight/HypergraphGo/internal/parser"

// Parse a term (returns error on failure)
term, err := parser.ParseTerm("(Lam x 0)")

// Parse a term (panics on failure)
term := parser.MustParse("(Lam x 0)")

// Parse multiple terms
terms, err := parser.ParseMultiple("Type (Lam x 0)")
```

### Formatting

```go
// Convert term to S-expression string
str := parser.FormatTerm(term)
```

### Grammar Summary

See `internal/parser/grammar.go` for complete BNF grammar.

```
; Atoms
0, 1, 2, ...           ; de Bruijn variables
Type, Type0, Type1     ; universes
foo, Bar               ; globals

; Standard forms
(Pi x A B)             ; function type
(Lam x body)           ; lambda
(App f arg)            ; application
(Sigma x A B)          ; pair type
(Pair a b)             ; pair
(Id A x y)             ; identity type
(Refl A x)             ; reflexivity
(J A C d x y p)        ; J eliminator

; Cubical forms
(Path A x y)           ; path type
(PathP A x y)          ; dependent path
(PathLam i body)       ; path abstraction
(PathApp p r)          ; path application
(Transport A e)        ; transport
```

## Substitution (`kernel/subst`)

```go
import "github.com/watchthelight/HypergraphGo/kernel/subst"

// Shift variables >= cutoff by d
shifted := subst.Shift(d, cutoff, term)

// Substitute s for variable j
substituted := subst.Subst(j, s, term)

// Interval variable operations (cubical)
iShifted := subst.IShift(d, cutoff, term)
iSubstituted := subst.ISubst(j, s, term)
```

## Global Environment

### Built-in Types

When using `NewCheckerWithPrimitives()`:

```
Nat  : Type
zero : Nat
succ : Nat → Nat

Bool  : Type
true  : Bool
false : Bool

natElim  : (P : Nat → Type) → P zero → ((n : Nat) → P n → P (succ n)) → (n : Nat) → P n
boolElim : (P : Bool → Type) → P true → P false → (b : Bool) → P b
```

### Custom Definitions

```go
env := check.NewGlobalEnv()

// Define a type
env.DefineType("MyType", ast.Sort{U: 0})

// Define a constant with type and value
env.DefineConst("myConst", myType, myValue)

// Look up type
ty := env.LookupType("myConst")

// Look up value
val := env.LookupValue("myConst")
```

## Inductive Types

### Checking Positivity

```go
import "github.com/watchthelight/HypergraphGo/kernel/check"

// Check single inductive
err := check.CheckPositivity("List", constructors)

// Check mutual inductives
err := check.CheckMutualPositivity(indNames, constructorMap)
```

### Recursor Generation

```go
ind := &check.Inductive{
    Name:         "Nat",
    Type:         natType,
    Constructors: []check.Constructor{...},
    Eliminator:   "natElim",
}

recursorType := check.GenerateRecursorTypeSimple(ind)
```

## Examples

See the `examples/` directory for complete working examples:

- `examples/typecheck/` - Basic type checking
- `examples/cubical/` - Cubical type theory features
- `examples/inductive/` - Inductive types and eliminators

## References

- Pierce, B. C. (2002). *Types and Programming Languages*. MIT Press.
- The Univalent Foundations Program (2013). *Homotopy Type Theory: Univalent Foundations of Mathematics*.
- Cohen, C. et al. (2018). *Cubical Type Theory: A Constructive Interpretation of the Univalence Axiom*. TYPES 2015.
