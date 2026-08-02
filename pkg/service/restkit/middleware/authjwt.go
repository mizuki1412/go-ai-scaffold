package middleware

import (
	"time"

	"github.com/example/go-ai-scaffold/pkg/service/cachekit"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/context"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/router"
)

// AuthJWT 用户名密码校验
func AuthJWT() router.Handler {
	return func(ctx *context.Context) {
		jwt := ctx.GetJwt()
		// 获取 jwt
		if !jwt.IsValid() || jwt.ExpiresAt.Before(time.Now()) || cachekit.Get("token:"+jwt.Token()) == "" {
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
