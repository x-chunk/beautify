package beautify

import (
	"errors"
	"testing"

	"go.xchunk.org/beautify/dra"
)

func TestContextWidth(t *testing.T) {
	base := NewContext(nil)
	tests := []struct {
		name    string
		ctx     Context
		width   int
		bounded bool
	}{
		{"zero value", Context{}, 0, false},
		{"new", base, 0, false},
		{"bounded", base.WithWidth(20), 20, true},
		{"bounded zero", base.WithWidth(0), 0, true},
		{"negative is clamped", base.WithWidth(-5), 0, true},
		{"rebound", base.WithWidth(20).WithWidth(10), 10, true},
		{"unbounded again", base.WithWidth(20).Unbounded(), 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if w, ok := tt.ctx.Width(); w != tt.width || ok != tt.bounded {
				t.Errorf("Width() = %d, %v, want %d, %v", w, ok, tt.width, tt.bounded)
			}
		})
	}
	bounded := base.WithWidth(20)
	bounded.WithWidth(5)
	bounded.Unbounded()
	if w, _ := bounded.Width(); w != 20 {
		t.Errorf("the receiver changed to %d", w)
	}
}

func TestContextFactory(t *testing.T) {
	if (Context{}).Factory() == nil || NewContext(nil).Factory() == nil {
		t.Error("Factory() = nil")
	}
	f := dra.NewFactory(dra.WithMode(dra.ModeLong))
	ctx := NewContext(f)
	if ctx.Factory() != f || ctx.WithWidth(3).Factory() != f || ctx.Unbounded().Factory() != f {
		t.Error("the factory was not kept")
	}
}

func TestComposableFunc(t *testing.T) {
	f := ComposableFunc(func(ctx Context) (dra.Model, error) {
		return ctx.Factory().Text("hi"), nil
	})
	if m, err := f.Render(Context{}); err != nil || m.String() != "hi" {
		t.Errorf("Render() = %q, %v", m.String(), err)
	}
	if _, err := ComposableFunc(nil).Render(Context{}); !errors.Is(err, ErrNilComposable) {
		t.Errorf("nil function: %v, want ErrNilComposable", err)
	}
}
