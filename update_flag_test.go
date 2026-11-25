package beautify

// Registers the -update flag so "go test ./... -update" works for every
// package.
import _ "go.xchunk.org/beautify/internal/golden"
