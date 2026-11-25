package blocks_test

import (
	"fmt"
	"os"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/blocks"
	"go.xchunk.org/beautify/dra"
)

// show prints a block without styles at a fixed width, so the output is
// stable.
func show(c beautify.Composable) {
	composer := beautify.NewComposer(os.Stdout,
		beautify.WithProfile(dra.ProfilePlain),
		beautify.WithWidth(40),
	)
	if err := composer.Render(beautify.NewTemplate(c)); err != nil {
		fmt.Println(err)
	}
}

func ExampleBorder() {
	show(blocks.Border(
		blocks.Text("All systems are running."),
		blocks.BorderTitle("status"),
		blocks.BorderCaption("ok"),
	))
	// Output:
	// ╭─ status ──────────── ok ─╮
	// │ All systems are running. │
	// ╰──────────────────────────╯
}

func ExampleRow() {
	show(blocks.Row([]blocks.Cell{
		blocks.Fixed(8, blocks.Text("fixed")),
		blocks.Flex(1, blocks.Text("flexible")),
		blocks.Auto(blocks.Text("auto")),
	}, blocks.RowGap(1)))
	// Output:
	// fixed    flexible                   auto
}

func ExampleColumn() {
	show(blocks.Column([]beautify.Composable{
		blocks.Text("first"),
		blocks.Text("second"),
	}, blocks.ColumnGap(1)))
	// Output:
	// first
	//
	// second
}

func ExampleKeyValue() {
	show(blocks.KeyValue([]blocks.Pair{
		{Key: "Host", Value: blocks.Text("example")},
		{Key: "Kernel", Value: blocks.Text("a version string that is long")},
	}))
	// Output:
	// Host  : example
	// Kernel: a version string that is long
}

func ExampleList() {
	show(blocks.List([]beautify.Composable{
		blocks.Text("one"),
		blocks.Text("two\nlines"),
	}, blocks.ListOrdered()))
	// Output:
	// 1. one
	// 2. two
	//    lines
}

func ExampleTree() {
	show(blocks.Tree(blocks.Node{
		Label: blocks.Text("beautify"),
		Children: []blocks.Node{
			{Label: blocks.Text("dra"), Children: []blocks.Node{
				{Label: blocks.Text("model.go")},
			}},
			{Label: blocks.Text("blocks")},
		},
	}))
	// Output:
	// beautify
	// ├── dra
	// │   └── model.go
	// └── blocks
}

func ExampleSections() {
	show(blocks.Sections([]blocks.Section{
		{Title: "System", Body: blocks.Text("OS      Ubuntu 24.04")},
		{Title: "Hardware", Body: blocks.Text("CPU     Ryzen 7 5800X")},
	}, blocks.SectionsCaption("host")))
	// Output:
	// ╭─ System ─────── host ─╮
	// │ OS      Ubuntu 24.04  │
	// ├─ Hardware ────────────┤
	// │ CPU     Ryzen 7 5800X │
	// ╰───────────────────────╯
}

func ExampleTable() {
	show(blocks.Table(
		blocks.TableHeaders(blocks.Text("Fruit"), blocks.Text("Qty")),
		blocks.TableRow(blocks.Text("Apple"), blocks.Text("1")),
		blocks.TableRow(blocks.Text("Kiwi"), blocks.Text("12")),
		blocks.TableColumns(blocks.ColumnSpec{}, blocks.ColumnSpec{Align: dra.AlignRight}),
	))
	// Output:
	// ╭───────┬─────╮
	// │ Fruit │ Qty │
	// ├───────┼─────┤
	// │ Apple │   1 │
	// │ Kiwi  │  12 │
	// ╰───────┴─────╯
}
