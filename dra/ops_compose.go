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
