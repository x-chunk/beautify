// Package term answers questions about the terminal behind a writer.
package term

import (
	"io"
	"os"
)

// IsTerminal reports whether w is a terminal. Any character device counts,
// so it is also true for the null device. It is false for writers other than
// *os.File.
func IsTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok || f == nil {
		return false
	}
	info, err := f.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// Width returns the terminal width in cells for w. It returns 0 and false
// when w is not a terminal, when the platform cannot tell, and when the
// terminal reports a width of 0.
func Width(w io.Writer) (int, bool) {
	f, ok := w.(*os.File)
	if !ok || f == nil {
		return 0, false
	}
	n := fileWidth(f)
	return n, n > 0
}
