package exception

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
)

// 业务错误码分段约定（对照响应 RestRet.Result）：
//   - 0       未指定，Recover 兜底按 500 处理（兼容历史 panic(exception.New(...)) 行为）
//   - 401     鉴权类错误，HTTP 状态同步为 401
//   - 300~999 业务错误段，随 result 字段原样透传给前端
//   - >=1000  保留
const (
	CodeNone        = 0
	CodeAuthErr     = 401
	CodeBusiness    = 300 // 业务码段起点：300~999
	CodeBusinessEnd = 999 // 业务码段终点
)

type Exception struct {
	Msg   string
	Code  int
	File  string
	Line  int
	Stack []string
	Cause error
}

func New(msg string, skip1 ...int) Exception {
	return NewCode(CodeNone, msg, skip1...)
}

// NewCode 带业务错误码构造异常，code 随 Recover 透传到响应 result。
func NewCode(code int, msg string, skip1 ...int) Exception {
	var skip = 1
	if skip1 != nil && len(skip1) > 0 {
		skip = skip1[0]
	}
	_, file, line, _ := runtime.Caller(skip)
	stack := make([]string, 0, 3)
	stack = append(stack, getStackInfo(file, line))
	for i := 1; i < 3; i++ {
		_, file1, line1, ok := runtime.Caller(skip + i)
		if ok {
			stack = append(stack, getStackInfo(file1, line1))
		}
	}
	return Exception{
		Msg:   msg,
		Code:  code,
		File:  file,
		Line:  line,
		Stack: stack,
	}
}

func NewE(err error, skip ...int) Exception {
	if err == nil {
		return New("", skip...)
	}
	return New(err.Error(), skip...)
}

// Wrap 包装底层 err 形成错误链：消息拼接根因，错误码继承链上已有的业务码，
// Cause 保留根因供 errors.As/errors.Is 继续追溯。
func Wrap(err error, msg string, skip ...int) Exception {
	if err == nil {
		return New(msg, skip...)
	}
	e := NewCode(CodeNone, fmt.Sprintf("%s: %v", msg, err), skip...)
	e.Code = CodeOf(err)
	e.Cause = err
	return e
}

// CodeOf 从错误链中提取第一个非零业务码；无则返回 CodeNone。
func CodeOf(err error) int {
	for err != nil {
		var e Exception
		if errors.As(err, &e) && e.Code != CodeNone {
			return e.Code
		}
		err = errors.Unwrap(err)
	}
	return CodeNone
}

// Unwrap 支持 errors.As/Is 沿 Wrap 链回溯。
func (th Exception) Unwrap() error {
	return th.Cause
}

func getStackInfo(file string, line int) string {
	return file + ":" + strconv.Itoa(line)
}

// Deprecated: 旧的错误日志信息
func (th Exception) Error() string {
	ret := fmt.Sprintf(`%s
Exception StackTrace: 
`, th.Msg)
	for _, e := range th.Stack {
		ret += e + "\n"
	}
	return ret
}
