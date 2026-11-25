// Package beautify builds beautiful terminal output: styled text, frames,
// tables, trees and custom layouts.
//
// This package is the high level of the library. A [Composable] is an element
// of output, a [Template] is an ordered set of them, and a [Composer] renders
// templates to a writer:
//
//	c := beautify.NewComposer(os.Stdout)
//	err := c.Render(beautify.NewTemplate(
//		blocks.Border(blocks.Text("hello"), blocks.BorderTitle("greeting")),
//	))
//
// Standard elements are in package blocks. They have no privileges: a custom
// element implements [Composable] with the operations of package dra, the
// low level, and is used the same way. [ComposableFunc] turns a function into
// an element.
//
// The [Composer] detects the color profile and the width of its terminal.
// Output to anything else is plain text without escape sequences, unless the
// environment or an option asks for color.
package beautify
