package beautify_test

import (
	"fmt"
	"os"
	"strings"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

// text is a Composable of plain text.
func text(s string) beautify.Composable {
	return beautify.ComposableFunc(func(ctx beautify.Context) (dra.Model, error) {
		return ctx.Factory().Text(s), nil
	})
}

func Example() {
	c := beautify.NewComposer(os.Stdout,
		beautify.WithProfile(dra.ProfilePlain),
		beautify.WithWidth(40),
	)
	t := beautify.NewTemplate(text("one"), text("two"), text("three")).
		With(beautify.WithSeparator(text(" | ")))
	if err := c.Render(t); err != nil {
		fmt.Println(err)
	}
	// Output: one | two | three
}

// progress is a custom Composable written with dra operations.
type progress struct {
	done, total int
}

// Render draws a bar that fills the available width.
func (p progress) Render(ctx beautify.Context) (dra.Model, error) {
	width, bounded := ctx.Width()
	if !bounded {
		width = 20
	}
	f := ctx.Factory()
	label := f.Text(fmt.Sprintf(" %d/%d", p.done, p.total))
	room := max(width-label.Width(), 0)
	filled := room * p.done / max(p.total, 1)
	bar := dra.Concat(
		dra.Repeat(f.Styled("#", dra.Foreground(dra.Green)), filled),
		dra.Repeat(f.Text("."), room-filled),
		label,
	)
	// The result must not be wider than the bound.
	return dra.Truncate(bar, width, dra.Model{}), nil
}

func ExampleComposable() {
	c := beautify.NewComposer(os.Stdout,
		beautify.WithProfile(dra.ProfilePlain),
		beautify.WithWidth(24),
	)
	if err := c.Render(beautify.NewTemplate(progress{done: 3, total: 4})); err != nil {
		fmt.Println(err)
	}
	// Output: ###############..... 3/4
}

func ExampleTemplate() {
	header := beautify.NewTemplate(text("name"), text("value")).
		With(beautify.WithSeparator(text("  ")))
	rule := text(strings.Repeat("-", 11))
	t := beautify.NewTemplate(header, rule, text("key   42")).
		With(beautify.WithDirection(beautify.Stacked))

	c := beautify.NewComposer(os.Stdout, beautify.WithProfile(dra.ProfilePlain))
	if err := c.Render(t); err != nil {
		fmt.Println(err)
	}
	// Output:
	// name  value
	// -----------
	// key   42
}
