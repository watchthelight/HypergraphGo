# Security Policy

## Reporting Vulnerabilities

If you discover a security vulnerability, **do not open a public issue**.

### Private Disclosure

Email: **security@watchthelight.org**

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if you have one)

### What to Expect

- Acknowledgment within 48 hours
- Assessment within 7 days
- Fix timeline depends on severity

### Public Disclosure

We practice coordinated disclosure:
1. Reporter notifies us privately
2. We assess and develop a fix
3. We release the fix
4. Public disclosure after patch is available

### Scope

This policy covers:
- The HoTTGo kernel (`kernel/`)
- The evaluation engine (`internal/eval/`)
- CLI tools (`cmd/`)

For issues in dependencies, report upstream.

## Supported Versions

Only the latest release is actively supported with security updates.

## Not Security Issues

- Performance problems
- Feature requests
- Documentation errors

Use regular GitHub issues for these.
