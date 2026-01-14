// Package version provides build and version information for HoTTGo.
//
// Version information is set at build time via ldflags:
//
//	go build -ldflags "-X github.com/watchthelight/HypergraphGo/internal/version.Version=1.0.0 \
//	                   -X github.com/watchthelight/HypergraphGo/internal/version.Commit=abc123 \
//	                   -X github.com/watchthelight/HypergraphGo/internal/version.Date=2026-01-01"
//
// # Variables
//
//   - Version: semantic version string (e.g., "1.8.2")
//   - Commit: git commit hash
//   - Date: build date
//
// When not set via ldflags, defaults are used (empty strings for
// Commit and Date, current version for Version).
package version
