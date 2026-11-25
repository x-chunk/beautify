package beautify

import (
	"errors"
	"slices"
	"strings"
	"sync"
	"testing"

	"go.xchunk.org/beautify/dra"
)

// txt is a Composable of fixed text that ignores its bound.
func txt(s string) Composable {
	return ComposableFunc(func(ctx Context) (dra.Model, error) {
		return ctx.Factory().Text(s), nil
	})
}

// recorder remembers the widths it was rendered with; -1 stands for
// unbounded.
type recorder struct {
	text   string
	widths *[]int
}

func (r recorder) Render(ctx Context) (dra.Model, error) {
	w, bounded := ctx.Width()
	if !bounded {
		w = -1
	}
	*r.widths = append(*r.widths, w)
	return ctx.Factory().Text(r.text), nil
}

var errItem = errors.New("item failed")

func failing() Composable {
	return ComposableFunc(func(Context) (dra.Model, error) { return dra.Model{}, errItem })
}

func TestTemplateRender(t *testing.T) {
	inline := func(items ...Composable) Template { return NewTemplate(items...) }
	stacked := func(items ...Composable) Template {
		return NewTemplate(items...).With(WithDirection(Stacked))
	}
	sep := WithSeparator(txt("|"))
	tests := []struct {
		name          string
		tpl           Template
		want          string
		width, height int
	}{
		{"zero template", Template{}, "", 0, 0},
		{"no items with a separator", inline().With(sep), "", 0, 0},
		{"one item", inline(txt("a")).With(sep), "a", 1, 1},
		{"inline", inline(txt("a"), txt("b")), "ab", 2, 1},
		{"inline separator", inline(txt("a"), txt("b"), txt("c")).With(sep), "a|b|c", 5, 1},
		{"stacked", stacked(txt("a"), txt("bb")), "a\nbb", 2, 2},
		{"stacked separator", stacked(txt("a"), txt("b")).With(sep), "a\n|\nb", 1, 3},
		{"multi-line separator", stacked(txt("a"), txt("b")).With(WithSeparator(txt("-\n="))), "a\n-\n=\nb", 1, 4},
		{"multi-line item side by side", inline(txt("a\nbb"), txt("|"), txt("c\nd")), "a |c\nbb d", 4, 2},
		{"unknown direction is inline", inline(txt("a"), txt("b")).With(WithDirection(Direction(9))), "ab", 2, 1},
		{"empty item keeps its separators", inline(txt("a"), txt(""), txt("b")).With(sep), "a||b", 4, 1},
		{"empty item keeps its separators, stacked", stacked(txt("a"), txt(""), txt("b")).With(sep), "a\n|\n|\nb", 1, 4},
		{"skip empty", inline(txt(""), txt("a"), txt(""), txt("b"), txt("")).With(sep, WithSkipEmpty(true)), "a|b", 3, 1},
		{"skip empty, all empty", inline(txt(""), txt("")).With(sep, WithSkipEmpty(true)), "", 0, 0},
		{"skip empty off again", inline(txt(""), txt("a")).With(sep, WithSkipEmpty(true), WithSkipEmpty(false)), "|a", 2, 1},
		{"nested", inline(txt("["), stacked(txt("a"), txt("b")), txt("]")), "[a]\n b", 3, 2},
		{"nil option", inline(txt("a")).With(nil), "a", 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := tt.tpl.Render(Context{})
			if err != nil {
				t.Fatal(err)
			}
			got := m.String()
			if got != tt.want || m.Width() != tt.width || m.Height() != tt.height {
				t.Errorf("got %q %dx%d, want %q %dx%d", got, m.Width(), m.Height(), tt.want, tt.width, tt.height)
			}
		})
	}
}

