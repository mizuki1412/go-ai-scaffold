package middleware

import (
	"fmt"
	"os"
	"time"

	"github.com/example/go-ai-scaffold/pkg/service/logkit"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/context"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/router"
)

const (
	colorGreen  = "\033[42;37m"
	colorYellow = "\033[43;37m"
	colorRed    = "\033[41;37m"
	colorReset  = "\033[0m"
)

func Log() router.Handler {
	return func(ctx *context.Context) {
		t := time.Now()
		logkit.Info("request", "url", ctx.Request.URL.String())

		ctx.Proxy.Next()

		latency := float64(time.Since(t).Microseconds()) / 1000
		status := ctx.Proxy.Writer.Status()
		logkit.InfoFile("response", "url", ctx.Request.URL.String(), "latency", latency, "status", status)
		msg := fmt.Sprintf("msg=response %s%d%s %s%s%s url=%s",
			statusColor(status), status, colorReset,
			latencyColor(latency), fmtLatency(latency), colorReset,
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
		return "\033[44;37m"
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
