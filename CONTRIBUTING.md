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

## Commit Messages

Be clear. Be concise. Start with a verb.

```
feat(kernel): add J eliminator for identity types
fix(nbe): correct reification for nested Pi types
docs: update roadmap with Phase 5 status
```

## Questions?

Open an issue. Don't email.
