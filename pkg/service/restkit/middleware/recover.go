package middleware

import (
	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/service/logkit"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/context"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/router"
	"github.com/spf13/cast"
)

// Recover 错误处理。
// B4: 用 Writer.Written() 判断响应是否已写入，替代 IsAborted()。
// 原逻辑在 IsAborted() 为 true 时直接 return，会吞掉未写入响应的错误：
//   - handler 调用 Abort() 后 panic → 响应未写 → return 后客户端拿到空响应
//
// 改为：只要响应未写入就补写 JsonError，已写入则跳过避免 gin 重复写告警。
func Recover() router.Handler {
	return func(ctx *context.Context) {
		defer func() {
			if err := recover(); err != nil {
				var msg string
				if e, ok := err.(exception.Exception); ok {
					msg = e.Msg
					// 带代码位置信息
					logkit.ErrorException(e)
				} else {
					msg = cast.ToString(err)
					logkit.ErrorException(exception.New(msg, 3))
				}
				if !ctx.Proxy.Writer.Written() {
					ctx.JsonError(msg)
				}
			}
		}()
		ctx.Proxy.Next()
	}
}
