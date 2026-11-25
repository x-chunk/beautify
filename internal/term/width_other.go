//go:build !(linux || darwin || freebsd || netbsd || openbsd || dragonfly)

package term

import "os"

// fileWidth reports an unknown width on platforms without a size query.
func fileWidth(*os.File) int {
	return 0
}
