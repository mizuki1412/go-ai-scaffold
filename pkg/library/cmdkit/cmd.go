package cmdkit

import (
	"bufio"
	"errors"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/spf13/cast"
)

type RunParams struct {
	Timeout int  `comment:"超时时间s"`
	Async   bool `comment:"异步处理返回值"`
}

// Run
// example: []string{"/bin/bash", "-c", "xxx xxx"}, []string{"/bin/sh", "-c", "xxx.sh xxx"}, []string{"xxx","-h"}
// example: []string{"cmd", "/C", "xxx xxx"}, []string{"xxx.bat"}
func Run(command []string, params ...RunParams) (string, error) {
	if len(command) == 0 {
		panic(exception.New("cmd need command"))
	}
	var param RunParams
	if len(params) == 0 {
		param = RunParams{}
	} else {
		param = params[0]
	}
	name := command[0]
	var args []string
	if len(command) > 1 {
		args = command[1:]
	}
	cmd := exec.Command(name, args...)
	if !param.Async {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return "", err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return "", err
		}
		if err = cmd.Start(); err != nil {
			return "", err
		}
		if param.Timeout > 0 {
			// P1 修复：to 改带缓冲——超时分支返回后，结果 goroutine 仍能完成投递并退出，
			// 原无缓冲写法会让它在 `to <-` 上永久阻塞（goroutine 泄漏）
			to := make(chan map[string]any, 1)
			go func() {
				ret0, err2 := getRet(stdout, stderr, cmd)
				to <- map[string]any{"ret": ret0, "err": err2}
			}()
			select {
			case <-time.After(time.Duration(param.Timeout) * time.Second):
				// P1 修复：超时必须杀掉子进程，否则子进程残留继续运行
				if cmd.Process != nil {
					_ = cmd.Process.Kill()
				}
				return "", errors.New("cmd timeout:" + name)
			case m := <-to:
				ret := m["ret"].(string)
				var err error
				if m["err"] != nil {
					err = m["err"].(error)
				}
				return cast.ToString(ret), err
			}
		} else {
			ret, err := getRet(stdout, stderr, cmd)
			return ret, err
		}
	} else {
		if err := cmd.Start(); err != nil {
			return "", err
		}
	}
	return "", nil
}

func getRet(stdout io.ReadCloser, stderr io.ReadCloser, cmd *exec.Cmd) (string, error) {
	// P2 修复：ReadString 在「无尾换行的最后一行」时同时返回数据与 io.EOF，
	// 原实现先判错 break 后拼接，会丢掉这一行；改为先追加再判错。
	// 循环拼接改用 strings.Builder（§11.4：O(n) 拼接替代 O(n²) 的 ret += line）。
	var sb strings.Builder
	reader := bufio.NewReader(stdout)
	var readErr error
	for {
		line, err2 := reader.ReadString('\n')
		sb.WriteString(line)
		if err2 != nil {
			if !errors.Is(err2, io.EOF) {
				readErr = err2
			}
			break
		}
	}
	ret := sb.String()
	bytesErr, err := io.ReadAll(stderr)
	if err != nil {
		return ret, err
	}
	if len(bytesErr) != 0 {
		return ret, errors.New(string(bytesErr))
	}
	if err = cmd.Wait(); err != nil {
		return ret, err
	}
	return ret, readErr
}
