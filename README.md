# beautify
Make your output your own. A fully customizable library for displaying anything beautifully пиздато

## Installation

```sh
go get go.xchunk.org/beautify
```

Requires Go 1.24 or newer. The library has no dependencies besides the
standard library.

## Quick start

```go
package main

import (
	"log"
	"os"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/blocks"
	"go.xchunk.org/beautify/dra"
)

func main() {
	accent := dra.Foreground(dra.Cyan)

	report := beautify.NewTemplate(
		blocks.Border(
			blocks.KeyValue([]blocks.Pair{
				{Key: "Status", Value: blocks.Badge("running", dra.Background(dra.Green))},
				{Key: "Uptime", Value: blocks.Text("3 days")},
			}, blocks.KeyStyle(accent)),
			blocks.BorderTitle("server"),
			blocks.BorderColor(accent),
		),
		blocks.Table(
			blocks.TableHeaders(blocks.Text("Fruit"), blocks.Text("Qty")),
			blocks.TableRow(blocks.Text("Apple"), blocks.Text("1")),
			blocks.TableRow(blocks.Text("Kiwi"), blocks.Text("12")),
		),
	).With(beautify.WithDirection(beautify.Stacked))

	if err := beautify.NewComposer(os.Stdout).Render(report); err != nil {
		log.Fatal(err)
	}
}
```

```
╭─ server ──────────╮
│ Status:  running  │
│ Uptime: 3 days    │
╰───────────────────╯
╭───────┬─────╮
│ Fruit │ Qty │
├───────┼─────┤
│ Apple │ 1   │
│ Kiwi  │ 12  │
╰───────┴─────╯
```

The `Composer` detects the colors and the width of the terminal. When the
output is not a terminal it writes plain text without escape sequences.
`NO_COLOR`, `FORCE_COLOR` and `CLICOLOR_FORCE` are honored, and options fix
the profile or the width.

## Two levels

| Package | Level | What it is for |
|---------|-------|----------------|
| `beautify` | high | `Composable`, `Template` and `Composer`: what to show and where to write it |
| `beautify/blocks` | high | Ready-made elements: text, rules, borders, rows, tables, trees, lists and more |
| `beautify/dra` | low | `Model`, `Style`, layout operations and the `Engine` that produces bytes |

Every element lowers itself to a `dra.Model`, which is styled text. The
`Engine` renders a model for a color profile and converts colors the profile
cannot show. Everything the blocks do goes through the public API of `dra`,
so your own elements can do the same.

Styles inherit by one rule: an unset color means "same as the parent", and
attributes are only added. Text is sanitized when a model is built, so
external input cannot inject escape sequences.

## Real Mode and Long Mode

| | Real Mode (default) | Long Mode |
|---|---|---|
| Unit | one rune is one cell | a grapheme cluster of 0, 1 or 2 cells |
| Good for | Latin, Cyrillic, Greek, box drawing | CJK, emoji, combining marks, flags |
| Cost | no tables, fastest | table lookups for non-ASCII text |
| Limits | wide and combining text is misaligned | simplified Unicode tables, see `goals.md` |

The mode belongs to the factory:

```go
c := beautify.NewComposer(os.Stdout,
	beautify.WithFactory(dra.NewFactory(dra.WithMode(dra.ModeLong))),
)
```

## Write your own block

A block is anything that implements `Composable`. It gets a `Context` with a
factory and the available width, and returns a model that is not wider than
that.

```go
// Progress is a bar that fills the available width.
type Progress struct {
	Done, Total int
}

func (p Progress) Render(ctx beautify.Context) (dra.Model, error) {
	width, bounded := ctx.Width()
	if !bounded {
		width = 20
	}
	f := ctx.Factory()
	label := f.Text(fmt.Sprintf(" %d/%d", p.Done, p.Total))
	room := max(width-label.Width(), 0)
	filled := room * p.Done / max(p.Total, 1)
	bar := dra.Concat(
		dra.Repeat(f.Styled("█", dra.Foreground(dra.Green)), filled),
		dra.Repeat(f.Styled("░", dra.Foreground(dra.BrightBlack)), room-filled),
		label,
	)
	return dra.Truncate(bar, width, dra.Model{}), nil
}
```

`Progress{3, 4}` can now be used wherever a block is expected, for example
inside `blocks.Border` or a cell of `blocks.Row`.

## Status

`goals.md` lists what is deferred, such as live output and full Unicode
conformance, and what is deliberately left out.
