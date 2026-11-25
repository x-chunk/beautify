package dra

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
