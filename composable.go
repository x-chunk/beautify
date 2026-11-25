package beautify

import (
	"errors"

	"go.xchunk.org/beautify/dra"
)

// ErrNilComposable is returned when a nil Composable is rendered.
var ErrNilComposable = errors.New("beautify: nil composable")

// Composable is an element of high-level output. Implementations must be
// safe for concurrent use when they are rendered concurrently.
type Composable interface {
	// Render lowers the element to a model. When ctx is bounded, the
	// returned model must not be wider than the bound.
	Render(ctx Context) (dra.Model, error)
}

// ComposableFunc adapts a function to Composable.
type ComposableFunc func(ctx Context) (dra.Model, error)

// Render calls f. A nil f returns ErrNilComposable.
func (f ComposableFunc) Render(ctx Context) (dra.Model, error) {
	if f == nil {
		return dra.Model{}, ErrNilComposable
	}
	return f(ctx)
}
