// Package blocks is a library of standard elements for package beautify.
//
// Every block is a constructor that returns a beautify.Composable and takes
// options of its own type:
//
//	blocks.Border(
//		blocks.Text("hello"),
//		blocks.BorderTitle("greeting"),
//		blocks.BorderColor(dra.Foreground(dra.Blue)),
//	)
//
// The primitives are [Text], [Space], [Rule] and [Model]. [Styled], [Padding],
// [Border], [Align], [Sized] and [Badge] decorate a single element. [Row],
// [Column], [KeyValue], [List], [Tree], [Table] and [Sections] arrange
// several.
//
// All blocks follow the same rules. In a bounded context the result is never
// wider than the bound, and a block without any room renders nothing; lack of
// space is never an error. Styles default to inheriting from the enclosing
// element. A nil child is reported as beautify.ErrNilComposable unless the
// block says otherwise. Block values are immutable and can be rendered
// repeatedly and concurrently, provided that their children can.
//
// Blocks use only the public API of beautify and dra, so anything they do is
// also possible in a custom element.
package blocks
