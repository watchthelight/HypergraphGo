package script

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/watchthelight/HypergraphGo/kernel/check"
)

// BenchmarkProofVerification benchmarks the verification of all proof files.
func BenchmarkProofVerification(b *testing.B) {
	wd, err := os.Getwd()
	if err != nil {
		b.Fatalf("failed to get working directory: %v", err)
	}

	projectRoot := filepath.Join(wd, "..", "..")
	proofsDir := filepath.Join(projectRoot, "examples", "proofs")

	if _, err := os.Stat(proofsDir); os.IsNotExist(err) {
		b.Skipf("examples/proofs directory not found at %s", proofsDir)
	}

	// Collect all .htt files
	var proofFiles []struct {
		path    string
		content string
	}
	err = filepath.Walk(proofsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".htt" {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			proofFiles = append(proofFiles, struct {
				path    string
				content string
			}{path: path, content: string(content)})
		}
		return nil
	})
	if err != nil {
		b.Fatalf("failed to walk proofs directory: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pf := range proofFiles {
			script, err := ParseString(pf.content)
			if err != nil {
				b.Fatalf("parse error in %s: %v", pf.path, err)
			}
			checker := check.NewCheckerWithStdlib()
			Execute(script, checker)
		}
	}
}

// BenchmarkSingleProofFile benchmarks individual proof files.
func BenchmarkSingleProofFile(b *testing.B) {
	wd, err := os.Getwd()
	if err != nil {
		b.Fatalf("failed to get working directory: %v", err)
	}

	projectRoot := filepath.Join(wd, "..", "..")
	proofsDir := filepath.Join(projectRoot, "examples", "proofs")

	if _, err := os.Stat(proofsDir); os.IsNotExist(err) {
		b.Skipf("examples/proofs directory not found at %s", proofsDir)
	}

	// Benchmark specific files
	files := []string{
		"hott/path_algebra.htt",
		"hott/equivalences.htt",
		"integration/groups.htt",
		"integration/peano.htt",
	}

	for _, file := range files {
		path := filepath.Join(proofsDir, file)
		content, err := os.ReadFile(path)
		if err != nil {
			continue // Skip if file doesn't exist
		}

		b.Run(file, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				script, err := ParseString(string(content))
				if err != nil {
					b.Fatalf("parse error: %v", err)
				}
				checker := check.NewCheckerWithStdlib()
				Execute(script, checker)
			}
		})
	}
}

// BenchmarkParseProofFile benchmarks just the parsing of proof files.
func BenchmarkParseProofFile(b *testing.B) {
	wd, err := os.Getwd()
	if err != nil {
		b.Fatalf("failed to get working directory: %v", err)
	}

	projectRoot := filepath.Join(wd, "..", "..")
	path := filepath.Join(projectRoot, "examples", "proofs", "integration", "groups.htt")

	content, err := os.ReadFile(path)
	if err != nil {
		b.Skipf("file not found: %s", path)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseString(string(content))
		if err != nil {
			b.Fatalf("parse error: %v", err)
		}
	}
}
