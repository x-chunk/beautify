// Package dra is the low level of beautify: the Direct Render API. It holds
// styles, models, layout operations on models, and the engine that turns a
// model into bytes.
//
// # Models
//
// A [Model] is styled text. A [Factory] creates models; it is the only way
// to turn a string into one:
//
//	f := dra.NewFactory()
//	m := f.Styled("done", dra.Foreground(dra.Green).WithAttrs(dra.Bold))
//
// A model made by a factory has one slot between two pieces of text.
// [Model.Fill] puts another model there, so a blank can be built once and
// filled many times:
//
//	blank := f.MustNew(dra.WithFrame("[", "]"), dra.WithStyle(dra.Foreground(dra.Blue)))
//	fmt.Println(dra.NewEngine(dra.ProfileANSI16).String(blank.Fill(m)))
//
// Parts of the text can have different styles: [WithPalette] lists them and
// [WithColormap] assigns one to every rune.
//
// # Styles
//
// A [Style] has a foreground, a background and attributes. Inheritance has
// one rule: an unset color means "same as the parent", and attributes are
// only added. [TerminalDefault] is a color that resets to the default of the
// terminal instead of inheriting.
//
// # Layout
//
// Operations such as [Stack], [Join], [Pad], [Wrap], [Truncate] and [Restyle]
// take models and return a new one. They cannot fail: sizes are clamped, and
// an unfilled slot counts as empty. Their results have no slot.
//
// # Rendering
//
// An [Engine] renders a model for one color [Profile]. Colors that the
// profile cannot show are converted to the nearest one it can.
//
// # Width modes
//
// [ModeReal] counts every rune as one cell. [ModeLong] measures grapheme
// clusters, so CJK text, emoji and combining marks line up, at the cost of
// table lookups. The mode belongs to the factory; models of both modes can
// be mixed.
//
// All types are immutable and safe for concurrent use. Text is sanitized by
// default, so external input cannot inject escape sequences.
package dra
