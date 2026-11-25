package dra

// Pad extends every line to width cells with unstyled spaces, placing the
// text according to align. Lines at or above width are unchanged: Pad never
// truncates. An empty model has no lines to pad and stays empty.
func Pad(m Model, width int, align Align) Model {
	src := flatten(&m)
	lines := make([]line, len(src))
	for i, l := range src {
		gap := max(width-l.width, 0)
		left := align.lead(gap)
		lines[i] = padLine(l, left, gap-left)
	}
	return composed(lines)
}

// PadHeight adds empty lines so that the model has at least height lines,
// placing the text according to valign. It never removes lines. On an empty
// model with a positive height the result is Blank(0, height).
func PadHeight(m Model, height int, valign VAlign) Model {
	src := flatten(&m)
	gap := max(height-len(src), 0)
	lines := make([]line, len(src)+gap)
	copy(lines[valign.lead(gap):], src)
	return composed(lines)
}

// Repeat returns a single line that repeats the first line of unit to fill
// exactly width cells. A remainder is filled with a cut copy of the unit,
// and cells that still remain because a wide cluster does not fit are filled
// with spaces. A unit of width 0 gives Blank(width, 1).
func Repeat(unit Model, width int) Model {
	l := firstLine(unit)
	if l.width == 0 || width <= 0 {
		return Blank(width, 1)
	}
	var b builder
	for range width / l.width {
		b.addLine(l)
	}
	b.addCut(l, width-b.width)
	b.add(spaces(width - b.width))
	b.endLine()
	return b.model()
}
