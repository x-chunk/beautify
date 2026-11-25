package blocks

import (
	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

// defaultRuleChar is the character a Rule repeats.
const defaultRuleChar = "─"

// RuleOption configures Rule.
type RuleOption func(*rule)

// RuleChar sets the text that the rule repeats. The default is "─", which
// also replaces an empty string.
func RuleChar(s string) RuleOption {
	return func(r *rule) {
		if s == "" {
			s = defaultRuleChar
		}
		r.char = s
	}
}

// RuleStyle sets the style of the rule.
func RuleStyle(st dra.Style) RuleOption {
	return func(r *rule) { r.style = st }
}

// RuleWidth sets the width of the rule in an unbounded context. The default
// is 40.
func RuleWidth(n int) RuleOption {
	return func(r *rule) { r.width = n }
}

type rule struct {
	char  string
	style dra.Style
	width int
}

// Rule returns a horizontal line that fills the available width.
func Rule(opts ...RuleOption) beautify.Composable {
	r := rule{char: defaultRuleChar, width: 40}
	applyOptions(&r, opts)
	return r
}

// Render implements beautify.Composable.
func (r rule) Render(ctx beautify.Context) (dra.Model, error) {
	if noRoom(ctx) {
		return dra.Model{}, nil
	}
	width := r.width
	if w, bounded := ctx.Width(); bounded {
		width = w
	}
	return dra.Repeat(ctx.Factory().Styled(r.char, r.style), width), nil
}
