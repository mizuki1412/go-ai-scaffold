package stringkit

import "testing"

func TestCompareVersion(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.0", "v1.0.0", 0},
		{"2.10.8", "v2.10.7", 1},
		{"1.9", "1.10", -1},
		{"v1", "1", 0},
	}
	for _, c := range cases {
		if got := CompareVersion(c.a, c.b); got != c.want {
			t.Errorf("CompareVersion(%q,%q)=%d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestLevenshteinAndSimilarText(t *testing.T) {
	if d := Levenshtein("kitten", "sitting", 1, 1, 1); d != 3 {
		t.Errorf("Levenshtein(kitten,sitting)=%d, want 3", d)
	}
	var percent float64
	if n := SimilarText("hello world", "hello go", &percent); n <= 0 || percent <= 0 {
		t.Errorf("SimilarText 应有公共段, n=%d percent=%.2f", n, percent)
	}
}

func TestIsNumeric(t *testing.T) {
	for _, s := range []string{"123", "-1", "+1.5", "12.34"} {
		if !IsNumeric(s) {
			t.Errorf("IsNumeric(%q) 应为 true", s)
		}
	}
	for _, s := range []string{"", "-", "1.2.3", "12a", "."} {
		if IsNumeric(s) {
			t.Errorf("IsNumeric(%q) 应为 false", s)
		}
	}
}

func TestSplitAndTrimAndEqualFold(t *testing.T) {
	got := SplitAndTrim(" a , b ,,c ", ",")
	if len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Fatalf("SplitAndTrim 结果不符: %v", got)
	}
	if !EqualFoldWithoutChars("user_Name", "USER-NAME") {
		t.Fatal("忽略大小写与符号后应相等")
	}
	if EqualFoldWithoutChars("abc", "abd") {
		t.Fatal("不同内容不应相等")
	}
}
