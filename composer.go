package beautify

import (
	"io"
	"os"
	"strconv"
	"sync"

	"go.xchunk.org/beautify/dra"
	"go.xchunk.org/beautify/internal/term"
)

// maxPooledBuffer is the capacity above which an output buffer is dropped
// instead of reused, so one large render does not pin its memory.
const maxPooledBuffer = 64 << 10

// buffers holds output buffers between renders.
var buffers = sync.Pool{New: func() any { return new([]byte) }}

// Composer renders Templates to a writer. It is safe for concurrent use: the
// output of concurrent Render calls never interleaves. Nothing else may
// write to the writer at the same time.
type Composer struct {
	w       io.Writer
	engine  dra.Engine
	factory *dra.Factory
	newline bool

	fixedWidth bool
	width      int // used with fixedWidth; 0 or less means unbounded

	mu sync.Mutex // serializes writes to w
}

type composerConfig struct {
	profile      dra.Profile
	fixedProfile bool
	width        int
	fixedWidth   bool
	factory      *dra.Factory
	newline      bool
}

// ComposerOption configures a Composer.
type ComposerOption func(*composerConfig)

// WithProfile fixes the color profile instead of detecting it.
func WithProfile(p dra.Profile) ComposerOption {
	return func(c *composerConfig) {
		c.profile, c.fixedProfile = p, true
	}
}

// WithWidth fixes the available width to n cells instead of detecting it.
// With n <= 0 the width is unbounded.
func WithWidth(n int) ComposerOption {
	return func(c *composerConfig) {
		c.width, c.fixedWidth = n, true
	}
}

// WithFactory sets the factory that Composables build their models with. Use
// it to choose the width mode or a base style. A nil f means a default
// factory.
func WithFactory(f *dra.Factory) ComposerOption {
	return func(c *composerConfig) {
		c.factory = f
	}
}

// WithNewline sets whether a newline follows each render. The default is
// on.
func WithNewline(on bool) ComposerOption {
	return func(c *composerConfig) {
		c.newline = on
	}
}

// NewComposer returns a Composer that writes to w. The color profile is
// detected once, with DetectProfile, unless an option fixes it. A nil w
// discards the output. Nil options are skipped.
func NewComposer(w io.Writer, opts ...ComposerOption) *Composer {
	cfg := composerConfig{newline: true}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if w == nil {
		w = io.Discard
	}
	if !cfg.fixedProfile {
		cfg.profile = DetectProfile(w)
	}
	return &Composer{
		w:          w,
		engine:     dra.NewEngine(cfg.profile),
		factory:    cfg.factory,
		newline:    cfg.newline,
		fixedWidth: cfg.fixedWidth,
		width:      cfg.width,
	}
}

// Profile returns the color profile the composer renders for.
func (c *Composer) Profile() dra.Profile {
	return c.engine.Profile()
}

// Render renders t and writes the result with a single Write call. An empty
// result writes nothing.
//
// The width is the fixed one if set, otherwise that of the terminal, which
// is queried on every call, otherwise the COLUMNS variable, otherwise
// unbounded.
//
// When t fails to render, its error is returned and nothing is written. A
// write error is returned as is; a short write without an error returns
// io.ErrShortWrite. A panic in a Composable is not recovered.
func (c *Composer) Render(t Template) error {
	m, err := t.Render(c.context())
	if err != nil {
		return err
	}
	if m.IsEmpty() {
		return nil
	}
	p := buffers.Get().(*[]byte)
	buf := c.engine.Append((*p)[:0], m)
	if c.newline {
		buf = append(buf, '\n')
	}
	err = c.write(buf)
	if cap(buf) <= maxPooledBuffer {
		*p = buf
		buffers.Put(p)
	}
	return err
}

// context returns the context for one render.
func (c *Composer) context() Context {
	ctx := NewContext(c.factory)
	if c.fixedWidth {
		if c.width <= 0 {
			return ctx
		}
		return ctx.WithWidth(c.width)
	}
	if n, ok := term.Width(c.w); ok {
		return ctx.WithWidth(n)
	}
	if n, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && n > 0 {
		return ctx.WithWidth(n)
	}
	return ctx
}

// write sends buf to the writer in one call.
func (c *Composer) write(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	n, err := c.w.Write(buf)
	if err == nil && n < len(buf) {
		err = io.ErrShortWrite
	}
	return err
}
