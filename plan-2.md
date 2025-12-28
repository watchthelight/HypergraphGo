# Parallel Claude Instance Prompts for HypergraphGo - Phase 2

Based on the comprehensive codebase audit, here are 6 prompts for separate Claude instances to work on substantial improvements in parallel. These address issues NOT covered by plan.md.

Each prompt is fully self-contained with all necessary context.

---

## Prompt 1: AST Raw Terms and Print Optimization

```
You are working on the HypergraphGo project at /Users/bash/Documents/Projects/hottgo

=== AVAILABLE CLAUDE CODE RESOURCES ===

You have access to powerful agents, tools, MCP servers, and skills. Use these to work efficiently.

AGENTS (via Task tool with subagent_type parameter):
- Explore: Fast codebase exploration, finding files, understanding patterns
  Example: Task tool with subagent_type='Explore' prompt="Find all files that define raw term types"
- Plan: Designing implementation strategies before coding
  Example: Task tool with subagent_type='Plan' prompt="Plan the implementation of RId, RRefl, RJ"
- general-purpose: Complex multi-step research tasks
  Example: Task tool with subagent_type='general-purpose' prompt="Analyze all AST term types and their relationships"

BUILT-IN TOOLS:
- Read: Read file contents (Go files, images, PDFs)
- Write: Create new files
- Edit: Make targeted edits to existing files
- Glob: Find files by pattern (e.g., "**/*_test.go", "internal/ast/*.go")
- Grep: Search file contents with regex (e.g., pattern="func.*RawTerm")
- Bash: Run shell commands (go test, go build, git, etc.)
- WebSearch: Search the web for documentation or solutions
- WebFetch: Fetch and analyze web page content

MCP SERVERS (configured in .mcp.json):
- GitHub MCP: Work with PRs, issues, code reviews
  - Run /mcp to authenticate via OAuth
  - Then ask: "Show open PRs", "Create issue for bug", "Check CI status"

SKILLS:
- /mermaid: Generate diagrams (architecture, flow, sequence)
  Example: /mermaid to create diagram of AST term hierarchy

BEST PRACTICES:
1. Start with Explore agent to understand unfamiliar code before making changes
2. Use Glob before Grep to narrow down files
3. Make multiple parallel tool calls when searches are independent
4. Use GitHub MCP to check for existing issues/PRs before starting
5. Generate diagrams with /mermaid when documenting complex relationships

=== END RESOURCES ===

SETUP:
1. Use Explore agent to understand the AST structure and raw term patterns
2. Create and checkout branch: `git checkout -b fix/ast-raw-terms`
3. All commits must be authored by watchthelight: use `git commit --author="watchthelight <admin@watchthelight.org>"`
4. Commit frequently after each logical piece of work (aim for 8-12 commits)

YOUR TASK: Add missing raw term types and fix performance issues in the AST package.

CONTEXT - Issues Found:
- internal/ast/raw.go is missing raw term types for identity types (Id, Refl, J)
- internal/ast/resolve.go cannot handle identity raw terms
- collectSpine in print.go:15-26 is O(n^2) due to slice prepending
- raw.go has no documentation

SCOPE (files to modify):
- internal/ast/raw.go (ADD RId, RRefl, RJ types - ~30 lines)
- internal/ast/resolve.go (ADD cases for identity raw terms - ~40 lines)
- internal/ast/print.go (FIX collectSpine performance - ~15 lines)
- internal/ast/raw_test.go (NEW - test raw term resolution - ~100 lines)
- CHANGELOG.md (update with your changes)

SPECIFIC CHANGES:

1. Add Raw Identity Types (raw.go):
   - RId struct with A, X, Y fields (raw terms for endpoints)
   - RRefl struct with A, X fields
   - RJ struct with A, C, D, X, Y, P fields
   - Add corresponding isRawTerm() marker methods

2. Update Resolver (resolve.go):
   - Add cases in the main resolve switch for RId, RRefl, RJ
   - Convert raw identity terms to core identity terms
   - Handle variable resolution in all subterms

3. Fix collectSpine Performance (print.go:15-26):
   Current O(n^2) code:
   ```go
   args = append([]Term{app.U}, args...)  // Creates new slice each iteration
   ```
   Fix: Collect in order, then reverse at the end:
   ```go
   args = append(args, app.U)  // Collect forward
   // ... after loop:
   slices.Reverse(args)  // Single O(n) reverse
   ```

4. Add Tests (raw_test.go):
   - Test resolution of RId, RRefl, RJ
   - Test nested identity types
   - Test identity types with bound variables

VERIFICATION:
- Run `go test ./internal/ast/ -v` after each commit
- Run `go build ./...` to ensure no compilation errors
- Verify collectSpine still produces correct output

Do NOT merge to main - leave the branch for review.
```

