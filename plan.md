# Implementation plan: beautify v1

This plan describes the first version of `beautify` completely: packages,
public API, exact behavior, edge cases, tests and the order of work. Read
`AGENTS.md` first; its rules apply to every step. `goals.md` lists what is
deliberately not part of this version.

## Contents

1. Scope and principles
2. Global requirements
3. Repository layout
4. Milestone 0: scaffolding
5. Milestone 1: `internal/ansi`
6. Milestone 2: `internal/width`
7. Milestone 3: `dra` styles and profiles
8. Milestone 4: `dra` Model and Factory
9. Milestone 5: `dra` Engine
10. Milestone 6: `dra` layout operations
11. Milestone 7: `beautify` core
12. Milestone 8: `blocks`
13. Milestone 9: documentation and examples
14. Cross-cutting edge-case matrix
15. Final verification checklist
16. Decision log

---

## 1. Scope and principles

### 1.1 What is built

A Go library with two levels:

- **Low level, package `dra`.** The universal `Model`, the `Factory` that
  creates models, the `Engine` (the Direct Render API) that turns a model
  into bytes, and a set of layout operations on models.
- **High level, packages `beautify` and `blocks`.** `Composable`,
  `Template`, `Composer`, and a library of standard blocks.

Pipeline:

```
Composer.Render(Template)
  -> Template.Render(ctx)            one Model for the whole template
       -> item.Render(ctx) -> Model  for every Composable in it
  -> Engine.Append(buf, model)       bytes
  -> one Write call to the io.Writer
```

### 1.2 Principles that decide ambiguous cases

When this plan does not cover a situation, decide with these rules, in this
order:

1. **Predictability over cleverness.** One rule without exceptions beats a
   smart default that surprises.
2. **`dra` is powerful, `Composer` is minimal.** A new capability goes into
   `dra` or `blocks`.
3. **Blocks have no privileges.** `blocks` uses only the public API. If a
   block cannot be written cleanly, extend `dra` and record the addition in
   the decision log (section 16).
4. **Validate once, render fast.** Only `Factory.New` validates. The Engine
   and the layout operations trust their inputs and cannot fail.
5. **Never panic on user input.** Out-of-range sizes are clamped, nil
   values produce errors or are treated as empty, as specified per case.

### 1.3 Out of scope

Redraw and live output, full Unicode conformance, Windows console setup,
semantic themes, a `log/slog` adapter, overlays, absolute positioning and a
flexbox-style engine. See `goals.md` and the non-goals in `AGENTS.md`.

---

## 2. Global requirements

- **Language and module.** Go 1.24. Module path
  `go.xchunk.org/beautify`. `go.mod` has no `require` entries.
- **Dependencies.** Standard library only, in code and in tests.
- **Immutability.** `Model`, `Factory`, `Style`, `Color`, `Engine`,
  `Template` and `Context` are immutable after creation. Every function
  that takes a slice from the caller copies it before storing it. Every
  function that returns a slice returns a copy or a slice the caller owns.
- **Concurrency.** All immutable types are safe for concurrent use without
  locking. `Composer` is safe for concurrent `Render` calls.
- **Errors.** Sentinel errors are exported variables; wrap them with
  `fmt.Errorf("...: %w", ...)` to add detail so callers can use
  `errors.Is`.
- **Documentation.** Every exported identifier has a doc comment per
  `AGENTS.md`. Every package has a `doc.go` with a package comment.
- **Formatting and static checks.** `gofmt -l .` prints nothing.
  `go vet ./...` is clean.
- **Tests.** `go test -race ./...` passes. Table-driven tests are the
  default. Rendering tests use golden files (section 4.3). Line coverage is
  at least 90% for `dra` and `internal/*`, and at least 85% for `beautify`
  and `blocks`.
- **Commits.** Atomic, per `AGENTS.md`. Each milestone lists its expected
  commits. Every commit builds and passes tests. Push to the working branch
  after each milestone.

---

## 3. Repository layout

```
go.mod
doc.go                    package beautify
composable.go             Composable, ComposableFunc
context.go                Context
template.go               Template, TemplateOption
composer.go               Composer, ComposerOption
profile.go                DetectProfile
*_test.go, example_test.go
blocks/
  doc.go
  text.go space.go rule.go model.go          primitives
  styled.go padding.go border.go align.go
  sized.go badge.go                          wrappers
  row.go column.go table.go tree.go
  list.go keyvalue.go sections.go            containers
  imports_test.go                            enforces "no internal imports"
  *_test.go, testdata/
dra/
  doc.go
  color.go attr.go style.go profile.go
  mode.go
  model.go metrics.go sanitize.go
  factory.go option.go colormap.go
  engine.go
  flat.go                                    composed-model representation
  ops_compose.go ops_pad.go ops_lines.go
  ops_truncate.go ops_wrap.go
  measure.go                                 StringWidth
  gradient.go                                Gradient, Axis
  *_test.go, testdata/
internal/
  ansi/     SGR encoding and color downgrade
  width/    rune widths and cluster segmentation
  term/     terminal detection and size
  golden/   golden-file test helper
.github/workflows/ci.yml
```

File names are a guide; split or merge files when it keeps them readable.
Package boundaries and import directions are fixed.

Import rules:

- `internal/ansi`, `internal/width`, `internal/term`, `internal/golden`
  import nothing from this module.
- `dra` imports `internal/ansi` and `internal/width`.
- `beautify` imports `dra` and `internal/term`.
- `blocks` imports `beautify` and `dra` only. Its test files may import
  `internal/golden`.

---

## 4. Milestone 0: scaffolding

### 4.1 Work

- Create `go.mod` (`module go.xchunk.org/beautify`, `go 1.24`).
- Add `doc.go` for the root package with a short package comment.
- Add `.github/workflows/ci.yml`: on push and pull request, with Go 1.24 on
  `ubuntu-latest`, `macos-latest` and `windows-latest`, run the `gofmt`
  check (Linux only), `go vet ./...` and `go test -race ./...`.

### 4.2 Golden helper

`internal/golden` exposes one function:

```go
// Assert compares got with the content of testdata/<name>.golden relative
// to the test's package directory. With the -update flag it writes got to
// the file instead.
func Assert(t *testing.T, name string, got []byte)
```

- The `-update` flag is registered once in this package.
- Golden files store raw bytes, including escape sequences.
- On mismatch, the failure message shows both values with `%q` so escape
  sequences are readable.
- A missing golden file fails the test with a message that tells the reader
  to run with `-update`.

### 4.3 Golden conventions

- One golden file per case, named `<test>_<case>.golden`.
- Review golden diffs by hand before committing; never regenerate blindly.

### 4.4 Checks

- `go build ./...`, `go vet ./...` and `go test ./...` succeed.

### 4.5 Commits

1. `Initialize Go module`
2. `Add golden-file test helper`
3. `Add CI workflow`

---

## 5. Milestone 1: `internal/ansi`

Encoding of SGR sequences and color conversion. This package has its own
primitive types so it does not import `dra`.

### 5.1 Types

```go
type Level uint8 // Plain, NoColor, ANSI16, ANSI256, TrueColor

type ColorKind uint8 // None, Default, Indexed16, Indexed256, RGB

type Color struct {
    Kind    ColorKind
    Index   uint8   // Indexed16: 0-15, Indexed256: 0-255
    R, G, B uint8   // RGB
}

type Style struct {
    FG, BG Color
    Attrs  uint8 // bit set, see 5.2
}
```

`None` means no color information. `Default` means the terminal default
color. Both produce no color parameter; they differ only for inheritance,
which is resolved in `dra` before encoding.

### 5.2 Attributes

Bits and SGR codes, encoded in this fixed order:

| Attribute | SGR |
|-----------|-----|
| Bold      | 1   |
| Faint     | 2   |
| Italic    | 3   |
| Underline | 4   |
| Blink     | 5   |
| Reverse   | 7   |
| Strike    | 9   |

### 5.3 Functions

```go
// Downgrade converts s to what level can express. The result is
// canonical: styles that look the same at this level compare equal.
func Downgrade(s Style, level Level) Style

// AppendSGR appends the sequence that switches the terminal to s from any
// previous state. s must already be downgraded.
func AppendSGR(dst []byte, s Style) []byte

// AppendReset appends ESC [ 0 m.
func AppendReset(dst []byte) []byte
```

### 5.4 Encoding rules

- Every style switch is a full reset followed by the new style in one
  sequence: `ESC [ 0 ; <attrs> ; <fg> ; <bg> m`. A switch to the empty
  style is `ESC [ 0 m`. Incremental diffs are not used: SGR cannot turn
  bold and faint off independently, and a full reset is deterministic.
- Parameter order: `0`, attributes in table order, foreground, background.
- Foreground: `Indexed16` 0-7 -> `30`-`37`, 8-15 -> `90`-`97`;
  `Indexed256` -> `38;5;N`; `RGB` -> `38;2;R;G;B`.
- Background: `Indexed16` 0-7 -> `40`-`47`, 8-15 -> `100`-`107`;
  `Indexed256` -> `48;5;N`; `RGB` -> `48;2;R;G;B`.
- `None` and `Default` emit no parameter (the leading `0` already restores
  the default).
- Numbers are written without leading zeros and without allocation.

Reference outputs (must be pinned by tests):

| Style | Bytes |
|-------|-------|
| empty | `\x1b[0m` |
| bold, fg Indexed16 1 | `\x1b[0;1;31m` |
| fg Indexed16 9, bg Indexed16 4 | `\x1b[0;91;44m` |
| bold+underline, fg Indexed256 196 | `\x1b[0;1;4;38;5;196m` |
| fg RGB 255,128,0, bg RGB 0,0,0 | `\x1b[0;38;2;255;128;0;48;2;0;0;0m` |

### 5.5 Downgrade rules

- `Plain`: result is the empty style (no colors, no attributes).
- `NoColor`: colors become `None`, attributes are kept.
- `TrueColor`: unchanged, except canonicalization below.
- `ANSI256`: `RGB` -> `Indexed256`. `Indexed16` stays.
- `ANSI16`: `RGB` and `Indexed256` -> `Indexed16`.
- Canonicalization at every level: `Default` becomes `None`; `Indexed256`
  with index 0-15 becomes `Indexed16` with the same index.

