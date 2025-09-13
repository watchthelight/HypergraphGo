# Contributing

- Do not add inference or sugar to `internal/kernel`. All non-kernel code must re-check through the kernel.
- Keep error text stable; if you must change messages, update golden tests.
- No panics in kernel or checker code; return typed errors.
- Prefer small, isolated commits with tests. If a rule isn’t documented, it isn’t implemented.
