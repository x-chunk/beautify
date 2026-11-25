package dra

// StringWidth returns the width in cells of the widest line of s. The text
// is normalized first the way a Factory with default settings does it, so
// the result matches the width of a model built from s.
func StringWidth(s string, mode Mode) int {
	text, _ := normalize(s, nil, defaultTabWidth, true)
	return measure(text, mode).max
}
