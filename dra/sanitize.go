package dra

import (
	"strings"
	"unicode/utf8"

	"go.xchunk.org/beautify/internal/width"
)

// defaultTabWidth is the number of spaces a tab becomes unless an option
// says otherwise.
const defaultTabWidth = 4

// normalize prepares text for a Model. Invalid UTF-8 bytes become U+FFFD,
// CRLF becomes LF, tabs become tabWidth spaces and, when sanitize is set,
// unsafe runes become U+FFFD.
//
// cmap holds one entry per rune of the original text, or is nil. The
// returned colormap matches the returned text; it is cmap itself when the
// text needs no change, so callers must treat it as read-only.
func normalize(text string, cmap []int8, tabWidth int, sanitize bool) (string, []int8) {
	if isNormal(text, sanitize) {
		return text, cmap
	}
	var b strings.Builder
	b.Grow(len(text))
	var out []int8
	if cmap != nil {
		out = make([]int8, 0, len(cmap))
	}
	add := func(r rune, i int) {
		b.WriteRune(r)
		if cmap != nil {
			out = append(out, cmap[i])
		}
	}
	i := 0 // rune index in the original text
	for pos, r := range text {
		switch {
		case r == '\r' && strings.HasPrefix(text[pos+1:], "\n"):
			// The newline that follows is added on its own turn.
		case r == '\t':
			for range tabWidth {
				add(' ', i)
			}
		case sanitize && isUnsafe(r):
			add(utf8.RuneError, i)
		default:
			// Ranging over a string already decodes each invalid byte as
			// U+FFFD.
			add(r, i)
		}
		i++
	}
	return b.String(), out
}

// isNormal reports whether normalize would leave text unchanged.
func isNormal(text string, sanitize bool) bool {
	if !utf8.ValidString(text) {
		return false
	}
	for _, r := range text {
		if r == '\t' || r == '\r' || sanitize && isUnsafe(r) {
			return false
		}
	}
	return true
}

// isUnsafe reports whether r can change the state of the terminal or the
// order in which text is displayed.
func isUnsafe(r rune) bool {
	switch {
	case r < 0x20:
		return r != '\n'
	case r < 0x7F:
		return false
	case r <= 0x9F: // DEL and C1 controls
		return true
	case r >= 0x2028 && r <= 0x202E: // separators, bidi embeddings and overrides
		return true
	default:
		return r >= 0x2066 && r <= 0x2069 // bidi isolates
	}
}

// unifyClusters gives every rune of a Long Mode cluster the colormap entry
// of the cluster's first rune, in place. A style change inside a cluster
// would split it with an escape sequence and break its rendering.
func unifyClusters(text string, cmap []int8) {
	if cmap == nil {
		return
	}
	for i := 0; text != ""; {
		size, _ := width.Next(text, width.Long)
		n := utf8.RuneCountInString(text[:size])
		for j := i + 1; j < i+n; j++ {
			cmap[j] = cmap[i]
		}
		i += n
		text = text[size:]
	}
}