RGB to 256:

- Cube levels are `0, 95, 135, 175, 215, 255`. For each channel pick the
  nearest level (ties go to the lower level). Cube index is
  `16 + 36*r + 6*g + b`.
- Gray ramp: index `232 + i` has value `8 + 10*i` for `i` in 0-23. Pick the
  nearest gray to the channel average (ties go to the lower index).
- Compare the squared RGB distance of the cube candidate and the gray
  candidate to the original color. The smaller wins; on a tie the cube
  wins.
- RGB never maps to indexes 0-15 here, because those depend on the
  terminal theme.

To 16 colors:

- Convert `Indexed256` 16-255 to its RGB value (cube and gray formulas
  above), then pick the nearest of the 16 colors by squared RGB distance
  using the xterm defaults below. Ties go to the lower index.

| Idx | RGB | Idx | RGB |
|-----|--------|-----|--------|
| 0 | 000000 | 8  | 808080 |
| 1 | 800000 | 9  | ff0000 |
| 2 | 008000 | 10 | 00ff00 |
| 3 | 808000 | 11 | ffff00 |
| 4 | 000080 | 12 | 0000ff |
| 5 | 800080 | 13 | ff00ff |
| 6 | 008080 | 14 | 00ffff |
| 7 | c0c0c0 | 15 | ffffff |

### 5.6 Edge cases

- Pure black, pure white and exact cube colors map to themselves.
- Grays that are equally close to a cube color and a ramp color follow the
  tie rule.
- A style with attributes only encodes with no color parameters.
- Two different RGB colors that land on the same index downgrade to equal
  styles (this is what lets the Engine skip redundant switches).
- `AppendSGR` and `AppendReset` do not allocate when `dst` has capacity.

### 5.7 Tests

- Table tests for every row of 5.4 and each rule of 5.5.
- Exhaustive test: every `Indexed256` value downgrades to a valid
  `Indexed16` value; every cube and gray index round-trips through its RGB
  value back to itself at `ANSI256`.
- `testing.AllocsPerRun` asserts zero allocations for `AppendSGR`.

### 5.8 Commits

1. `Add SGR encoding to internal/ansi`
2. `Add color downgrade to internal/ansi`

---

## 6. Milestone 2: `internal/width`

Rune widths and cluster segmentation for the two width modes.

### 6.1 API

```go
type Mode uint8 // Real, Long

// Next returns the byte length and cell width of the first cluster of s.
// For empty s it returns 0, 0.
func Next(s string, mode Mode) (size, width int)

// String returns the cell width of s, which must not contain '\n'.
func String(s string, mode Mode) int
```

All layout code iterates text through `Next`, so cluster rules live in one
place.

### 6.2 Real Mode

- Every rune is its own cluster with width 1.
- This includes combining marks, zero-width characters, CJK and emoji.
  Their measured width can differ from what a terminal shows. This is the
  documented trade-off of Real Mode: no tables, maximum speed.
- Invalid UTF-8: each invalid byte is one cluster of width 1 (it decodes as
  U+FFFD).

### 6.3 Long Mode: segmentation

A cluster is a base rune plus everything that attaches to it. After reading
the base rune, keep consuming while the next rune is an extender:

- A combining mark: Unicode categories `Mn`, `Me` or `Mc`.
- A variation selector: U+FE00-U+FE0F, U+E0100-U+E01EF.
- An emoji modifier (skin tone): U+1F3FB-U+1F3FF.
- A tag character: U+E0020-U+E007F.
- A Hangul medial vowel or final consonant: U+1160-U+11FF,
  U+D7B0-U+D7FF.
- U+200D (zero-width joiner). After a ZWJ, the following rune, if any, is
  consumed as part of the same cluster regardless of what it is, and
  extender matching continues after it.

Regional indicators (U+1F1E6-U+1F1FF) pair up: a regional indicator
followed by another forms one cluster of both; extender matching continues
after the pair. A regional indicator not followed by another is a cluster
of its own.

A cluster never spans `'\n'`: a newline is never consumed as an extender or
as the rune after a ZWJ.

### 6.4 Long Mode: width

The width of a cluster is decided in this order:

1. If the cluster is a regional-indicator pair: 2.
2. If the cluster contains U+FE0F (emoji presentation): 2.
3. Otherwise the width of its first rune.

Width of a single rune:

- 0 for categories `Mn`, `Me` and `Cf`, for U+200B (zero-width space) and
  for the Hangul extender ranges above. Exception: U+00AD (soft hyphen)
  is 1.
- 2 for the wide ranges in 6.5.
- 1 for everything else, including ambiguous-width characters and a lone
  regional indicator.

### 6.5 Wide ranges

Hand-written, sorted, searched by binary search:

```
1100-115F   2E80-303E   3041-33FF   3400-4DBF   4E00-9FFF
A000-A4CF   A960-A97F   AC00-D7A3   F900-FAFF   FE10-FE19
FE30-FE6F   FF00-FF60   FFE0-FFE6
1F300-1F64F 1F680-1F6FF 1F900-1F9FF 1FA70-1FAFF
20000-2FFFD 30000-3FFFD
```

Single code points and short ranges with default emoji presentation in the
Basic Multilingual Plane:

```
231A-231B 23E9-23EC 23F0 23F3 25FD-25FE 2614-2615 2648-2653 267F 2693
26A1 26AA-26AB 26BD-26BE 26C4-26C5 26CE 26D4 26EA 26F2-26F3 26F5 26FA
26FD 2705 270A-270B 2728 274C 274E 2753-2755 2757 2795-2797 27B0 27BF
2B1B-2B1C 2B50 2B55
```

The emoji blocks are coarse on purpose. Differences from the Unicode
standard are a known limitation recorded in `goals.md`. Do not add a
generator in this version.

### 6.6 Edge cases

- Empty string: `Next` returns `0, 0`; `String` returns 0.
- A string that starts with an extender (a combining mark with no base):
  the extender is the base of its own cluster, so its width follows the
  single-rune rule (0 for `Mn`).
- A trailing ZWJ with nothing after it: part of the current cluster.
- `ZWJ` chains (family emoji): one cluster, width 2.
- A keycap sequence such as `1` + U+FE0F + U+20E3: one cluster, width 2
  (rule 2).
- A flag followed by a third regional indicator: one pair cluster (2) and
  one lone cluster (1).
- Text + U+FE0E (text presentation selector): width of the first rune.
- Invalid UTF-8 bytes: each is a cluster of width 1 in both modes.
- ASCII-only input takes a fast path: when the first byte is below 0x80
  and the next byte is below 0x80 or absent, return `1, 1` without table
  lookups.

### 6.7 Tests

- Table tests for every edge case above, in both modes.
- Property test: for any string, summing `size` over repeated `Next` calls
  equals `len(s)`, and every `size` is at least 1 for non-empty input.
- Fuzz test `FuzzNext`: never panics, always makes progress, never returns
  a negative width or a width above 2.
- Test that the wide table is sorted and non-overlapping.
- Benchmark `String` on ASCII, Cyrillic, CJK and emoji text in both modes.

### 6.8 Commits

1. `Add rune width tables to internal/width`
2. `Add cluster segmentation to internal/width`

---

## 7. Milestone 3: `dra` styles and profiles

### 7.1 Color

```go
// Color is a terminal color. The zero value is "unset" and inherits.
type Color struct{ /* unexported */ }

func ANSI(n uint8) Color        // 16-color palette; n is taken modulo 16
func ANSI256(n uint8) Color     // 256-color palette
func RGB(r, g, b uint8) Color   // 24-bit color
func Hex(s string) (Color, error) // "#rrggbb", "rrggbb", "#rgb", "rgb"

// TerminalDefault is the terminal's own default color. Unlike the zero
// Color it does not inherit: it resets to the default inside a colored
// parent.
var TerminalDefault Color

func (c Color) IsSet() bool
```

Named constants for the 16 colors: `Black`, `Red`, `Green`, `Yellow`,
`Blue`, `Magenta`, `Cyan`, `White`, and `BrightBlack` through
`BrightWhite`. These are package-level variables of type `Color`.

`Hex` returns `ErrInvalidColor` for any other input (wrong length,
non-hex digits, empty string). Shorthand `#abc` expands to `#aabbcc`.
Parsing is case-insensitive.

`Color` is comparable with `==`.

### 7.2 Attr

```go
type Attr uint8

const (
    Bold Attr = 1 << iota
    Faint
    Italic
    Underline
    Blink
    Reverse
    Strike
)
```

### 7.3 Style

```go
// Style is a set of visual properties. The zero value inherits everything.
type Style struct {
    FG    Color
    BG    Color
    Attrs Attr
}

// Over returns s layered on top of parent: unset colors of s are taken
// from parent, attributes are combined.
func (s Style) Over(parent Style) Style

func (s Style) IsZero() bool
```

Inheritance is this one rule and nothing else:

- `FG`: if `s.FG` is unset, use `parent.FG`. Same for `BG`.
- `Attrs`: `s.Attrs | parent.Attrs`. A child cannot remove an attribute.
- `TerminalDefault` counts as set, so it stops inheritance.

Convenience constructors, all returning `Style`: `Foreground(c)`,
`Background(c)`, and methods `WithFG(c)`, `WithBG(c)`, `WithAttrs(a)`
(adds attributes) that return a modified copy.

### 7.4 Profile

```go
type Profile uint8

const (
    ProfilePlain     Profile = iota // no escape sequences at all
    ProfileNoColor                  // attributes only
    ProfileANSI16
    ProfileANSI256
    ProfileTrueColor
)

func (p Profile) String() string
```

The zero value is `ProfilePlain`, so an uninitialized Engine emits plain
text. Values outside the range are treated as `ProfilePlain`.

### 7.5 Mode

```go
type Mode uint8

const (
    ModeReal Mode = iota // one rune is one cell; default
    ModeLong             // grapheme clusters of width 0, 1 or 2
)
```

`StringWidth(s string, mode Mode) int` (file `measure.go`) returns the
width of the widest line of `s` after the same normalization the Factory
applies with default settings (section 8.4). It exists so authors of custom
blocks can measure plain strings without building a model.

