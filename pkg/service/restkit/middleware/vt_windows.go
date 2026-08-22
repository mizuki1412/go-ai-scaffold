//go:build windows

package middleware

import "golang.org/x/sys/windows"

// init 为 stdout/stderr 启用 VT（ANSI 转义）处理。
// 旧版 conhost（cmd.exe 默认宿主）默认不解析 ANSI 颜色码，会把 `\033[102;30m`
// 原样输出为 `←[102;30m` 乱码；Windows Terminal 默认已开启，重复设置无副作用。
// 输出被重定向（非控制台）时 GetConsoleMode 失败，静默跳过——此时
// colorEnabled 已判非终端，颜色本来就不输出。
func init() {
	enableVT(windows.Stdout)
	enableVT(windows.Stderr)
}

func enableVT(handle windows.Handle) {
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err == nil {
		_ = windows.SetConsoleMode(handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	}
}
