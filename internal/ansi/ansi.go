// Package ansi encodes SGR escape sequences and converts colors to what a
// terminal can show. It has its own primitive types so that it does not
// depend on package dra.
package ansi

// Level is the color capability of a terminal.
type Level uint8

// Levels in order of increasing capability. The zero value is Plain.
const (
	Plain     Level = iota // no escape sequences
	NoColor                // attributes only
	ANSI16                 // 16-color palette
	ANSI256                // 256-color palette
	TrueColor              // 24-bit colors
)

// ColorKind tells which fields of a Color are in use.
type ColorKind uint8

// Kinds of color. None and Default both encode to no parameter; they differ
// only for style inheritance, which is resolved before encoding.
const (
	None       ColorKind = iota // no color information
	Default                     // the terminal's default color
	Indexed16                   // Index in 0-15
	Indexed256                  // Index in 0-255
	RGB                         // R, G and B
)

// Color is a terminal color. Fields that the kind does not use are zero, so
// colors can be compared with ==.
type Color struct {
	Kind    ColorKind
	Index   uint8
	R, G, B uint8
}

// Attribute bits of Style.Attrs.
const (
	Bold uint8 = 1 << iota
	Faint
	Italic
	Underline
	Blink
	Reverse
	Strike
)

// attrCodes holds the SGR code of each attribute bit, lowest bit first.
var attrCodes = [...]uint8{1, 2, 3, 4, 5, 7, 9}

// Style is a set of visual properties.
type Style struct {
	FG, BG Color
	Attrs  uint8 // combination of attribute bits
}

// AppendSGR appends to dst the sequence that switches the terminal to s
// from any previous state and returns the extended slice. s must already be
// downgraded.
//
// Every switch starts with a full reset. SGR cannot turn off bold and faint
// independently, so incremental switches would depend on the previous state.
func AppendSGR(dst []byte, s Style) []byte {
	dst = append(dst, "\x1b[0"...)
	for i, code := range attrCodes {
		if s.Attrs&(1<<i) != 0 {
			dst = append(dst, ';', '0'+code)
		}
	}
	dst = appendColor(dst, s.FG, 30)
	dst = appendColor(dst, s.BG, 40)
	return append(dst, 'm')
}

// AppendReset appends ESC [ 0 m to dst and returns the extended slice.
func AppendReset(dst []byte) []byte {
	return append(dst, "\x1b[0m"...)
}

// appendColor appends the parameters for c. base is 30 for the foreground
// and 40 for the background; the other codes are at fixed offsets from it.
func appendColor(dst []byte, c Color, base uint8) []byte {
	switch c.Kind {
	case Indexed16:
		code := base + c.Index
		if c.Index >= 8 {
			code = base + 60 + c.Index - 8
		}
		dst = append(dst, ';')
		dst = appendUint8(dst, code)
	case Indexed256:
		dst = append(dst, ';')
		dst = appendUint8(dst, base+8)
		dst = append(dst, ";5;"...)
		dst = appendUint8(dst, c.Index)
	case RGB:
		dst = append(dst, ';')
		dst = appendUint8(dst, base+8)
		dst = append(dst, ";2;"...)
		dst = appendUint8(dst, c.R)
		dst = append(dst, ';')
		dst = appendUint8(dst, c.G)
		dst = append(dst, ';')
		dst = appendUint8(dst, c.B)
	}
	return dst
}

// appendUint8 appends n in decimal without leading zeros.
func appendUint8(dst []byte, n uint8) []byte {
	if n >= 100 {
		dst = append(dst, '0'+n/100)
	}
	if n >= 10 {
		dst = append(dst, '0'+n/10%10)
	}
	return append(dst, '0'+n%10)
}
