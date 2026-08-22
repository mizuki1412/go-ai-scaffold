package middleware

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/example/go-ai-scaffold/pkg/service/logkit"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/context"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/router"
)

// 色板按 Windows conhost 默认 16 色调色板的对比度选取：亮背景配黑字、深背景配亮白字。
// 色相含义不变：绿=成功/快、黄=警告/较慢、红=错误/慢、蓝=3xx 重定向。
const (
	colorGreen  = "\033[102;30m" // 亮绿底 + 黑字（conhost 对比度约 15:1；原 102;97 白字仅约 1.3:1）
	colorYellow = "\033[103;30m" // 亮黄底 + 黑字（约 19:1，警示牌配色；原 43;37 橄榄底灰字发糊）
	colorRed    = "\033[41;97m"  // 深红底 + 亮白字（约 8.6:1，错误白字红底的惯例）
	colorBlue   = "\033[44;97m"  // 深蓝底 + 亮白字（约 12:1，3xx 重定向）
	colorReset  = "\033[0m"
)

// colorEnabled 仅当 stderr 为终端（控制台）时输出 ANSI 颜色：
// 重定向到文件/管道时自动退化为纯文本，避免转义码污染日志。
var colorEnabled = func() bool {
	fi, err := os.Stderr.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}()

// paint 为 s 染色；未启用颜色（非终端）时原样返回。
func paint(color, s string) string {
	if !colorEnabled {
		return s
	}
	return color + s + colorReset
}

func Log() router.Handler {
	return func(ctx *context.Context) {
		t := time.Now()
		logkit.Info("request", "url", ctx.Request.URL.String())

		ctx.Proxy.Next()

		latency := float64(time.Since(t).Microseconds()) / 1000
		status := ctx.Proxy.Writer.Status()
		logkit.InfoFile("response", "url", ctx.Request.URL.String(), "latency", latency, "status", status)
		msg := fmt.Sprintf("msg=response %s %s url=%s",
			paint(statusColor(status), strconv.Itoa(status)),
			paint(latencyColor(latency), fmtLatency(latency)),
			ctx.Request.URL.String())
		fmt.Fprintf(os.Stderr, "time=%s level=INFO %s\n",
			t.Format("2006/01/02-15:04:05"), msg)
	}
}

func statusColor(code int) string {
	if code >= 200 && code < 300 {
		return colorGreen
	}
	if code >= 300 && code < 400 {
		return colorBlue
	}
	return colorRed
}

func latencyColor(ms float64) string {
	if ms < 1000 {
		return colorGreen
	}
	if ms < 3000 {
		return colorYellow
	}
	return colorRed
}

func fmtLatency(ms float64) string {
	if ms >= 1000 {
		return fmt.Sprintf("%7.2fs", ms/1000)
	}
	return fmt.Sprintf("%7.2fms", ms)
}
