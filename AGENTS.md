# beautify

A Go library for building beautiful terminal output: styled text, frames,
tables, trees and custom layouts. Module path: `github.com/x-chunk/beautify`.

Read `plan.md` for the implementation plan and `goals.md` for deferred work
and deliberate non-goals before changing the design.

## Architecture

The library has two levels and three public packages. Imports only go
downward; a lower package never imports a higher one.

```
beautify/          Composer, Composable, Template, Context   high level
beautify/blocks/   standard blocks built on the public API   high level
beautify/dra/      Model, Style, Factory, Engine, layout ops low level
beautify/internal/ ANSI encoding, width tables, terminal     details
```

Render pipeline:

```
Composer.Render(Template)
  -> each Composable.Render(ctx) -> dra.Model
     -> dra.Engine (the Direct Render API) -> bytes -> io.Writer
```

Core concepts:

- `dra.Model` is the universal base model for everything that is rendered.
  It is styled text with exactly one slot (placeholder), a palette and a
  colormap (`[]int8` indexes into the palette). It knows nothing about
  layout. It is not a `Composable` and is rendered only through the DRA.
- `dra.Factory` is the only way to create a `Model`. It holds base
  parameters and overrides them through functional options.
- `dra.Engine` is the DRA: a pure function from a `Model` and a color
  profile to bytes.
- `beautify.Composable` is the interface of every high-level element. Its
  `Render` method lowers the element to a `Model`.
- `beautify.Template` is an ordered set of Composables and is the unit the
  `Composer` renders. A Template is itself a Composable, so templates nest.
- `beautify.Composer` renders Templates to an `io.Writer`.

## Design rules

- Keep `Composer` minimal: create, render a Template, a few options. New
  capabilities go into `dra` or `blocks`, not into `Composer`. Extend its
  API only when there is no other place for the feature.
- Make `dra` as powerful as needed. It is the layer where layout and
  rendering capability lives, and its API may grow.
- Standard blocks have no privileges. Everything in `blocks` is written
  only against the public API of `beautify` and `dra`, the same way a user
  writes a custom block. `blocks` must never import `internal/`. If a block
  needs something that is not public, that is a gap in `dra`: close it in
  `dra` for everyone instead of working around it.
- `Model` fields stay unexported. Invariants are validated once in
  `Factory.New`; the Engine trusts every `Model` and never validates.
- `Model`, `Factory`, `Style` and `Engine` are immutable and safe for
  concurrent use. Operations return new values and never mutate inputs.
- The Engine is pure and stateless: the same `Model` and profile always
  produce the same bytes. It never detects the terminal and never reads the
  environment. Detection belongs to `Composer`.
- Only `Composer` writes to the `io.Writer`.
- Style inheritance has one rule with no exceptions: an unset color means
  "same as the parent", and attributes are only added, never removed.
- Text that enters a `Model` is sanitized by default so external input
  cannot inject escape sequences.
- No third-party dependencies. Only the Go standard library.
- The public API must stay simple and predictable. Prefer one obvious way
  to do a thing over several configurable ones.
- Width modes: Real Mode (default, one rune is one cell) and Long Mode
  (grapheme clusters with widths 0, 1 or 2). The `Model` layout is the same
  in both modes.

## Non-goals

These are deliberate decisions, not omissions. Do not add them without an
explicit request from the maintainer.

- Overlaying blocks on top of each other and absolute positioning.
- A flexbox-style layout engine (shrink priorities, min-content sizing).
  Fixed sizes, content sizes and weighted shares are enough for now.
- A `log/slog` adapter. Logging is one possible use case, not the purpose
  of the library, and the adapter is out of scope for now.

## Working in this repository

- Format with `gofmt`. Run `go vet ./...` and `go test -race ./...` before
  every commit.
- Rendering tests compare against golden files in `testdata/`. Regenerate
  them with `go test ./... -update` and review the diff before committing.
- Every new layout operation or block comes with tests for its edge cases:
  empty input, zero and negative widths, multi-line content, wide runes and
  unfilled slots.

## Code comments and documentation

- Write all code comments in American English, in plain but technical language.
- Write comments the way an engineer writes them in a production codebase: short, direct and specific. No filler, no chatty tone, no narrating the obvious.
- Explain why, not what. Use comments for intent, constraints, invariants, edge cases, side effects and non-obvious trade-offs.
- Code must be well documented by default. Every exported or public package, type, function, method, constant and variable gets a doc comment that follows the language's convention (for example, in Go the comment starts with the identifier name).
- In doc comments, describe parameters, return values, returned errors and, where relevant, concurrency, ownership and lifetime expectations.
- Keep comments in sync with the code. Update or delete a stale comment in the same change that makes it stale.
- Do not leave commented-out code. Do not leave a TODO without context on what is missing and why.

## Commits

- Make every change as atomic commits: one logical change per commit.
- Each commit must leave the project in a working state (it builds and tests pass).
- Do not mix unrelated changes in one commit. Keep features, bug fixes, refactors, formatting and dependency updates in separate commits.
- Write commit messages in American English, imperative mood, with a concise subject line (72 characters or fewer). Add a body only when the reason for the change is not obvious from the diff.
