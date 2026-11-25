package dra

import "go.xchunk.org/beautify/internal/width"

// WrapMode selects where Wrap may break a line. Unknown values behave like
// WrapWords.
type WrapMode uint8

const (
	// WrapWords breaks at spaces and splits only words that are wider than
	// the line.
	WrapWords WrapMode = iota
	// WrapAnywhere breaks between any two clusters.
	WrapAnywhere
)

// Wrap re-flows the lines of m that are wider than width. Styles are kept
// across breaks and empty lines are preserved. With a width of 0 or less the
// text is unchanged.
//
// WrapWords drops the spaces at a break. Only U+0020 counts as a space. The
// leading spaces of a source line indent its first output line; they are
// dropped when the first word does not fit after them.
//
// No result line is wider than width, with one exception: a single cluster
// wider than width, which is a wide cluster at width 1, is placed alone on
// its line.
func Wrap(m Model, width int, mode WrapMode) Model {
	if width <= 0 || m.Width() <= width {
		return composed(flatten(&m))
	}
	w := wrapper{limit: width}
	for _, l := range flatten(&m) {
		switch {
		case l.width <= width:
			w.b.addLine(l)
		case mode == WrapAnywhere:
			w.split(w.clusters(l))
		default:
			w.words(w.clusters(l))
		}
		w.b.endLine()
	}
	return w.b.model()
}

// wrapper fills output lines of at most limit cells.
type wrapper struct {
	b     builder
	limit int
	buf   []span // reused by clusters
}

// clusters splits l into spans of one cluster each. The result is valid
// until the next call.
func (w *wrapper) clusters(l line) []span {
	w.buf = w.buf[:0]
	for _, s := range l.spans {
		for s.text != "" {
			size, cells := width.Next(s.text, s.mode.width())
			w.buf = append(w.buf, span{text: s.text[:size], style: s.style, mode: s.mode, width: cells})
			s.text = s.text[size:]
		}
	}
	return w.buf
}

// split appends cells to the open line and breaks it wherever the next cell
// does not fit. A cell of width 0 always fits.
func (w *wrapper) split(cells []span) {
	for _, c := range cells {
		if w.b.width > 0 && c.width > 0 && w.b.width+c.width > w.limit {
			w.b.endLine()
		}
		w.b.add(c)
	}
}

// words appends cells to the open line, which must be empty, and breaks at
// runs of spaces.
func (w *wrapper) words(cells []span) {
	hasWord := false // whether the open line holds a word
	for len(cells) > 0 {
		gap, rest := cut(cells, true)
		word, rest := cut(rest, false)
		cells = rest

		gapWidth, wordWidth := total(gap), total(word)
		switch {
		case len(word) == 0:
			// Trailing spaces. A line of nothing but spaces is indentation
			// and follows the same rule.
			if w.b.width+gapWidth <= w.limit {
				w.append(gap)
			}
			continue
		case w.b.width+gapWidth+wordWidth <= w.limit:
			w.append(gap)
			w.append(word)
		default:
			if hasWord {
				w.b.endLine()
			}
			w.split(word)
		}
		hasWord = true
	}
}

func (w *wrapper) append(cells []span) {
	for _, c := range cells {
		w.b.add(c)
	}
}

// cut splits cells after the leading run of spaces, or of everything else.
func cut(cells []span, spaces bool) (run, rest []span) {
	n := 0
	for n < len(cells) && (cells[n].text == " ") == spaces {
		n++
	}
	return cells[:n], cells[n:]
}

// total returns the combined width of cells.
func total(cells []span) int {
	sum := 0
	for _, c := range cells {
		sum += c.width
	}
	return sum
}
