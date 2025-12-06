//go:build !cubical

package ast

import "bytes"

// tryWriteExtension is the default implementation when no extensions are enabled.
// Returns false, indicating the term was not handled.
func tryWriteExtension(b *bytes.Buffer, t Term) bool {
	return false
}
