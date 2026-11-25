package dra

import "slices"

// Option configures a Factory or a single model. Options are applied in
// order and later options win. Slices passed to options are copied.
type Option func(*config)

// config collects the effect of options. Its slices are shared between
// copies and never written to.
type config struct {
	prefix, suffix string
	palette        []Style
	cmap           []int8
	base           Style
	mode           Mode
	tabWidth       int
	sanitize       bool
}

func defaultConfig() config {
	return config{tabWidth: defaultTabWidth, sanitize: true}
}

// apply runs opts on c. Nil options are skipped.
func (c *config) apply(opts []Option) {
	for _, opt := range opts {
		if opt != nil {
			opt(c)
		}
	}
}

// WithText sets the text before the slot to s and the text after it to
// empty.
func WithText(s string) Option {
	return WithFrame(s, "")
}

// WithFrame sets the text before and after the slot.
func WithFrame(prefix, suffix string) Option {
	return func(c *config) {
		c.prefix, c.suffix = prefix, suffix
	}
}

// WithPalette replaces the palette that colormap entries refer to. A model
// can have at most 128 styles.
func WithPalette(styles ...Style) Option {
	styles = slices.Clone(styles)
	return func(c *config) {
		c.palette = styles
	}
}

// WithColormap replaces the colormap: one palette index per rune of the
// text before and then after the slot, or NoStyle. It is matched against
// the text as given, before tabs and line endings are normalized. An invalid
// UTF-8 byte counts as one rune. A nil colormap leaves all runes without
// their own style.
//
// A colormap set on a Factory fits only texts of its length; New fails with
// ErrColormapLength for other texts.
func WithColormap(cmap []int8) Option {
	cmap = slices.Clone(cmap)
	return func(c *config) {
		c.cmap = cmap
	}
}

// WithStyle replaces the base style, which applies to the whole model
// including the slot content.
func WithStyle(base Style) Option {
	return func(c *config) {
		c.base = base
	}
}

// WithMode sets the width mode. The default is ModeReal.
func WithMode(mode Mode) Option {
	return func(c *config) {
		c.mode = mode
	}
}

// WithTabWidth sets how many spaces replace a tab. The default is 4. Tabs
// are not aligned to tab stops because the final column of a tab is unknown
// until layout.
func WithTabWidth(n int) Option {
	return func(c *config) {
		c.tabWidth = n
	}
}

// WithSanitize enables or disables the replacement of control characters,
// bidirectional controls and Unicode line separators with U+FFFD. It is on
// by default.
//
// With sanitization off the caller takes responsibility for the text: raw
// control characters reach the terminal, are measured like any other rune
// and can desynchronize the style tracking of the Engine.
func WithSanitize(on bool) Option {
	return func(c *config) {
		c.sanitize = on
	}
}
