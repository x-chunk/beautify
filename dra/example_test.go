package dra_test

import (
	"fmt"

	"go.xchunk.org/beautify/dra"
)

// show prints a model without styles, so the output is stable.
func show(m dra.Model) {
	fmt.Println(dra.NewEngine(dra.ProfilePlain).String(m))
}

func Example() {
	f := dra.NewFactory()
	m := f.Styled("done", dra.Foreground(dra.Green).WithAttrs(dra.Bold))
	fmt.Printf("%q\n", dra.NewEngine(dra.ProfileANSI16).String(m))
	fmt.Printf("%q\n", dra.NewEngine(dra.ProfilePlain).String(m))
	// Output:
	// "\x1b[0;1;32mdone\x1b[0m"
	// "done"
}

func ExampleFactory_New() {
	m, err := dra.NewFactory().New(
		dra.WithText("INFO ready"),
		dra.WithPalette(dra.Foreground(dra.Green)),
		dra.WithColormap(dra.Runs(
			dra.Run{N: 4, Index: 0},
			dra.Run{N: 6, Index: dra.NoStyle},
		)),
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("%q\n", dra.NewEngine(dra.ProfileANSI16).String(m))
	// Output: "\x1b[0;32mINFO\x1b[0m ready"
}

func ExampleModel_Fill() {
	f := dra.NewFactory()
	blank := f.MustNew(dra.WithFrame("[ ", " ]"))
	for _, name := range []string{"one", "two"} {
		show(blank.Fill(f.Text(name)))
	}
	// Output:
	// [ one ]
	// [ two ]
}

func ExampleStack() {
	f := dra.NewFactory()
	show(dra.Stack(dra.AlignRight, f.Text("right"), f.Text("aligned")))
	// Output:
	//   right
	// aligned
}

func ExampleJoin() {
	f := dra.NewFactory()
	show(dra.Join(dra.AlignMiddle, f.Text("a\nb\nc"), f.Text(" <- "), f.Text("x\ny\nz")))
	// Output:
	// a    x
	// b <- y
	// c    z
}

func ExampleWrap() {
	m := dra.NewFactory().Text("The quick brown fox jumps over the lazy dog")
	show(dra.Wrap(m, 16, dra.WrapWords))
	// Output:
	// The quick brown
	// fox jumps over
	// the lazy dog
}

func ExampleTruncate() {
	f := dra.NewFactory()
	show(dra.Truncate(f.Text("a long line\nshort"), 8, f.Text("…")))
	// Output:
	// a long …
	// short
}

func ExampleRestyleFunc() {
	g := dra.NewGradient(dra.AxisHorizontal, dra.RGB(0, 0, 0), dra.RGB(200, 0, 0))
	m := dra.NewFactory().Text("abc")
	m = dra.RestyleFunc(m, func(x, y int) dra.Style {
		return dra.Foreground(g.At(x, y, 3, 1))
	})
	fmt.Printf("%q\n", dra.NewEngine(dra.ProfileTrueColor).String(m))
	// Output: "\x1b[0;38;2;0;0;0ma\x1b[0;38;2;100;0;0mb\x1b[0;38;2;200;0;0mc\x1b[0m"
}
