package dra

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestNewValidation(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
		want error
		msg  string // must appear in the message
	}{
		{"palette too large", []Option{WithPalette(make([]Style, 129)...)}, ErrPaletteSize, "129"},
		{"palette at the limit", []Option{WithPalette(make([]Style, 128)...)}, nil, ""},
		{"colormap too short", []Option{WithText("abc"), WithColormap([]int8{-1})}, ErrColormapLength, "1"},
		{"colormap too long", []Option{WithText("a"), WithColormap([]int8{-1, -1, -1})}, ErrColormapLength, "3"},
		{"colormap covers prefix and suffix", []Option{WithFrame("ab", "c"), WithColormap([]int8{-1, -1, -1})}, nil, ""},
		{"empty colormap with text", []Option{WithText("a"), WithColormap([]int8{})}, ErrColormapLength, "0"},
		{"empty colormap without text", []Option{WithColormap([]int8{})}, nil, ""},
		{"nil colormap", []Option{WithText("abc"), WithColormap(nil)}, nil, ""},
		{"index too large", []Option{WithText("ab"), WithPalette(Style{}), WithColormap([]int8{0, 1})}, ErrPaletteIndex, "position 1"},
		{"index below -1", []Option{WithText("ab"), WithPalette(Style{}), WithColormap([]int8{-2, 0})}, ErrPaletteIndex, "-2"},
		{"index without a palette", []Option{WithText("a"), WithColormap([]int8{0})}, ErrPaletteIndex, "0"},
		{"last palette index", []Option{WithText("a"), WithPalette(make([]Style, 128)...), WithColormap([]int8{127})}, nil, ""},
		{"zero tab width", []Option{WithTabWidth(0)}, ErrTabWidth, "0"},
		{"negative tab width", []Option{WithTabWidth(-3)}, ErrTabWidth, "-3"},
		{"invalid mode", []Option{WithMode(Mode(7))}, ErrInvalidMode, "7"},
		{"length is checked before normalization", []Option{WithText("a\tb"), WithColormap([]int8{-1, -1, -1})}, nil, ""},
		{"invalid bytes count as runes", []Option{WithText("\xff\xfe"), WithColormap([]int8{-1, -1})}, nil, ""},
		{"unused palette", []Option{WithText("a"), WithPalette(Style{FG: Red})}, nil, ""},
		{"nil option is ignored", []Option{nil, WithText("a")}, nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFactory().New(tt.opts...)
			if !errors.Is(err, tt.want) {
				t.Fatalf("New() error = %v, want %v", err, tt.want)
			}
			if err != nil && !strings.Contains(err.Error(), tt.msg) {
				t.Errorf("error %q does not mention %q", err, tt.msg)
			}
			// Base options are validated only when a model is built.
			if _, err := NewFactory(tt.opts...).New(); !errors.Is(err, tt.want) {
				t.Errorf("as base options: error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestNewText(t *testing.T) {
	f := NewFactory()
	tests := []struct {
		name   string
		opts   []Option
		text   string
		height int
	}{
		{"no options", nil, "", 0},
		{"empty text", []Option{WithText("")}, "", 0},
		{"newline", []Option{WithText("\n")}, "\n", 2},
		{"frame", []Option{WithFrame("[", "]")}, "[]", 1},
		{"text after frame clears the suffix", []Option{WithFrame("[", "]"), WithText("x")}, "x", 1},
		{"frame after text replaces both", []Option{WithText("x"), WithFrame("[", "]")}, "[]", 1},
		{"tabs", []Option{WithText("\t"), WithTabWidth(2)}, "  ", 1},
		{"only controls", []Option{WithText("\x00\x01\x1b")}, "���", 1},
		{"sanitize off", []Option{WithText("\x1b"), WithSanitize(false)}, "\x1b", 1},
		{"sanitize back on", []Option{WithText("\x1b"), WithSanitize(false), WithSanitize(true)}, "�", 1},
		{"parts are normalized separately", []Option{WithFrame("a\r", "\nb")}, "a�\nb", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := f.New(tt.opts...)
			if err != nil {
				t.Fatal(err)
			}
			if m.String() != tt.text || m.Height() != tt.height {
				t.Errorf("got %q, height %d, want %q, %d", m.String(), m.Height(), tt.text, tt.height)
			}
			if !m.HasSlot() || m.IsFilled() {
				t.Errorf("HasSlot=%v IsFilled=%v", m.HasSlot(), m.IsFilled())
			}
		})
	}
}

func TestNewMapsColormap(t *testing.T) {
	m, err := NewFactory().New(
		WithFrame("a\t", "\r\nb"),
		WithTabWidth(2),
		WithPalette(make([]Style, 5)...),
		WithColormap([]int8{0, 1, 2, 3, 4}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if want := []int8{0, 1, 1, 3, 4}; !slices.Equal(m.leaf.cmap, want) {
		t.Errorf("cmap = %v, want %v", m.leaf.cmap, want)
	}
	if m.leaf.prefixRunes != 3 {
		t.Errorf("prefixRunes = %d, want 3", m.leaf.prefixRunes)
	}
}

func TestNewUnifiesClustersInLongMode(t *testing.T) {
	opts := []Option{
		WithFrame("é", "ó"),
		WithPalette(make([]Style, 4)...),
		WithColormap([]int8{0, 1, 2, 3}),
	}
	long, err := NewFactory(WithMode(ModeLong)).New(opts...)
	if err != nil {
		t.Fatal(err)
	}
	if want := []int8{0, 0, 2, 2}; !slices.Equal(long.leaf.cmap, want) {
		t.Errorf("long cmap = %v, want %v", long.leaf.cmap, want)
	}
	real, err := NewFactory().New(opts...)
	if err != nil {
		t.Fatal(err)
	}
	if want := []int8{0, 1, 2, 3}; !slices.Equal(real.leaf.cmap, want) {
		t.Errorf("real cmap = %v, want %v", real.leaf.cmap, want)
	}
}

func TestFactoryOptionPrecedence(t *testing.T) {
	base := NewFactory(WithText("base"), WithMode(ModeLong))
	derived := base.With(WithText("derived"))
	if got := base.MustNew().String(); got != "base" {
		t.Errorf("base = %q; With changed its receiver", got)
	}
	if got := derived.MustNew().String(); got != "derived" {
		t.Errorf("derived = %q", got)
	}
	if got := derived.MustNew(WithText("call")).String(); got != "call" {
		t.Errorf("call = %q", got)
	}
	if derived.Mode() != ModeLong {
		t.Errorf("derived mode = %v", derived.Mode())
	}
	if base.MustNew(WithMode(ModeReal)).Mode() != ModeReal {
		t.Error("call option did not override the mode")
	}
}

func TestFactoryColormapWithOtherText(t *testing.T) {
	f := NewFactory(WithText("abc"), WithColormap([]int8{-1, -1, -1}))
	if _, err := f.New(WithText("ab")); !errors.Is(err, ErrColormapLength) {
		t.Errorf("New() = %v, want ErrColormapLength", err)
	}
	if got := f.Text("ab").String(); got != "ab" {
		t.Errorf("Text() = %q; the factory colormap must be ignored", got)
	}
}

func TestTextNeverFails(t *testing.T) {
	f := NewFactory(
		WithFrame("[", "]"), WithMode(Mode(9)), WithTabWidth(0),
		WithPalette(make([]Style, 300)...), WithColormap([]int8{99}), WithStyle(Style{FG: Red}),
	)
	m := f.Text("a\tb")
	if m.String() != "a    b" || m.Mode() != ModeReal || m.leaf.base != (Style{FG: Red}) {
		t.Errorf("Text() = %q, mode %v, base %+v", m.String(), m.Mode(), m.leaf.base)
	}
	if f.Mode() != ModeReal {
		t.Errorf("Mode() = %v, want ModeReal for an invalid mode", f.Mode())
	}
}

func TestStyledLayersOverBase(t *testing.T) {
	f := NewFactory(WithStyle(Style{FG: Red, BG: Blue}))
	m := f.Styled("x", Style{FG: Green, Attrs: Bold})
	if want := (Style{FG: Green, BG: Blue, Attrs: Bold}); m.leaf.base != want {
		t.Errorf("base = %+v, want %+v", m.leaf.base, want)
	}
}

func TestNilFactory(t *testing.T) {
	var f *Factory
	if m, err := f.New(WithText("a")); err != nil || m.String() != "a" {
		t.Errorf("New() = %q, %v", m.String(), err)
	}
	if got := f.With(WithMode(ModeLong)).Mode(); got != ModeLong {
		t.Errorf("With().Mode() = %v", got)
	}
	if f.MustNew(WithText("a")).String() != "a" || f.Text("a").String() != "a" ||
		f.Styled("a", Style{}).String() != "a" || f.Mode() != ModeReal {
		t.Error("nil factory does not behave like NewFactory()")
	}
}

func TestMustNewPanics(t *testing.T) {
	defer func() {
		if err, ok := recover().(error); !ok || !errors.Is(err, ErrTabWidth) {
			t.Errorf("recovered %v, want ErrTabWidth", err)
		}
	}()
	NewFactory().MustNew(WithTabWidth(0))
}

func TestOptionsCopySlices(t *testing.T) {
	palette := []Style{{FG: Red}}
	cmap := []int8{0}
	withPalette, withCmap := WithPalette(palette...), WithColormap(cmap)
	f := NewFactory(WithText("a"), withPalette, withCmap)
	before := f.MustNew()
	palette[0], cmap[0] = Style{FG: Blue}, 5
	for name, m := range map[string]Model{
		"model": before, "factory": f.MustNew(),
		"option": NewFactory().MustNew(WithText("a"), withPalette, withCmap),
	} {
		if m.leaf.palette[0] != (Style{FG: Red}) || m.leaf.cmap[0] != 0 {
			t.Errorf("%s sees the changed slice: %+v %v", name, m.leaf.palette, m.leaf.cmap)
		}
	}
}

func TestNewDoesNotChangeSharedColormap(t *testing.T) {
	f := NewFactory(WithMode(ModeLong), WithText("é"),
		WithPalette(make([]Style, 2)...), WithColormap([]int8{0, 1}))
	f.MustNew()
	if m := f.MustNew(WithMode(ModeReal)); !slices.Equal(m.leaf.cmap, []int8{0, 1}) {
		t.Errorf("cmap = %v; a Long Mode build changed the factory", m.leaf.cmap)
	}
}

func FuzzNew(f *testing.F) {
	f.Add("a\tb\r\n\x1b[31m", "‮\xff", uint8(0))
	f.Add("", "́", uint8(1))
	f.Fuzz(func(t *testing.T, prefix, suffix string, mode uint8) {
		cmap := make([]int8, utf8.RuneCountInString(prefix)+utf8.RuneCountInString(suffix))
		m, err := NewFactory().New(WithFrame(prefix, suffix), WithMode(Mode(mode%2)),
			WithPalette(Style{FG: Red}), WithColormap(cmap))
		if err != nil {
			t.Fatal(err)
		}
		text := m.String()
		if !utf8.ValidString(text) {
			t.Fatalf("%q is not valid UTF-8", text)
		}
		for _, r := range text {
			if isUnsafe(r) || r == '\t' {
				t.Fatalf("%q contains %U", text, r)
			}
		}
		if len(m.leaf.cmap) != utf8.RuneCountInString(text) {
			t.Fatalf("%d colormap entries for %q", len(m.leaf.cmap), text)
		}
		if m.Width() < 0 || m.Height() != 0 && m.Height() != strings.Count(text, "\n")+1 {
			t.Fatalf("size %dx%d for %q", m.Width(), m.Height(), text)
		}
	})
}
