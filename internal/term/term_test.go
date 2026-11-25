package term

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	_ "go.xchunk.org/beautify/internal/golden" // registers the -update flag
)

func TestNotATerminal(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	defer w.Close()
	file, err := os.Create(filepath.Join(t.TempDir(), "out"))
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	closed, err := os.Create(filepath.Join(t.TempDir(), "closed"))
	if err != nil {
		t.Fatal(err)
	}
	closed.Close()

	var nilFile *os.File
	tests := []struct {
		name string
		w    interface{ Write([]byte) (int, error) }
	}{
		{"buffer", &bytes.Buffer{}},
		{"nil writer", nil},
		{"nil file", nilFile},
		{"pipe", w},
		{"regular file", file},
		{"closed file", closed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsTerminal(tt.w) {
				t.Error("IsTerminal() = true")
			}
			if n, ok := Width(tt.w); n != 0 || ok {
				t.Errorf("Width() = %d, %v, want 0, false", n, ok)
			}
		})
	}
}

func TestCharacterDevice(t *testing.T) {
	null, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Skip(err)
	}
	defer null.Close()
	// A character device without a size, such as the null device, must not
	// report a width.
	if n, ok := Width(null); n != 0 || ok {
		t.Errorf("Width() = %d, %v, want 0, false", n, ok)
	}
}
