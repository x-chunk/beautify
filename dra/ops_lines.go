package dra

// PrefixLines puts first before the first line of m and rest before every
// other line. Only the first line of each prefix is used. An empty model
// stays empty and gets no prefix.
func PrefixLines(m Model, first, rest Model) Model {
	return decorateLines(m, first, rest, true)
}

// SuffixLines puts first after the first line of m and rest after every
// other line. The suffix is attached to the end of each line; pad the model
// first when the suffixes must align. Only the first line of each suffix is
// used. An empty model stays empty and gets no suffix.
func SuffixLines(m Model, first, rest Model) Model {
	return decorateLines(m, first, rest, false)
}

func decorateLines(m, first, rest Model, before bool) Model {
	firstDecor, restDecor := firstLine(first), firstLine(rest)
	var b builder
	for i, l := range flatten(&m) {
		decor := restDecor
		if i == 0 {
			decor = firstDecor
		}
		if before {
			b.addLine(decor)
			b.addLine(l)
		} else {
			b.addLine(l)
			b.addLine(decor)
		}
		b.endLine()
	}
	return b.model()
}
