package dra

import (
	"errors"
	"fmt"

	"go.xchunk.org/beautify/internal/ansi"
)

// ErrInvalidColor is returned by Hex for input that is not a hex color.
var ErrInvalidColor = errors.New("dra: invalid color")

// Color is a terminal color. The zero value is "unset" and inherits. Colors
// are comparable with ==.
type Color struct {
	c ansi.Color
}

// TerminalDefault is the terminal's own default color. Unlike the zero
// Color it does not inherit: it resets to the default inside a colored
// parent.
var TerminalDefault = Color{ansi.Color{Kind: ansi.Default}}

// The 16 colors of the ANSI palette. Their exact look depends on the
// terminal theme.
var (
	Black         = ANSI(0)
	Red           = ANSI(1)
	Green         = ANSI(2)
	Yellow        = ANSI(3)
	Blue          = ANSI(4)
	Magenta       = ANSI(5)
	Cyan          = ANSI(6)
	White         = ANSI(7)
	BrightBlack   = ANSI(8)
	BrightRed     = ANSI(9)
	BrightGreen   = ANSI(10)
	BrightYellow  = ANSI(11)
	BrightBlue    = ANSI(12)
	BrightMagenta = ANSI(13)
	BrightCyan    = ANSI(14)
	BrightWhite   = ANSI(15)
)

// ANSI returns a color of the 16-color palette. n is taken modulo 16.
func ANSI(n uint8) Color {
	return Color{ansi.Color{Kind: ansi.Indexed16, Index: n % 16}}
}

// ANSI256 returns a color of the 256-color palette.
func ANSI256(n uint8) Color {
	return Color{ansi.Color{Kind: ansi.Indexed256, Index: n}}
}

// RGB returns a 24-bit color.
func RGB(r, g, b uint8) Color {
	return Color{ansi.Color{Kind: ansi.RGB, R: r, G: g, B: b}}
}

// Hex parses a color written as "#rrggbb", "rrggbb", "#rgb" or "rgb". Case
// does not matter; the short form repeats every digit. For any other input
// it returns the unset Color and an error that wraps ErrInvalidColor.
func Hex(s string) (Color, error) {
	digits := s
	if digits != "" && digits[0] == '#' {
		digits = digits[1:]
	}
	var v [6]uint8
	switch len(digits) {
	case 3, 6:
		step := 6 / len(digits)
		for i := range v {
			d, ok := hexDigit(digits[i/step])
			if !ok {
				return Color{}, fmt.Errorf("%w: %q", ErrInvalidColor, s)
			}
			v[i] = d
		}
	default:
		return Color{}, fmt.Errorf("%w: %q", ErrInvalidColor, s)
	}
	return RGB(v[0]<<4|v[1], v[2]<<4|v[3], v[4]<<4|v[5]), nil
}

func hexDigit(c byte) (uint8, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	default:
		return 0, false
	}
}

// IsSet reports whether c is a color rather than "inherit".
// TerminalDefault counts as set.
func (c Color) IsSet() bool {
	return c.c.Kind != ansi.None
}
