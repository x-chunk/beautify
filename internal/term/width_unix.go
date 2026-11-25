//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly

package term

import (
	"os"
	"syscall"
	"unsafe"
)

// fileWidth queries the terminal size and returns the number of columns, or
// 0 when f is not a terminal.
func fileWidth(f *os.File) int {
	conn, err := f.SyscallConn()
	if err != nil {
		return 0
	}
	// Matches struct winsize of the kernel.
	var size struct{ rows, cols, xpixel, ypixel uint16 }
	var errno syscall.Errno
	// Control keeps the descriptor open during the call. File.Fd would also
	// switch the file to blocking mode.
	err = conn.Control(func(fd uintptr) {
		_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, fd,
			syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&size)))
	})
	if err != nil || errno != 0 {
		return 0
	}
	return int(size.cols)
}
