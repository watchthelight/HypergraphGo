# HoTT Kernel Architecture Diagrams

This document provides visual representations of the HypergraphGo HoTT (Homotopy Type Theory) kernel architecture, data flow, and key algorithms.

---

## MASTER DIAGRAM: Complete HoTT Kernel Architecture

```mermaid
flowchart TB
    subgraph InputLayer["INPUT LAYER"]
        direction LR
        SourceTerm["Source Term<br/>(ast.Term)"]
        TypingCtx["Typing Context<br/>(kernel/ctx.Ctx)"]
        GlobalEnv["Global Environment<br/>(kernel/check.GlobalEnv)"]
    end

    subgraph KernelCheck["KERNEL: TYPE CHECKER (kernel/check)"]
        direction TB

        subgraph CheckerCore["Checker Core"]
            Checker["Checker{globals, conv}"]
            PublicAPI["Public API:<br/>Synth() / Check() / CheckIsType()"]
        end

        subgraph SynthesisMode["SYNTHESIS MODE (Infer Type)"]
            direction TB
            SynthDispatch{"synth()<br/>dispatch"}

            subgraph AtomicSynth["Atomic Terms"]
                SynthVar["synthVar<br/>ctx.LookupVar(ix)<br/>+ Shift"]
                SynthSort["synthSort<br/>Sort U → Sort U+1"]
                SynthGlobal["synthGlobal<br/>globals.LookupType"]
            end

            subgraph FunctionSynth["Function Types"]
                SynthPi["synthPi<br/>checkIsType(A,B)<br/>→ Sort max(U,V)"]
                SynthLam["synthLam<br/>(annotated only)<br/>→ Pi type"]
                SynthApp["synthApp<br/>synth(f) → Pi<br/>check(u,A)<br/>→ B[u/x]"]
            end

            subgraph ProductSynth["Product Types"]
                SynthSigma["synthSigma<br/>checkIsType(A,B)<br/>→ Sort max(U,V)"]
                SynthFst["synthFst<br/>synth(p) → Sigma<br/>→ A"]
                SynthSnd["synthSnd<br/>synth(p) → Sigma<br/>→ B[fst p/x]"]
            end

            subgraph IdentitySynth["Identity Types"]
                SynthId["synthId<br/>checkIsType(A)<br/>check(x,y : A)<br/>→ Sort U"]
                SynthRefl["synthRefl<br/>checkIsType(A)<br/>check(x : A)<br/>→ Id A x x"]
                SynthJ["synthJ<br/>check motive C<br/>check base d<br/>check proof p<br/>→ C y p"]
            end

            subgraph ControlSynth["Control Flow"]
                SynthLet["synthLet<br/>synth/check val<br/>extend ctx<br/>→ bodyTy[val/x]"]
            end
        end

        subgraph CheckingMode["CHECKING MODE (Verify Type)"]
            direction TB
            CheckDispatch{"check()<br/>dispatch"}

            CheckLam["checkLam<br/>(unannotated)<br/>ensurePi(expected)<br/>check(body, B)"]
            CheckPair["checkPair<br/>ensureSigma(expected)<br/>check(fst, A)<br/>check(snd, B[fst/x])"]
            CheckBySynth["checkBySynth<br/>synth(term)<br/>conv(inferred, expected)"]
        end

        subgraph Helpers["Helper Functions"]
            CheckIsType["checkIsType<br/>synth → ensureSort<br/>→ Level"]
            EnsurePi["ensurePi<br/>whnf → expect Pi"]
            EnsureSigma["ensureSigma<br/>whnf → expect Sigma"]
            EnsureSort["ensureSort<br/>whnf → expect Sort"]
            WHNF["whnf<br/>→ EvalNBE"]
            ConvCheck["conv<br/>→ core.Conv"]
            MkJMotiveType["mkJMotiveType<br/>Π(y:A).Π(p:Id A x y).Type"]
        end
    end

    subgraph KernelSubst["KERNEL: SUBSTITUTION (kernel/subst)"]
        direction LR
        Shift["Shift(d, c, term)<br/>Increment free vars by d"]
        Subst["Subst(ix, repl, term)<br/>Replace var ix with repl"]
    end

    subgraph InternalCore["INTERNAL: CONVERSION (internal/core)"]
        direction TB
        Conv["Conv(env, t, u, opts)<br/>Definitional Equality"]
        AlphaEq["AlphaEq(a, b)<br/>Structural Equality"]
        EtaEqual["etaEqual(a, b)<br/>η-Equality (optional)"]
        ShiftTerm["shiftTerm(d, c, term)<br/>Variable Shifting"]
    end

    subgraph InternalEval["INTERNAL: NbE ENGINE (internal/eval)"]
        direction TB

        subgraph EvalCore["Evaluation Core"]
            EvalNBE["EvalNBE(term)<br/>Full Normalization"]
            Eval["Eval(env, term)<br/>Syntax → Semantics"]
            Reify["Reify(value)<br/>Semantics → Syntax"]
        end

        subgraph EvalDispatch["Eval Dispatch"]
            EvalVar["Var → env.Lookup(ix)"]
            EvalGlobal["Global → VGlobal{name}"]
            EvalSort["Sort → VSort{level}"]
            EvalLam["Lam → VLam{Closure{env,body}}"]
            EvalApp["App → Apply(Eval(f), Eval(u))"]
            EvalPair["Pair → VPair{Eval(fst), Eval(snd)}"]
            EvalFst["Fst → doFst(Eval(p))"]
            EvalSnd["Snd → doSnd(Eval(p))"]
            EvalPi["Pi → VPi{Eval(A), Closure{env,B}}"]
            EvalSigma["Sigma → VSigma{Eval(A), Closure{env,B}}"]
            EvalLet["Let → Eval(env.Extend(val), body)"]
            EvalId["Id → VId{Eval(A,X,Y)}"]
            EvalRefl["Refl → VRefl{Eval(A,X)}"]
            EvalJ["J → evalJ(...)"]
        end

        subgraph ApplyCore["Application & Projections"]
            Apply["Apply(fun, arg)"]
            BetaReduce["VLam: Beta Reduction<br/>Eval(env.Extend(arg), body)"]
            NeutralApp["VNeutral: Extend Spine<br/>append(sp, arg)"]
            DoFst["doFst(pair)<br/>VPair → Fst<br/>VNeutral → extend sp"]
            DoSnd["doSnd(pair)<br/>VPair → Snd<br/>VNeutral → extend sp"]
        end

        subgraph JElim["J Elimination"]
            EvalJFn["evalJ(a,c,d,x,y,p)"]
            JCompRule["p = VRefl?<br/>YES → return d<br/>NO → VNeutral{J, spine}"]
        end

        subgraph ReifyDispatch["Reify Dispatch"]
            ReifyNeutral["VNeutral → reifyNeutral"]
            ReifyLam["VLam → Lam{Reify(Apply(lam,fresh))}"]
            ReifyPair["VPair → Pair{Reify(fst,snd)}"]
            ReifySort["VSort → Sort{level}"]
            ReifyGlobal["VGlobal → Global{name}"]
            ReifyPi["VPi → Pi{Reify(A), Reify(Apply(B,fresh))}"]
            ReifySigma["VSigma → Sigma{Reify(A), Reify(Apply(B,fresh))}"]
            ReifyId["VId → Id{Reify(A,X,Y)}"]
            ReifyRefl["VRefl → Refl{Reify(A,X)}"]
        end

        subgraph SemanticDomain["SEMANTIC DOMAIN (Values)"]
            direction LR
            VSort["VSort{Level}"]
            VGlobal["VGlobal{Name}"]
            VLam["VLam{*Closure}"]
            VPi["VPi{A, *Closure}"]
            VSigma["VSigma{A, *Closure}"]
            VPair["VPair{Fst, Snd}"]
            VNeutral["VNeutral{Neutral}"]
            VId["VId{A, X, Y}"]
            VRefl["VRefl{A, X}"]
        end

        subgraph NeutralTerms["Neutral Terms"]
            Neutral["Neutral{Head, Spine}"]
            HeadVar["Head.Var: int"]
            HeadGlob["Head.Glob: string"]
            Spine["Spine: []Value"]
        end

        subgraph Environment["Environment"]
            Env["Env{Bindings: []Value}"]
            EnvExtend["Extend(v) → prepend v"]
            EnvLookup["Lookup(ix) → Bindings[ix]"]
        end

        subgraph Closure["Closure"]
            ClosureStruct["Closure{*Env, ast.Term}"]
            ClosureApply["Apply: Eval(env.Extend(arg), term)"]
        end
    end

    subgraph InternalAST["INTERNAL: AST (internal/ast)"]
        direction TB

        subgraph TermInterface["Term Interface"]
            Term["Term interface<br/>isCoreTerm()"]
        end

        subgraph TermTypes["Term Types"]
            direction LR

            subgraph Atomic["Atomic"]
                TSort["Sort{U}"]
                TVar["Var{Ix}"]
                TGlobal["Global{Name}"]
            end

            subgraph Functions["Functions (Π)"]
                TPi["Pi{Binder,A,B}"]
                TLam["Lam{Binder,Ann,Body}"]
                TApp["App{T,U}"]
            end

            subgraph Products["Products (Σ)"]
                TSigma["Sigma{Binder,A,B}"]
                TPair["Pair{Fst,Snd}"]
                TFst["Fst{P}"]
                TSnd["Snd{P}"]
            end

            subgraph Identity["Identity (Id)"]
                TId["Id{A,X,Y}"]
                TRefl["Refl{A,X}"]
                TJ["J{A,C,D,X,Y,P}"]
            end

            subgraph Control["Control"]
                TLet["Let{Binder,Ann,Val,Body}"]
            end
        end

        subgraph Printing["Pretty Printing"]
            Sprint["Sprint(term) → string"]
            Write["write(buf, term)"]
        end
    end

    subgraph KernelCtx["KERNEL: CONTEXT (kernel/ctx)"]
        direction TB
        Ctx["Ctx{Tele: []Binding}"]
        CtxExtend["Extend(name, type)<br/>prepend binding"]
        CtxLookup["LookupVar(ix)<br/>→ type at index"]
        CtxDrop["Drop()<br/>remove newest"]
        Binding["Binding{Name, Type}"]
    end

    subgraph OutputLayer["OUTPUT LAYER"]
        direction LR
        InferredType["Inferred Type<br/>(ast.Term)"]
        TypeError["Type Error<br/>(*TypeError)"]
        NormalForm["Normal Form<br/>(ast.Term)"]
    end

    %% INPUT CONNECTIONS
    SourceTerm --> Checker
    TypingCtx --> Checker
    GlobalEnv --> Checker

    %% CHECKER CORE FLOW
    Checker --> PublicAPI
    PublicAPI --> SynthDispatch
    PublicAPI --> CheckDispatch

    %% SYNTHESIS DISPATCH
    SynthDispatch --> SynthVar
    SynthDispatch --> SynthSort
    SynthDispatch --> SynthGlobal
    SynthDispatch --> SynthPi
    SynthDispatch --> SynthLam
    SynthDispatch --> SynthApp
    SynthDispatch --> SynthSigma
    SynthDispatch --> SynthFst
    SynthDispatch --> SynthSnd
    SynthDispatch --> SynthId
    SynthDispatch --> SynthRefl
    SynthDispatch --> SynthJ
    SynthDispatch --> SynthLet

    %% CHECK DISPATCH
    CheckDispatch --> CheckLam
    CheckDispatch --> CheckPair
    CheckDispatch --> CheckBySynth

    %% HELPERS CONNECTIONS
    SynthPi --> CheckIsType
    SynthSigma --> CheckIsType
    SynthId --> CheckIsType
    SynthRefl --> CheckIsType
    SynthJ --> CheckIsType
    SynthJ --> MkJMotiveType
    SynthApp --> EnsurePi
    SynthFst --> EnsureSigma
    SynthSnd --> EnsureSigma
    CheckLam --> EnsurePi
    CheckPair --> EnsureSigma
    CheckIsType --> EnsureSort
    EnsurePi --> WHNF
    EnsureSigma --> WHNF
    EnsureSort --> WHNF
    CheckBySynth --> ConvCheck

    %% WHNF TO NBE
    WHNF --> EvalNBE

    %% SUBST CONNECTIONS
    SynthVar --> Shift
    SynthApp --> Subst
    SynthSnd --> Subst
    SynthLet --> Subst
    CheckPair --> Subst
    MkJMotiveType --> Shift

    %% CORE CONNECTIONS
    ConvCheck --> Conv
    Conv --> Eval
    Conv --> Reify
    Conv --> AlphaEq
    Conv --> EtaEqual
    AlphaEq --> ShiftTerm

    %% EVALNBE FLOW
    EvalNBE --> Eval
    EvalNBE --> Reify

    %% EVAL DISPATCH
    Eval --> EvalVar
    Eval --> EvalGlobal
    Eval --> EvalSort
    Eval --> EvalLam
    Eval --> EvalApp
    Eval --> EvalPair
    Eval --> EvalFst
    Eval --> EvalSnd
    Eval --> EvalPi
    Eval --> EvalSigma
    Eval --> EvalLet
    Eval --> EvalId
    Eval --> EvalRefl
    Eval --> EvalJ

    %% APPLY CONNECTIONS
    EvalApp --> Apply
    Apply --> BetaReduce
    Apply --> NeutralApp
    EvalFst --> DoFst
    EvalSnd --> DoSnd

    %% J ELIM
    EvalJ --> EvalJFn
    EvalJFn --> JCompRule

    %% REIFY DISPATCH
    Reify --> ReifyNeutral
    Reify --> ReifyLam
    Reify --> ReifyPair
    Reify --> ReifySort
    Reify --> ReifyGlobal
    Reify --> ReifyPi
    Reify --> ReifySigma
    Reify --> ReifyId
    Reify --> ReifyRefl

    %% VALUE TYPES
    EvalSort --> VSort
    EvalGlobal --> VGlobal
    EvalLam --> VLam
    EvalPi --> VPi
    EvalSigma --> VSigma
    EvalPair --> VPair
    NeutralApp --> VNeutral
    EvalId --> VId
    EvalRefl --> VRefl

    %% NEUTRAL STRUCTURE
    VNeutral --> Neutral
    Neutral --> HeadVar
    Neutral --> HeadGlob
    Neutral --> Spine

    %% ENVIRONMENT
    EvalVar --> EnvLookup
    EvalLet --> EnvExtend
    BetaReduce --> EnvExtend
    Env --> EnvExtend
    Env --> EnvLookup

    %% CLOSURE
    VLam --> ClosureStruct
    VPi --> ClosureStruct
    VSigma --> ClosureStruct
    ClosureStruct --> ClosureApply
    ClosureApply --> Eval

    %% CONTEXT
    SynthVar --> CtxLookup
    SynthPi --> CtxExtend
    SynthSigma --> CtxExtend
    SynthLam --> CtxExtend
    SynthLet --> CtxExtend
    CheckLam --> CtxExtend
    Ctx --> Binding

    %% AST TERM TYPES
    Term --> TSort
    Term --> TVar
    Term --> TGlobal
    Term --> TPi
    Term --> TLam
    Term --> TApp
    Term --> TSigma
    Term --> TPair
    Term --> TFst
    Term --> TSnd
    Term --> TId
    Term --> TRefl
    Term --> TJ
    Term --> TLet

    %% OUTPUT
    SynthDispatch --> InferredType
    CheckDispatch --> TypeError
    Reify --> NormalForm

    %% DARK COLOR SCHEME - All black/dark gray
    style InputLayer fill:#1a1a1a,stroke:#444,color:#fff
    style KernelCheck fill:#1a1a1a,stroke:#444,color:#fff
    style KernelSubst fill:#1a1a1a,stroke:#444,color:#fff
    style KernelCtx fill:#1a1a1a,stroke:#444,color:#fff
    style InternalCore fill:#1a1a1a,stroke:#444,color:#fff
    style InternalEval fill:#1a1a1a,stroke:#444,color:#fff
    style InternalAST fill:#1a1a1a,stroke:#444,color:#fff
    style OutputLayer fill:#1a1a1a,stroke:#444,color:#fff

    style CheckerCore fill:#222,stroke:#555,color:#fff
    style SynthesisMode fill:#222,stroke:#555,color:#fff
    style CheckingMode fill:#222,stroke:#555,color:#fff
    style Helpers fill:#222,stroke:#555,color:#fff

    style AtomicSynth fill:#2a2a2a,stroke:#666,color:#fff
    style FunctionSynth fill:#2a2a2a,stroke:#666,color:#fff
    style ProductSynth fill:#2a2a2a,stroke:#666,color:#fff
    style IdentitySynth fill:#2a2a2a,stroke:#666,color:#fff
    style ControlSynth fill:#2a2a2a,stroke:#666,color:#fff

    style EvalCore fill:#222,stroke:#555,color:#fff
    style EvalDispatch fill:#222,stroke:#555,color:#fff
    style ApplyCore fill:#222,stroke:#555,color:#fff
    style JElim fill:#222,stroke:#555,color:#fff
    style ReifyDispatch fill:#222,stroke:#555,color:#fff
    style SemanticDomain fill:#2a2a2a,stroke:#666,color:#fff
    style NeutralTerms fill:#2a2a2a,stroke:#666,color:#fff
    style Environment fill:#2a2a2a,stroke:#666,color:#fff
    style Closure fill:#2a2a2a,stroke:#666,color:#fff

    style TermInterface fill:#222,stroke:#555,color:#fff
    style TermTypes fill:#2a2a2a,stroke:#666,color:#fff
    style Atomic fill:#333,stroke:#777,color:#fff
    style Functions fill:#333,stroke:#777,color:#fff
    style Products fill:#333,stroke:#777,color:#fff
    style Identity fill:#333,stroke:#777,color:#fff
    style Control fill:#333,stroke:#777,color:#fff
    style Printing fill:#333,stroke:#777,color:#fff
```

