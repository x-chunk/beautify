package blocks

import (
	"slices"
	"strconv"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

// ListOption configures List.
type ListOption func(*list)

// ListBullet sets the marker of unordered items. The default is "•".
func ListBullet(s string) ListOption {
	return func(l *list) { l.bullet = s }
}

// ListOrdered numbers the items instead of marking them with a bullet.
func ListOrdered() ListOption {
	return func(l *list) { l.ordered = true }
}

// ListStart sets the number of the first item of an ordered list. The
// default is 1.
func ListStart(n int) ListOption {
	return func(l *list) { l.start = n }
}

// ListMarkerStyle sets the style of bullets and numbers.
func ListMarkerStyle(st dra.Style) ListOption {
	return func(l *list) { l.style = st }
}

// ListGap sets the number of empty lines between items. The default is 0.
func ListGap(n int) ListOption {
	return func(l *list) { l.gap = max(n, 0) }
}

type list struct {
	items   []beautify.Composable
	bullet  string
	ordered bool
	start   int
	style   dra.Style
	gap     int
}

// List shows items with a marker before each. Multi-line items hang under
// their first line. Numbers are aligned on the right.
func List(items []beautify.Composable, opts ...ListOption) beautify.Composable {
	l := list{items: slices.Clone(items), bullet: "•", start: 1}
	applyOptions(&l, opts)
	return l
}

// Render implements beautify.Composable.
func (l list) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) {
		return dra.Model{}, nil
	}
	markers := make([]dra.Model, len(l.items))
	width := 0
	for i := range l.items {
		markers[i] = ctx.Factory().Styled(l.marker(i), l.style)
		width = max(width, markers[i].Width())
	}
	itemCtx := shrink(ctx, width+1)
	rows := make([]dra.Model, 0, 2*len(l.items))
	for i, item := range l.items {
		m, err := renderChild(itemCtx, "list", i, item)
		if err != nil {
			return dra.Model{}, err
		}
		if i > 0 {
			rows = append(rows, dra.Blank(0, l.gap))
		}
		marker := dra.Pad(markers[i], width, dra.AlignRight)
		rows = append(rows, dra.Join(dra.AlignTop, marker, dra.Blank(1, 1), m))
	}
	return clip(ctx, dra.Stack(dra.AlignLeft, rows...)), nil
}

// marker returns the text before item i.
func (l list) marker(i int) string {
	if l.ordered {
		return strconv.Itoa(l.start+i) + "."
	}
	return oneLine(l.bullet)
}
