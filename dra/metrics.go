package dra

import (
	"strings"

	"go.xchunk.org/beautify/internal/width"
)

// metrics describes the shape of a piece of text. A height of 0 means there
// is no text at all, which differs from one empty line.
type metrics struct {
	first  int // width of the first line
	last   int // width of the last line
	max    int // width of the widest line
	height int // number of lines
}

// measure computes the metrics of normalized text.
func measure(text string, mode Mode) metrics {
	if text == "" {
		return metrics{}
	}
	var m metrics
	for {
		line, rest, more := strings.Cut(text, "\n")
		w := width.String(line, mode.width())
		if m.height == 0 {
			m.first = w
		}
		m.last = w
		m.max = max(m.max, w)
		m.height++
		if !more {
			return m
		}
		text = rest
	}
}

// concat returns the metrics of a followed by b in one flow: the first line
// of b continues the last line of a.
func (a metrics) concat(b metrics) metrics {
	if a.height == 0 {
		return b
	}
	if b.height == 0 {
		return a
	}
	joint := a.last + b.first
	m := metrics{
		first:  a.first,
		last:   b.last,
		max:    max(a.max, b.max, joint),
		height: a.height + b.height - 1,
	}
	if a.height == 1 {
		m.first = joint
	}
	if b.height == 1 {
		m.last = joint
	}
	return m
}
