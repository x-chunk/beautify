package blocks

import (
	"slices"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

// Section is one part of Sections. An empty title gives a plain edge.
type Section struct {
	Title string
	Body  beautify.Composable
}

// SectionsOption configures Sections.
type SectionsOption func(*sections)

// SectionsBorder sets the characters of the frame and the dividers. The
// default is BorderRounded.
func SectionsBorder(set BorderSet) SectionsOption {
	return func(s *sections) { s.set = set }
}

// SectionsBorderStyle sets the style of the frame and the dividers.
func SectionsBorderStyle(st dra.Style) SectionsOption {
	return func(s *sections) { s.style = st }
}

// SectionsBorderGradient colors the frame and the dividers by position, by
// the rules of BorderGradient.
func SectionsBorderGradient(g dra.Gradient) SectionsOption {
	return func(s *sections) { s.gradient = g }
}

// SectionsTitleAlign sets where in its edge a title sits. The default is the
// left end.
func SectionsTitleAlign(a dra.Align) SectionsOption {
	return func(s *sections) { s.titleAlign = a }
}

// SectionsTitleStyle sets the style of the titles, layered over
// SectionsBorderStyle.
func SectionsTitleStyle(st dra.Style) SectionsOption {
	return func(s *sections) { s.titleStyle = st }
}

// SectionsCaption sets a label at the right end of the top edge, by the
// rules of BorderCaption.
func SectionsCaption(text string) SectionsOption {
	return func(s *sections) { s.caption = text }
}

// SectionsCaptionStyle sets the style of the caption, layered over
// SectionsBorderStyle.
func SectionsCaptionStyle(st dra.Style) SectionsOption {
	return func(s *sections) { s.captionStyle = st }
}

// SectionsFooter sets a label at the right end of the bottom edge, by the
// rules of BorderFooter.
func SectionsFooter(m dra.Model) SectionsOption {
	return func(s *sections) { s.footer = m }
}

// SectionsPadding sets the blank space around every body. The default is 1
// cell on the left and the right and none above and below. Negative values
// count as 0.
func SectionsPadding(x, y int) SectionsOption {
	return func(s *sections) { s.padX, s.padY = max(x, 0), max(y, 0) }
}

type sections struct {
	frame
	list []Section
}

// Sections draws one frame around several bodies, with dividers between
// neighbors that connect to the frame. The title of the first section sits
// in the top edge, every other title in the divider above its section. The
// box is as wide as the widest body needs; labels never widen it.
func Sections(list []Section, opts ...SectionsOption) beautify.Composable {
	s := sections{frame: defaultFrame(), list: slices.Clone(list)}
	applyOptions(&s, opts)
	return s
}

// Render implements beautify.Composable.
func (s sections) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) || len(s.list) == 0 {
		return dra.Model{}, nil
	}
	s.padX = limit(ctx, s.padX)
	inner := s.inner(ctx)
	parts := make([]framePart, len(s.list))
	for i, section := range s.list {
		body, err := renderChild(inner, "sections", i, section.Body)
		if err != nil {
			return dra.Model{}, err
		}
		parts[i] = framePart{title: section.Title, body: body}
	}
	return s.render(ctx, parts), nil
}
