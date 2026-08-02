package middleware

import (
	"time"

	"github.com/example/go-ai-scaffold/pkg/service/cachekit"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/context"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/router"
)

// AuthJWT 用户名密码校验。
// B10: 用登录时签发的原始 token 字符串作 cache key，而非 jwt.Token() 重新签发。
func AuthJWT() router.Handler {
	return func(ctx *context.Context) {
		jwt := ctx.GetJwt()
		token := ctx.GetJwtToken()
		// 获取 jwt
		if !jwt.IsValid() || jwt.ExpiresAt.Before(time.Now()) || (token != "" && cachekit.Get("token:"+token) == "") {
			ctx.Json(context.RestRet{
				Result:  context.ResultAuthErr,
				Message: "登录失效",
			})
			ctx.Proxy.Abort()
		} else {
			ctx.Proxy.Next()
		}
	}
}
