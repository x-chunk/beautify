// Package width measures text in terminal cells and splits it into
// clusters, the smallest units that layout may not cut.
package width

import (
	"sort"
	"unicode"
)

// runeRange is an inclusive range of code points.
type runeRange struct{ lo, hi rune }

// wideRanges lists runes that occupy two cells, sorted and disjoint. The
// emoji blocks are coarse on purpose: exact tables need a generator, which
// goals.md defers.
var wideRanges = []runeRange{
	{0x1100, 0x115F},
	{0x231A, 0x231B}, {0x23E9, 0x23EC}, {0x23F0, 0x23F0}, {0x23F3, 0x23F3},
	{0x25FD, 0x25FE}, {0x2614, 0x2615}, {0x2648, 0x2653}, {0x267F, 0x267F},
	{0x2693, 0x2693}, {0x26A1, 0x26A1}, {0x26AA, 0x26AB}, {0x26BD, 0x26BE},
	{0x26C4, 0x26C5}, {0x26CE, 0x26CE}, {0x26D4, 0x26D4}, {0x26EA, 0x26EA},
	{0x26F2, 0x26F3}, {0x26F5, 0x26F5}, {0x26FA, 0x26FA}, {0x26FD, 0x26FD},
	{0x2705, 0x2705}, {0x270A, 0x270B}, {0x2728, 0x2728}, {0x274C, 0x274C},
	{0x274E, 0x274E}, {0x2753, 0x2755}, {0x2757, 0x2757}, {0x2795, 0x2797},
	{0x27B0, 0x27B0}, {0x27BF, 0x27BF}, {0x2B1B, 0x2B1C}, {0x2B50, 0x2B50},
	{0x2B55, 0x2B55},
	{0x2E80, 0x303E}, {0x3041, 0x33FF}, {0x3400, 0x4DBF}, {0x4E00, 0x9FFF},
	{0xA000, 0xA4CF}, {0xA960, 0xA97F}, {0xAC00, 0xD7A3}, {0xF900, 0xFAFF},
	{0xFE10, 0xFE19}, {0xFE30, 0xFE6F}, {0xFF00, 0xFF60}, {0xFFE0, 0xFFE6},
	{0x1F300, 0x1F64F}, {0x1F680, 0x1F6FF}, {0x1F900, 0x1F9FF},
	{0x1FA70, 0x1FAFF},
	{0x20000, 0x2FFFD}, {0x30000, 0x3FFFD},
}

// firstSpecial is the lowest code point whose width can differ from 1 or
// that can extend a cluster.
const firstSpecial = 0x0300

// runeWidth returns the cell width of a single rune in Long Mode.
func runeWidth(r rune) int {
	switch {
	case r < firstSpecial:
		// The soft hyphen is the only format character here and terminals
		// show it, so it has width 1 like the rest of the range.
		return 1
	case isHangulExtender(r), unicode.In(r, unicode.Mn, unicode.Me, unicode.Cf):
		// Cf covers U+200B (zero-width space) and U+200D.
		return 0
	case isWide(r):
		return 2
	default:
		return 1
	}
}

// isWide reports whether r is in wideRanges.
func isWide(r rune) bool {
	i := sort.Search(len(wideRanges), func(i int) bool { return wideRanges[i].hi >= r })
	return i < len(wideRanges) && wideRanges[i].lo <= r
}

// isHangulExtender reports whether r is a medial vowel or a final
// consonant, which terminals draw inside the preceding syllable.
func isHangulExtender(r rune) bool {
	return r >= 0x1160 && r <= 0x11FF || r >= 0xD7B0 && r <= 0xD7FF
}
