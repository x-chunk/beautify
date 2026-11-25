package blocks

import (
	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

// frame draws the box of Border and Sections. The blocks differ only in how
// many parts they put into it.
type frame struct {
	set          BorderSet
	style        dra.Style
	gradient     dra.Gradient
	titleAlign   dra.Align
	titleStyle   dra.Style
	caption      string
	captionStyle dra.Style
	footer       dra.Model
	padX, padY   int
}

// framePart is one body inside a frame and the title of the edge above it.
type framePart struct {
	title string
	body  dra.Model
}

func defaultFrame() frame {
	return frame{set: BorderRounded, padX: 1}
}

// inner returns the context for the bodies of the frame.
func (f *frame) inner(ctx beautify.Context) beautify.Context {
	return shrink(ctx, 2+2*f.padX)
}

// render draws the frame around parts, which must not be empty, and cuts the
// result to the bound of ctx.
func (f *frame) render(ctx beautify.Context, parts []framePart) dra.Model {
	content, height := 0, 1+len(parts)
	for _, p := range parts {
		content = max(content, p.body.Width())
		height += f.interiorHeight(p.body)
	}
	p := painter{
		frame:   f,
		factory: ctx.Factory(),
		inner:   content + 2*f.padX,
		height:  height,
	}
	rows := make([]dra.Model, 0, 2*len(parts)+1)
	y := 0
	for i, part := range parts {
		left, right, caption := f.set.LeftJoin, f.set.RightJoin, ""
		if i == 0 {
			left, right, caption = f.set.TopLeft, f.set.TopRight, f.caption
		}
		rows = append(rows, p.edge(y, left, f.set.Top, right, p.topLabels(part.title, caption)))
		y++
		rows = append(rows, p.interior(y, dra.Pad(part.body, content, dra.AlignLeft)))
		y += f.interiorHeight(part.body)
	}
	rows = append(rows, p.edge(y, f.set.BottomLeft, f.set.Bottom, f.set.BottomRight, p.bottomLabels()))
	return clip(ctx, dra.Stack(dra.AlignLeft, rows...))
}

// interiorHeight returns the number of lines between the edges around body.
func (f *frame) interiorHeight(body dra.Model) int {
	if body.IsEmpty() {
		return 0
	}
	return body.Height() + 2*f.padY
}

// painter draws the pieces of one frame whose size is known.
type painter struct {
	*frame
	factory *dra.Factory
	inner   int // width between the left and the right edge
	height  int // height of the whole box
}

// label is text inside an edge, with one space on each side.
type label struct {
	at   int // column of the left space, counted from the corner
	text dra.Model
	// paint colors the text with the gradient, given the column of its first
	// cell in the box and its line.
	paint func(m dra.Model, x, y int) dra.Model
}

func (l label) end() int {
	return l.at + l.text.Width() + 2
}

// topLabels lays out the title and the caption of an edge.
func (p *painter) topLabels(title, caption string) []label {
	var labels []label
	free := 0 // first column the caption may start in
	if title != "" && p.inner >= 3 {
		st := p.titleStyle.Over(p.style)
		text := dra.Truncate(p.factory.Styled(oneLine(title), st), p.inner-2, p.factory.Styled(ellipsis, st))
		if text.Width() > 0 {
			l := label{text: text, paint: p.painterOf(p.titleStyle)}
			l.at = p.place(text, p.titleAlign)
			labels = append(labels, l)
			free = l.end() + 1
		}
	}
	if caption != "" {
		text := p.factory.Styled(oneLine(caption), p.captionStyle.Over(p.style))
		at := p.place(text, dra.AlignRight)
		if text.Width() > 0 && at >= free {
			labels = append(labels, label{at: at, text: text, paint: p.painterOf(p.captionStyle)})
		}
	}
	return labels
}

// bottomLabels lays out the footer.
func (p *painter) bottomLabels() []label {
	lines := dra.Lines(p.footer)
	if len(lines) == 0 || lines[0].Width() == 0 {
		return nil
	}
	at := p.place(lines[0], dra.AlignRight)
	if at < 1 {
		return nil
	}
	return []label{{at: at, text: lines[0], paint: p.paintUnder}}
}

// place returns the column for a label in an edge. At the ends it leaves one
// edge cell next to the corner if there is room. The result is negative when
// the label does not fit.
func (p *painter) place(text dra.Model, align dra.Align) int {
	free := p.inner - text.Width() - 2
	if free < 0 {
		return free
	}
	switch align {
	case dra.AlignCenter:
		return free / 2
	case dra.AlignRight:
		return free - min(free, 1)
	default:
		return min(free, 1)
	}
}

// edge draws a horizontal edge or divider. labels must be ordered and must
// not overlap.
func (p *painter) edge(y int, left, fill, right string, labels []label) dra.Model {
	x := 0
	var pieces []dra.Model
	add := func(m dra.Model, paint func(dra.Model, int, int) dra.Model) {
		pieces = append(pieces, paint(m, x, y))
		x += m.Width()
	}
	fillCell, space := p.cell(fill), p.cell("")
	add(p.cell(left), p.paintOver)
	for _, l := range labels {
		add(dra.Repeat(fillCell, 1+l.at-x), p.paintOver)
		add(space, p.paintOver)
		add(l.text, l.paint)
		add(space, p.paintOver)
	}
	add(dra.Repeat(fillCell, 1+p.inner-x), p.paintOver)
	add(p.cell(right), p.paintOver)
	return dra.Concat(pieces...)
}

// interior draws the left and the right edge around body, whose lines must
// all have the content width.
func (p *painter) interior(y int, body dra.Model) dra.Model {
	if body.IsEmpty() {
		return dra.Model{}
	}
	blank := dra.Blank(p.inner, p.padY)
	body = dra.Stack(dra.AlignLeft, blank, pad(body, p.padX, p.padX), blank)
	column := func(s string, x int) dra.Model {
		c := p.cell(s)
		return p.paintOver(dra.PrefixLines(dra.Blank(0, body.Height()), c, c), x, y)
	}
	return dra.Join(dra.AlignTop, column(p.set.Left, 0), body, column(p.set.Right, 1+p.inner))
}

// cell returns s as exactly one cell in the style of the frame.
func (p *painter) cell(s string) dra.Model {
	lines := dra.Lines(dra.Truncate(p.factory.Styled(s, p.style), 1, dra.Model{}))
	if len(lines) == 0 || lines[0].Width() == 0 {
		return p.factory.Styled(" ", p.style)
	}
	return lines[0]
}

// painterOf returns how to color a label with style st: a foreground of its
// own wins over the gradient.
func (p *painter) painterOf(st dra.Style) func(dra.Model, int, int) dra.Model {
	if st.FG.IsSet() {
		return func(m dra.Model, _, _ int) dra.Model { return m }
	}
	return p.paintOver
}

// paintOver replaces the foreground of m, whose first cell is at x and y,
// with the color of the gradient.
func (p *painter) paintOver(m dra.Model, x, y int) dra.Model {
	if p.gradient.IsZero() {
		return m
	}
	return dra.RestyleFunc(m, func(dx, dy int) dra.Style {
		return dra.Foreground(p.colorAt(x+dx, y+dy))
	})
}

// paintUnder colors the cells of the single-line model m that have no
// foreground and layers the style of the frame under it.
func (p *painter) paintUnder(m dra.Model, x, y int) dra.Model {
	if !p.gradient.IsZero() {
		// RestyleFunc would override the model's own colors and cannot see
		// them, so every cluster is restyled on its own. Wrapping at width 1
		// is what takes a line apart into clusters.
		cells := dra.Lines(dra.Wrap(m, 1, dra.WrapAnywhere))
		for i, c := range cells {
			cells[i] = dra.Restyle(c, dra.Foreground(p.colorAt(x, y)))
			x += c.Width()
		}
		m = dra.Concat(cells...)
	}
	return dra.Restyle(m, p.style)
}

func (p *painter) colorAt(x, y int) dra.Color {
	return p.gradient.At(x, y, p.inner+2, p.height)
}
