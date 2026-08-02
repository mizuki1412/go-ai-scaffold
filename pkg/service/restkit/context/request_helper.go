package context

import (
	"github.com/example/go-ai-scaffold/pkg/library/c"
	"github.com/example/go-ai-scaffold/pkg/service/jwtkit"
)

var HeaderTokenKey = "Authorization"
var CookieTokenKey = "token"

func (ctx *Context) ReadToken() {
	token := ctx.Request.Header.Get(HeaderTokenKey)
	if token == "" || token == "undefined" {
		// 从cookie中获取
		token, _ = ctx.Proxy.Cookie(CookieTokenKey)
	}
	if token != "" && token != "undefined" {
		_ = c.RecoverFuncWrapper(func() {
			code := jwtkit.Parse(token)
			ctx.Set("jwt", code)
			ctx.Set("jwt-token", token)
		})
	}
}
