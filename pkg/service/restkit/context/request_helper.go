package context

import (
	"strings"

	"github.com/example/go-ai-scaffold/pkg/library/c"
	"github.com/example/go-ai-scaffold/pkg/service/jwtkit"
)

var HeaderTokenKey = "Authorization"
var CookieTokenKey = "token"

// BearerPrefix Swagger UI / OpenAPI http/bearer scheme 发送 "Bearer <token>"。
const BearerPrefix = "Bearer "

func (ctx *Context) ReadToken() {
	token := ctx.Request.Header.Get(HeaderTokenKey)
	if token == "" || token == "undefined" {
		// 从cookie中获取
		token, _ = ctx.Proxy.Cookie(CookieTokenKey)
	}
	if token != "" && token != "undefined" {
		// 兼容 Swagger UI Authorize：http/bearer scheme 发送 "Bearer <token>"，需剥离前缀
		token = strings.TrimPrefix(token, BearerPrefix)
		_ = c.RecoverFuncWrapper(func() {
			code := jwtkit.Parse(token)
			ctx.Set("jwt", code)
			ctx.Set("jwt-token", token)
		})
	}
}
