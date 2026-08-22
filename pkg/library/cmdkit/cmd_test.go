package cmdkit

import (
	"io"
	"os/exec"
	"strings"
	"testing"
)

// P2-12 起步单测：getRet 的 P2-4 回归——无尾换行的最后一行不能丢。

// newEchoCmd 起一个即刻退出的子进程（输出丢弃），仅用于满足 getRet 内的 cmd.Wait()。
// 用 go 自身保证跨平台可用。
func newEchoCmd(t *testing.T) *exec.Cmd {
	t.Helper()
	cmd := exec.Command("go", "version")
	if err := cmd.Start(); err != nil {
		t.Skipf("无法启动子进程（跳过）: %v", err)
	}
	return cmd
}

func TestGetRetKeepsLastLineWithoutNewline(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"常规多行带尾换行", "line1\nline2\n", "line1\nline2\n"},
		{"最后一行无尾换行（P2-4 回归）", "line1\nline2", "line1\nline2"},
		{"单行无换行", "only", "only"},
		{"空输出", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newEchoCmd(t)
			stdout := io.NopCloser(strings.NewReader(tt.in))
			stderr := io.NopCloser(strings.NewReader(""))
			got, err := getRet(stdout, stderr, cmd)
			if err != nil {
				t.Fatalf("getRet error: %v", err)
			}
			if got != tt.want {
				t.Errorf("getRet = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetRetStderrAsError(t *testing.T) {
	cmd := newEchoCmd(t)
	stdout := io.NopCloser(strings.NewReader("out\n"))
	stderr := io.NopCloser(strings.NewReader("boom"))
	_, err := getRet(stdout, stderr, cmd)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Errorf("stderr 内容应转为 error，got %v", err)
	}
}

func TestRunRejectsEmptyCommand(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Errorf("空 command 应 panic")
		}
	}()
	_, _ = Run(nil)
}
