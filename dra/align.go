package dra

// Align is a horizontal alignment. Unknown values behave like AlignLeft.
type Align uint8

// Horizontal alignments. Centering with an odd remainder puts the extra cell
// on the right.
const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

// VAlign is a vertical alignment. Unknown values behave like AlignTop.
type VAlign uint8

// Vertical alignments. Centering with an odd remainder puts the extra line
// at the bottom.
const (
	AlignTop VAlign = iota
	AlignMiddle
	AlignBottom
)

// lead returns how much of gap goes before the content.
func (a Align) lead(gap int) int {
	switch a {
	case AlignCenter:
		return gap / 2
	case AlignRight:
		return gap
	default:
		return 0
	}
}

// lead returns how much of gap goes before the content.
func (a VAlign) lead(gap int) int {
	switch a {
	case AlignMiddle:
		return gap / 2
	case AlignBottom:
		return gap
	default:
		return 0
	}
}