### 7.6 Edge cases

- `Style{}.Over(Style{})` is the zero style.
- `Over` is associative: `a.Over(b.Over(c)) == a.Over(b).Over(c)`. Test it
  on generated styles.
- `ANSI(17)` equals `ANSI(1)`.
- `TerminalDefault.Over` semantics: a child with `FG: TerminalDefault`
  inside a red parent ends up with the terminal default foreground.

### 7.7 Tests

- Table tests for constructors, `Hex` (valid and every invalid shape) and
  `Over`.
- Property test for associativity of `Over`.

### 7.8 Commits

1. `Add Color and Attr to dra`
2. `Add Style with inheritance to dra`
3. `Add Profile and Mode to dra`

---

## 8. Milestone 4: `dra` Model and Factory

### 8.1 Model: concept

A `Model` is styled text with exactly one slot. It has two internal forms
that share one public type; users cannot tell them apart except through
`HasSlot`.

**Leaf form**, created by the Factory:

- `prefix` and `suffix`: the text before and after the slot.
- `palette []Style`: at most 128 entries.
- `cmap []int8`: one entry per rune of `prefix` followed by one entry per
  rune of `suffix`. `-1` means "no own style". Any other value is an index
  into `palette`.
- `base Style`: applies to the whole model, including slot content.
- `slot *Model`: the content of the slot, or nil when unfilled.
- `mode Mode` and cached metrics.

**Composed form**, produced by layout operations (milestone 6): a list of
lines, each a list of spans `{text, style, mode}`. It has no slot. It
exists because merging the palettes of many models can exceed 128 styles;
the composed form stores styles directly and has no such limit. The
colormap stays the authoring format, as designed.

`Model` is a small struct passed by value. The zero `Model` is valid: it is
empty and renders nothing.

### 8.2 Model: style resolution

For a rune `i` of a leaf, with `inherited` being the style coming from the
enclosing models (zero at the top):

```
outer     = base.Over(inherited)
effective = palette[cmap[i]].Over(outer)     when cmap[i] >= 0
effective = outer                            when cmap[i] == -1
```

Slot content is resolved with `inherited = outer`. Slot content inherits
the model's base style only, never the style of the runes next to the slot.

For a composed model, each span's effective style is
`span.style.Over(inherited)`.

### 8.3 Model: text semantics

- The visible text of a leaf is `prefix + text(slot) + suffix`,
  concatenated as a flow. `'\n'` starts a new line wherever it appears,
  including inside slot content. A multi-line slot value continues on the
  prefix's last line and the suffix continues on the value's last line.
- An unfilled slot contributes nothing. A model with an unfilled slot is
  renderable; it is a reusable blank.
- Height is the number of lines: for non-empty text, the number of `'\n'`
  plus one. A trailing newline therefore adds an empty last line.
- A model with no text at all (empty prefix, empty suffix, empty or
  unfilled slot) has height 0 and width 0 and is called empty.
- Width is the width of the widest line. `'\n'` has no width.

### 8.4 Text normalization and sanitization

Applied by the Factory to `prefix` and `suffix` separately, once, at
creation. Steps in order:

1. Invalid UTF-8 bytes each become U+FFFD.
2. `"\r\n"` becomes `"\n"`.
3. `'\t'` becomes `tabWidth` spaces (default 4). Tabs are not aligned to
   tab stops: the final column is unknown until layout.
4. Sanitization, when enabled (the default). Each of these runes becomes
   U+FFFD:
   - C0 controls U+0000-U+001F except `'\n'`. This includes a lone `'\r'`
     and ESC.
   - U+007F (DEL).
   - C1 controls U+0080-U+009F.
   - Bidirectional controls U+202A-U+202E and U+2066-U+2069.
   - U+2028 and U+2029 (line and paragraph separators).

With sanitization disabled, step 4 is skipped; steps 1-3 always run.
Raw control characters then pass through to the output, are measured like
any other rune, and may desynchronize the Engine's style tracking. The doc
comment of the option must say that the caller takes responsibility.

The colormap is authored against the original text, before normalization.
Normalization maps it:

- A rune replaced one-to-one keeps its entry.
- A tab expanded to `n` spaces gives all `n` spaces the tab's entry.
- For `"\r\n"`, the `'\r'` entry is dropped and the `'\n'` entry is kept.
- An invalid byte counts as one rune in the original text.

In Long Mode, after normalization, every rune of a cluster takes the
colormap entry of the cluster's first rune. A style change inside a cluster
would split it with an escape sequence and break its rendering. This makes
the Engine mode-agnostic.

### 8.5 Metrics

Each model caches four numbers: `first` (width of the first line), `last`
(width of the last line), `max` (width of the widest line) and `height`.
They are computed at creation and at `Fill`, never at render time.

Concatenation of two pieces `a` then `b`:

```
if a.height == 0: result = b
if b.height == 0: result = a
joint  = a.last + b.first
height = a.height + b.height - 1
first  = joint if a.height == 1 else a.first
last   = joint if b.height == 1 else b.last
max    = max(a.max, b.max, joint)
```

A leaf's metrics are `concat(concat(prefix, slot), suffix)`. Widths are
measured per piece with that piece's own mode. A cluster cannot span the
boundary between prefix, slot content and suffix; a combining mark at the
start of slot content is measured on its own.

### 8.6 Model: public methods

```go
func (m Model) Width() int
func (m Model) Height() int
func (m Model) IsEmpty() bool      // Height() == 0
func (m Model) HasSlot() bool      // leaf form
func (m Model) IsFilled() bool     // leaf form with slot content
func (m Model) Mode() Mode         // leaf: its mode; composed: ModeReal
                                   // unless any span is ModeLong
func (m Model) Fill(content Model) Model
func (m Model) String() string     // plain text, no styles
```

`Fill`:

- Returns a new model with `content` in the slot. The receiver is not
  changed, so a blank can be filled many times, concurrently.
- Filling an already filled model replaces the content.
- Filling with an empty model is the same as leaving the slot unfilled,
  except that `IsFilled` reports true.
- On a model without a slot (composed form), `Fill` returns the receiver
  unchanged. Callers that need to know check `HasSlot`. Document this
  prominently.
- `content` may be of either form and either mode.
- Self-reference is impossible because models are values; no cycle check
  is needed.

`String` returns the visible text with `'\n'` between lines and without any
escape sequences. It is for tests and debugging.

### 8.7 Factory and options

```go
type Option func(*config) // config is unexported

func NewFactory(opts ...Option) *Factory
func (f *Factory) With(opts ...Option) *Factory
func (f *Factory) New(opts ...Option) (Model, error)
func (f *Factory) MustNew(opts ...Option) Model
func (f *Factory) Text(s string) Model
func (f *Factory) Styled(s string, st Style) Model
func (f *Factory) Mode() Mode
```

- `NewFactory` stores base parameters. It never fails: options are only
  validated when a model is built, because a base colormap is meaningless
  without the final text.
- `With` returns a derived factory: the receiver's options followed by the
  new ones. The receiver is unchanged.
- `New` applies the factory's options, then the call's options. Later
  options win. Then it normalizes, validates and builds a leaf model.
- `MustNew` panics on error. It is for models defined by constants in
  code. Say so in the doc comment.
- `Text(s)` is `New(WithText(s))` with any colormap from the factory
  ignored. It cannot fail, so it returns only a `Model`.
- `Styled(s, st)` is `Text(s)` with `st` layered over the factory's base
  style (`st.Over(factoryBase)`).
- A nil `*Factory` behaves like `NewFactory()`. Every method must handle a
  nil receiver.

Options:

| Option | Effect |
|--------|--------|
| `WithText(s string)` | Sets the prefix to `s` and the suffix to empty. |
| `WithFrame(prefix, suffix string)` | Sets both parts around the slot. |
| `WithPalette(styles ...Style)` | Replaces the palette. |
| `WithColormap(cmap []int8)` | Replaces the colormap. `nil` means all `-1`. |
| `WithStyle(base Style)` | Replaces the base style. |
| `WithMode(mode Mode)` | Sets the width mode. Default `ModeReal`. |
| `WithTabWidth(n int)` | Spaces per tab. Default 4. |
| `WithSanitize(on bool)` | Enables or disables step 4 of 8.4. Default on. |

Slices passed to options are copied when the option is applied.

Colormap helper:

```go
// Run describes N consecutive runes that share palette index Index.
type Run struct {
    N     int
    Index int8
}

// Runs expands runs into a colormap. Runs with N <= 0 are skipped.
func Runs(runs ...Run) []int8

// NoStyle is the colormap entry for "no own style".
const NoStyle int8 = -1
```

### 8.8 Validation in `New`

Checked against the original (pre-normalization) text:

| Condition | Error |
|-----------|-------|
| `len(palette) > 128` | `ErrPaletteSize` |
| colormap non-nil and its length differs from the rune count of prefix plus suffix | `ErrColormapLength` |
| an entry below `-1` or at or above `len(palette)` | `ErrPaletteIndex` |
| `tabWidth < 1` | `ErrTabWidth` |
| mode is not `ModeReal` or `ModeLong` | `ErrInvalidMode` |

Error messages include the offending numbers (expected and actual length,
the bad index and its position).

### 8.9 Edge cases

- `New()` with no options: an empty model with a slot.
- `WithText("")`: empty model, height 0.
- `WithText("\n")`: height 2, width 0.
- `WithFrame("[", "]")` unfilled: text `[]`, width 2.
- `WithFrame("a\n", "\nb")` filled with `"x\ny"`: lines `a`, `x`, `y`, `b`.
- An empty non-nil colormap with empty text is valid.
- A colormap given with text that contains a tab: validated against the
  original rune count, then expanded.
- A palette that is set but never referenced is valid.
- Text consisting only of sanitized characters becomes that many U+FFFD.
- `WithText` after `WithFrame` clears the suffix; `WithFrame` after
  `WithText` replaces both.
- A factory colormap combined with a call-level `WithText` of a different
  length fails with `ErrColormapLength`; this is expected and documented.
- Mutating a slice after passing it to an option does not affect the
  factory or its models.
