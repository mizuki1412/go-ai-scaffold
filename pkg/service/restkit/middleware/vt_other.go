//go:build !windows

package middleware

// 非 Windows 平台终端原生支持 ANSI 转义码，无需 VT 启用（见 vt_windows.go）。
