package eval_test

import (
	"testing"

	"github.com/watchthelight/HypergraphGo/internal/ast"
	"github.com/watchthelight/HypergraphGo/internal/eval"
)

func BenchmarkNormalize_SpineDepth32(b *testing.B) {
	args := make([]ast.Term, 32)
	for i := range args {
		args[i] = ast.Var{Ix: i % 3}
	}
	spine := ast.MkApps(ast.Global{Name: "f"}, args...)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = eval.Normalize(spine)
	}
}