- Long Mode with a colormap that changes inside a cluster: normalized to
  the first rune's entry, no error.

### 8.10 Tests

- Table tests for every row of 8.8 and every case of 8.9.
- Metrics: table of prefix, slot and suffix combinations against expected
  `Width` and `Height`, in both modes, including nested fills three levels
  deep.
- Property test: for random text, `Width` and `Height` computed from cached
  metrics equal the values computed by splitting `String()` and measuring
  each line.
- Fuzz test `FuzzNew`: with sanitization on, `String()` of any model never
  contains a rune from the sanitized set, and never panics.
- Immutability: fill one blank from many goroutines under `-race`.
- Slice-copy tests for `WithPalette` and `WithColormap`.

### 8.11 Commits

1. `Add text normalization and sanitization to dra`
2. `Add Model with metrics to dra`
3. `Add Factory and options to dra`
4. `Add colormap helpers to dra`

---

## 9. Milestone 5: `dra` Engine

The Direct Render API. It is the only code that produces escape sequences
from models.

### 9.1 API

```go
// Engine renders models for one color profile. The zero value renders
// plain text.
type Engine struct{ /* unexported: profile */ }

func NewEngine(p Profile) Engine
func (e Engine) Profile() Profile
func (e Engine) Append(dst []byte, m Model) []byte
func (e Engine) Render(w io.Writer, m Model) (int, error)
func (e Engine) String(m Model) string
```

- `Append` is the primitive. `Render` builds the bytes with `Append` and
  makes exactly one `Write` call; it makes no call for an empty model.
  `String` is `string(Append(nil, m))`.
- The Engine holds no state between calls and is safe for concurrent use.
- The Engine emits exactly the model: it adds no trailing newline.

### 9.2 Algorithm

Walk the model depth-first in text order. Keep `current`, the downgraded
style the terminal is in; it starts as the empty style.

For each run of text with effective style `s` (8.2):

1. `d = Downgrade(s, profile)`.
2. Split the run at `'\n'`. For each piece:
   - If the piece is non-empty and `d != current`: emit the switch to `d`
     and set `current = d`.
   - Emit the piece.
   - At each `'\n'`: if `current` is not empty, emit a reset and set
     `current` to empty. Then emit `'\n'`.
3. After the last run: if `current` is not empty, emit a reset.

Consequences that must hold:

- No escape sequence is emitted for unstyled text.
- Adjacent runs with equal downgraded styles produce one switch.
- A style never stays active across a newline, so a background color does
  not bleed to the right edge of the terminal.
- No switch is emitted before a newline or at the end for a run that has
  no visible characters on that line.
- Output with `ProfilePlain` contains no ESC byte and equals `m.String()`.
- Every output that contains a switch ends in a state equal to the empty
  style.

For a leaf, runs are maximal sequences of runes with the same colormap
entry in the prefix, then the slot content (recursively, with the inherited
style from 8.2), then runs of the suffix.

### 9.3 Reference outputs

`f` is a default factory; profile is `ProfileTrueColor` unless noted.

| Model | Bytes |
|-------|-------|
| `f.Text("hi")` | `hi` |
| `f.Styled("hi", bold+red)` | `\x1b[0;1;31mhi\x1b[0m` |
| base red, text `a\nb` | `\x1b[0;31ma\x1b[0m\n\x1b[0;31mb\x1b[0m` |
| frame `[`,`]` base red, filled with unstyled `x` | `\x1b[0;31m[x]\x1b[0m` |
| frame `[`,`]` unstyled, filled with green `x` | `[\x1b[0;32mx\x1b[0m]` |
| base red, text `a\n` | `\x1b[0;31ma\x1b[0m\n` |
| the bold red model at `ProfileNoColor` | `\x1b[0;1mhi\x1b[0m` |
| the bold red model at `ProfilePlain` | `hi` |

### 9.4 Edge cases

- Empty model: appends nothing.
- A model that is only newlines: only newlines, no escapes, even with a
  base style.
- A styled run of spaces: styled, because a background makes spaces
  visible.
- Slot content with its own base inside a styled parent: colors override,
  attributes add.
- `TerminalDefault` foreground inside a red parent: the switch for that run
  carries no foreground parameter.
- Two colors that downgrade to the same index: one switch, not two.
- Deep nesting (1000 levels of fill): renders correctly. Recursion is
  acceptable; add a test at this depth.
- Text with sanitization disabled that contains ESC: passed through
  unchanged.
- `Render` to a writer that returns an error or a short write: the error is
  returned with the byte count; short writes return `io.ErrShortWrite`.

### 9.5 Performance requirements

- `Append` on a leaf without slot content performs zero allocations when
  `dst` has enough capacity. Assert with `testing.AllocsPerRun`.
- `Append` on any model allocates nothing beyond growing `dst`.
- Benchmarks: plain text line, a line with five style runs, a three-level
  nested fill, a composed model of 50 lines.

### 9.6 Tests

- Golden tests for every row of 9.3 and every profile.
- Property test: for random models, stripping SGR sequences from the output
  equals `m.String()`, for every profile.
- Property test: the output never contains `'\n'` while a style is active
  (scan the bytes and track state).
- Fuzz test `FuzzEngine` over text, colormap and profile.
- Concurrency test under `-race`.

### 9.7 Commits

1. `Add Engine to dra`
2. `Add Engine benchmarks and allocation tests`

---

## 10. Milestone 6: `dra` layout operations

Package-level functions that take models and return a new model in the
composed form. They are what makes custom layout possible, so they must be
complete enough to build every block of milestone 8 without touching
internals.

### 10.1 General rules

- **Operations cannot fail.** They return only a `Model`.
- **Sizes are clamped.** Negative widths, heights and counts are treated
  as 0.
- **Slots are closed.** An unfilled slot is treated as empty. The result
  never has a slot.
- **Styles are preserved** and stay relative: the result still inherits
  from whatever it is later placed in. When flattening a leaf, each span
  gets `palette[i].Over(base)` (or `base`), and slot content spans get
  `span.Over(base)`.
- **Padding is unstyled.** Spaces created by an operation carry the zero
  style, so they inherit from the outer context. `Restyle` with a
  background therefore colors padding too. This is intended: it is how
  badges and filled frames work.
- **Single-line arguments.** Where an operation expects a single-line model
  (a prefix, a tail, a repeat unit), only the first line of the argument is
  used. This one rule replaces error returns.
- **Mixed modes are allowed.** Each span keeps the mode it was measured
  with; cutting and wrapping use the span's own mode.
- **No trailing-space trimming** unless an operation says so.
- Adjacent spans with equal style and mode are merged when building a
  result.

### 10.2 Types

```go
type Align uint8  // AlignLeft, AlignCenter, AlignRight
type VAlign uint8 // AlignTop, AlignMiddle, AlignBottom
type WrapMode uint8 // WrapWords, WrapAnywhere
type Axis uint8   // AxisVertical, AxisHorizontal, AxisDiagonal
```

Unknown values behave like the first constant of the type.

```go
// Gradient is a color that changes with the position inside a rectangle.
// The zero value has no stops and yields the unset Color everywhere.
type Gradient struct{ /* unexported */ }

func NewGradient(axis Axis, stops ...Color) Gradient
func (g Gradient) At(x, y, width, height int) Color
func (g Gradient) IsZero() bool // no stops
```

- Stops are spread evenly along the axis and interpolated linearly per
  RGB channel, rounding to the nearest value. The result of `At` is an
  RGB color; the Engine downgrades it like any other color, so on a
  16-color profile a gradient collapses into a few bands.
- Stops that are not RGB are converted with the tables of 5.5 when the
  gradient is created. Unset stops and `TerminalDefault` are dropped.
- No stops: `At` returns the unset `Color`. One stop: that color
  everywhere.
- Position `t` in the range 0-1: `AxisVertical` uses `y / (height-1)`,
  `AxisHorizontal` uses `x / (width-1)`, `AxisDiagonal` uses
  `(x + y) / (width-1 + height-1)`. A zero denominator gives `t = 0`.
  `t` is clamped, so coordinates outside the rectangle take the color of
  the nearest end.
- `Gradient` is an immutable value and copies the stops it is given.

Centering with an odd remainder puts the extra cell on the right
(horizontal) or at the bottom (vertical).

### 10.3 Operations

#### `Lines(m Model) []Model`

One single-line model per line of `m`. Empty model: `nil`. A line with no
text becomes a model of height 1 and width 0.

#### `Blank(width, height int) Model`

`height` lines of `width` spaces. `Blank(0, 1)` is one empty line: this is
how to produce an empty line, since text cannot express it. `height` 0
gives an empty model.

#### `Concat(ms ...Model) Model`

Flow concatenation with the text semantics of 8.3: the first line of each
model continues the last line of the previous one. Empty models are
skipped. No arguments: empty model.

#### `Stack(align Align, ms ...Model) Model`

Models placed under each other. Empty models contribute nothing. Lines are
not padded on the right. For `AlignCenter` and `AlignRight`, each line is
padded on the left relative to the widest line of the result; no right
padding is added. Result height is the sum of heights.

#### `Join(valign VAlign, ms ...Model) Model`

Models placed side by side. Empty models are skipped. Result height is the
maximum height. A shorter model is extended with empty lines according to
`valign`. Every line of every model except the last one is padded on the
right to that model's width; the last model's lines are not padded. When
all models have height 1 this equals `Concat`.

#### `Pad(m Model, width int, align Align) Model`

Pads every line to `width` cells with spaces according to `align`. Lines
already at or above `width` are unchanged; `Pad` never truncates. On an
empty model it returns an empty model (there are no lines to pad).

#### `PadHeight(m Model, height int, valign VAlign) Model`

Adds empty lines so the model has at least `height` lines. Never removes
lines. On an empty model with `height > 0` it returns `Blank(0, height)`.

#### `PrefixLines(m Model, first, rest Model) Model`

Puts `first` before the first line of `m` and `rest` before every other
line. Empty model: empty result (no prefix is emitted).

#### `SuffixLines(m Model, first, rest Model) Model`

Same on the right side. The suffix is attached directly to the end of each
line; pad first when the lines must align.

