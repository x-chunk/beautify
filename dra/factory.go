package dra

import (
	"errors"
	"fmt"
	"slices"
	"unicode/utf8"
)

// Errors returned by Factory.New, wrapped with the offending values.
var (
	// ErrPaletteSize means the palette has more than 128 styles.
	ErrPaletteSize = errors.New("dra: palette too large")
	// ErrColormapLength means the colormap does not have one entry per rune.
	ErrColormapLength = errors.New("dra: colormap length does not match text")
	// ErrPaletteIndex means a colormap entry is not a palette index.
	ErrPaletteIndex = errors.New("dra: palette index out of range")
	// ErrTabWidth means the tab width is below 1.
	ErrTabWidth = errors.New("dra: invalid tab width")
	// ErrInvalidMode means the width mode is unknown.
	ErrInvalidMode = errors.New("dra: invalid mode")
)

// maxPalette is the number of styles an int8 colormap can address.
const maxPalette = 128

// Factory creates models. It holds base options that every model starts
// from. A Factory is immutable and safe for concurrent use; a nil *Factory
// behaves like NewFactory().
type Factory struct {
	cfg config
}

// NewFactory returns a factory with the given base options. Options are
// validated when a model is built, because a colormap has no meaning without
// the final text.
func NewFactory(opts ...Option) *Factory {
	cfg := defaultConfig()
	cfg.apply(opts)
	return &Factory{cfg: cfg}
}

// With returns a factory with the options of f followed by opts. f is
// unchanged.
func (f *Factory) With(opts ...Option) *Factory {
	return &Factory{cfg: f.config(opts)}
}

// New builds a model from the options of f followed by opts.
//
// The returned error wraps ErrPaletteSize, ErrColormapLength,
// ErrPaletteIndex, ErrTabWidth or ErrInvalidMode.
func (f *Factory) New(opts ...Option) (Model, error) {
	cfg := f.config(opts)
	if err := cfg.validate(); err != nil {
		return Model{}, err
	}
	return cfg.build(), nil
}

// MustNew is like New but panics on error. It is for models defined by
// constants in code; use New for anything that depends on input.
func (f *Factory) MustNew(opts ...Option) Model {
	m, err := f.New(opts...)
	if err != nil {
		panic(err)
	}
	return m
}

// Text builds a model with the text s and the base style of f. The palette
// and the colormap of f are ignored. Text cannot fail: an invalid mode or
// tab width of f falls back to the default.
func (f *Factory) Text(s string) Model {
	return f.Styled(s, Style{})
}

// Styled is like Text with st layered over the base style of f.
func (f *Factory) Styled(s string, st Style) Model {
	cfg := f.config(nil)
	cfg.prefix, cfg.suffix = s, ""
	cfg.palette, cfg.cmap = nil, nil
	cfg.base = st.Over(cfg.base)
	cfg.mode = f.Mode()
	if cfg.tabWidth < 1 {
		cfg.tabWidth = defaultTabWidth
	}
	return cfg.build()
}

// Mode returns the width mode of the models f creates. An invalid mode is
// reported as ModeReal.
func (f *Factory) Mode() Mode {
	if f != nil && f.cfg.mode == ModeLong {
		return ModeLong
	}
	return ModeReal
}

// config returns the options of f followed by opts. It accepts a nil f.
func (f *Factory) config(opts []Option) config {
	cfg := defaultConfig()
	if f != nil {
		cfg = f.cfg
	}
	cfg.apply(opts)
	return cfg
}

// validate checks the invariants that the rest of the package relies on.
func (c *config) validate() error {
	if len(c.palette) > maxPalette {
		return fmt.Errorf("%w: %d styles, the limit is %d", ErrPaletteSize, len(c.palette), maxPalette)
	}
	if c.cmap != nil {
		runes := utf8.RuneCountInString(c.prefix) + utf8.RuneCountInString(c.suffix)
		if len(c.cmap) != runes {
			return fmt.Errorf("%w: %d entries for %d runes", ErrColormapLength, len(c.cmap), runes)
		}
	}
	for i, index := range c.cmap {
		if index < NoStyle || int(index) >= len(c.palette) {
			return fmt.Errorf("%w: index %d at position %d, palette has %d styles",
				ErrPaletteIndex, index, i, len(c.palette))
		}
	}
	if c.tabWidth < 1 {
		return fmt.Errorf("%w: %d", ErrTabWidth, c.tabWidth)
	}
	if c.mode != ModeReal && c.mode != ModeLong {
		return fmt.Errorf("%w: %d", ErrInvalidMode, c.mode)
	}
	return nil
}

// build creates the model for a valid config.
func (c *config) build() Model {
	var prefixMap, suffixMap []int8
	if c.cmap != nil {
		n := utf8.RuneCountInString(c.prefix)
		prefixMap, suffixMap = c.cmap[:n], c.cmap[n:]
	}
	prefix, prefixMap := normalize(c.prefix, prefixMap, c.tabWidth, c.sanitize)
	suffix, suffixMap := normalize(c.suffix, suffixMap, c.tabWidth, c.sanitize)
	var cmap []int8
	if c.cmap != nil {
		cmap = slices.Concat(prefixMap, suffixMap)
		if c.mode == ModeLong {
			unifyClusters(prefix, cmap[:len(prefixMap)])
			unifyClusters(suffix, cmap[len(prefixMap):])
		}
	}
	return newLeaf(leaf{
		prefix:  prefix,
		suffix:  suffix,
		palette: c.palette,
		cmap:    cmap,
		base:    c.base,
		mode:    c.mode,
	})
}
