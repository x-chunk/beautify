package blocks

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
	"go.xchunk.org/beautify/internal/golden"
)

var (
	red  = dra.Style{FG: dra.Red}
	blue = dra.Style{FG: dra.Blue, Attrs: dra.Bold}
	onBG = dra.Style{BG: dra.Green}

	realFactory = dra.NewFactory()
	longFactory = dra.NewFactory(dra.WithMode(dra.ModeLong))
)

// unbounded stands for "no bound" where tests take a width.
const unbounded = -1

func context(f *dra.Factory, width int) beautify.Context {
	ctx := beautify.NewContext(f)
	if width != unbounded {
		ctx = ctx.WithWidth(width)
	}
	return ctx
}

// render renders c with the default factory and fails the test on error.
func render(t *testing.T, c beautify.Composable, width int) dra.Model {
	t.Helper()
	return renderWith(t, realFactory, c, width)
}

func renderWith(t *testing.T, f *dra.Factory, c beautify.Composable, width int) dra.Model {
	t.Helper()
	m, err := c.Render(context(f, width))
	if err != nil {
		t.Fatalf("Render() at width %d: %v", width, err)
	}
	return m
}

// checkText compares the visible text of c with lines joined by newlines.
func checkText(t *testing.T, c beautify.Composable, width int, lines ...string) {
	t.Helper()
	want := strings.Join(lines, "\n")
	if got := render(t, c, width).String(); got != want {
		t.Errorf("width %d:\n got %q\nwant %q\n%s", width, got, want, got)
	}
}

// checkEmpty verifies that c renders an empty model.
func checkEmpty(t *testing.T, c beautify.Composable, width int) {
	t.Helper()
	if m := render(t, c, width); !m.IsEmpty() {
		t.Errorf("width %d: got %q, want an empty model", width, m.String())
	}
}

// checkGolden compares c with golden files at a plain and a styled profile.
func checkGolden(t *testing.T, name string, c beautify.Composable, width int) {
	t.Helper()
	m := render(t, c, width)
	for _, p := range []dra.Profile{dra.ProfilePlain, dra.ProfileANSI256} {
		golden.Assert(t, name+"_"+p.String(), dra.NewEngine(p).Append(nil, m))
	}
}

// checkRectangle verifies that all lines of m have the same width.
func checkRectangle(t *testing.T, m dra.Model) {
	t.Helper()
	for _, l := range dra.Lines(m) {
		if l.Width() != m.Width() {
			t.Errorf("line %q is %d wide, the model is %d", l.String(), l.Width(), m.Width())
		}
	}
}

// children are contents that stress width handling.
var children = map[string]beautify.Composable{
	"empty": raw(""),
	"short": raw("ab"),
	// raw ignores its bound, so it also tests that blocks enforce theirs.
	"long word":  raw("incomprehensibilities"),
	"multi-line": raw("one\ntwo words\n\nlast"),
	"wide":       raw("漢字 and e\u0301 \U0001F468\u200d\U0001F469\u200d\U0001F467 \U0001F1E6\U0001F1E7"),
	"newlines":   raw("\n\n"),
	"trailing":   raw("end\n"),
	"hostile":    raw("\x1b[31m\ta\r\nb\rc\xff\u202e"),
	"model":      fixedModel(realFactory.Text("漢字 prepared").Fill(longFactory.Text("漢字"))),
	"wrapping":   wrapping("some text that wraps at word boundaries"),
}

// raw is fixed text that ignores the bound it is given.
func raw(s string) beautify.Composable {
	return beautify.ComposableFunc(func(ctx beautify.Context) (dra.Model, error) {
		return ctx.Factory().Text(s), nil
	})
}

// wrapping is text that honors its bound.
func wrapping(s string) beautify.Composable {
	return beautify.ComposableFunc(func(ctx beautify.Context) (dra.Model, error) {
		m := ctx.Factory().Text(s)
		if w, ok := ctx.Width(); ok {
			m = dra.Truncate(dra.Wrap(m, max(w, 1), dra.WrapWords), w, dra.Model{})
		}
		return m, nil
	})
}

// fixedModel is a child of fixed content in any width mode.
func fixedModel(m dra.Model) beautify.Composable {
	return beautify.ComposableFunc(func(beautify.Context) (dra.Model, error) { return m, nil })
}

var errChild = errors.New("child failed")

func failing() beautify.Composable {
	return beautify.ComposableFunc(func(beautify.Context) (dra.Model, error) {
		return dra.Model{}, errChild
	})
}

// checkContract verifies the rules that hold for every block: the result is
// never wider than the bound, a bound of 0 gives an empty model, every
// profile renders the same text, and concurrent renders agree.
func checkContract(t *testing.T, name string, c beautify.Composable) {
	t.Helper()
	for _, f := range []*dra.Factory{realFactory, longFactory} {
		for _, width := range []int{0, 1, 2, 3, 5, 10, 20, 40, unbounded} {
			label := fmt.Sprintf("%s, mode %d, width %d", name, f.Mode(), width)
			m, err := c.Render(context(f, width))
			if err != nil {
				t.Errorf("%s: %v", label, err)
				continue
			}
			if width != unbounded && m.Width() > width {
				t.Errorf("%s: result is %d wide:\n%s", label, m.Width(), m.String())
			}
			if width == 0 && !m.IsEmpty() {
				t.Errorf("%s: got %q, want an empty model", label, m.String())
			}
			if out := dra.NewEngine(dra.ProfilePlain).String(m); strings.ContainsAny(out, "\x1b\r\t") {
				t.Errorf("%s: unsafe output %q", label, out)
			}
		}
	}

	want := render(t, c, 12).String()
	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if m, err := c.Render(context(realFactory, 12)); err != nil || m.String() != want {
				t.Errorf("%s: concurrent render gave %q, %v, want %q", name, m.String(), err, want)
			}
		}()
	}
	wg.Wait()
}

// checkChildContract runs checkContract for a block around every kind of
// child.
func checkChildContract(t *testing.T, build func(child beautify.Composable) beautify.Composable) {
	t.Helper()
	for name, child := range children {
		checkContract(t, name, build(child))
	}
}

// checkChildErrors verifies how a block reports a nil and a failing child.
// The messages must mention the block and the index of the child.
func checkChildErrors(t *testing.T, block string, index int, build func(child beautify.Composable) beautify.Composable) {
	t.Helper()
	for _, width := range []int{1, 20, unbounded} {
		for want, child := range map[error]beautify.Composable{
			beautify.ErrNilComposable: nil,
			errChild:                  failing(),
		} {
			m, err := build(child).Render(context(realFactory, width))
			if !errors.Is(err, want) {
				t.Errorf("width %d: error = %v, want %v", width, err, want)
				continue
			}
			if msg := fmt.Sprintf("%s child %d", block, index); !strings.Contains(err.Error(), msg) {
				t.Errorf("width %d: error %q does not mention %q", width, err, msg)
			}
			if !m.IsEmpty() {
				t.Errorf("width %d: model %q returned with an error", width, m.String())
			}
		}
	}
}
