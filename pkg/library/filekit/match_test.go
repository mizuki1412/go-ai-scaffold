package filekit

import "testing"

func TestMatchGlob(t *testing.T) {
	cases := []struct {
		pattern, name string
		want          bool
	}{
		{"*.go", "main.go", true},
		{"src/**/*.go", "src/foo/bar/main.go", true},
		{"src/**/*.go", "src/main.go", true},  // ** 匹配 0 段
		{"src/**/*.go", "doc/main.go", false}, // 前缀不匹配
		{"**", "any/path/file.go", true},
		{"a**b", "axxb", true}, // 非完整段，退化为两个 *
		{"a**b", "a/b", false}, // 不应跨分隔符
		{"[abc]", "a", true},
	}
	for _, c := range cases {
		got, err := MatchGlob(c.pattern, c.name)
		if err != nil {
			t.Fatalf("MatchGlob(%q,%q) 意外错误: %v", c.pattern, c.name, err)
		}
		if got != c.want {
			t.Errorf("MatchGlob(%q,%q)=%v, want %v", c.pattern, c.name, got, c.want)
		}
	}
	if _, err := MatchGlob("[abc", "a"); err == nil {
		t.Error("坏模式应返回 error")
	}
}

func TestReadLines(t *testing.T) {
	f := t.TempDir() + "/lines.txt"
	if err := WriteFile(f, []byte("a\n\nb")); err != nil {
		t.Fatal(err)
	}
	var lines []string
	if err := ReadLines(f, func(line string) error {
		lines = append(lines, line)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(lines) != 3 || lines[0] != "a" || lines[1] != "" || lines[2] != "b" {
		t.Fatalf("逐行读取结果不符: %q", lines)
	}
}

func TestSearchNotFoundListsPaths(t *testing.T) {
	_, err := Search("no-such-file-xyz.txt", t.TempDir())
	if err == nil {
		t.Fatal("找不到文件应返回 error")
	}
}