#### `Repeat(unit Model, width int) Model`

A single line that repeats `unit` to fill exactly `width` cells. If the
unit has width 0, or `width` is 0, the result is `Blank(width, 1)`. When
`width` is not a multiple of the unit width, the remainder is filled by a
truncated copy of the unit, and any cells still left (a wide cluster that
does not fit) are filled with spaces.

#### `Truncate(m Model, width int, tail Model) Model`

Cuts every line that is wider than `width`. Height is unchanged. For a line
to cut:

- If `tail` is wider than `width`: keep clusters while the total stays at
  or below `width`; no tail.
- Otherwise: keep clusters while the total stays at or below
  `width - tail.Width()`, then append `tail`.
- Cuts happen only at cluster boundaries. A wide cluster that does not fit
  is dropped, so a line can end up one cell narrower than `width`; it is
  not padded.
- `width` 0: every line becomes empty.

Lines at or below `width` are untouched and get no tail. Pass the zero
`Model` as `tail` for a hard cut.

#### `TruncateHeight(m Model, height int) Model`

Keeps the first `height` lines.

#### `Wrap(m Model, width int, mode WrapMode) Model`

Re-flows every line so that no line exceeds `width`. `width` 0 or less:
returns the model unchanged in composed form.

`WrapAnywhere`: break at cluster boundaries whenever the next cluster does
not fit.

`WrapWords`, for each source line:

1. A line at or below `width` is kept as is.
2. Otherwise split it into alternating runs of spaces (U+0020 only) and
   words (everything else). No-break space is not a break opportunity.
3. Fill output lines greedily.
   - A word that fits is appended.
   - A word that does not fit starts a new output line. Spaces before the
     break are dropped.
   - A word wider than `width` is placed at the start of a line and split
     at cluster boundaries with the `WrapAnywhere` rule; its last piece
     stays open for more content.
   - A space run that fits is appended. A space run that does not fit is
     dropped and forces a break.
4. Leading spaces of the source line are indentation of the first output
   line only. If the indentation alone is at or above `width`, it is
   dropped.
5. Trailing spaces of an output line created by a break are dropped;
   trailing spaces at the end of the source line are kept if they fit.

In both modes, a single cluster wider than `width` (a wide cluster with
`width` 1) is placed alone on its line. This is the only case where a
result line exceeds `width`; document it.

Empty source lines are preserved. Styles are preserved across breaks.

#### `Restyle(m Model, st Style) Model`

Layers `st` under the model: every span becomes `span.style.Over(st)`. The
model's own styles win; `st` fills what is unset and adds attributes.

#### `RestyleFunc(m Model, fn func(x, y int) Style) Model`

Layers a style chosen by position over the model: every cell becomes
`fn(x, y).Over(cell style)`, where `x` is the column of the cell and `y`
its line, both from 0. Unlike `Restyle`, the returned style wins over the
model's own; a zero `Style` leaves the cell unchanged. This is the
primitive for gradients and for any coloring that depends on position.

- `fn` is called once per cluster, with the column of its first cell, in
  text order. It must be a pure function: the result may be cached and
  the call order is not part of the contract.
- A nil `fn` returns the model unchanged in composed form.
- Text, width and height are never changed.
- Adjacent cells that end up with equal styles are merged into one span,
  so a function that returns few distinct styles costs few spans.

### 10.4 Edge cases

Test each of these explicitly.

- Every operation with zero arguments, with one argument, and with only
  empty models.
- Every operation with negative and zero sizes.
- `Stack` and `Join` with models of different widths and heights and each
  alignment value.
- `Join` where the tallest model is first, middle and last.
- `Join` of a model with trailing spaces: the spaces are kept.
- `Pad` to a width smaller than the content.
- `PrefixLines` with a multi-line prefix (first line used) and with an
  empty prefix.
- `PrefixLines` on a model with empty lines: the prefix is still added.
- `Truncate` where the cut lands inside a wide cluster, where the tail is
  exactly `width`, where the tail is wider than `width`, and with width 1.
- `Truncate` in Long Mode cutting an emoji ZWJ sequence: the whole cluster
  stays or goes.
- `Wrap` with a word longer than the width, with only spaces, with
  multiple consecutive spaces at a break, with leading indentation, and
  with a style change in the middle of a word.
- `Wrap` with `width` 1 and wide clusters.
- `Repeat` with a wide unit and an odd width, and with a multi-cell unit
  such as `"-="`.
- `Restyle` over a model that already has a foreground: the model's
  foreground wins, the background is added.
- An unfilled blank passed to any operation: treated as its prefix plus
  suffix.
- Operations on mixed Real and Long models.
- Results of operations used as slot content of a leaf, and rendered:
  inheritance still works.
- More than 128 distinct styles across the inputs of `Stack`: works (this
  is the reason for the composed form).
- `RestyleFunc` with a nil function, with a function that returns the zero
  style, over wide clusters in Long Mode (one call per cluster), and with
  a different style for every cell of a 40 by 20 model.
- `Gradient.At` with no stops, one stop, two and three stops, each axis,
  a 1 by 1 rectangle, and coordinates outside the rectangle.

### 10.5 Tests

- Table tests per operation, asserting `String()`, `Width()` and
  `Height()`.
- Golden tests for styled cases, rendered through the Engine.
- Property tests on random models and sizes:
  - `Truncate(m, w, tail)`: every line width is at most `w`; height
    unchanged.
  - `Wrap(m, w, WrapAnywhere)` in Real Mode: every line width is at most
    `w` for `w >= 1`; removing newlines gives the original text without
    its newlines.
  - `Wrap(m, w, WrapWords)`: the sequence of non-space runes is unchanged.
  - `Pad(m, w, a)`: every line width is at least `w` for a non-empty
    model.
  - `Stack` height is the sum of heights; `Join` height is the maximum.
  - `Join` width is the sum of widths of non-empty models.
  - `Lines` then `Stack(AlignLeft, ...)` gives the same `String()`.
- Fuzz tests for `Truncate` and `Wrap`: never panic, always terminate.
- Benchmarks for `Join`, `Wrap` and `Truncate` on a 100-line model.

### 10.6 Commits

1. `Add composed model form to dra`
2. `Add Lines, Blank and Concat operations`
3. `Add Stack and Join operations`
4. `Add Pad, PadHeight and Repeat operations`
5. `Add PrefixLines and SuffixLines operations`
6. `Add Truncate and TruncateHeight operations`
7. `Add Wrap operation`
8. `Add Restyle operation`
9. `Add Gradient and RestyleFunc operation`

---

## 11. Milestone 7: `beautify` core

### 11.1 `internal/term`

```go
// IsTerminal reports whether w is a terminal.
func IsTerminal(w io.Writer) bool

// Width returns the terminal width in cells for w.
func Width(w io.Writer) (int, bool)
```

- `IsTerminal`: true when `w` is an `*os.File` whose `Stat` mode has
  `os.ModeCharDevice` set. Any other writer, a nil file or a `Stat` error:
  false. This also reports true for `/dev/null`; that is accepted.
- `Width`: on Linux, macOS and the BSDs (build tag
  `linux || darwin || freebsd || netbsd || openbsd || dragonfly`), use the
  `TIOCGWINSZ` ioctl through the standard `syscall` package. On other
  platforms a stub returns `0, false`. A reported width of 0 is treated as
  unknown.

### 11.2 Profile detection

```go
// DetectProfile returns the color profile suitable for w, based on
// whether w is a terminal and on the environment.
func DetectProfile(w io.Writer) dra.Profile
```

Implement the decision as an unexported pure function
`detect(isTerminal bool, goos string, getenv func(string) string)` so it
can be tested without a terminal. Rules, first match wins:

1. `forced` is true when `FORCE_COLOR` or `CLICOLOR_FORCE` is set to a
   non-empty value other than `"0"`.
2. Not a terminal and not `forced`: `ProfilePlain`.
3. `TERM` is `"dumb"` and not `forced`: `ProfilePlain`.
4. `NO_COLOR` is non-empty: `ProfileNoColor`. `NO_COLOR` wins over the
   force variables.
5. On Windows, when neither `WT_SESSION` nor `TERM` is set and not
   `forced`: `ProfilePlain`.
6. `COLORTERM` is `"truecolor"` or `"24bit"` (case-insensitive):
   `ProfileTrueColor`.
7. `TERM` ends with `"-direct"`: `ProfileTrueColor`.
8. `TERM` contains `"256color"`: `ProfileANSI256`.
9. Otherwise: `ProfileANSI16`.

### 11.3 Context

```go
// Context carries what a Composable needs to render.
type Context struct{ /* unexported */ }

func NewContext(f *dra.Factory) Context
func (c Context) Factory() *dra.Factory
func (c Context) Width() (width int, bounded bool)
func (c Context) WithWidth(width int) Context
func (c Context) Unbounded() Context
```

- A new context is unbounded.
- `WithWidth` clamps negative values to 0. A bounded width of 0 is valid
  and means "no room".
- `Factory` never returns nil: the zero `Context` and a nil factory both
  yield a shared default factory.
- Unbounded is explicit, not encoded as a magic width. `Width` returns
  `0, false` when unbounded.

### 11.4 Composable

```go
// Composable is an element of high-level output.
type Composable interface {
    // Render lowers the element to a model. When ctx is bounded, the
    // returned model must not be wider than the bound.
    Render(ctx Context) (dra.Model, error)
}

// ComposableFunc adapts a function to Composable.
type ComposableFunc func(ctx Context) (dra.Model, error)
```

The width contract is the element's responsibility. Containers that lay
out children enforce it defensively (12.1).

### 11.5 Template

```go
type Direction uint8 // Inline (default), Stacked

type TemplateOption func(*templateConfig)

func NewTemplate(items ...Composable) Template
func (t Template) With(opts ...TemplateOption) Template
func (t Template) Append(items ...Composable) Template
func (t Template) Len() int
func (t Template) Render(ctx Context) (dra.Model, error)
```

Options: `WithDirection(d Direction)`, `WithSeparator(sep Composable)`,
`WithSkipEmpty(skip bool)`.

`Template` is an immutable value. `With` and `Append` return modified
copies and copy the item slice.