---

## Table of Contents

1. [Package Architecture](#1-package-architecture)
2. [Term Type Hierarchy](#2-term-type-hierarchy)
3. [Value Type Hierarchy (NbE)](#3-value-type-hierarchy-nbe)
4. [Bidirectional Type Checking Flow](#4-bidirectional-type-checking-flow)
5. [Normalization by Evaluation (NbE) Pipeline](#5-normalization-by-evaluation-nbe-pipeline)
6. [Eval Function Flow](#6-eval-function-flow)
7. [Apply Function (Beta Reduction)](#7-apply-function-beta-reduction)
8. [Reify Function Flow](#8-reify-function-flow)
9. [J Elimination (Path Induction)](#9-j-elimination-path-induction)
10. [Conversion Checking](#10-conversion-checking)
11. [Context and Environment Management](#11-context-and-environment-management)
12. [Complete Type Checking Pipeline](#12-complete-type-checking-pipeline)

---

## 1. Package Architecture

```mermaid
flowchart TB
    subgraph cmd["Command Layer"]
        hg["cmd/hg<br/>CLI Entry Point"]
        hottgo["cmd/hottgo<br/>CLI Entry Point"]
    end

    subgraph kernel["Kernel Layer (Trusted Core)"]
        check["kernel/check<br/>Bidirectional Type Checker"]
        ctx["kernel/ctx<br/>Typing Context"]
        subst["kernel/subst<br/>Substitution Operations"]
    end

    subgraph internal["Internal Layer"]
        ast["internal/ast<br/>Core AST Terms"]
        eval["internal/eval<br/>NbE Evaluation"]
        core["internal/core<br/>Conversion Checking"]
    end

    subgraph hypergraph["Hypergraph Package"]
        hgraph["hypergraph/<br/>Generic Hypergraph"]
    end

    hg --> check
    hottgo --> check

    check --> ctx
    check --> subst
    check --> ast
    check --> eval
    check --> core

    ctx --> ast
    subst --> ast

    core --> eval
    core --> ast

    eval --> ast

    style kernel fill:#1a1a1a,stroke:#444,color:#fff
    style internal fill:#1a1a1a,stroke:#444,color:#fff
    style cmd fill:#1a1a1a,stroke:#444,color:#fff
    style hypergraph fill:#1a1a1a,stroke:#444,color:#fff
```

### Package Dependencies (Detailed)

```mermaid
flowchart LR
    subgraph Kernel
        check[kernel/check]
        ctx[kernel/ctx]
        subst[kernel/subst]
    end

    subgraph Internal
        ast[internal/ast]
        eval[internal/eval]
        core[internal/core]
    end

    check -->|GlobalEnv, Checker| ast
    check -->|Synth, Check| ctx
    check -->|Shift, Subst| subst
    check -->|EvalNBE| eval
    check -->|Conv| core

    ctx -->|Term types| ast
    subst -->|Term types| ast

    core -->|Eval, Reify| eval
    core -->|AlphaEq| ast

    eval -->|Term, Value| ast

    style Kernel fill:#1a1a1a,stroke:#444,color:#fff
    style Internal fill:#1a1a1a,stroke:#444,color:#fff
    style check fill:#222,stroke:#555,color:#fff
    style ctx fill:#222,stroke:#555,color:#fff
    style subst fill:#222,stroke:#555,color:#fff
    style ast fill:#222,stroke:#555,color:#fff
    style eval fill:#222,stroke:#555,color:#fff
    style core fill:#222,stroke:#555,color:#fff
```

---

## 2. Term Type Hierarchy

```mermaid
classDiagram
    class Term {
        <<interface>>
        +isCoreTerm()
    }

    class Sort {
        +U Level
    }

    class Var {
        +Ix int
    }

    class Global {
        +Name string
    }

    class Pi {
        +Binder string
        +A Term
        +B Term
    }

    class Lam {
        +Binder string
        +Ann Term
        +Body Term
    }

    class App {
        +T Term
        +U Term
    }

    class Sigma {
        +Binder string
        +A Term
        +B Term
    }

    class Pair {
        +Fst Term
        +Snd Term
    }

    class Fst {
        +P Term
    }

    class Snd {
        +P Term
    }

    class Let {
        +Binder string
        +Ann Term
        +Val Term
        +Body Term
    }

    class Id {
        +A Term
        +X Term
        +Y Term
    }

    class Refl {
        +A Term
        +X Term
    }

    class J {
        +A Term
        +C Term
        +D Term
        +X Term
        +Y Term
        +P Term
    }

    Term <|.. Sort : implements
    Term <|.. Var : implements
    Term <|.. Global : implements
    Term <|.. Pi : implements
    Term <|.. Lam : implements
    Term <|.. App : implements
    Term <|.. Sigma : implements
    Term <|.. Pair : implements
    Term <|.. Fst : implements
    Term <|.. Snd : implements
    Term <|.. Let : implements
    Term <|.. Id : implements
    Term <|.. Refl : implements
    Term <|.. J : implements
```

### Term Categories

```mermaid
flowchart TB
    subgraph Atomic["Atomic Terms"]
        Sort["Sort<br/>Universe Levels"]
        Var["Var<br/>De Bruijn Index"]
        Global["Global<br/>Named Constants"]
    end

    subgraph Function["Function Types (Π)"]
        Pi["Pi<br/>(x:A) → B"]
        Lam["Lam<br/>λx. body"]
        App["App<br/>f u"]
    end

    subgraph Product["Product Types (Σ)"]
        Sigma["Sigma<br/>Σ(x:A). B"]
        Pair["Pair<br/>(a, b)"]
        Fst["Fst<br/>π₁"]
        Snd["Snd<br/>π₂"]
    end

    subgraph Identity["Identity Types (Id)"]
        Id["Id<br/>Id A x y"]
        Refl["Refl<br/>refl A x"]
        J["J<br/>Path Induction"]
    end

    subgraph Control["Control Flow"]
        Let["Let<br/>let x = v in e"]
    end

    Pi --> Lam
    Lam --> App
    Sigma --> Pair
    Pair --> Fst
    Pair --> Snd
    Id --> Refl
    Refl --> J

    style Atomic fill:#1a1a1a,stroke:#444,color:#fff
    style Function fill:#1a1a1a,stroke:#444,color:#fff
    style Product fill:#1a1a1a,stroke:#444,color:#fff
    style Identity fill:#1a1a1a,stroke:#444,color:#fff
    style Control fill:#1a1a1a,stroke:#444,color:#fff
```

---

## 3. Value Type Hierarchy (NbE)

```mermaid
classDiagram
    class Value {
        <<interface>>
        +isValue()
    }

    class VNeutral {
        +N Neutral
    }

    class Neutral {
        +Head Head
        +Sp []Value
    }

    class Head {
        +Var int
        +Glob string
    }

    class VLam {
        +Body *Closure
    }

    class VPi {
        +A Value
        +B *Closure
    }

    class VSigma {
        +A Value
        +B *Closure
    }

    class VPair {
        +Fst Value
        +Snd Value
    }

    class VSort {
        +Level int
    }

    class VGlobal {
        +Name string
    }

    class VId {
        +A Value
        +X Value
        +Y Value
    }

    class VRefl {
        +A Value
        +X Value
    }

    class Closure {
        +Env *Env
        +Term ast.Term
    }

    class Env {
        +Bindings []Value
        +Extend(v Value) *Env
        +Lookup(ix int) Value
    }

    Value <|.. VNeutral
    Value <|.. VLam
    Value <|.. VPi
    Value <|.. VSigma
    Value <|.. VPair
    Value <|.. VSort
    Value <|.. VGlobal
    Value <|.. VId
    Value <|.. VRefl

    VNeutral --> Neutral
    Neutral --> Head

    VLam --> Closure
    VPi --> Closure
    VSigma --> Closure

    Closure --> Env
```

---

## 4. Bidirectional Type Checking Flow

```mermaid
flowchart TB
    Start([Term to Check]) --> Mode{Mode?}

    Mode -->|Synth| Synth["synth(ctx, span, term)"]
    Mode -->|Check| Check["check(ctx, span, term, expected)"]

    subgraph SynthMode["Synthesis Mode (Infer Type)"]
        Synth --> SynthSwitch{Term Type?}
        SynthSwitch -->|Var| SynthVar["ctx.LookupVar(ix)<br/>+ Shift"]
        SynthSwitch -->|Sort| SynthSort["Sort U → Sort U+1"]
        SynthSwitch -->|Global| SynthGlobal["globals.LookupType"]
        SynthSwitch -->|Pi/Sigma| SynthType["checkIsType for A, B<br/>→ Sort max(U,V)"]
        SynthSwitch -->|Lam| SynthLam["Requires annotation<br/>→ Pi type"]
        SynthSwitch -->|App| SynthApp["synth(f) → Pi<br/>check(u, A)<br/>→ B[u/x]"]
        SynthSwitch -->|Fst/Snd| SynthProj["synth(p) → Sigma<br/>→ component type"]
        SynthSwitch -->|Let| SynthLet["synth/check val<br/>extend ctx<br/>synth body"]
        SynthSwitch -->|Id| SynthId["checkIsType(A)<br/>check(x,A), check(y,A)<br/>→ Sort U"]
        SynthSwitch -->|Refl| SynthRefl["checkIsType(A)<br/>check(x,A)<br/>→ Id A x x"]
        SynthSwitch -->|J| SynthJ["Build motive type<br/>Check all args<br/>→ C y p"]
    end

    subgraph CheckMode["Checking Mode (Verify Type)"]
        Check --> CheckSwitch{Term Type?}
        CheckSwitch -->|Lam unannotated| CheckLam["ensurePi(expected)<br/>extend ctx with A<br/>check(body, B)"]
        CheckSwitch -->|Pair| CheckPair["ensureSigma(expected)<br/>check(fst, A)<br/>check(snd, B[fst/x])"]
        CheckSwitch -->|Default| CheckBySynth["synth(term)<br/>conv(inferred, expected)"]
    end

    SynthVar --> Result([Type or Error])
    SynthSort --> Result
    SynthGlobal --> Result
    SynthType --> Result
    SynthLam --> Result
    SynthApp --> Result
    SynthProj --> Result
    SynthLet --> Result
    SynthId --> Result
    SynthRefl --> Result
    SynthJ --> Result

    CheckLam --> ResultCheck([nil or Error])
    CheckPair --> ResultCheck
    CheckBySynth --> ResultCheck

    style SynthMode fill:#1a1a1a,stroke:#444,color:#fff
    style CheckMode fill:#1a1a1a,stroke:#444,color:#fff
```

---

## 5. Normalization by Evaluation (NbE) Pipeline

```mermaid
flowchart LR
    subgraph Syntax["Syntax Domain"]
        Term["ast.Term<br/>(Source)"]
        NormalForm["ast.Term<br/>(Normal Form)"]
    end

    subgraph Semantics["Semantic Domain"]
        Value["eval.Value<br/>(WHNF)"]
    end

    Term -->|"Eval(env, term)"| Value
    Value -->|"Reify(value)"| NormalForm

    Term -.->|"EvalNBE(term)"| NormalForm

    style Syntax fill:#1a1a1a,stroke:#444,color:#fff
    style Semantics fill:#1a1a1a,stroke:#444,color:#fff
```

### NbE Complete Pipeline

```mermaid
sequenceDiagram
    participant User
    participant EvalNBE
    participant Eval
    participant Env
    participant Apply
    participant Reify

    User->>EvalNBE: EvalNBE(term)
    EvalNBE->>Eval: Eval(empty_env, term)

    loop For each subterm
        Eval->>Env: Lookup/Extend
        Env-->>Eval: Value or new Env

        alt Application
            Eval->>Apply: Apply(fun, arg)
            Apply-->>Eval: Reduced Value
        end
    end

    Eval-->>EvalNBE: Value (WHNF)
    EvalNBE->>Reify: Reify(value)

    loop For each sub-value
        Reify->>Reify: Recursive reify
    end

    Reify-->>EvalNBE: ast.Term (Normal Form)
    EvalNBE-->>User: Normalized Term
```

---

## 6. Eval Function Flow

```mermaid
flowchart TB
    Start([Eval env term]) --> NilCheck{term nil?}
    NilCheck -->|Yes| ReturnNil["VGlobal{nil}"]
    NilCheck -->|No| EnvCheck{env nil?}
    EnvCheck -->|Yes| CreateEnv["env = empty Env"]
    EnvCheck -->|No| Switch
    CreateEnv --> Switch

    Switch{Term Type?}

    Switch -->|Var| EvalVar["env.Lookup(ix)"]
    Switch -->|Global| EvalGlobal["vGlobal(name)"]
    Switch -->|Sort| EvalSort["VSort{Level: U}"]
    Switch -->|Lam| EvalLam["VLam{Closure{env, body}}"]
    Switch -->|App| EvalApp["fun := Eval(env, T)<br/>arg := Eval(env, U)<br/>Apply(fun, arg)"]
    Switch -->|Pair| EvalPair["VPair{<br/>  Fst: Eval(env, fst),<br/>  Snd: Eval(env, snd)<br/>}"]
    Switch -->|Fst| EvalFst["p := Eval(env, P)<br/>Fst(p)"]
    Switch -->|Snd| EvalSnd["p := Eval(env, P)<br/>Snd(p)"]
    Switch -->|Pi| EvalPi["VPi{<br/>  A: Eval(env, A),<br/>  B: Closure{env, B}<br/>}"]
    Switch -->|Sigma| EvalSigma["VSigma{<br/>  A: Eval(env, A),<br/>  B: Closure{env, B}<br/>}"]
    Switch -->|Let| EvalLet["val := Eval(env, Val)<br/>newEnv := env.Extend(val)<br/>Eval(newEnv, Body)"]
    Switch -->|Id| EvalId["VId{<br/>  A: Eval(env, A),<br/>  X: Eval(env, X),<br/>  Y: Eval(env, Y)<br/>}"]
    Switch -->|Refl| EvalRefl["VRefl{<br/>  A: Eval(env, A),<br/>  X: Eval(env, X)<br/>}"]
    Switch -->|J| EvalJ["evalJ(<br/>  Eval(env, A),<br/>  Eval(env, C),<br/>  Eval(env, D),<br/>  Eval(env, X),<br/>  Eval(env, Y),<br/>  Eval(env, P)<br/>)"]

    EvalVar --> Result([Value])
    EvalGlobal --> Result
    EvalSort --> Result
    EvalLam --> Result
    EvalApp --> Result
    EvalPair --> Result
    EvalFst --> Result
    EvalSnd --> Result
    EvalPi --> Result
    EvalSigma --> Result
    EvalLet --> Result
    EvalId --> Result
    EvalRefl --> Result
    EvalJ --> Result
    ReturnNil --> Result

    style Switch fill:#1a1a1a,stroke:#444,color:#fff
```

---

## 7. Apply Function (Beta Reduction)

```mermaid
flowchart TB
    Start([Apply fun arg]) --> FunType{fun type?}

    FunType -->|VLam| BetaReduce
    FunType -->|VNeutral| ExtendSpine
    FunType -->|Other| BadApp

    subgraph BetaReduce["Beta Reduction"]
        BR1["newEnv := closure.Env.Extend(arg)"]
        BR2["Eval(newEnv, closure.Term)"]
        BR1 --> BR2
    end

    subgraph ExtendSpine["Neutral Application"]
        ES1["newSp := append(neutral.Sp, arg)"]
        ES2["VNeutral{Head: head, Sp: newSp}"]
        ES1 --> ES2
    end

    subgraph BadApp["Error Case"]
        BA1["VNeutral{<br/>  Head: 'bad_app',<br/>  Sp: [fun, arg]<br/>}"]
    end

    BetaReduce --> Result([Value])
    ExtendSpine --> Result
    BadApp --> Result

    style BetaReduce fill:#1a1a1a,stroke:#444,color:#fff
    style ExtendSpine fill:#1a1a1a,stroke:#444,color:#fff
    style BadApp fill:#1a1a1a,stroke:#444,color:#fff
```

### Beta Reduction Example

```mermaid
sequenceDiagram
    participant Apply
    participant Closure
    participant Env
    participant Eval

    Note over Apply: Apply(VLam{body}, arg)
    Apply->>Closure: Get body closure
    Closure-->>Apply: {env, term}
    Apply->>Env: env.Extend(arg)
    Env-->>Apply: newEnv with arg at index 0
    Apply->>Eval: Eval(newEnv, term)
    Note over Eval: Body evaluated with<br/>arg bound to Var{0}
    Eval-->>Apply: Reduced Value
```

---

## 8. Reify Function Flow

```mermaid
flowchart TB
    Start([Reify value]) --> ValType{Value Type?}

    ValType -->|VNeutral| ReifyNeutral["reifyNeutral(neutral)"]
    ValType -->|VLam| ReifyLam
    ValType -->|VPair| ReifyPair["Pair{<br/>  Fst: Reify(fst),<br/>  Snd: Reify(snd)<br/>}"]
    ValType -->|VSort| ReifySort["Sort{U: level}"]
    ValType -->|VGlobal| ReifyGlobal["Global{Name: name}"]
    ValType -->|VPi| ReifyPi
    ValType -->|VSigma| ReifySigma
    ValType -->|VId| ReifyId["Id{<br/>  A: Reify(A),<br/>  X: Reify(X),<br/>  Y: Reify(Y)<br/>}"]
    ValType -->|VRefl| ReifyRefl["Refl{<br/>  A: Reify(A),<br/>  X: Reify(X)<br/>}"]

    subgraph ReifyLam["Reify Lambda"]
        RL1["freshVar := vVar(0)"]
        RL2["bodyVal := Apply(VLam, freshVar)"]
        RL3["bodyTerm := Reify(bodyVal)"]
        RL4["Lam{Binder: '_', Body: bodyTerm}"]
        RL1 --> RL2 --> RL3 --> RL4
    end

    subgraph ReifyPi["Reify Pi Type"]
        RP1["a := Reify(A)"]
        RP2["freshVar := vVar(0)"]
        RP3["bVal := Apply(VLam{B}, freshVar)"]
        RP4["b := Reify(bVal)"]
        RP5["Pi{Binder: '_', A: a, B: b}"]
        RP1 --> RP2 --> RP3 --> RP4 --> RP5
    end

    subgraph ReifySigma["Reify Sigma Type"]
        RS1["a := Reify(A)"]
        RS2["freshVar := vVar(0)"]
        RS3["bVal := Apply(VLam{B}, freshVar)"]
        RS4["b := Reify(bVal)"]
        RS5["Sigma{Binder: '_', A: a, B: b}"]
        RS1 --> RS2 --> RS3 --> RS4 --> RS5
    end

    ReifyNeutral --> Result([ast.Term])
    ReifyLam --> Result
    ReifyPair --> Result
    ReifySort --> Result
    ReifyGlobal --> Result
    ReifyPi --> Result
    ReifySigma --> Result
    ReifyId --> Result
    ReifyRefl --> Result

    style ReifyLam fill:#1a1a1a,stroke:#444,color:#fff
    style ReifyPi fill:#1a1a1a,stroke:#444,color:#fff
    style ReifySigma fill:#1a1a1a,stroke:#444,color:#fff
```

### Reify Neutral Terms

```mermaid
flowchart TB
    Start([reifyNeutral neutral]) --> HeadType{Head Type?}

    HeadType -->|Var| CreateVar["head := Var{Ix: var}"]
    HeadType -->|Global 'fst'| CreateFst["Fst{P: Reify(sp[0])}"]
    HeadType -->|Global 'snd'| CreateSnd["Snd{P: Reify(sp[0])}"]
    HeadType -->|Global other| CreateGlobal["head := Global{Name: glob}"]

    CreateVar --> ApplySpine
    CreateGlobal --> ApplySpine

    subgraph ApplySpine["Apply Spine Arguments"]
        AS1["result := head"]
        AS2["for arg in spine:"]
        AS3["  result := App{T: result, U: Reify(arg)}"]
        AS1 --> AS2 --> AS3
    end

    ApplySpine --> Result([ast.Term])
    CreateFst --> Result
    CreateSnd --> Result

    style ApplySpine fill:#1a1a1a,stroke:#444,color:#fff
```

---

## 9. J Elimination (Path Induction)

```mermaid
flowchart TB
    Start([evalJ a c d x y p]) --> CheckProof{p is VRefl?}

    CheckProof -->|Yes| ComputationRule
    CheckProof -->|No| StuckNeutral

    subgraph ComputationRule["Computation Rule Applies"]
        CR1["J A C d x x (refl A x) → d"]
        CR2["return d"]
        CR1 --> CR2
    end

    subgraph StuckNeutral["Stuck - Return Neutral"]
        SN1["head := Head{Glob: 'J'}"]
        SN2["spine := [a, c, d, x, y, p]"]
        SN3["VNeutral{N: Neutral{Head: head, Sp: spine}}"]
        SN1 --> SN2 --> SN3
    end

    ComputationRule --> Result([Value])
    StuckNeutral --> Result

    style ComputationRule fill:#1a1a1a,stroke:#444,color:#fff
    style StuckNeutral fill:#1a1a1a,stroke:#444,color:#fff
```

### J Typing Rules

```mermaid
flowchart TB
    subgraph JType["J Elimination Type"]
        A["A : Type"]
        x["x : A"]
        y["y : A"]
        C["C : (y : A) → Id A x y → Type"]
        d["d : C x (refl A x)"]
        p["p : Id A x y"]
        result["J A C d x y p : C y p"]

        A --> C
        x --> C
        x --> d
        y --> p
        C --> d
        C --> result
        p --> result
    end

    style JType fill:#1a1a1a,stroke:#444,color:#fff
```

---

## 10. Conversion Checking

```mermaid
flowchart TB
    Start([Conv env t u opts]) --> EnvCheck{env nil?}
    EnvCheck -->|Yes| CreateEnv["env := NewEnv()"]
    EnvCheck -->|No| EvalBoth
    CreateEnv --> EvalBoth

    subgraph EvalBoth["Evaluate Both Terms"]
        E1["valT := Eval(env, t)"]
        E2["valU := Eval(env, u)"]
        E1 --> E2
    end

    EvalBoth --> Reify

    subgraph Reify["Reify to Normal Forms"]
        R1["nfT := Reify(valT)"]
        R2["nfU := Reify(valU)"]
        R1 --> R2
    end

    Reify --> EtaCheck{EnableEta?}

    EtaCheck -->|Yes| EtaEqual["etaEqual(nfT, nfU)"]
    EtaCheck -->|No| AlphaEq["AlphaEq(nfT, nfU)"]

    EtaEqual --> Result([bool])
    AlphaEq --> Result

    style EvalBoth fill:#1a1a1a,stroke:#444,color:#fff
    style Reify fill:#1a1a1a,stroke:#444,color:#fff
```

### Alpha Equality

```mermaid
flowchart TB
    Start([AlphaEq a b]) --> NilCheck{both nil?}
    NilCheck -->|Yes| ReturnTrue([true])
    NilCheck -->|No| OneNil{one nil?}
    OneNil -->|Yes| ReturnFalse([false])
    OneNil -->|No| TypeMatch

    TypeMatch{Same Type?}
    TypeMatch -->|No| ReturnFalse
    TypeMatch -->|Yes| Compare

    subgraph Compare["Structural Comparison"]
        C1["Sort: a.U == b.U"]
        C2["Var: a.Ix == b.Ix"]
        C3["Global: a.Name == b.Name"]
        C4["Pi: AlphaEq(A) && AlphaEq(B)"]
        C5["Lam: AlphaEq(Body)"]
        C6["App: AlphaEq(T) && AlphaEq(U)"]
        C7["Sigma: AlphaEq(A) && AlphaEq(B)"]
        C8["Pair: AlphaEq(Fst) && AlphaEq(Snd)"]
        C9["Id: AlphaEq(A,X,Y)"]
        C10["Refl: AlphaEq(A,X)"]
        C11["J: AlphaEq(A,C,D,X,Y,P)"]
    end

    Compare --> Result([bool])

    style Compare fill:#1a1a1a,stroke:#444,color:#fff
```

---

## 11. Context and Environment Management

### Typing Context (kernel/ctx)

```mermaid
flowchart LR
    subgraph Ctx["Ctx Structure"]
        Tele["Tele: []Binding"]
    end

    subgraph Operations["Context Operations"]
        Extend["Extend(name, type)<br/>Add binding at front"]
        Lookup["LookupVar(ix)<br/>Get type by index"]
        Drop["Drop()<br/>Remove newest"]
        Len["Len()<br/>Number of bindings"]
    end

    Ctx --> Operations

    style Ctx fill:#1a1a1a,stroke:#444,color:#fff
    style Operations fill:#1a1a1a,stroke:#444,color:#fff
```

### De Bruijn Environment (internal/eval)

```mermaid
flowchart TB
    subgraph Env["Env Structure"]
        Bindings["Bindings: []Value"]
    end

    subgraph EnvOps["Environment Operations"]
        direction LR
        Extend["Extend(v)<br/>Prepend value"]
        Lookup["Lookup(ix)<br/>Get by index"]
    end

    subgraph Example["Example: [A, B, C]"]
        E0["Index 0 → A (newest)"]
        E1["Index 1 → B"]
        E2["Index 2 → C (oldest)"]

        After["After Extend(D):<br/>[D, A, B, C]"]

        E0 --> After
    end

    Env --> EnvOps
    EnvOps --> Example

    style Env fill:#1a1a1a,stroke:#444,color:#fff
    style EnvOps fill:#1a1a1a,stroke:#444,color:#fff
    style Example fill:#1a1a1a,stroke:#444,color:#fff
```

### Global Environment (kernel/check)

```mermaid
classDiagram
    class GlobalEnv {
        -axioms map[string]*Axiom
        -defs map[string]*Definition
        -inductives map[string]*Inductive
        -primitives map[string]*Primitive
        -order []string
        +NewGlobalEnv() *GlobalEnv
        +NewGlobalEnvWithPrimitives() *GlobalEnv
        +AddAxiom(name, type)
        +AddDefinition(name, type, body, trans)
        +AddInductive(name, type, constrs, elim)
        +LookupType(name) Term
        +LookupDefinitionBody(name) Term, bool
        +Has(name) bool
    }

    class Axiom {
        +Name string
        +Type Term
    }

    class Definition {
        +Name string
        +Type Term
        +Body Term
        +Transparency Transparency
    }

    class Inductive {
        +Name string
        +Type Term
        +Constructors []Constructor
        +Eliminator string
    }

    class Primitive {
        +Name string
        +Type Term
    }

    GlobalEnv --> Axiom
    GlobalEnv --> Definition
    GlobalEnv --> Inductive
    GlobalEnv --> Primitive
```

---

## 12. Complete Type Checking Pipeline

```mermaid
flowchart TB
    subgraph Input["Input"]
        Term["ast.Term"]
        Ctx["Typing Context"]
        Globals["Global Environment"]
    end

    subgraph TypeChecker["Type Checker (kernel/check)"]
        Checker["Checker"]
        Synth["synth()"]
        Check["check()"]

        Checker --> Synth
        Checker --> Check
    end

    subgraph Helpers["Helper Operations"]
        CheckIsType["checkIsType()"]
        EnsurePi["ensurePi()"]
        EnsureSigma["ensureSigma()"]
        EnsureSort["ensureSort()"]
        WHNF["whnf() → EvalNBE"]
    end

    subgraph CoreOps["Core Operations"]
        Conv["conv() → core.Conv()"]
        Subst["subst.Subst()"]
        Shift["subst.Shift()"]
    end

    subgraph NbE["NbE Engine"]
        Eval["eval.Eval()"]
        Apply["eval.Apply()"]
        Reify["eval.Reify()"]
        EvalJ["eval.evalJ()"]
    end

    subgraph Output["Output"]
        Type["Inferred Type"]
        Error["TypeError"]
    end

    Term --> Checker
    Ctx --> Checker
    Globals --> Checker

    TypeChecker --> Helpers
    TypeChecker --> CoreOps

    Helpers --> WHNF
    WHNF --> NbE
    CoreOps --> NbE

    TypeChecker --> Type
    TypeChecker --> Error

    style Input fill:#1a1a1a,stroke:#444,color:#fff
    style TypeChecker fill:#1a1a1a,stroke:#444,color:#fff
    style NbE fill:#1a1a1a,stroke:#444,color:#fff
    style Helpers fill:#1a1a1a,stroke:#444,color:#fff
    style CoreOps fill:#1a1a1a,stroke:#444,color:#fff
    style Output fill:#1a1a1a,stroke:#444,color:#fff
```

### Complete Example: Type Checking Identity Function

```mermaid
sequenceDiagram
    participant User
    participant Checker
    participant Synth
    participant CheckIsType
    participant Subst
    participant Conv
    participant NbE

    Note over User: λ(A:Type0). λ(x:A). x
    User->>Checker: Synth(ctx, term)

    Checker->>Synth: synth(ctx, Lam A:Type0. ...)
    Synth->>CheckIsType: checkIsType(ctx, Type0)
    CheckIsType->>NbE: whnf(Type0)
    NbE-->>CheckIsType: Sort{1}
    CheckIsType-->>Synth: level 0 ✓

    Note over Synth: Extend ctx with A:Type0
    Synth->>Synth: synth(ctx+A, Lam x:A. x)
    Synth->>CheckIsType: checkIsType(ctx+A, A)
    CheckIsType->>NbE: whnf(A)
    NbE-->>CheckIsType: Sort{0}
    CheckIsType-->>Synth: level 0 ✓

    Note over Synth: Extend ctx with x:A
    Synth->>Synth: synth(ctx+A+x, x)
    Note over Synth: Var lookup: x → A (shifted)
    Synth->>Subst: Shift(2, 0, A)
    Subst-->>Synth: A (at correct index)

    Synth-->>Synth: Pi{x, A, A}
    Synth-->>Synth: Pi{A, Type0, Pi{x, A, A}}
    Synth-->>Checker: Π(A:Type0). Π(x:A). A
    Checker-->>User: Type inferred ✓
```

---

## Summary

This document provides a comprehensive visual guide to the HoTT kernel architecture:

1. **Package Structure**: Clear separation between kernel (trusted), internal (implementation), and command layers
2. **Term Hierarchy**: 14 term constructors covering dependent types, pairs, and identity types
3. **Value Hierarchy**: 9 value types for the NbE semantic domain
4. **Bidirectional Type Checking**: Synth/Check modes with case analysis for each term type
5. **NbE Pipeline**: Eval → Apply → Reify for normalization
6. **J Elimination**: Path induction with computation rule for reflexivity
7. **Conversion**: Definitional equality via normalization and structural comparison
8. **Context Management**: De Bruijn indices with proper shifting and substitution

The kernel implements a sound intensional type theory with identity types (Id, refl, J), supporting the foundations for homotopy type theory.
