package dra

import (
	"io"
	"strings"

	"go.xchunk.org/beautify/internal/ansi"
)

// Engine renders models for one color profile. It is a pure function of the
// model and the profile: it keeps no state between calls, never detects the
// terminal and never reads the environment. An Engine is safe for concurrent
// use. The zero value renders plain text.
type Engine struct {
	profile Profile
}

// NewEngine returns an engine for the profile p.
func NewEngine(p Profile) Engine {
	return Engine{profile: p}
}

// Profile returns the profile the engine renders for.
func (e Engine) Profile() Profile {
	return e.profile
}

// Append appends the rendered model to dst and returns the extended slice.
// It emits exactly the model and adds no trailing newline. Styles are reset
// before every newline and at the end, so nothing leaks into the output
// that follows. Append allocates only to grow dst.
func (e Engine) Append(dst []byte, m Model) []byte {
	w := encoder{dst: dst, level: e.profile.level()}
	w.model(&m, Style{})
	w.reset()
	return w.dst
}

// Render writes the rendered model to w with a single Write call, or with
// none when the model is empty. It returns the number of bytes written and
// the error of the writer, or io.ErrShortWrite when the writer accepted only
// a part without reporting an error.
func (e Engine) Render(w io.Writer, m Model) (int, error) {
	if m.IsEmpty() {
		return 0, nil
	}
	buf := e.Append(nil, m)
	n, err := w.Write(buf)
	if err == nil && n < len(buf) {
		err = io.ErrShortWrite
	}
	return n, err
}

// String returns the rendered model as a string.
func (e Engine) String(m Model) string {
	return string(e.Append(nil, m))
}

// encoder tracks the terminal state of one Append call.
type encoder struct {
	dst     []byte
	level   ansi.Level
	current ansi.Style // downgraded style the terminal is in
}

// model emits m with the style inherited from the enclosing models.
func (w *encoder) model(m *Model, inherited Style) {
	l := m.leaf
	if l == nil {
		w.lines(m.lines, inherited)
		return
	}
	outer := l.base.Over(inherited)
	w.mapped(l.prefix, l.prefixMap(), l, outer)
	if m.slot != nil {
		w.model(m.slot, outer)
	}
	w.mapped(l.suffix, l.suffixMap(), l, outer)
}

// lines emits the content of a composed model.
func (w *encoder) lines(lines []line, inherited Style) {
	for i, l := range lines {
		if i > 0 {
			w.newline()
		}
		for _, s := range l.spans {
			w.line(s.text, w.downgrade(s.style.Over(inherited)))
		}
	}
}

// mapped emits one part of a leaf as runs of equal style.
func (w *encoder) mapped(text string, cmap []int8, l *leaf, outer Style) {
	for text != "" {
		size, runes, st := l.nextRun(text, cmap, outer)
		w.text(text[:size], st)
		text, cmap = text[size:], cmap[min(runes, len(cmap)):]
	}
}

// downgrade converts s to what the profile can show.
func (w *encoder) downgrade(s Style) ansi.Style {
	return ansi.Downgrade(ansi.Style{FG: s.FG.c, BG: s.BG.c, Attrs: uint8(s.Attrs)}, w.level)
}

// text emits a run that has one style.
func (w *encoder) text(text string, s Style) {
	d := w.downgrade(s)
	for {
		i := strings.IndexByte(text, '\n')
		piece := text
		if i >= 0 {
			piece = text[:i]
		}
		w.line(piece, d)
		if i < 0 {
			return
		}
		w.newline()
		text = text[i+1:]
	}
}

// line emits text without newlines. Empty text causes no style switch.
func (w *encoder) line(text string, d ansi.Style) {
	if text == "" {
		return
	}
	if d != w.current {
		w.dst = ansi.AppendSGR(w.dst, d)
		w.current = d
	}
	w.dst = append(w.dst, text...)
}

// newline ends the line. The style is reset first so a background color does
// not extend to the right edge of the terminal.
func (w *encoder) newline() {
	w.reset()
	w.dst = append(w.dst, '\n')
}

// reset returns the terminal to the empty style if it is not in it.
func (w *encoder) reset() {
	if w.current != (ansi.Style{}) {
		w.dst = ansi.AppendReset(w.dst)
		w.current = ansi.Style{}
	}
}