Rendering:

1. Render the separator once, if set, with the template's context.
2. Render items in order.
   - `Stacked`: every item gets the template's context.
   - `Inline`, unbounded: every item gets an unbounded context.
   - `Inline`, bounded: each item gets the width that is left after the
     items and separators already placed, clamped at 0.
3. With `WithSkipEmpty(true)`, items that render to an empty model are
   dropped together with their separator. Default is false: an empty item
   still gets separators around it.
4. Combine: `Inline` uses `dra.Join(dra.AlignTop, ...)`, `Stacked` uses
   `dra.Stack(dra.AlignLeft, ...)`, with the separator model between
   items, never before the first or after the last.

Errors:

- A nil item or nil separator interface value: `ErrNilComposable`, wrapped
  with the item index.
- An item's error is returned wrapped with its index:
  `fmt.Errorf("beautify: template item %d: %w", i, err)`. Rendering stops
  at the first error.

Edge cases:

- No items: empty model, no error, separator not rendered.
- One item: no separator.
- A template nested as an item of another template.
- A multi-line item in an `Inline` template: placed side by side.
- A multi-line separator in a `Stacked` template: all its lines appear.
- The zero `Template` is valid and renders an empty model.

### 11.6 Composer

```go
type ComposerOption func(*composerConfig)

func NewComposer(w io.Writer, opts ...ComposerOption) *Composer
func (c *Composer) Render(t Template) error
func (c *Composer) Profile() dra.Profile
```

Options:

| Option | Effect |
|--------|--------|
| `WithProfile(p dra.Profile)` | Fixes the profile instead of detecting it. |
| `WithWidth(n int)` | Fixes the width. `n <= 0` means unbounded. |
| `WithFactory(f *dra.Factory)` | Factory given to Composables. |
| `WithNewline(on bool)` | Whether a newline follows each render. Default on. |

That is the whole API. Do not add `Print`, `Sprintf`-style helpers or
per-call options.

Behavior:

- The profile is resolved once in `NewComposer`.
- A nil writer is replaced by `io.Discard`.
- `Render`:
  1. Resolve the width for this call: the fixed width if set; otherwise the
     terminal width if the writer is a terminal and the size is known;
     otherwise the `COLUMNS` variable if it parses as a positive integer;
     otherwise unbounded. The terminal is queried on every call so resizes
     are picked up.
  2. Build the context and call `t.Render`.
  3. On error, return it and write nothing.
  4. Append the model with the Engine into a buffer taken from a
     `sync.Pool`. With newline on, append `'\n'` when the model is not
     empty.
  5. Lock a mutex, write the buffer with one `Write` call, unlock.
  6. Return the buffer to the pool. Buffers that grew beyond 64 KiB are
     dropped instead of pooled.
- An empty model writes nothing, not even the newline.
- A write error is returned as is. A short write returns
  `io.ErrShortWrite`.
- A panic inside a Composable is not recovered. Document it.
- Rendering runs outside the lock; only the write is serialized. Output of
  concurrent `Render` calls never interleaves.

### 11.7 Tests

- `detect`: table test covering every rule and every precedence pair
  (for example `NO_COLOR` with `FORCE_COLOR`, `TERM=dumb` with
  `COLORTERM=truecolor`, Windows with and without `WT_SESSION`).
- `Context`: zero value, clamping, bounded zero versus unbounded.
- `Template`: every edge case of 11.5, both directions, width passed to
  items (use a recording `ComposableFunc`).
- `Composer`:
  - output to a `bytes.Buffer` is plain by default;
  - `WithProfile` forces styles;
  - newline on and off, and no output for an empty template;
  - nothing is written when an item fails;
  - a writer that fails and a writer that writes short;
  - exactly one `Write` call per render (counting writer);
  - 100 goroutines rendering multi-line templates under `-race`, then
    verify that every rendered block is intact in the output.
- `internal/term`: `IsTerminal` is false for a buffer, a pipe and a regular
  file.

### 11.8 Commits

1. `Add terminal detection to internal/term`
2. `Add profile detection`
3. `Add Context and Composable`
4. `Add Template`
5. `Add Composer`

---

## 12. Milestone 8: `blocks`

Every block is a constructor that returns a `beautify.Composable`, with
functional options specific to that block. Blocks use only the public API
of `beautify` and `dra`.

### 12.1 Rules for all blocks

- **Width contract.** With a bounded context, the result is never wider
  than the bound. A block that hands a narrower width to a child and gets
  back a wider model truncates it with `dra.Truncate` (hard cut, no tail).
- **No room.** With a bounded width of 0, a block returns an empty model
  and no error.
- **Insufficient room for decoration.** When the bound is smaller than the
  block's own decoration (borders, padding, guides), the block renders the
  decoration it can and truncates the result to the bound. It never
  returns an error for lack of space.
- **Nil children.** A nil child is `beautify.ErrNilComposable`, wrapped
  with the block name and child index.
- **Child errors** are wrapped with the block name and child index and
  stop rendering.
- **Styles** default to the zero style (inherit). Blocks take
  `dra.Style` values through options.
- **Models are built through `ctx.Factory()`**, so the width mode of the
  factory applies everywhere.
- **Immutability.** A block value is immutable after construction and can
  be rendered concurrently and repeatedly with different contexts.
- **Option naming.** Each block has its own option type named
  `<Block>Option` and constructors prefixed with the block name where a
  shared name would collide (`BorderStyle`, `TableBorder`). Keep names
  consistent across blocks: `...Style`, `...Gap`, `...Align`.

If a block cannot be implemented cleanly with the current `dra` API, add
the missing operation to `dra` in its own commit with tests, and record it
in section 16. Do not reach into internals.

### 12.2 Enforcing "no privileges"

`blocks/imports_test.go` parses every non-test Go file of the package with
`go/parser` and fails when an import path contains `/internal/` or is
anything from this module other than the root package and `dra`.

### 12.3 Primitives

#### `Text(s string, opts ...TextOption)`

Options: `TextStyle(dra.Style)`, `TextOverflow(Overflow)` with
`OverflowWrap` (default), `OverflowTruncate`, `OverflowKeep`,
`TextEllipsis(string)` (default `"…"`).

- Unbounded: the text as is.
- Bounded: wrap by words, or truncate each line with the ellipsis, or keep
  as is and then hard-truncate to honor the contract. `OverflowKeep` exists
  for content that must not be re-flowed; it still never exceeds the
  bound.
- Empty string: empty model.
- Multi-line input keeps its line breaks.

#### `Space(width, height int)`

A blank area. Negative values are 0. With a bound, the width is clamped to
it.

#### `Rule(opts ...RuleOption)`

A horizontal line that fills the available width. Options:
`RuleChar(string)` (default `"─"`), `RuleStyle(dra.Style)`,
`RuleWidth(int)` (width when the context is unbounded, default 40).

- Bounded: exactly the bound.
- An empty rule character falls back to the default.
- A multi-cell character is repeated with `dra.Repeat`.

#### `Model(m dra.Model)`

Wraps a prepared model. Bounded: truncates to the bound.

### 12.4 Wrappers (one child)

#### `Styled(child, st dra.Style)`

Applies `st` under the child with `dra.Restyle`.

#### `Padding(child, opts ...PaddingOption)`

Options: `PaddingAll(n)`, `PaddingX(n)`, `PaddingY(n)`,
`PaddingSides(top, right, bottom, left)`. Negative values are 0.

- The child is rendered with the bound reduced by left plus right.
- All lines are padded to the same width first, so the right padding
  forms a straight edge.
- An empty child with non-zero padding renders the padding area
  (`top + bottom` lines of `left + right` spaces); with all-zero padding it
  is empty.

#### `Border(child, opts ...BorderOption)`

Options: `BorderStyle(BorderSet)`, `BorderColor(dra.Style)`,
`BorderGradient(dra.Gradient)`, `BorderTitle(string)`, `BorderTitleAlign(dra.Align)`,
`BorderTitleStyle(dra.Style)`, `BorderCaption(string)`,
`BorderCaptionStyle(dra.Style)`, `BorderFooter(dra.Model)`,
`BorderPadding(x, y int)` (default 1, 0).

`BorderSet` is an exported struct with fields `TopLeft`, `Top`, `TopRight`,
`Left`, `Right`, `BottomLeft`, `Bottom`, `BottomRight`, `LeftJoin` and
`RightJoin` (strings). Provided sets: `BorderRounded` (default),
`BorderSquare`, `BorderDouble`, `BorderHeavy`, `BorderASCII`. Users can
pass their own.

`LeftJoin` and `RightJoin` are the cells where a horizontal divider meets
the left and the right edge. `Border` itself does not draw them; they
belong to the set so that `Sections` (12.5) and custom blocks can draw
dividers that connect to the frame. Their values in the provided sets:

| Set | `LeftJoin` | `RightJoin` |
|-----|------------|-------------|
| `BorderRounded`, `BorderSquare` | `├` | `┤` |
| `BorderDouble` | `╠` | `╣` |
| `BorderHeavy` | `┣` | `┫` |
| `BorderASCII` | `+` | `+` |

- The child gets the bound minus two border cells and the horizontal
  padding.
- The box is as wide as the content needs, not the full bound.
- The title sits in the top edge with one space on each side. A title
  wider than the top edge is truncated with an ellipsis. With no room for
  even one title cell plus its spaces, the title is omitted.
- The caption is a second label in the top edge. It is placed the way a
  right-aligned title is, whatever `BorderTitleAlign` says. The title is
  laid out first by the rules above. The caption is drawn only when it
  fits whole to the right of the title with at least one edge cell
  between them; otherwise it is omitted. It is never truncated and never
  widens the box. With a right-aligned title the caption therefore never
  appears. Newlines in the caption are replaced by spaces.
- `BorderTitleStyle` and `BorderCaptionStyle` are layered over
  `BorderColor` (`style.Over(borderColor)`) and apply to the label text
  only, not to the spaces around it. Both default to the zero style, so a
  label looks like the border unless styled.
