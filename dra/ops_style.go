package dra

import "go.xchunk.org/beautify/internal/width"

// Restyle layers st under the model: the model's own styles win, st fills
// the colors that are unset and adds its attributes. Padding created by
// layout operations is unstyled, so a background in st colors it too.
func Restyle(m Model, st Style) Model {
	var b builder
	for _, l := range flatten(&m) {
		for _, s := range l.spans {
			s.style = s.style.Over(st)
			b.add(s)
		}
		b.endLine()
	}
	return b.model()
}

// RestyleFunc layers a style chosen by position over the model. fn receives
// the column and the line of a cell, both from 0. Unlike with Restyle, the
// colors that fn returns win over the model's own; a zero Style leaves the
// cell unchanged. Text and size never change.
//
// fn is called once per cluster with the column of its first cell. It must
// be a pure function: the order and number of calls can change. A nil fn
// changes nothing.
func RestyleFunc(m Model, fn func(x, y int) Style) Model {
	if fn == nil {
		return composed(flatten(&m))
	}
	var b builder
	for y, l := range flatten(&m) {
		x := 0
		for _, s := range l.spans {
			x = restyleSpan(&b, s, x, y, fn)
		}
		b.endLine()
	}
	return b.model()
}

// restyleSpan appends s to b with the styles of fn and returns the column
// after it.
func restyleSpan(b *builder, s span, x, y int, fn func(x, y int) Style) int {
	run := span{mode: s.mode} // clusters of s with equal styles
	start, end := 0, 0        // byte range of run in s.text
	for end < len(s.text) {
		size, cells := width.Next(s.text[end:], s.mode.width())
		if st := fn(x, y).Over(s.style); end == start || st != run.style {
			run.text = s.text[start:end]
			b.add(run)
			run.style, run.width, start = st, 0, end
		}
		end += size
		x += cells
		run.width += cells
	}
	run.text = s.text[start:]
	b.add(run)
	return x
}
