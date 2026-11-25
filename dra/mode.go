package dra

import "go.xchunk.org/beautify/internal/width"

// Mode selects how text is measured and cut.
type Mode uint8

const (
	// ModeReal counts every rune as one cell. It is the default and the
	// fastest mode. Combining marks, CJK and emoji are measured differently
	// from what a terminal shows.
	ModeReal Mode = iota
	// ModeLong groups runes into grapheme clusters of width 0, 1 or 2.
	ModeLong
)

// width converts m to the mode of the width package.
func (m Mode) width() width.Mode {
	if m == ModeLong {
		return width.Long
	}
	return width.Real
}
