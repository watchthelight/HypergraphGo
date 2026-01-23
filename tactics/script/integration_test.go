package script

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/watchthelight/HypergraphGo/kernel/check"
)

// TestAllExampleProofs verifies all .htt proof files in examples/proofs/.
// This is an integration test that ensures the entire proof suite remains valid.
func TestAllExampleProofs(t *testing.T) {
	// Find project root (go up from tactics/script to project root)
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// Navigate up to project root
	projectRoot := filepath.Join(wd, "..", "..")
	proofsDir := filepath.Join(projectRoot, "examples", "proofs")

	// Check if the directory exists
	if _, err := os.Stat(proofsDir); os.IsNotExist(err) {
		t.Skipf("examples/proofs directory not found at %s", proofsDir)
	}

	// Collect all .htt files
	var proofFiles []string
	err = filepath.Walk(proofsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".htt" {
			proofFiles = append(proofFiles, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk proofs directory: %v", err)
	}

	if len(proofFiles) == 0 {
		t.Fatal("no .htt files found in examples/proofs")
	}

	t.Logf("Found %d proof files to verify", len(proofFiles))

	totalTheorems := 0
	passedTheorems := 0

	for _, file := range proofFiles {
		relPath, _ := filepath.Rel(proofsDir, file)
		t.Run(relPath, func(t *testing.T) {
			content, err := os.ReadFile(file)
			if err != nil {
				t.Fatalf("failed to read file: %v", err)
			}

			script, err := ParseString(string(content))
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			checker := check.NewCheckerWithStdlib()
			result := Execute(script, checker)

			fileTheorems := 0
			filePassed := 0

			for _, item := range result.Items {
				if item.Kind == ItemTheorem {
					fileTheorems++
					if item.Success {
						filePassed++
					} else {
						t.Errorf("theorem %s failed: %v", item.Name, item.Error)
					}
				}
			}

			totalTheorems += fileTheorems
			passedTheorems += filePassed

			if filePassed < fileTheorems {
				t.Errorf("%d/%d theorems passed", filePassed, fileTheorems)
			} else {
				t.Logf("%d/%d theorems verified", filePassed, fileTheorems)
			}
		})
	}

	t.Logf("Total: %d/%d theorems verified across %d files", passedTheorems, totalTheorems, len(proofFiles))
}

// TestProofFileCount ensures we have the expected number of proof files.
// This prevents accidental deletion of proof files.
func TestProofFileCount(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	projectRoot := filepath.Join(wd, "..", "..")
	proofsDir := filepath.Join(projectRoot, "examples", "proofs")

	if _, err := os.Stat(proofsDir); os.IsNotExist(err) {
		t.Skipf("examples/proofs directory not found at %s", proofsDir)
	}

	var count int
	err = filepath.Walk(proofsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".htt" {
			count++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk proofs directory: %v", err)
	}

	// We expect at least 20 proof files (the current count)
	minExpected := 20
	if count < minExpected {
		t.Errorf("expected at least %d proof files, found %d", minExpected, count)
	}

	t.Logf("Found %d proof files", count)
}