- The footer is a label in the bottom edge. It is placed the way the
  caption is placed in the top edge: at the right end, with one space on
  each side. It is a model, not a string, so that it can carry several
  styles (a row of colored status dots, for example); only its first line
  is used. It is drawn only when it fits whole with at least one edge cell
  on each side; otherwise it is omitted. It is never truncated and never
  widens the box. `BorderColor` is layered under it with `dra.Restyle`,
  so its own styles win. An empty model means no footer.
- `BorderGradient` colors the frame by position: for every border cell
  the foreground of `BorderColor` is replaced by the gradient's color at
  that cell within the whole box, and everything else is resolved as
  before. The title and the caption therefore follow the gradient unless
  their own style sets a foreground, and so do the cells of the footer.
  The child is not touched. A gradient
  with no stops is the same as no gradient. Implement it with
  `dra.RestyleFunc` on the cells the block knows to be frame cells.
- An empty child renders a box with an empty interior: two lines.
- Each field of a custom `BorderSet` is expected to be one cell wide. An
  empty field is replaced by a space. Wider fields are truncated to one
  cell.
- With a bound below 2, apply the "insufficient room" rule of 12.1.

#### `Align(child, h dra.Align)`

- Bounded: pads every line to the bound.
- Unbounded: aligns lines relative to the child's own width.

#### `Sized(child, opts ...SizedOption)`

Options: `MinWidth(n)`, `MaxWidth(n)`, `Width(n)` (both), `MinHeight(n)`,
`MaxHeight(n)`, `Height(n)`.

- The child is rendered with a bound of `min(context bound, MaxWidth)`.
- The result is padded up to the minimum and truncated down to the
  maximum, then clamped to the context bound, which always wins.
- When the minimum exceeds the maximum, the maximum wins.

#### `Badge(text string, st dra.Style)`

The text with one space on each side, with `st` applied to the whole
badge. Newlines in the text are replaced by spaces. Bounded: truncated with
an ellipsis inside the padding.

### 12.5 Containers

#### `Row(children []Cell, opts ...RowOption)`

Children placed side by side.

```go
func Fixed(n int, c beautify.Composable) Cell   // exactly n cells
func Auto(c beautify.Composable) Cell           // as wide as its content
func Flex(weight int, c beautify.Composable) Cell // a share of what is left
```

Options: `RowGap(n int)` (default 0), `RowAlign(dra.VAlign)`.

Width algorithm:

- Unbounded: `Fixed` cells get their width; `Auto` and `Flex` cells render
  unbounded.
- Bounded, with `avail` being the bound minus the gaps between cells:
  1. `Fixed` cells take their width, in order, each clamped to what is
     left.
  2. `Auto` cells render in order, each with what is left as its bound,
     and take their rendered width.
  3. `Flex` cells share what is left in proportion to their weights with
     integer division; the remainder is handed out one cell at a time from
     left to right. Weights below 1 are treated as 1.
- A cell whose allotted width is 0 is omitted together with its gap.
- `Fixed` and `Flex` cells are padded to their allotted width so columns
  line up; `Auto` cells are not padded.
- Every cell is truncated to its allotted width.
- If the gaps alone exceed the bound, cells are dropped from the right
  until the row fits.
- No children: empty model.

#### `Column(children []beautify.Composable, opts ...ColumnOption)`

Children stacked. Options: `ColumnGap(n int)` (empty lines between
children), `ColumnAlign(dra.Align)`. Every child gets the context's bound.
Empty children are skipped and do not produce a gap.

#### `KeyValue(pairs []Pair, opts ...KeyValueOption)`

`Pair` is `{Key string; Value beautify.Composable}`. Options:
`KeyValueSeparator(string)` (default `": "`), `KeyStyle(dra.Style)`,
`KeyAlign(dra.Align)`.

- Keys are padded to the widest key so values line up.
- A multi-line value hangs: its continuation lines are indented to the
  value column.
- The value gets the bound minus the key column and separator. When that
  leaves less than 1 cell, the value moves to the next line, indented by
  two spaces, and gets the bound minus 2.
- A nil value renders as an empty value, not an error: this block is a
  convenience for data display.
- Newlines in keys are replaced by spaces.

#### `List(items []beautify.Composable, opts ...ListOption)`

Options: `ListBullet(string)` (default `"•"`), `ListOrdered()`,
`ListStart(n int)` (default 1), `ListMarkerStyle(dra.Style)`,
`ListGap(n int)`.

- Unordered: the bullet, one space, then the item.
- Ordered: `1.`, `2.` and so on, right-aligned to the widest number.
- A multi-line item hangs under its first line.
- The item gets the bound minus the marker column.

#### `Tree(root Node, opts ...TreeOption)`

```go
type Node struct {
    Label    beautify.Composable
    Children []Node
}
```

Options: `TreeGuides(GuideSet)`, `TreeGuideStyle(dra.Style)`,
`TreeMaxDepth(n int)`.

`GuideSet` is an exported struct with fields `Branch` (default `"├── "`),
`Last` (`"└── "`), `Pipe` (`"│   "`) and `Space` (`"    "`). Provided:
`GuidesUnicode` (default), `GuidesASCII` (`"|-- "`, `` "`-- " ``,
`"|   "`, `"    "`), `GuidesRounded` (`"╰── "` for `Last`).

- The root label has no guide.
- A child's first line gets `Branch`, or `Last` for the last child. Its
  following lines and its whole subtree get `Pipe`, or `Space` under the
  last child. Build this with `dra.PrefixLines`.
- A label gets the bound minus the indentation at its depth, with a
  minimum of 1.
- `TreeMaxDepth(n)` with `n > 0`: nodes deeper than `n` are replaced by a
  single `…` child. The root is depth 0. `n <= 0` means no limit.
- A nil label renders as an empty line, so the structure stays visible.
- A label that renders to an empty model also occupies one line.
- The fields of a custom `GuideSet` should have equal widths; if they do
  not, each is padded to the widest.
- Recursion depth equals tree depth. Add a test with depth 10000 built as
  a chain to confirm there is no quadratic blowup; if recursion is a
  problem at that depth, switch to an explicit stack.

#### `Sections(sections []Section, opts ...SectionsOption)`

One frame around several children, with a divider between neighbors that
connects to the frame. Each child can have a title.

```
╭─ System ──────── host ─╮
│ OS      Ubuntu 24.04   │
├─ Hardware ─────────────┤
│ CPU     Ryzen 7 5800X  │
╰────────────────────────╯
```

`Section` is `{Title string; Body beautify.Composable}`. Options:
`SectionsBorder(BorderSet)` (default `BorderRounded`),
`SectionsBorderStyle(dra.Style)`,
`SectionsBorderGradient(dra.Gradient)`, `SectionsTitleAlign(dra.Align)`,
`SectionsTitleStyle(dra.Style)`, `SectionsCaption(string)`,
`SectionsCaptionStyle(dra.Style)`, `SectionsFooter(dra.Model)`,
`SectionsPadding(x, y int)` (default 1, 0).

- The title of the first section sits in the top edge. The title of every
  other section sits in the divider above it. Titles follow the title
  rules of `Border`: one space on each side, truncated with an ellipsis,
  omitted when there is no room, aligned by `SectionsTitleAlign`. A
  section with an empty title gets a plain edge or divider.
- A divider is `LeftJoin`, then the `Top` cell repeated across the inner
  width including the horizontal padding, then `RightJoin`.
- The caption belongs to the top edge only and follows the caption rules
  of `Border`. The footer follows the footer rules of `Border`.
- `SectionsTitleStyle` and `SectionsCaptionStyle` are layered over
  `SectionsBorderStyle`, as in `Border`.
- `SectionsBorderGradient` follows the rules of `BorderGradient`. The
  dividers and their join cells are frame cells.
- Every body gets the bound minus two border cells and the horizontal
  padding. The box is as wide as the widest body needs, not the full
  bound, and every body is padded to that width so the right edge is
  straight. Titles and the caption never widen the box.
- The vertical padding is applied inside every section, above and below
  its body.
- A body that renders to an empty model contributes no interior lines:
  its edge or divider is still drawn.
- No sections: empty model. One section renders exactly what `Border`
  renders for the same child and the equivalent options.
- A nil body is `beautify.ErrNilComposable` (12.1).
- The fields of a custom `BorderSet` follow the rules of `Border`: an
  empty field, `LeftJoin` and `RightJoin` included, is replaced by a
  space, and a wider field is truncated to one cell.
- With a bound below 2, apply the "insufficient room" rule of 12.1.

#### `Table(opts ...TableOption)`

Options: `TableHeaders(cells ...beautify.Composable)`,
`TableRow(cells ...beautify.Composable)` (repeatable),
`TableRows(rows [][]beautify.Composable)`,
`TableColumns(specs ...ColumnSpec)`, `TableBorder(TableBorderSet)`,
`TableBorderStyle(dra.Style)`, `TableHeaderStyle(dra.Style)`,
`TablePadding(n int)` (horizontal cell padding, default 1).

`ColumnSpec` is `{Width int; MinWidth int; MaxWidth int; Align dra.Align}`
where a zero `Width` means "fit content".

Provided border sets: `TableRounded` (default), `TableSquare`,
`TableASCII`, `TableMinimal` (a rule under the header, columns separated
by two spaces, no outer frame), `TableNone`.

Layout:

1. The column count is the maximum cell count over the headers and all
   rows. Missing cells are empty. Missing column specs are the zero spec.
2. Natural width of a column: its fixed `Width`, or else the widest cell
   rendered unbounded, clamped to `MinWidth` and `MaxWidth`.
3. Unbounded context: use natural widths.
4. Bounded: `avail` is the bound minus borders and padding. If the natural
   widths fit, use them. Otherwise shrink: repeatedly take one cell from
   the widest shrinkable column (leftmost on ties) until the total fits. A
   column is shrinkable while it is above `max(MinWidth, 1)` and has no
   fixed `Width`. If no column can shrink further, stop; the table is then
   truncated to the bound (12.1).
5. Render every cell again with its final column width as the bound.
6. Row height is the tallest cell in the row; shorter cells are padded at
   the bottom. Cells are aligned per the column spec.

Edge cases:

