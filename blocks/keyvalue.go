package blocks

import (
	"slices"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

// Pair is an entry of KeyValue. A nil Value is shown as an empty value.
type Pair struct {
	Key   string
	Value beautify.Composable
}

// KeyValueOption configures KeyValue.
type KeyValueOption func(*keyValue)

// KeyValueSeparator sets the text between keys and values. The default is
// ": ".
func KeyValueSeparator(s string) KeyValueOption {
	return func(k *keyValue) { k.separator = s }
}

// KeyStyle sets the style of the keys.
func KeyStyle(st dra.Style) KeyValueOption {
	return func(k *keyValue) { k.style = st }
}

// KeyAlign sets how keys are aligned with the widest one. The default is the
// left.
func KeyAlign(a dra.Align) KeyValueOption {
	return func(k *keyValue) { k.align = a }
}

type keyValue struct {
	pairs     []Pair
	separator string
	style     dra.Style
	align     dra.Align
}

// KeyValue shows pairs with the values lined up in a column. When the bound
// leaves no room for that column, every value moves to the line below its
// key. Line breaks in keys and in the separator become spaces.
func KeyValue(pairs []Pair, opts ...KeyValueOption) beautify.Composable {
	k := keyValue{pairs: slices.Clone(pairs), separator: ": "}
	applyOptions(&k, opts)
	return k
}

// Render implements beautify.Composable.
func (k keyValue) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) {
		return dra.Model{}, nil
	}
	f := ctx.Factory()
	keys := make([]dra.Model, len(k.pairs))
	width := 0
	for i, p := range k.pairs {
		keys[i] = f.Styled(oneLine(p.Key), k.style)
		width = max(width, keys[i].Width())
	}
	separator := f.Text(oneLine(k.separator))

	const indent = 2 // of values below their key
	column := width + separator.Width()
	valueCtx, below := shrink(ctx, column), false
	if w, bounded := valueCtx.Width(); bounded && w < 1 {
		valueCtx, below = shrink(ctx, indent), true
	}

	rows := make([]dra.Model, 0, len(k.pairs))
	for i, p := range k.pairs {
		key := keys[i]
		if key.IsEmpty() {
			// An empty model has no line to pad.
			key = dra.Blank(0, 1)
		}
		key = dra.Concat(dra.Pad(key, width, k.align), separator)
		var value dra.Model
		if p.Value != nil {
			var err error
			if value, err = renderChild(valueCtx, "keyvalue", i, p.Value); err != nil {
				return dra.Model{}, err
			}
		}
		if below {
			blank := dra.Blank(indent, 1)
			rows = append(rows, key, dra.PrefixLines(value, blank, blank))
		} else {
			rows = append(rows, dra.Join(dra.AlignTop, key, value))
		}
	}
	return clip(ctx, dra.Stack(dra.AlignLeft, rows...)), nil
}
