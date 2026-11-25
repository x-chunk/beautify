package golden

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckMatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.golden")
	if err := os.WriteFile(path, []byte("\x1b[0mhi\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := check(path, []byte("\x1b[0mhi\r\n"), false); err != nil {
		t.Errorf("check() = %v, want nil", err)
	}
}

func TestCheckMismatchQuotesBothValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.golden")
	if err := os.WriteFile(path, []byte("\x1b[0mwant"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := check(path, []byte("\x1b[0mgot"), false)
	if err == nil {
		t.Fatal("check() = nil, want an error")
	}
	for _, sub := range []string{`"\x1b[0mwant"`, `"\x1b[0mgot"`} {
		if !strings.Contains(err.Error(), sub) {
			t.Errorf("error %q does not contain %s", err, sub)
		}
	}
}

func TestCheckMissingFileMentionsUpdate(t *testing.T) {
	err := check(filepath.Join(t.TempDir(), "none.golden"), nil, false)
	if err == nil || !strings.Contains(err.Error(), "-update") {
		t.Errorf("check() = %v, want a hint about -update", err)
	}
}

func TestCheckUpdateWritesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "testdata", "x.golden")
	if err := check(path, []byte("new"), true); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil || string(got) != "new" {
		t.Errorf("file = %q, %v, want %q", got, err, "new")
	}
}

func TestAssertReadsTestdata(t *testing.T) {
	Assert(t, "assert_bytes", []byte("\x1b[0;1mbold\x1b[0m\n"))
}
