package dra

// Attr is a set of text attributes. Combine attributes with |.
type Attr uint8

// Text attributes. Support depends on the terminal. The bits match those of
// internal/ansi, so a style converts without a lookup.
const (
	Bold Attr = 1 << iota
	Faint
	Italic
	Underline
	Blink
	Reverse
	Strike
)
