package width

import (
	"unicode"
	"unicode/utf8"
)

// Mode selects how text is split into clusters and measured.
type Mode uint8

const (
	// Real treats every rune as one cluster of width 1. It needs no tables;
	// the trade-off is that combining marks, CJK and emoji are measured
	// differently from what a terminal shows.
	Real Mode = iota
	// Long groups runes into grapheme clusters of width 0, 1 or 2.
	Long
)

const (
	zeroWidthJoiner   = 0x200D
	emojiPresentation = 0xFE0F
)

// Next returns the byte length and cell width of the first cluster of s.
// For empty s it returns 0, 0. Modes other than Long behave like Real. An
// invalid UTF-8 byte is always a cluster of its own with width 1.
func Next(s string, mode Mode) (size, width int) {
	if s == "" {
		return 0, 0
	}
	if s[0] < utf8.RuneSelf && (len(s) == 1 || s[1] < utf8.RuneSelf) {
		return 1, 1
	}
	base, size := utf8.DecodeRuneInString(s)
	if mode != Long || base == '\n' || base == utf8.RuneError && size == 1 {
		return size, 1
	}

	width = runeWidth(base)
	if isRegionalIndicator(base) {
		if r, n := utf8.DecodeRuneInString(s[size:]); isRegionalIndicator(r) {
			size += n
			width = 2
		}
	}
	afterJoiner := false
	for size < len(s) {
		r, n := utf8.DecodeRuneInString(s[size:])
		if r == '\n' || r == utf8.RuneError && n == 1 {
			break
		}
		if !afterJoiner && !isExtender(r) {
			break
		}
		if r == emojiPresentation {
			width = 2
		}
		afterJoiner = r == zeroWidthJoiner
		size += n
	}
	return size, width
}

// String returns the cell width of s, which must not contain '\n'.
func String(s string, mode Mode) int {
	if mode != Long {
		return utf8.RuneCountInString(s)
	}
	total := 0
	for s != "" {
		size, w := Next(s, mode)
		total += w
		s = s[size:]
	}
	return total
}

// Separable reports whether a and b can be joined into one string without
// merging the clusters at the boundary, so that the width of the result is
// the sum of both widths. It can report false for strings that would join
// safely.
func Separable(a, b string, mode Mode) bool {
	if mode != Long || a == "" || b == "" {
		return true
	}
	first, _ := utf8.DecodeRuneInString(b)
	if first < firstSpecial {
		if last := a[len(a)-1]; last < utf8.RuneSelf {
			return true
		}
	}
	last, _ := utf8.DecodeLastRuneInString(a)
	switch {
	case last == zeroWidthJoiner, isExtender(first):
		return false
	default:
		return !(isRegionalIndicator(last) && isRegionalIndicator(first))
	}
}

// isExtender reports whether r attaches to the cluster before it.
func isExtender(r rune) bool {
	if r < firstSpecial {
		return false
	}
	switch {
	case r == zeroWidthJoiner,
		r >= 0xFE00 && r <= 0xFE0F, r >= 0xE0100 && r <= 0xE01EF, // variation selectors
		r >= 0x1F3FB && r <= 0x1F3FF, // skin tones
		r >= 0xE0020 && r <= 0xE007F, // tags
		isHangulExtender(r):
		return true
	default:
		return unicode.In(r, unicode.Mn, unicode.Me, unicode.Mc)
	}
}

// isRegionalIndicator reports whether r is one half of a flag.
func isRegionalIndicator(r rune) bool {
	return r >= 0x1F1E6 && r <= 0x1F1FF
}
