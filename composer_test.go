package beautify

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"go.xchunk.org/beautify/dra"
)

// red is a Composable of red text.
func red(s string) Composable {
	return ComposableFunc(func(ctx Context) (dra.Model, error) {
		return ctx.Factory().Styled(s, dra.Style{FG: dra.Red}), nil
	})
}

// writer records calls and can fail or write short.
type writer struct {
	bytes.Buffer
	calls int
	limit int // bytes accepted per call, 0 for all
	err   error
}

func (w *writer) Write(p []byte) (int, error) {
	w.calls++
	if w.limit > 0 && len(p) > w.limit {
		p = p[:w.limit]
	}
	n, _ := w.Buffer.Write(p)
	return n, w.err
}

// cleanEnv removes the variables that influence a Composer.
func cleanEnv(t *testing.T) {
	for _, key := range []string{"FORCE_COLOR", "CLICOLOR_FORCE", "NO_COLOR", "COLUMNS"} {
		t.Setenv(key, "")
	}
}

func TestComposerRender(t *testing.T) {
	cleanEnv(t)
	tests := []struct {
		name  string
		opts  []ComposerOption
		tpl   Template
		want  string
		calls int
	}{
		{"plain by default", nil, NewTemplate(red("hi")), "hi\n", 1},
		{"profile", []ComposerOption{WithProfile(dra.ProfileANSI16)}, NewTemplate(red("hi")), "\x1b[0;31mhi\x1b[0m\n", 1},
		{"no newline", []ComposerOption{WithNewline(false)}, NewTemplate(txt("hi")), "hi", 1},
		{"newline on again", []ComposerOption{WithNewline(false), WithNewline(true)}, NewTemplate(txt("hi")), "hi\n", 1},
		{"empty template", nil, Template{}, "", 0},
		{"empty item", nil, NewTemplate(txt("")), "", 0},
		{"one empty line", nil, NewTemplate(ComposableFunc(func(Context) (dra.Model, error) {
			return dra.Blank(0, 1), nil
		})), "\n", 1},
		{"one empty line without newline", []ComposerOption{WithNewline(false)}, NewTemplate(
			ComposableFunc(func(Context) (dra.Model, error) { return dra.Blank(0, 1), nil })), "", 0},
		{"multi-line", nil, NewTemplate(txt("a\nb")), "a\nb\n", 1},
		{"nil option", []ComposerOption{nil}, NewTemplate(txt("a")), "a\n", 1},
		{"factory", []ComposerOption{WithFactory(dra.NewFactory(dra.WithTabWidth(1)))}, NewTemplate(txt("\t")), " \n", 1},
		{"nil factory", []ComposerOption{WithFactory(nil)}, NewTemplate(txt("a")), "a\n", 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w writer
			if err := NewComposer(&w, tt.opts...).Render(tt.tpl); err != nil {
				t.Fatal(err)
			}
			if w.String() != tt.want || w.calls != tt.calls {
				t.Errorf("wrote %q in %d calls, want %q in %d", w.String(), w.calls, tt.want, tt.calls)
			}
		})
	}
}

func TestComposerProfile(t *testing.T) {
	cleanEnv(t)
	if got := NewComposer(&writer{}).Profile(); got != dra.ProfilePlain {
		t.Errorf("detected %v for a buffer", got)
	}
	if got := NewComposer(&writer{}, WithProfile(dra.ProfileANSI256)).Profile(); got != dra.ProfileANSI256 {
		t.Errorf("Profile() = %v", got)
	}
	t.Setenv("FORCE_COLOR", "1")
	t.Setenv("COLORTERM", "truecolor")
	c := NewComposer(&writer{})
	t.Setenv("FORCE_COLOR", "")
	if got := c.Profile(); got != dra.ProfileTrueColor {
		t.Errorf("Profile() = %v; it must be resolved once", got)
	}
}

func TestComposerWidth(t *testing.T) {
	cleanEnv(t)
	tests := []struct {
		name    string
		columns string
		opts    []ComposerOption
		want    int // -1 stands for unbounded
	}{
		{"unknown", "", nil, -1},
		{"COLUMNS", "72", nil, 72},
		{"COLUMNS is not a number", "wide", nil, -1},
		{"COLUMNS is zero", "0", nil, -1},
		{"COLUMNS is negative", "-5", nil, -1},
		{"fixed", "72", []ComposerOption{WithWidth(30)}, 30},
		{"fixed unbounded", "72", []ComposerOption{WithWidth(0)}, -1},
		{"fixed negative", "72", []ComposerOption{WithWidth(-1)}, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var widths []int
			c := NewComposer(&writer{}, tt.opts...)
			t.Setenv("COLUMNS", tt.columns)
			if err := c.Render(NewTemplate(recorder{"x", &widths})); err != nil {
				t.Fatal(err)
			}
			if len(widths) != 1 || widths[0] != tt.want {
				t.Errorf("widths %v, want %d", widths, tt.want)
			}
		})
	}
}

func TestComposerErrors(t *testing.T) {
	var w writer
	if err := NewComposer(&w).Render(NewTemplate(txt("a"), failing())); !errors.Is(err, errItem) || w.calls != 0 {
		t.Errorf("failing item: %v, %d writes", err, w.calls)
	}
	failure := errors.New("disk full")
	w = writer{limit: 1, err: failure}
	if err := NewComposer(&w).Render(NewTemplate(txt("abc"))); err != failure {
		t.Errorf("failing writer: %v", err)
	}
	w = writer{limit: 1}
	if err := NewComposer(&w).Render(NewTemplate(txt("abc"))); err != io.ErrShortWrite {
		t.Errorf("short write: %v", err)
	}
}

func TestComposerNilWriter(t *testing.T) {
	if err := NewComposer(nil).Render(NewTemplate(txt("a"))); err != nil {
		t.Error(err)
	}
}

func TestComposerReusesBuffersSafely(t *testing.T) {
	var w writer
	c := NewComposer(&w, WithProfile(dra.ProfilePlain))
	large := strings.Repeat("x", 100<<10)
	for _, s := range []string{large, "a", large, "bb"} {
		w.Reset()
		if err := c.Render(NewTemplate(txt(s))); err != nil || w.String() != s+"\n" {
			t.Fatalf("wrote %d bytes, %v, want %d", w.Len(), err, len(s)+1)
		}
	}
}

func TestComposerConcurrent(t *testing.T) {
	var w writer
	c := NewComposer(&w, WithProfile(dra.ProfileANSI16), WithWidth(40))
	const goroutines = 100
	block := func(i int) string { return fmt.Sprintf("begin %d\nmiddle %d\nend %d", i, i, i) }
	var wg sync.WaitGroup
	for i := range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := c.Render(NewTemplate(txt(block(i)))); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if w.calls != goroutines {
		t.Errorf("%d writes, want %d", w.calls, goroutines)
	}
	out := "\n" + w.String()
	for i := range goroutines {
		if !strings.Contains(out, "\n"+block(i)+"\n") {
			t.Errorf("block %d is not intact", i)
		}
	}
}
