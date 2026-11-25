package dra

// Style is a set of visual properties. The zero value inherits everything.
// Styles are comparable with ==.
type Style struct {
	FG    Color // foreground; unset inherits
	BG    Color // background; unset inherits
	Attrs Attr  // added to the attributes of the parent
}

// Foreground returns a style that sets only the foreground color.
func Foreground(c Color) Style {
	return Style{FG: c}
}

// Background returns a style that sets only the background color.
func Background(c Color) Style {
	return Style{BG: c}
}

// WithFG returns a copy of s with the foreground color c.
func (s Style) WithFG(c Color) Style {
	s.FG = c
	return s
}

// WithBG returns a copy of s with the background color c.
func (s Style) WithBG(c Color) Style {
	s.BG = c
	return s
}

// WithAttrs returns a copy of s with the attributes a added.
func (s Style) WithAttrs(a Attr) Style {
	s.Attrs |= a
	return s
}

// Over returns s layered on top of parent: unset colors of s are taken from
// parent, attributes are combined. A child cannot remove an attribute.
func (s Style) Over(parent Style) Style {
	if !s.FG.IsSet() {
		s.FG = parent.FG
	}
	if !s.BG.IsSet() {
		s.BG = parent.BG
	}
	s.Attrs |= parent.Attrs
	return s
}

// IsZero reports whether s changes nothing.
func (s Style) IsZero() bool {
	return s == Style{}
}