func TestTemplateWidths(t *testing.T) {
	tests := []struct {
		name  string
		ctx   Context
		texts []string
		opts  []TemplateOption
		want  []int
	}{
		{"inline unbounded", Context{}, []string{"abc", "de"}, nil, []int{-1, -1}},
		{"inline bounded", Context{}.WithWidth(10), []string{"abc", "de", "f"}, nil, []int{10, 7, 5}},
		{"inline with separator", Context{}.WithWidth(10), []string{"abc", "de", "f"},
			[]TemplateOption{WithSeparator(txt(", "))}, []int{10, 5, 1}},
		{"inline clamps at 0", Context{}.WithWidth(4), []string{"abc", "de", "f"}, nil, []int{4, 1, 0}},
		{"inline bound 0", Context{}.WithWidth(0), []string{"", ""}, nil, []int{0, 0}},
		{"skipped items take no room", Context{}.WithWidth(10), []string{"abc", "", "f"},
			[]TemplateOption{WithSeparator(txt(", ")), WithSkipEmpty(true)}, []int{10, 5, 5}},
		{"empty items take a separator", Context{}.WithWidth(10), []string{"abc", "", "f"},
			[]TemplateOption{WithSeparator(txt(", "))}, []int{10, 5, 3}},
		{"stacked bounded", Context{}.WithWidth(10), []string{"abc", "de"},
			[]TemplateOption{WithDirection(Stacked)}, []int{10, 10}},
		{"stacked unbounded", Context{}, []string{"abc"}, []TemplateOption{WithDirection(Stacked)}, []int{-1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []int
			var items []Composable
			for _, s := range tt.texts {
				items = append(items, recorder{s, &got})
			}
			if _, err := NewTemplate(items...).With(tt.opts...).Render(tt.ctx); err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("widths %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateSeparatorContext(t *testing.T) {
	var widths []int
	tpl := NewTemplate(txt("a"), txt("b"), txt("c")).With(WithSeparator(recorder{"|", &widths}))
	if _, err := tpl.Render(Context{}.WithWidth(10)); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(widths, []int{10}) {
		t.Errorf("separator rendered with %v, want once with the template's width", widths)
	}
	widths = nil
	if _, err := NewTemplate().With(WithSeparator(recorder{"|", &widths})).Render(Context{}); err != nil || widths != nil {
		t.Errorf("separator of an empty template was rendered: %v, %v", widths, err)
	}
}

func TestTemplateKeepsBound(t *testing.T) {
	items := []Composable{txt("abcdef"), txt("漢字\nlonger text")}
	for _, d := range []Direction{Inline, Stacked} {
		tpl := NewTemplate(items...).With(WithDirection(d), WithSeparator(txt(" | ")))
		for _, w := range []int{0, 1, 2, 5, 9, 12, 100} {
			m, err := tpl.Render(Context{}.WithWidth(w))
			if err != nil || m.Width() > w {
				t.Errorf("direction %d, bound %d: width %d, %v", d, w, m.Width(), err)
			}
		}
	}
}

func TestTemplateErrors(t *testing.T) {
	tests := []struct {
		name string
		tpl  Template
		want error
		msg  string
	}{
		{"nil item", NewTemplate(txt("a"), nil), ErrNilComposable, "item 1"},
		{"nil separator", NewTemplate(txt("a"), txt("b")).With(WithSeparator(nil)), ErrNilComposable, "separator"},
		{"failing item", NewTemplate(txt("a"), txt("b"), failing()), errItem, "item 2"},
		{"failing separator", NewTemplate(txt("a"), txt("b")).With(WithSeparator(failing())), errItem, "separator"},
		{"nested failure", NewTemplate(NewTemplate(failing())), errItem, "item 0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, d := range []Direction{Inline, Stacked} {
				m, err := tt.tpl.With(WithDirection(d)).Render(Context{})
				if !errors.Is(err, tt.want) || !strings.Contains(err.Error(), tt.msg) {
					t.Errorf("Render() error = %v, want %v mentioning %q", err, tt.want, tt.msg)
				}
				if !m.IsEmpty() {
					t.Errorf("model %q returned with an error", m.String())
				}
			}
		})
	}
}

func TestTemplateStopsAtFirstError(t *testing.T) {
	var widths []int
	_, err := NewTemplate(failing(), recorder{"x", &widths}).Render(Context{})
	if !errors.Is(err, errItem) || widths != nil {
		t.Errorf("error %v, later items rendered: %v", err, widths)
	}
}

func TestTemplateIsImmutable(t *testing.T) {
	items := []Composable{txt("a"), txt("b")}
	base := NewTemplate(items...)
	items[0] = txt("X")
	left, right := base.Append(txt("l")), base.Append(txt("r"), txt("s"))
	styled := base.With(WithSeparator(txt("-")))
	more := []Composable{txt("m")}
	appended := base.Append(more...)
	more[0] = txt("X")
	for want, tpl := range map[string]Template{
		"ab": base, "abl": left, "abrs": right, "a-b": styled, "abm": appended,
	} {
		m, err := tpl.Render(Context{})
		if err != nil || m.String() != want {
			t.Errorf("got %q, %v, want %q", m.String(), err, want)
		}
	}
	if base.Len() != 2 || right.Len() != 4 || (Template{}).Len() != 0 {
		t.Errorf("Len() = %d, %d", base.Len(), right.Len())
	}
}

func TestTemplateConcurrent(t *testing.T) {
	tpl := NewTemplate(txt("a"), txt("b")).With(WithSeparator(txt("|")))
	var wg sync.WaitGroup
	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if m, err := tpl.Render(Context{}.WithWidth(5)); err != nil || m.String() != "a|b" {
				t.Errorf("got %q, %v", m.String(), err)
			}
			tpl.Append(txt("c"))
		}()
	}
	wg.Wait()
}