- No headers and no rows: empty model.
- Headers only: header row and borders.
- Rows only: no header separator.
- Rows of different lengths.
- A nil cell is an empty cell.
- Zero columns after counting: empty model.
- Multi-line cells and cells that wrap after shrinking.
- More column specs than columns: extra specs are ignored.

### 12.6 Tests

For every block:

- Golden tests for the default look and each option, at `ProfilePlain` and
  one styled profile.
- Width contract: render with bounds 0, 1, 2, 3, 5, 10, 20, 40 and
  unbounded; assert `Width() <= bound` every time. Run this as one shared
  table-driven helper over all blocks, with children that include long
  words, wide runes and multi-line text.
- Empty input, nil children, failing children.
- Long Mode: a block with CJK and emoji content keeps its right edge
  straight (assert all lines have equal width where the block promises a
  rectangle).
- Concurrency: render the same block value from several goroutines.

Composition tests:

- A tree inside a border inside a row cell.
- A table whose cells contain lists.
- A border inside a border with a title.
- `Sections` with one section against `Border` with the same child and
  options: the output is byte-for-byte equal, for every provided
  `BorderSet`.
- A template of blocks rendered by a `Composer` at a fixed width.

### 12.7 Commits

One commit per block, in this order, each with its tests and golden files:

1. `Add import guard test for blocks`
2. `Add Text block`
3. `Add Space, Rule and Model blocks`
4. `Add Styled block`
5. `Add Padding block`
6. `Add Border block`
7. `Add Align and Sized blocks`
8. `Add Badge block`
9. `Add Row block`
10. `Add Column block`
11. `Add KeyValue block`
12. `Add List block`
13. `Add Tree block`
14. `Add Table block`
15. `Add Sections block`
16. `Add block composition tests`

---

## 13. Milestone 9: documentation and examples

- Package comments in every `doc.go` that explain the level, the main
  types and one short example.
- `example_test.go` files with runnable examples that have `// Output:`
  blocks. Use `ProfilePlain` so the output is stable:
  - `dra`: building a model with a palette and colormap; filling a blank;
    `Stack`, `Join`, `Wrap`, `Truncate`.
  - `beautify`: a `Composer` rendering a template; a custom Composable
    written from scratch with `dra` operations.
  - `blocks`: one example per container block.
- Extend `README.md`: keep its current first lines unchanged and add
  installation, a short quick start, an overview of the two levels, Real
  Mode versus Long Mode with their trade-offs, a "write your own block"
  section, and a link to `goals.md`.
- Check that `go doc` output of each package reads well top to bottom.

Commits:

1. `Add package documentation and examples for dra`
2. `Add package documentation and examples for beautify`
3. `Add examples for blocks`
4. `Document usage in README`

---

## 14. Cross-cutting edge-case matrix

Each row must be covered by at least one test in every layer it applies to.

| Case | dra | Engine | ops | Template | blocks |
|------|-----|--------|-----|----------|--------|
| Empty input | x | x | x | x | x |
| Only newlines | x | x | x |   | x |
| Trailing newline | x | x | x |   | x |
| Unfilled slot | x | x | x |   |   |
| Nested fill, 3+ levels | x | x | x |   |   |
| Style inheritance through a slot | x | x | x |   | x |
| `TerminalDefault` inside a colored parent | x | x |   |   |   |
| Background across a newline |   | x | x |   | x |
| Tab characters | x |   | x |   | x |
| CR and CRLF | x |   |   |   | x |
| ESC and other control characters | x | x |   |   | x |
| Invalid UTF-8 | x | x | x |   | x |
| Combining marks | x |   | x |   |   |
| CJK wide runes | x |   | x |   | x |
| Emoji ZWJ sequences | x |   | x |   | x |
| Flags (regional indicators) | x |   | x |   |   |
| Real and Long models mixed | x | x | x |   | x |
| Width 0 bound |   |   | x | x | x |
| Width 1 bound |   |   | x | x | x |
| Negative sizes |   |   | x | x | x |
| Unbounded width |   |   |   | x | x |
| More than 128 distinct styles | x | x | x |   |   |
| Nil Composable |   |   |   | x | x |
| Failing Composable |   |   |   | x | x |
| Concurrent use | x | x |   | x | x |
| Every color profile |   | x |   |   | x |

---

## 15. Final verification checklist

Run all of these before declaring the work complete.

- [ ] `gofmt -l .` prints nothing.
- [ ] `go vet ./...` is clean.
- [ ] `go build ./...` succeeds.
- [ ] `go test -race ./...` passes.
- [ ] `go test -run=^$ -fuzz=<name> -fuzztime=30s` passes for every fuzz
      target.
- [ ] `go test -cover ./...` meets the thresholds of section 2.
- [ ] `go test -bench=. -benchmem ./dra/` shows zero allocations for
      `Engine.Append` on a leaf model.
- [ ] `GOOS=windows go build ./...` and `GOOS=darwin go build ./...`
      succeed.
- [ ] `go.mod` has no `require` entries.
- [ ] `blocks` imports nothing from `internal/` (the guard test passes).
- [ ] Every exported identifier has a doc comment.
- [ ] Every row of section 14 is covered.
- [ ] A manual run of a small program that prints a bordered table, a tree
      and a key-value block looks correct in a real terminal in both width
      modes, and when piped to a file contains no escape sequences.
- [ ] `git log` shows atomic commits that each build and pass tests.
- [ ] Any `dra` API added beyond this plan is recorded in section 16.

---

## 16. Decision log

Decisions already made. Do not revisit them without the maintainer.

| Decision | Reason |
|----------|--------|
| Slot is the gap between `prefix` and `suffix`, not a marker in the text | Nothing to escape. |
| Colormap entries are palette indexes; `-1` is "no own style" | `int8` cannot hold colors, and indexes make restyling cheap. |
| Composed models store styles directly instead of a palette | Merging palettes can exceed 128 styles. |
| `Fill` is lazy and keeps the child by reference | No palette merging or copying on the hot path. |
| Slot content inherits only the model's base style | One rule; the style next to the slot is an accident of position. |
| Attributes are additive | Avoids tri-state attributes. |
| Layout operations cannot fail; single-line arguments use their first line | Blocks stay free of error plumbing. |
| `Fill` on a slotless model returns it unchanged | Same reason; `HasSlot` is available. |
| Every style switch is a full reset plus the new style | Deterministic; SGR cannot unset bold and faint separately. |
| Styles are reset before each newline | Prevents background bleed. |
| Text is sanitized at model creation, on by default | Prevents escape injection and keeps the Engine fast. |
| Real Mode counts one rune as one cell | Speed and simplicity; documented trade-off. |
| Long Mode uses simplified tables | Zero dependencies and no generator in v1. |
| The engine type is `dra.Engine` | Avoids the stutter of `dra.DRA`; the package name carries the term. |
| Width detection uses `syscall` behind build tags | Zero dependencies. |
| `Composer.Render` takes a `Template` | The template is the unit of high-level rendering. |
| A newline follows each render by default | Output is line and block oriented. |

Record additions below this line while implementing: the date, what was
added or changed, and why.

- 2026-11-13. `Border` gets `BorderCaption`, `BorderCaptionStyle` and
  `BorderTitleStyle`. A box often needs two labels in its top edge (a
  name and a status, a section and an owner), and a label that must stand
  out from the border line. Neither can be built from outside the block:
  a child cannot draw on the border cells.
- 2026-11-13. `dra` gets `Gradient`, `Axis` and the `RestyleFunc`
  operation. Coloring by position (gradients on frames and on ASCII art)
  could not be expressed: `Restyle` applies one style to the whole model.
  `RestyleFunc` is the general primitive and `Gradient` is the common
  case on top of it, so custom blocks get the same capability as the
  standard ones.
- 2026-11-13. `Border` and `Sections` get a gradient option. A caller
  cannot apply `RestyleFunc` to the frame alone from outside: it would
  have to know which cells of the rendered box belong to the frame, which
  depends on the padding, the titles and the dividers.
- 2026-11-13. `Border` and `Sections` get a footer: a label in the bottom
  edge. It takes a `dra.Model` instead of a string because the motivating
  content, a row of differently colored marks, needs more than one style.
- 2026-11-13. `BorderSet` gets `LeftJoin` and `RightJoin`, and `blocks`
  gets the `Sections` container. A frame split by dividers that connect
  to it cannot be composed from `Border`: the divider has to cover the
  border cells and the padding, which only the block that draws the frame
  can reach. The join cells live in `BorderSet` so that every provided
  set, and every custom one, has matching dividers.
- 2026-11-25. Decisions made during the implementation of v1. No public
  `dra` API was added beyond this plan.
  - `internal/width` gets `Separable`. The builder of composed models merges
    equal neighboring spans only when that keeps the clusters at the
    boundary apart; otherwise a merge could change a width in Long Mode.
    `internal/ansi` gets `ToRGB` for the stops of `Gradient`.
  - In `internal/width` an invalid byte and a newline never take part in a
    longer cluster, not even after a ZWJ, and a ZWJ without a base does not
    attach the rune after it.
  - `Wrap` with `WrapWords` drops the indentation when the first word does
    not fit after it, instead of emitting an empty line. A zero-width cluster
    never causes a break.
  - A nil `ComposableFunc` returns `ErrNilComposable` instead of panicking.
  - A bounded `Template` cuts its result to the bound, because separators
    alone can exceed it. The separator is rendered, and can fail, whenever
    there is at least one item.
  - A `Border` title keeps one edge cell next to the corner when there is
    room. Its text can take the inner width minus its two spaces. Vertical
    padding is not drawn around an empty child.
  - The cells of a footer under a gradient are restyled one by one, because
    `RestyleFunc` would override the footer's own colors.
  - Table border sets without vertical lines get no special padding: the two
    spaces between the columns of `TableMinimal` are the default padding of
    both neighbors, so the outer cells have a margin of one space. A table
    with headers and no rows has no line under the header.
  - `Row` limits `Flex` weights to 1,000,000 and does not hand the gaps of
    omitted cells to other cells. Horizontal decoration larger than the bound
    is limited to it.
  - Every test package imports `internal/golden`, so that
    `go test ./... -update` works, and `.gitattributes` protects golden files
    from line-ending conversion.
