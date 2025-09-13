# Contributing

Thank you for your interest in HypergraphGo.

## Ground Rules
- The **kernel** (`internal/kernel`) must remain minimal, panic-free, and fully documented. All sugar/elaboration lives outside the kernel and must re-check through it.
- Error messages should be deterministic; update golden tests if wording changes.
- Prefer small, isolated commits with tests.

## Development
```bash
go test ./...
go run ./cmd/hottgo -version
```

## Code of Conduct
Please read and follow our Code of Conduct.

## Security
Report security issues privately via the instructions in SECURITY.md.
