package dra

import (
	"strings"
	"unicode/utf8"

	"go.xchunk.org/beautify/internal/width"
)

// span is a piece of a line with one style. Its text has no newlines and
// starts and ends at cluster boundaries.
type span struct {
	text  string
	style Style // relative: inherits from what the model is placed in
	mode  Mode  // mode the text was measured with
	width int
}

// line is one line of a composed model. Lines are shared between models and
// never changed.
type line struct {
	spans []span
	width int
}

// composed returns a model of the given lines. It takes ownership of the
// slice. Styles are stored per span instead of in a palette, because merging
// the palettes of many models can exceed what a colormap can address.
func composed(lines []line) Model {
	m := Model{lines: lines}
	if len(lines) == 0 {
		return Model{}
	}
	m.size = metrics{
		first:  lines[0].width,
		last:   lines[len(lines)-1].width,
		height: len(lines),
	}
	for _, l := range lines {
		m.size.max = max(m.size.max, l.width)
		for _, s := range l.spans {
			if s.mode == ModeLong {
				m.mode = ModeLong
			}
		}
	}
	return m
}

// flatten returns the lines of m. An unfilled slot counts as empty.
func flatten(m *Model) []line {
	if m.leaf == nil {
		return m.lines
	}
	if m.IsEmpty() {
		return nil
	}
	var b builder
	b.nested(m, Style{})
	b.endLine()
	return b.lines
}

// firstLine returns the first line of m, or an empty line. Operations that
// expect a single-line argument use it to ignore the rest.
func firstLine(m Model) line {
	if lines := flatten(&m); len(lines) > 0 {
		return lines[0]
	}
	return line{}
}

// spaces returns an unstyled span of n spaces, so that padding inherits its
// look from the outer context.
func spaces(n int) span {
	const cached = "                                                                "
	if n <= 0 {
		return span{}
	}
	if n <= len(cached) {
		return span{text: cached[:n], width: n}
	}
	return span{text: strings.Repeat(" ", n), width: n}
}

// blankLine returns a line of n spaces.
func blankLine(n int) line {
	if n <= 0 {
		return line{}
	}
	return line{spans: []span{spaces(n)}, width: n}
}

// builder assembles lines and merges adjacent spans that look the same.
type builder struct {
	lines []line
	spans []span // finished spans of the open line
	width int    // width of the open line, the pending span included

	// The pending span can still grow. Its text is in buf once something was
	// merged into it, so repeated merging does not copy quadratically.
	pending span
	buf     []byte
}

// add appends s to the open line.
func (b *builder) add(s span) {
	if s.text == "" {
		return
	}
	b.width += s.width
	if b.pending.text != "" && b.pending.style == s.style && b.pending.mode == s.mode && b.separable(s) {
		if len(b.buf) == 0 {
			b.buf = append(b.buf, b.pending.text...)
		}
		b.buf = append(b.buf, s.text...)
		b.pending.width += s.width
		return
	}
	b.flush()
	b.pending = s
}

// separable reports whether s can be joined to the pending span without
// changing how either is measured.
func (b *builder) separable(s span) bool {
	if s.mode != ModeLong {
		return true
	}
	if len(b.buf) == 0 {
		return width.Separable(b.pending.text, s.text, width.Long)
	}
	_, n := utf8.DecodeLastRune(b.buf)
	return width.Separable(string(b.buf[len(b.buf)-n:]), s.text, width.Long)
}

// addLine appends the spans of l to the open line.
func (b *builder) addLine(l line) {
	for _, s := range l.spans {
		b.add(s)
	}
}

// addCut appends as many whole clusters from the start of l as fit into
// limit cells. It stops at the first cluster that does not fit, and adds
// nothing when limit is not positive.
func (b *builder) addCut(l line, limit int) {
	if limit <= 0 {
		return
	}
	for _, s := range l.spans {
		if s.width <= limit {
			b.add(s)
			limit -= s.width
			continue
		}
		end, used := 0, 0
		for {
			size, w := width.Next(s.text[end:], s.mode.width())
			if size == 0 || used+w > limit {
				break
			}
			end += size
			used += w
		}
		b.add(span{text: s.text[:end], style: s.style, mode: s.mode, width: used})
		return
	}
}

// flush moves the pending span to the finished ones.
func (b *builder) flush() {
	if b.pending.text == "" {
		return
	}
	if len(b.buf) > 0 {
		b.pending.text = string(b.buf)
		b.buf = b.buf[:0]
	}
	b.spans = append(b.spans, b.pending)
	b.pending = span{}
}

// endLine closes the open line, which may be empty, and starts a new one.
func (b *builder) endLine() {
	b.flush()
	b.lines = append(b.lines, line{spans: b.spans, width: b.width})
	b.spans, b.width = nil, 0
}

// model returns the closed lines as a model.
func (b *builder) model() Model {
	return composed(b.lines)
}

// flow appends lines to the open line: the first one continues it and the
// last one stays open.
func (b *builder) flow(lines []line, under Style) {
	for i, l := range lines {
		if i > 0 {
			b.endLine()
		}
		for _, s := range l.spans {
			s.style = s.style.Over(under)
			b.add(s)
		}
	}
}

// nested appends the content of m as a flow. inherited is the combined style
// of the enclosing models.
func (b *builder) nested(m *Model, inherited Style) {
	l := m.leaf
	if l == nil {
		b.flow(m.lines, inherited)
		return
	}
	outer := l.base.Over(inherited)
	b.mapped(l.prefix, l.prefixMap(), l, outer)
	if m.slot != nil {
		b.nested(m.slot, outer)
	}
	b.mapped(l.suffix, l.suffixMap(), l, outer)
}

// mapped appends one part of a leaf as a flow.
func (b *builder) mapped(text string, cmap []int8, l *leaf, outer Style) {
	for text != "" {
		size, runes, st := l.nextRun(text, cmap, outer)
		run := text[:size]
		for {
			piece, rest, more := strings.Cut(run, "\n")
			b.add(span{text: piece, style: st, mode: l.mode, width: width.String(piece, l.mode.width())})
			if !more {
				break
			}
			b.endLine()
			run = rest
		}
		text, cmap = text[size:], cmap[min(runes, len(cmap)):]
	}
}
