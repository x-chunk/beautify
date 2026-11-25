package dra

// Truncate cuts every line that is wider than width and puts the first line
// of tail at the cut. A tail wider than width is left out. Lines at or below
// width are untouched, and the height never changes. Pass the zero Model as
// tail for a hard cut.
//
// Cuts happen only between clusters, so a cut line can be one cell narrower
// than width; it is not padded. With a width of 0 or less every line becomes
// empty.
func Truncate(m Model, width int, tail Model) Model {
	width = max(width, 0)
	if m.Width() <= width {
		return composed(flatten(&m))
	}
	end := firstLine(tail)
	if width == 0 || end.width > width {
		end = line{}
	}
	var b builder
	for _, l := range flatten(&m) {
		if l.width <= width {
			b.addLine(l)
		} else {
			b.addCut(l, width-end.width)
			b.addLine(end)
		}
		b.endLine()
	}
	return b.model()
}

// TruncateHeight keeps the first height lines. Zero and negative heights
// give an empty model.
func TruncateHeight(m Model, height int) Model {
	lines := flatten(&m)
	return composed(lines[:min(max(height, 0), len(lines))])
}
