package blocks

import (
	"strings"
	"testing"

	"go.xchunk.org/beautify/dra"
)

func TestSpace(t *testing.T) {
	checkText(t, Space(3, 2), unbounded, "   ", "   ")
	checkText(t, Space(3, 2), 2, "  ", "  ")
	checkText(t, Space(0, 2), unbounded, "", "")
	checkText(t, Space(-4, 2), 5, "", "")
	checkEmpty(t, Space(3, 0), unbounded)
	checkEmpty(t, Space(3, -1), 5)
	checkEmpty(t, Space(3, 2), 0)
	checkContract(t, "space", Space(7, 2))
	checkGolden(t, "space_default", Space(3, 2), unbounded)
}

func TestRule(t *testing.T) {
	checkText(t, Rule(), 5, "─────")
	checkText(t, Rule(), unbounded, strings.Repeat("─", 40))
	checkText(t, Rule(RuleWidth(3)), unbounded, "───")
	checkText(t, Rule(RuleWidth(3)), 5, "─────")
	checkText(t, Rule(RuleWidth(-3)), unbounded, "")
	checkText(t, Rule(RuleChar("=")), 4, "====")
	checkText(t, Rule(RuleChar("-=")), 5, "-=-=-")
	checkText(t, Rule(RuleChar("")), 3, "───")
	checkText(t, Rule(RuleChar("-\n=")), 3, "---")
	checkText(t, Rule(RuleChar("\n")), 3, "   ")
	checkEmpty(t, Rule(), 0)
	if m := renderWith(t, longFactory, Rule(RuleChar("漢")), 5); m.String() != "漢漢 " || m.Width() != 5 {
		t.Errorf("wide rule = %q, width %d", m.String(), m.Width())
	}
	checkContract(t, "rule", Rule())
	checkContract(t, "wide rule", Rule(RuleChar("漢-")))
	checkGolden(t, "rule_default", Rule(), 6)
	checkGolden(t, "rule_style", Rule(RuleStyle(red)), 6)
	checkGolden(t, "rule_char", Rule(RuleChar("-="), RuleStyle(blue)), 5)
	checkGolden(t, "rule_width", Rule(RuleWidth(4)), unbounded)
}

func TestModel(t *testing.T) {
	prepared := dra.NewFactory().Styled("prepared\nmodel", red)
	checkText(t, Model(prepared), unbounded, "prepared", "model")
	checkText(t, Model(prepared), 4, "prep", "mode")
	checkEmpty(t, Model(prepared), 0)
	checkEmpty(t, Model(dra.Model{}), 5)
	checkContract(t, "model", Model(prepared))
	checkContract(t, "long model", Model(longFactory.Text("漢字漢字")))
	checkGolden(t, "model_default", Model(prepared), 6)
}
