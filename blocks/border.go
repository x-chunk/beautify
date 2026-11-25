package blocks

import (
	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

// BorderSet holds the characters a frame is drawn with. Every field is
// expected to be one cell wide: an empty field is replaced by a space and a
// wider one is cut to one cell.
type BorderSet struct {
	TopLeft, Top, TopRight          string
	Left, Right                     string
	BottomLeft, Bottom, BottomRight string

	// LeftJoin and RightJoin are the cells where a horizontal divider meets
	// the left and the right edge. Border does not draw them; Sections and
	// custom blocks use them.
	LeftJoin, RightJoin string
}

// Provided border sets.
var (
	// BorderRounded has light lines and round corners. It is the default.
	BorderRounded = BorderSet{
		TopLeft: "╭", Top: "─", TopRight: "╮",
		Left: "│", Right: "│",
		BottomLeft: "╰", Bottom: "─", BottomRight: "╯",
		LeftJoin: "├", RightJoin: "┤",
	}
	// BorderSquare has light lines and square corners.
	BorderSquare = BorderSet{
		TopLeft: "┌", Top: "─", TopRight: "┐",
		Left: "│", Right: "│",
		BottomLeft: "└", Bottom: "─", BottomRight: "┘",
		LeftJoin: "├", RightJoin: "┤",
	}
	// BorderDouble has double lines.
	BorderDouble = BorderSet{
		TopLeft: "╔", Top: "═", TopRight: "╗",
		Left: "║", Right: "║",
		BottomLeft: "╚", Bottom: "═", BottomRight: "╝",
		LeftJoin: "╠", RightJoin: "╣",
	}
	// BorderHeavy has thick lines.
	BorderHeavy = BorderSet{
		TopLeft: "┏", Top: "━", TopRight: "┓",
		Left: "┃", Right: "┃",
		BottomLeft: "┗", Bottom: "━", BottomRight: "┛",
		LeftJoin: "┣", RightJoin: "┫",
	}
	// BorderASCII uses only ASCII characters.
	BorderASCII = BorderSet{
		TopLeft: "+", Top: "-", TopRight: "+",
		Left: "|", Right: "|",
		BottomLeft: "+", Bottom: "-", BottomRight: "+",
		LeftJoin: "+", RightJoin: "+",
	}
)

// BorderOption configures Border.
type BorderOption func(*border)

// BorderStyle sets the characters of the frame. The default is
// BorderRounded.
func BorderStyle(set BorderSet) BorderOption {
	return func(b *border) { b.set = set }
}

// BorderColor sets the style of the frame.
func BorderColor(st dra.Style) BorderOption {
	return func(b *border) { b.style = st }
}

// BorderGradient colors the frame by position within the box: it replaces
// the foreground of BorderColor. Labels follow it unless they have a
// foreground of their own. A gradient without stops has no effect.
func BorderGradient(g dra.Gradient) BorderOption {
	return func(b *border) { b.gradient = g }
}

// BorderTitle sets a label in the top edge. It is truncated when it is wider
// than the edge and never widens the box.
func BorderTitle(s string) BorderOption {
	return func(b *border) { b.title = s }
}

// BorderTitleAlign sets where in the top edge the title sits. The default is
// the left end.
func BorderTitleAlign(a dra.Align) BorderOption {
	return func(b *border) { b.titleAlign = a }
}

// BorderTitleStyle sets the style of the title text, layered over
// BorderColor.
func BorderTitleStyle(st dra.Style) BorderOption {
	return func(b *border) { b.titleStyle = st }
}

// BorderCaption sets a second label at the right end of the top edge. It is
// omitted when it does not fit whole next to the title.
func BorderCaption(s string) BorderOption {
	return func(b *border) { b.caption = s }
}

// BorderCaptionStyle sets the style of the caption text, layered over
// BorderColor.
func BorderCaptionStyle(st dra.Style) BorderOption {
	return func(b *border) { b.captionStyle = st }
}

// BorderFooter sets a label at the right end of the bottom edge. Its own
// styles win over BorderColor. Only its first line is used, and it is
// omitted when it does not fit whole. An empty model means no footer.
func BorderFooter(m dra.Model) BorderOption {
	return func(b *border) { b.footer = m }
}

// BorderPadding sets the blank space between the frame and the child. The
// default is 1 cell on the left and the right and none above and below.
// Negative values count as 0.
func BorderPadding(x, y int) BorderOption {
	return func(b *border) { b.padX, b.padY = max(x, 0), max(y, 0) }
}

type border struct {
	frame
	child beautify.Composable
	title string
}

// Border draws a frame around child. The box is as wide as the child needs.
// An empty child gives a box of two lines.
func Border(child beautify.Composable, opts ...BorderOption) beautify.Composable {
	b := border{frame: defaultFrame(), child: child}
	applyOptions(&b, opts)
	return b
}

// Render implements beautify.Composable.
func (b border) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) {
		return dra.Model{}, nil
	}
	b.padX = limit(ctx, b.padX)
	body, err := renderChild(b.inner(ctx), "border", 0, b.child)
	if err != nil {
		return dra.Model{}, err
	}
	return b.render(ctx, []framePart{{title: b.title, body: body}}), nil
}