---

## Prompt 2: Evaluation Engine Correctness Fixes

```
You are working on the HypergraphGo project at /Users/bash/Documents/Projects/hottgo

=== AVAILABLE CLAUDE CODE RESOURCES ===

You have access to powerful agents, tools, MCP servers, and skills. Use these to work efficiently.

AGENTS (via Task tool with subagent_type parameter):
- Explore: Fast codebase exploration, finding files, understanding patterns
  Example: Task tool with subagent_type='Explore' prompt="Find all files in internal/eval/"
- Plan: Designing implementation strategies before coding
  Example: Task tool with subagent_type='Plan' prompt="Plan the alpha-equality implementation"
- general-purpose: Complex multi-step research tasks
  Example: Task tool with subagent_type='general-purpose' prompt="Analyze the NbE implementation and identify all edge cases"

BUILT-IN TOOLS:
- Read: Read file contents (Go files, images, PDFs)
- Write: Create new files
- Edit: Make targeted edits to existing files
- Glob: Find files by pattern (e.g., "internal/eval/*.go")
- Grep: Search file contents with regex (e.g., pattern="func.*reifyNeutral")
- Bash: Run shell commands (go test, go build, git, etc.)
- WebSearch: Search the web for documentation or solutions
- WebFetch: Fetch and analyze web page content

MCP SERVERS (configured in .mcp.json):
- GitHub MCP: Work with PRs, issues, code reviews
  - Run /mcp to authenticate via OAuth
  - Then ask: "Show open PRs", "Create issue for bug", "Check CI status"

SKILLS:
- /mermaid: Generate diagrams (architecture, flow, sequence)
  Example: /mermaid to visualize NbE data flow

BEST PRACTICES:
1. Start with Explore agent to understand unfamiliar code before making changes
2. Use Glob before Grep to narrow down files
3. Make multiple parallel tool calls when searches are independent
4. Use GitHub MCP to check for existing issues/PRs before starting
5. Generate diagrams with /mermaid when documenting complex relationships

=== END RESOURCES ===

SETUP:
1. Use Explore agent to understand the NbE implementation thoroughly
2. Create and checkout branch: `git checkout -b fix/eval-correctness`
3. All commits must be authored by watchthelight: use `git commit --author="watchthelight <admin@watchthelight.org>"`
4. Commit frequently after each logical piece of work (aim for 10-15 commits)

YOUR TASK: Fix critical correctness issues in the evaluation engine.

CONTEXT - Issues Found:
- alphaEqCubical (nbe_cubical.go:898) uses string comparison - incorrect for alpha-equality
- EvalFill (nbe_cubical.go:750) always returns stuck - should compute at endpoints
- EvalUABeta (nbe_cubical.go:840) always stuck - univalence computation not implemented
- Code duplication between reifyNeutralAt and reifyNeutralCubicalAt (~80% shared)

SCOPE (files to modify):
- internal/eval/nbe_cubical.go (FIX alphaEqCubical, EvalFill, EvalUABeta - ~150 lines)
- internal/eval/nbe.go (REFACTOR reifyNeutralAt to share code - ~50 lines)
- internal/eval/alpha_eq.go (NEW - proper alpha-equality implementation - ~100 lines)
- internal/eval/alpha_eq_test.go (NEW - tests for alpha-equality - ~80 lines)
- CHANGELOG.md (update with your changes)

SPECIFIC CHANGES:

1. Implement Proper Alpha-Equality (alpha_eq.go):
   - Create AlphaEq(a, b ast.Term) bool function
   - Use de Bruijn indices - terms are alpha-equal if structurally identical
   - Handle all AST node types recursively
   - Handle binders correctly (they introduce new scope levels)

2. Fix alphaEqCubical (nbe_cubical.go:898):
   - Replace: `return ast.Sprint(a) == ast.Sprint(b)`
   - With: `return AlphaEq(a, b)`

3. Implement EvalFill Endpoint Cases (nbe_cubical.go:750-753):
   ```go
   func EvalFill(aClosure *IClosure, phi FaceValue, tubeClosure *IClosure, base Value) Value {
       // Fill at i=0 should return base
       // Fill at i=1 should return Comp result
       // Otherwise stuck
   }
   ```

4. Implement EvalUABeta Computation (nbe_cubical.go:840-844):
   The computation rule: transport (ua e) a = e.fst a
   ```go
   func EvalUABeta(equiv, arg Value) Value {
       // Extract fst from equiv (the forward function)
       // Apply it to arg
       fwd := Fst(equiv)
       return Apply(fwd, arg)
   }
   ```

5. Reduce reifyNeutral Duplication:
   - Extract shared spine handling into helper function
   - Call helper from both reifyNeutralAt and reifyNeutralCubicalAt

VERIFICATION:
- Run `go test ./internal/eval/ -v` after each commit
- Run `go test ./internal/eval/ -cover` to track coverage
- Ensure existing tests still pass

Do NOT merge to main - leave the branch for review.
```

