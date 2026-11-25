package dra

import "strings"

// Model is styled text with at most one slot. It is the universal input of
// the Engine and of the layout operations.
//
// A Model made by a Factory has a slot that Fill can put another model in.
// A Model made by a layout operation has none. The two behave the same
// otherwise.
//
// Text is a flow: '\n' starts a new line wherever it appears, including
// inside slot content. Multi-line slot content continues the last line of
// the text before the slot, and the text after the slot continues the last
// line of the content.
//
// Model is an immutable value and is safe for concurrent use. The zero Model
// is empty and renders nothing.
type Model struct {
	leaf *leaf  // nil for models without a slot
	slot *Model // content of the slot, nil when unfilled
	size metrics
	mode Mode
}

// leaf is the part of a model that Fill shares between its results.
type leaf struct {
	prefix, suffix string  // text before and after the slot
	palette        []Style // at most 128 entries
	cmap           []int8  // one entry per rune of prefix+suffix, nil for all NoStyle
	base           Style   // applies to the whole model, slot content included
	mode           Mode

	// Derived by newLeaf.
	prefixRunes            int // where the entries of suffix start in cmap
	prefixSize, suffixSize metrics
}

// newLeaf builds an unfilled model from normalized, validated parts. It
// takes ownership of the slices in l.
func newLeaf(l leaf) Model {
	for range l.prefix {
		l.prefixRunes++
	}
	l.prefixSize = measure(l.prefix, l.mode)
	l.suffixSize = measure(l.suffix, l.mode)
	return Model{leaf: &l, size: l.prefixSize.concat(l.suffixSize), mode: l.mode}
}

// Width returns the width in cells of the widest line.
func (m Model) Width() int {
	return m.size.max
}

// Height returns the number of lines. A trailing newline adds an empty last
// line. Height is 0 only for an empty model.
func (m Model) Height() int {
	return m.size.height
}

// IsEmpty reports whether the model has no text at all. A model of one
// empty line is not empty.
func (m Model) IsEmpty() bool {
	return m.size.height == 0
}

// HasSlot reports whether Fill has an effect on the model. It is true for
// models made by a Factory and false for results of layout operations.
func (m Model) HasSlot() bool {
	return m.leaf != nil
}

// IsFilled reports whether the model has a slot and the slot was filled,
// even with an empty model.
func (m Model) IsFilled() bool {
	return m.slot != nil
}

// Mode returns the width mode of the model itself; slot content can have
// another mode. For a result of a layout operation it is ModeLong when any
// part of the text was measured in ModeLong, and ModeReal otherwise.
func (m Model) Mode() Mode {
	return m.mode
}

// Fill returns a copy of the model with content in the slot, replacing any
// previous content. The receiver is unchanged, so one blank can be filled
// many times, also concurrently.
//
// A model without a slot is returned unchanged and content is discarded.
// Use HasSlot to tell the cases apart.
func (m Model) Fill(content Model) Model {
	if m.leaf == nil {
		return m
	}
	m.slot = &content
	m.size = m.leaf.prefixSize.concat(content.size).concat(m.leaf.suffixSize)
	return m
}

// String returns the visible text without styles, with '\n' between lines.
// It is meant for tests and debugging; use Engine to render.
func (m Model) String() string {
	var b strings.Builder
	m.writeText(&b)
	return b.String()
}

func (m *Model) writeText(b *strings.Builder) {
	if m.leaf == nil {
		return
	}
	b.WriteString(m.leaf.prefix)
	if m.slot != nil {
		m.slot.writeText(b)
	}
	b.WriteString(m.leaf.suffix)
}
