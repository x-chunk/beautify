package beautify

import "go.xchunk.org/beautify/dra"

// defaultFactory serves contexts that were not given a factory.
var defaultFactory = dra.NewFactory()

// Context carries what a Composable needs to render: the factory to build
// models with and the width that is available. It is an immutable value and
// is safe for concurrent use. The zero value is unbounded and uses a default
// factory.
type Context struct {
	factory *dra.Factory
	width   int
	bounded bool
}

// NewContext returns an unbounded context with the factory f. A nil f means
// a default factory.
func NewContext(f *dra.Factory) Context {
	return Context{factory: f}
}

// Factory returns the factory that models should be built with, so that
// settings such as the width mode apply everywhere. It is never nil.
func (c Context) Factory() *dra.Factory {
	if c.factory == nil {
		return defaultFactory
	}
	return c.factory
}

// Width returns the available width in cells and whether there is a bound.
// It returns 0 and false for an unbounded context. A bounded width of 0 is
// valid and means there is no room.
func (c Context) Width() (width int, bounded bool) {
	return c.width, c.bounded
}

// WithWidth returns a copy of c that is bounded to width cells. Negative
// values count as 0.
func (c Context) WithWidth(width int) Context {
	c.width, c.bounded = max(width, 0), true
	return c
}

// Unbounded returns a copy of c without a width bound.
func (c Context) Unbounded() Context {
	c.width, c.bounded = 0, false
	return c
}
