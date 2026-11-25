package beautify

import (
	"fmt"
	"slices"

	"go.xchunk.org/beautify/dra"
)

// Direction is the way a Template arranges its items. Unknown values behave
// like Inline.
type Direction uint8

const (
	// Inline places items side by side. It is the default.
	Inline Direction = iota
	// Stacked places items under each other.
	Stacked
)

// Template is an ordered set of Composables and the unit that a Composer
// renders. A Template is itself a Composable, so templates nest. It is an
// immutable value and can be rendered concurrently when its items can. The
// zero value has no items and renders an empty model.
type Template struct {
	items []Composable
	cfg   templateConfig
}

type templateConfig struct {
	direction    Direction
	separator    Composable
	hasSeparator bool
	skipEmpty    bool
}

// TemplateOption configures a Template.
type TemplateOption func(*templateConfig)

// WithDirection sets how the items are arranged. The default is Inline.
func WithDirection(d Direction) TemplateOption {
	return func(c *templateConfig) {
		c.direction = d
	}
}

// WithSeparator sets an element that is placed between items, never before
// the first or after the last one. It is rendered once with the context of
// the template. A nil sep makes rendering fail with ErrNilComposable.
func WithSeparator(sep Composable) TemplateOption {
	return func(c *templateConfig) {
		c.separator, c.hasSeparator = sep, true
	}
}

// WithSkipEmpty sets whether items that render to an empty model are dropped
// together with their separator. By default they keep their separators.
func WithSkipEmpty(skip bool) TemplateOption {
	return func(c *templateConfig) {
		c.skipEmpty = skip
	}
}

// NewTemplate returns a template of the given items.
func NewTemplate(items ...Composable) Template {
	return Template{items: slices.Clone(items)}
}

// With returns a copy of t with the options applied. Nil options are
// skipped.
func (t Template) With(opts ...TemplateOption) Template {
	for _, opt := range opts {
		if opt != nil {
			opt(&t.cfg)
		}
	}
	return t
}

// Append returns a copy of t with items added at the end.
func (t Template) Append(items ...Composable) Template {
	t.items = slices.Concat(t.items, items)
	return t
}

// Len returns the number of items.
func (t Template) Len() int {
	return len(t.items)
}

// Render combines the models of all items. In a bounded context, items of an
// Inline template get the width left by the items and separators before
// them, and items of a Stacked template get the whole bound. A result wider
// than the bound is cut.
//
// Rendering stops at the first error, which is returned wrapped with the
// index of its item. A nil item gives ErrNilComposable.
func (t Template) Render(ctx Context) (dra.Model, error) {
	if len(t.items) == 0 {
		return dra.Model{}, nil
	}
	var sep dra.Model
	if t.cfg.hasSeparator {
		var err error
		if sep, err = render(t.cfg.separator, ctx); err != nil {
			return dra.Model{}, fmt.Errorf("beautify: template separator: %w", err)
		}
	}

	bound, bounded := ctx.Width()
	stacked := t.cfg.direction == Stacked
	placed := make([]dra.Model, 0, 2*len(t.items))
	used := 0 // width of everything in placed, side by side
	for i, item := range t.items {
		gap := 0
		if len(placed) > 0 {
			gap = sep.Width()
		}
		itemCtx := ctx
		if bounded && !stacked {
			itemCtx = ctx.WithWidth(bound - used - gap)
		}
		m, err := render(item, itemCtx)
		if err != nil {
			return dra.Model{}, fmt.Errorf("beautify: template item %d: %w", i, err)
		}
		if t.cfg.skipEmpty && m.IsEmpty() {
			continue
		}
		if len(placed) > 0 {
			placed = append(placed, sep)
		}
		placed = append(placed, m)
		used += gap + m.Width()
	}

	var m dra.Model
	if stacked {
		m = dra.Stack(dra.AlignLeft, placed...)
	} else {
		m = dra.Join(dra.AlignTop, placed...)
	}
	if bounded {
		m = dra.Truncate(m, bound, dra.Model{})
	}
	return m, nil
}

// render renders c and reports a nil c as an error.
func render(c Composable, ctx Context) (dra.Model, error) {
	if c == nil {
		return dra.Model{}, ErrNilComposable
	}
	return c.Render(ctx)
}
