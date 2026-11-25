package blocks

import (
	"errors"
	"strings"
	"testing"

	"go.xchunk.org/beautify"
	"go.xchunk.org/beautify/dra"
)

func TestKeyValue(t *testing.T) {
	pairs := []Pair{{"OS", raw("Linux")}, {"Kernel", raw("7.3\nrt")}, {"", raw("x")}}
	checkText(t, KeyValue(pairs), unbounded,
		"OS    : Linux",
		"Kernel: 7.3",
		"        rt",
		"      : x")
	checkText(t, KeyValue(pairs[:2], KeyValueSeparator(" = "), KeyAlign(dra.AlignRight)), unbounded,
		"    OS = Linux",
		"Kernel = 7.3",
		"         rt")
	checkText(t, KeyValue(pairs[:1], KeyValueSeparator("")), unbounded, "OSLinux")
	checkText(t, KeyValue(pairs[:1], KeyValueSeparator("\n"), nil), unbounded, "OS Linux")
	checkText(t, KeyValue([]Pair{{"a\nb", raw("x")}, {"c\r\nd", nil}}), unbounded, "a b: x", "c d: ")
	checkText(t, KeyValue([]Pair{{"k", raw("")}}), unbounded, "k: ")
	checkText(t, KeyValue([]Pair{{"key", wrapping("aa bb cc")}}), 10, "key: aa bb", "     cc")

	// Without a cell for the value it moves to the next line.
	checkText(t, KeyValue([]Pair{{"key", wrapping("aa bb")}}), 6, "key: a", "     a", "     b", "     b")
	checkText(t, KeyValue([]Pair{{"key", wrapping("aa bb")}}), 5, "key: ", "  aa", "  bb")
	checkText(t, KeyValue([]Pair{{"key", wrapping("aa")}, {"k", raw("")}}), 4, "key:", "  aa", "k  :")
	checkText(t, KeyValue([]Pair{{"key", raw("aa")}}), 3, "key", "  a")
	checkText(t, KeyValue([]Pair{{"key", Text("aa")}}), 2, "ke")

	checkEmpty(t, KeyValue(nil), unbounded)
	checkEmpty(t, KeyValue(pairs), 0)

	in := []Pair{{"a", raw("x")}}
	kv := KeyValue(in)
	in[0].Key = "changed"
	checkText(t, kv, unbounded, "a: x")
}

func TestKeyValueErrors(t *testing.T) {
	kv := KeyValue([]Pair{{"a", nil}, {"b", failing()}})
	_, err := kv.Render(context(realFactory, 20))
	if !errors.Is(err, errChild) || !strings.Contains(err.Error(), "keyvalue child 1") {
		t.Errorf("error = %v", err)
	}
}

func TestKeyValueContract(t *testing.T) {
	checkChildContract(t, func(c beautify.Composable) beautify.Composable {
		return KeyValue([]Pair{{"key", c}, {"漢字\x1b", c}, {"", nil}})
	})
	long := renderWith(t, longFactory, KeyValue([]Pair{{"漢字", raw("a")}, {"key", raw("b")}}), unbounded)
	if got, want := long.String(), "漢字: a\nkey : b"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	pairs := []Pair{{"OS", Text("Linux", TextStyle(red))}, {"Kernel", Text("7.3\nrt")}}
	checkGolden(t, "keyvalue_default", Styled(KeyValue(pairs), onBG), unbounded)
	checkGolden(t, "keyvalue_style", KeyValue(pairs, KeyStyle(blue)), unbounded)
	checkGolden(t, "keyvalue_separator", KeyValue(pairs, KeyValueSeparator(" -> ")), unbounded)
	checkGolden(t, "keyvalue_right", KeyValue(pairs, KeyAlign(dra.AlignRight)), unbounded)
	checkGolden(t, "keyvalue_narrow", KeyValue(pairs, KeyStyle(blue)), 8)
}
