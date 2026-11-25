package blocks

import (
	"go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"
	"testing"

	_ "go.xchunk.org/beautify/internal/golden" // registers the -update flag
)

// TestNoPrivileges verifies that blocks are written only against the public
// API, the same way a user writes a custom block.
func TestNoPrivileges(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), name, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !allowedImport(path) {
				t.Errorf("%s imports %s", name, path)
			}
		}
		checked++
	}
	if checked == 0 {
		t.Error("no files were checked")
	}
}

// allowedImport reports whether a block may import path.
func allowedImport(path string) bool {
	const module = "go.xchunk.org/beautify"
	if path == module || path == module+"/dra" {
		return true
	}
	return !strings.HasPrefix(path, module+"/") && !strings.Contains(path, "/internal/")
}

func TestAllowedImport(t *testing.T) {
	tests := map[string]bool{
		"fmt":                                  true,
		"go.xchunk.org/beautify":               true,
		"go.xchunk.org/beautify/dra":           true,
		"go.xchunk.org/beautify/internal/ansi": false,
		"go.xchunk.org/beautify/other":         false,
		"example.com/x/internal/y":             false,
	}
	for path, want := range tests {
		if got := allowedImport(path); got != want {
			t.Errorf("allowedImport(%q) = %v, want %v", path, got, want)
		}
	}
}
