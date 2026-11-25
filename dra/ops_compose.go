package dra

// Lines returns one single-line model per line of m. An empty model has no
// lines and gives nil. A line without text becomes a model of height 1 and
// width 0.
func Lines(m Model) []Model {
	lines := flatten(&m)
	if len(lines) == 0 {
		return nil
	}
	ms := make([]Model, len(lines))
	for i := range lines {
		ms[i] = composed(lines[i : i+1 : i+1])
	}
	return ms
}

// Blank returns height lines of width unstyled spaces. Blank(0, 1) is one
// empty line, which text alone cannot express. Negative sizes count as 0; a
// height of 0 gives an empty model.
func Blank(width, height int) Model {
	if height <= 0 {
		return Model{}
	}
	lines := make([]line, height)
	l := blankLine(width)
	for i := range lines {
		lines[i] = l
	}
	return composed(lines)
}

// Concat joins models in one flow: the first line of each model continues
// the last line of the one before. Empty models are skipped.
func Concat(ms ...Model) Model {
	var b builder
	started := false
	for i := range ms {
		if ms[i].IsEmpty() {
			continue
		}
		b.nested(&ms[i], Style{})
		started = true
	}
	if started {
		b.endLine()
	}
	return b.model()
}

// Stack places models under each other. Empty models add nothing. Lines are
// padded on the left to align them with the widest line of the result and
// are never padded on the right; use Pad for that.
func Stack(align Align, ms ...Model) Model {
	widest, height := 0, 0
	for _, m := range ms {
		widest = max(widest, m.Width())
		height += m.Height()
	}
	lines := make([]line, 0, height)
	for i := range ms {
		for _, l := range flatten(&ms[i]) {
			lines = append(lines, padLine(l, align.lead(widest-l.width), 0))
		}
	}
	return composed(lines)
}

// Join places models side by side. Empty models are skipped. Shorter models
// are extended with empty lines according to valign. The lines of every
// model except the last one are padded on the right to the width of that
// model, so the next model starts in a straight column.
func Join(valign VAlign, ms ...Model) Model {
	height := 0
	columns := make([]Model, 0, len(ms))
	for _, m := range ms {
		if !m.IsEmpty() {
			columns = append(columns, m)
			height = max(height, m.Height())
		}
	}
	flat := make([][]line, len(columns))
	for i := range columns {
		flat[i] = flatten(&columns[i])
	}
	var b builder
	for y := range height {
		for i, lines := range flat {
			var l line
			if row := y - valign.lead(height-len(lines)); row >= 0 && row < len(lines) {
				l = lines[row]
			}
			b.addLine(l)
			if i < len(flat)-1 {
				b.add(spaces(columns[i].Width() - l.width))
			}
		}
		b.endLine()
	}
	return b.model()
}

// padLine returns l with unstyled spaces around it.
func padLine(l line, left, right int) line {
	if left <= 0 && right <= 0 {
		return l
	}
	var b builder
	b.add(spaces(left))
	b.addLine(l)
	b.add(spaces(right))
	b.endLine()
	return b.lines[0]
}