---

## Prompt 3: Hypergraph Package Fixes

```
You are working on the HypergraphGo project at /Users/bash/Documents/Projects/hottgo

=== AVAILABLE CLAUDE CODE RESOURCES ===

You have access to powerful agents, tools, MCP servers, and skills. Use these to work efficiently.

AGENTS (via Task tool with subagent_type parameter):
- Explore: Fast codebase exploration, finding files, understanding patterns
  Example: Task tool with subagent_type='Explore' prompt="Understand the hypergraph package structure"
- Plan: Designing implementation strategies before coding
  Example: Task tool with subagent_type='Plan' prompt="Plan making GreedyHittingSet deterministic"
- general-purpose: Complex multi-step research tasks
  Example: Task tool with subagent_type='general-purpose' prompt="Find all uses of sentinel errors in hypergraph"

BUILT-IN TOOLS:
- Read: Read file contents (Go files, images, PDFs)
- Write: Create new files
- Edit: Make targeted edits to existing files
- Glob: Find files by pattern (e.g., "hypergraph/*_test.go")
- Grep: Search file contents with regex (e.g., pattern="EdgeMembers|Copy")
- Bash: Run shell commands (go test, go build, git, etc.)
- WebSearch: Search the web for documentation or solutions
- WebFetch: Fetch and analyze web page content

MCP SERVERS (configured in .mcp.json):
- GitHub MCP: Work with PRs, issues, code reviews
  - Run /mcp to authenticate via OAuth
  - Then ask: "Show open PRs", "Create issue for bug", "Check CI status"

SKILLS:
- /mermaid: Generate diagrams (architecture, flow, sequence)
  Example: /mermaid to visualize hypergraph data structures

BEST PRACTICES:
1. Start with Explore agent to understand unfamiliar code before making changes
2. Use Glob before Grep to narrow down files
3. Make multiple parallel tool calls when searches are independent
4. Use GitHub MCP to check for existing issues/PRs before starting
5. Generate diagrams with /mermaid when documenting complex relationships

=== END RESOURCES ===

SETUP:
1. Use Explore agent to understand the hypergraph package structure
2. Create and checkout branch: `git checkout -b fix/hypergraph-issues`
3. All commits must be authored by watchthelight: use `git commit --author="watchthelight <admin@watchthelight.org>"`
4. Commit frequently after each logical piece of work (aim for 8-12 commits)

YOUR TASK: Fix bugs and improve test coverage in the hypergraph package.

CONTEXT - Issues Found:
- doc.go:43 references non-existent Primal method
- EdgeMembers and Copy methods have 0% test coverage
- ErrUnknownEdge and ErrUnknownVertex are defined but never used
- GreedyHittingSet is non-deterministic (map iteration order)
- Missing tests for duplicate vertices in AddEdge input

SCOPE (files to modify):
- hypergraph/doc.go (FIX incorrect Primal reference - 1 line)
- hypergraph/algorithms.go (FIX GreedyHittingSet determinism - ~10 lines)
- hypergraph/hypergraph.go (USE or REMOVE unused errors - ~5 lines)
- hypergraph/errors.go (REMOVE if unused - or document why kept)
- hypergraph/hypergraph_test.go (ADD tests for EdgeMembers, Copy, duplicates - ~100 lines)
- hypergraph/algorithms_test.go (ADD determinism test - ~20 lines)
- CHANGELOG.md (update with your changes)

SPECIFIC CHANGES:

1. Fix doc.go (line 43):
   - Remove or correct the reference to non-existent Primal method
   - Either: delete the line mentioning Primal
   - Or: add a Primal method as an alias for TwoSection

2. Make GreedyHittingSet Deterministic (algorithms.go:16-47):
   Current code iterates over map (non-deterministic):
   ```go
   for v := range h.vertices {
   ```
   Fix: Sort vertices first for reproducible results:
   ```go
   vertices := h.Vertices()
   slices.Sort(vertices)  // or sort.Strings if []string
   for _, v := range vertices {
   ```

3. Handle Unused Sentinel Errors:
   Option A: Use them - make RemoveVertex return ErrUnknownVertex when vertex not found
   Option B: Remove them from errors.go if they're intentionally unused
   Document the decision in a commit message

4. Add Missing Tests (hypergraph_test.go):
   - Test EdgeMembers with existing edge
   - Test EdgeMembers with non-existent edge
   - Test Copy method - verify deep copy semantics
   - Test AddEdge with duplicate vertices in input: ["A", "A", "B"]
   - Test VertexDegree for non-existent vertex
   - Test EdgeSize for non-existent edge

5. Add Determinism Test (algorithms_test.go):
   - Run GreedyHittingSet multiple times
   - Verify same result each time

VERIFICATION:
- Run `go test ./hypergraph/ -v` after each commit
- Run `go test ./hypergraph/ -cover` to verify coverage improvement
- Target: 95%+ coverage

Do NOT merge to main - leave the branch for review.
```

---

## Prompt 4: Context Package and Kernel Improvements

```
You are working on the HypergraphGo project at /Users/bash/Documents/Projects/hottgo

=== AVAILABLE CLAUDE CODE RESOURCES ===

You have access to powerful agents, tools, MCP servers, and skills. Use these to work efficiently.

AGENTS (via Task tool with subagent_type parameter):
- Explore: Fast codebase exploration, finding files, understanding patterns
  Example: Task tool with subagent_type='Explore' prompt="Find all callers of ctx.Extend"
- Plan: Designing implementation strategies before coding
  Example: Task tool with subagent_type='Plan' prompt="Plan removing custom itoa safely"
- general-purpose: Complex multi-step research tasks
  Example: Task tool with subagent_type='general-purpose' prompt="Analyze all usages of kernel/ctx"

BUILT-IN TOOLS:
- Read: Read file contents (Go files, images, PDFs)
- Write: Create new files
- Edit: Make targeted edits to existing files
- Glob: Find files by pattern (e.g., "kernel/**/*.go")
- Grep: Search file contents with regex (e.g., pattern="itoa|Extend|Drop")
- Bash: Run shell commands (go test, go build, git, etc.)
- WebSearch: Search the web for documentation or solutions
- WebFetch: Fetch and analyze web page content

MCP SERVERS (configured in .mcp.json):
- GitHub MCP: Work with PRs, issues, code reviews
  - Run /mcp to authenticate via OAuth
  - Then ask: "Show open PRs", "Create issue for bug", "Check CI status"

SKILLS:
- /mermaid: Generate diagrams (architecture, flow, sequence)
  Example: /mermaid to visualize context extension patterns

BEST PRACTICES:
1. Start with Explore agent to understand unfamiliar code before making changes
2. Use Glob before Grep to narrow down files
3. Make multiple parallel tool calls when searches are independent
4. Use GitHub MCP to check for existing issues/PRs before starting
5. Generate diagrams with /mermaid when documenting complex relationships

=== END RESOURCES ===

SETUP:
1. Use Explore agent to understand the context package and its callers
2. Create and checkout branch: `git checkout -b fix/kernel-ctx`
3. All commits must be authored by watchthelight: use `git commit --author="watchthelight <admin@watchthelight.org>"`
4. Commit frequently after each logical piece of work (aim for 8-12 commits)

YOUR TASK: Improve the kernel/ctx package and fix code pattern issues in kernel/.

CONTEXT - Issues Found:
- kernel/ctx has 50% test coverage - Len and Drop untested
- Mixed pointer/value receivers in Ctx type
- No nil validation in Extend method
- kernel/check/span.go:70-87 reimplements strconv.Itoa unnecessarily
- Inconsistent binder naming ("_" vs "" for anonymous)

SCOPE (files to modify):
- kernel/ctx/ctx.go (ADD nil validation, DOCUMENT receiver choice - ~10 lines)
- kernel/ctx/ctx_test.go (ADD comprehensive tests - ~80 lines)
- kernel/check/span.go (REMOVE custom itoa, use strconv - ~20 lines removed)
- kernel/check/errors.go (UPDATE to use strconv.Itoa - ~5 lines)
- CHANGELOG.md (update with your changes)

SPECIFIC CHANGES:

1. Add Nil Validation to Extend (ctx.go:33-35):
   ```go
   func (c *Ctx) Extend(name string, ty ast.Term) {
       if ty == nil {
           panic("ctx.Extend: nil type")
       }
       c.Tele = append(c.Tele, Binding{Name: name, Ty: ty})
   }
   ```

2. Document Receiver Convention (ctx.go):
   Add comment explaining why Extend uses pointer receiver while others use value:
   ```go
   // Extend modifies the context in place, hence pointer receiver.
   // Other methods return values or are read-only, hence value receivers.
   ```

3. Add Comprehensive Tests (ctx_test.go):
   - Test Len() on empty context
   - Test Len() after multiple Extend calls
   - Test Drop() on empty context (should return empty)
   - Test Drop() restores previous state
   - Test LookupVar(-1) returns false
   - Test LookupVar on empty context
   - Test chained Extend/Drop operations

4. Remove Custom itoa (span.go:70-87):
   - Delete the entire itoa function
   - Import "strconv" at top of file
   - Replace itoa(n) calls with strconv.Itoa(n)

5. Update errors.go:
   - If it uses itoa, update to strconv.Itoa
   - Verify all usages are updated

VERIFICATION:
- Run `go test ./kernel/ctx/ -v` after each commit
- Run `go test ./kernel/ctx/ -cover` - target 90%+ coverage
- Run `go test ./kernel/check/ -v` to ensure itoa removal doesn't break anything
- Run `go build ./...` to verify compilation

Do NOT merge to main - leave the branch for review.
```

---

## Prompt 5: CLI/REPL Robustness Improvements

```
You are working on the HypergraphGo project at /Users/bash/Documents/Projects/hottgo

=== AVAILABLE CLAUDE CODE RESOURCES ===

You have access to powerful agents, tools, MCP servers, and skills. Use these to work efficiently.

AGENTS (via Task tool with subagent_type parameter):
- Explore: Fast codebase exploration, finding files, understanding patterns
  Example: Task tool with subagent_type='Explore' prompt="Understand the CLI and REPL structure in cmd/hg"
- Plan: Designing implementation strategies before coding
  Example: Task tool with subagent_type='Plan' prompt="Plan signal handling implementation"
- general-purpose: Complex multi-step research tasks
  Example: Task tool with subagent_type='general-purpose' prompt="Analyze all error handling in cmd/hg"

BUILT-IN TOOLS:
- Read: Read file contents (Go files, images, PDFs)
- Write: Create new files
- Edit: Make targeted edits to existing files
- Glob: Find files by pattern (e.g., "cmd/hg/*.go")
- Grep: Search file contents with regex (e.g., pattern="modified|scanner")
- Bash: Run shell commands (go test, go build, git, etc.)
- WebSearch: Search the web for documentation or solutions
  Example: WebSearch query="Go signal handling best practices SIGINT"
- WebFetch: Fetch and analyze web page content

MCP SERVERS (configured in .mcp.json):
- GitHub MCP: Work with PRs, issues, code reviews
  - Run /mcp to authenticate via OAuth
  - Then ask: "Show open PRs", "Create issue for bug", "Check CI status"

SKILLS:
- /mermaid: Generate diagrams (architecture, flow, sequence)
  Example: /mermaid to visualize REPL state machine

BEST PRACTICES:
1. Start with Explore agent to understand unfamiliar code before making changes
2. Use Glob before Grep to narrow down files
3. Make multiple parallel tool calls when searches are independent
4. Use GitHub MCP to check for existing issues/PRs before starting
5. Generate diagrams with /mermaid when documenting complex relationships
6. Use WebSearch to find Go best practices for signal handling and atomic writes

=== END RESOURCES ===

SETUP:
1. Use Explore agent to understand the CLI and REPL structure
2. Create and checkout branch: `git checkout -b fix/cli-robustness`
3. All commits must be authored by watchthelight: use `git commit --author="watchthelight <admin@watchthelight.org>"`
4. Commit frequently after each logical piece of work (aim for 10-15 commits)

YOUR TASK: Add signal handling and fix robustness issues in the CLI/REPL.

CONTEXT - Issues Found:
- No signal handling - Ctrl+C loses unsaved changes without warning
- :quit safety mechanism incorrectly clears modified flag (repl.go:83-89)
- File save is not atomic (potential data corruption on write failure)
- Default scanner buffer (64KB) too small for some inputs
- scanner.Err() never checked after scan loop

SCOPE (files to modify):
- cmd/hg/repl.go (ADD signal handling, FIX quit safety, FIX scanner - ~60 lines)
- cmd/hg/io.go (FIX atomic file writes - ~20 lines)
- cmd/hg/repl_test.go (ADD tests for new behavior - ~50 lines)
- CHANGELOG.md (update with your changes)

SPECIFIC CHANGES:

1. Add Signal Handling (repl.go):
   ```go
   import (
       "os/signal"
       "syscall"
   )

   func runREPL(hg *hypergraph.Hypergraph[string], filename string) error {
       // Setup signal handling
       sigChan := make(chan os.Signal, 1)
       signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

       go func() {
           <-sigChan
           if state.modified {
               fmt.Println("\nWarning: Unsaved changes. Press Ctrl+C again to force exit.")
               <-sigChan
           }
           os.Exit(130)  // Standard SIGINT exit code
       }()
       // ... rest of REPL
   }
   ```

2. Fix :quit Safety Mechanism (repl.go:83-89):
   Current bug: clears modified flag, losing the unsaved state
   ```go
   // BAD:
   state.modified = false  // Don't do this!

   // GOOD: Use separate confirmation state
   case ":quit", ":q":
       if state.modified && !state.quitConfirmed {
           fmt.Println("Warning: unsaved changes. Use :save or :quit again to exit.")
           state.quitConfirmed = true
           return nil
       }
       return errQuit
   ```
   Add `quitConfirmed bool` to replState struct.

3. Implement Atomic File Writes (io.go):
   ```go
   func saveGraph(hg *hypergraph.Hypergraph[string], filename string) error {
       // Write to temp file first
       tmpFile := filename + ".tmp"
       f, err := os.Create(tmpFile)
       if err != nil {
           return err
       }

       if err := hg.SaveJSON(f); err != nil {
           f.Close()
           os.Remove(tmpFile)
           return err
       }
       f.Close()

       // Atomic rename
       return os.Rename(tmpFile, filename)
   }
   ```

4. Increase Scanner Buffer and Check Errors (repl.go):
   ```go
   scanner := bufio.NewScanner(os.Stdin)
   scanner.Buffer(make([]byte, 64*1024), 1024*1024)  // Allow 1MB lines

   for {
       // ... scanning loop
   }

   if err := scanner.Err(); err != nil {
       return fmt.Errorf("input error: %w", err)
   }
   ```

5. Reset quitConfirmed on Any Other Command:
   After processing any command that's not :quit, reset:
   ```go
   state.quitConfirmed = false
   ```

VERIFICATION:
- Run `go test ./cmd/hg/ -v` after each commit
- Test signal handling manually: run REPL, make changes, press Ctrl+C
- Test atomic save: verify no partial files on simulated write failure
- Run `go build ./cmd/hg/` to verify compilation

Do NOT merge to main - leave the branch for review.
```

---

## Prompt 6: CI/CD and Build Configuration Fixes

```
You are working on the HypergraphGo project at /Users/bash/Documents/Projects/hottgo

=== AVAILABLE CLAUDE CODE RESOURCES ===

You have access to powerful agents, tools, MCP servers, and skills. Use these to work efficiently.

AGENTS (via Task tool with subagent_type parameter):
- Explore: Fast codebase exploration, finding files, understanding patterns
  Example: Task tool with subagent_type='Explore' prompt="Understand the CI/CD workflow structure"
- Plan: Designing implementation strategies before coding
  Example: Task tool with subagent_type='Plan' prompt="Plan fixing the publish-deb job"
- general-purpose: Complex multi-step research tasks
  Example: Task tool with subagent_type='general-purpose' prompt="Analyze all GitHub workflows for issues"

BUILT-IN TOOLS:
- Read: Read file contents (Go files, images, PDFs, YAML)
- Write: Create new files
- Edit: Make targeted edits to existing files
- Glob: Find files by pattern (e.g., ".github/workflows/*.yml")
- Grep: Search file contents with regex (e.g., pattern="codecov|publish-deb")
- Bash: Run shell commands (go test, make, git, etc.)
- WebSearch: Search the web for documentation or solutions
  Example: WebSearch query="codecov-action v5 changelog 2024"
- WebFetch: Fetch and analyze web page content
  Example: WebFetch url="https://github.com/codecov/codecov-action/releases"

MCP SERVERS (configured in .mcp.json):
- GitHub MCP: Work with PRs, issues, code reviews, CI status
  - Run /mcp to authenticate via OAuth
  - Then ask: "Show open PRs", "Check CI status for branch", "List workflow runs"

SKILLS:
- /mermaid: Generate diagrams (architecture, flow, sequence)
  Example: /mermaid to visualize CI/CD pipeline

BEST PRACTICES:
1. Start with Explore agent to understand unfamiliar code before making changes
2. Use Glob before Grep to narrow down files
3. Make multiple parallel tool calls when searches are independent
4. Use GitHub MCP to check CI status and existing workflow runs
5. Use WebSearch/WebFetch to find latest GitHub Actions versions
6. Generate diagrams with /mermaid when documenting CI/CD flows

=== END RESOURCES ===

SETUP:
1. Use Explore agent to understand the current CI/CD configuration
2. Create and checkout branch: `git checkout -b fix/ci-build`
3. All commits must be authored by watchthelight: use `git commit --author="watchthelight <admin@watchthelight.org>"`
4. Commit frequently after each logical piece of work (aim for 8-12 commits)

YOUR TASK: Fix CI/CD issues and improve build configuration.

CONTEXT - Issues Found:
- .github/workflows/go.yml uses outdated codecov/codecov-action@v3
- .github/workflows/release.yml publish-deb job is broken (artifacts not shared)
- Missing .github/dependabot.yml for automated updates
- Makefile missing standard targets (build, clean, coverage, test-race)
- ci-linux.yml and ci-windows.yml missing race detection and cubical tests

SCOPE (files to modify):
- .github/workflows/go.yml (UPDATE codecov action - 1 line)
- .github/workflows/release.yml (FIX publish-deb job - ~15 lines)
- .github/workflows/ci-linux.yml (ADD race detection, cubical tests - ~10 lines)
- .github/workflows/ci-windows.yml (ADD cubical tests - ~5 lines)
- .github/dependabot.yml (NEW - ~15 lines)
- Makefile (ADD missing targets - ~30 lines)
- CHANGELOG.md (update with your changes)

SPECIFIC CHANGES:

1. Update Codecov Action (go.yml:52):
   Change: `uses: codecov/codecov-action@v3`
   To: `uses: codecov/codecov-action@v5`

2. Fix publish-deb Job (release.yml:125-133):
   Add artifact download step before using dist/*.deb:
   ```yaml
   publish-deb:
     needs: [goreleaser]
     runs-on: ubuntu-latest
     steps:
       - name: Download release artifacts
         run: |
           gh release download ${{ github.ref_name }} \
             --pattern '*.deb' \
             --dir dist/
         env:
           GH_TOKEN: ${{ secrets.GITHUB_TOKEN }}
       # ... rest of publish steps
   ```

3. Add Dependabot Configuration (.github/dependabot.yml):
   ```yaml
   version: 2
   updates:
     - package-ecosystem: "github-actions"
       directory: "/"
       schedule:
         interval: "weekly"
       commit-message:
         prefix: "ci"

     - package-ecosystem: "gomod"
       directory: "/"
       schedule:
         interval: "weekly"
       commit-message:
         prefix: "deps"
   ```

4. Update ci-linux.yml:
   Add race detection to test step:
   ```yaml
   - name: Test with race detection
     run: go test -race -v ./...

   - name: Test cubical features
     run: go test -v -tags cubical ./...
   ```

5. Update ci-windows.yml:
   Add cubical tests (race detection not reliable on Windows):
   ```yaml
   - name: Test cubical features
     run: go test -v -tags cubical ./...
   ```

6. Expand Makefile:
   ```makefile
   .PHONY: build clean test test-race test-cubical coverage lint check help

   build:
   	go build -o bin/hg ./cmd/hg

   clean:
   	rm -rf bin/ coverage.out

   test:
   	go test ./...

   test-race:
   	go test -race ./...

   test-cubical:
   	go test -tags cubical ./...

   coverage:
   	go test -coverprofile=coverage.out ./...
   	go tool cover -html=coverage.out -o coverage.html

   lint:
   	golangci-lint run

   check: test-race test-cubical lint
   	go vet ./...

   help:
   	@echo "Available targets:"
   	@echo "  build        - Build the hg binary"
   	@echo "  clean        - Remove build artifacts"
   	@echo "  test         - Run tests"
   	@echo "  test-race    - Run tests with race detection"
   	@echo "  test-cubical - Run cubical feature tests"
   	@echo "  coverage     - Generate coverage report"
   	@echo "  lint         - Run golangci-lint"
   	@echo "  check        - Run all checks"
   ```

VERIFICATION:
- Verify YAML syntax: `python -c "import yaml; yaml.safe_load(open('.github/workflows/go.yml'))"`
- Run `make help` to verify Makefile
- Run `make test` and `make test-race` locally
- Push branch and verify CI runs successfully

Do NOT merge to main - leave the branch for review.
```

---

## Summary

| Instance | Branch | Focus Area | Key Fixes |
|----------|--------|------------|-----------|
| 1 | fix/ast-raw-terms | AST raw terms, print optimization | Add RId/RRefl/RJ, fix O(n^2) collectSpine |
| 2 | fix/eval-correctness | Evaluation engine | Fix alphaEq, EvalFill, EvalUABeta |
| 3 | fix/hypergraph-issues | Hypergraph package | Fix doc.go, add tests, determinism |
| 4 | fix/kernel-ctx | Context and kernel | Tests, remove custom itoa, nil validation |
| 5 | fix/cli-robustness | CLI/REPL | Signal handling, atomic saves, quit fix |
| 6 | fix/ci-build | CI/CD configuration | Codecov v5, fix publish-deb, dependabot |

Each prompt is designed to be:
- **Self-contained**: All resources and context included in each prompt
- **Independent**: No dependencies between the 6 work streams
- **Substantial**: Each requires significant exploration and implementation
- **Well-scoped**: Clear deliverables and verification criteria
- **Consistent**: Same commit author and workflow patterns
