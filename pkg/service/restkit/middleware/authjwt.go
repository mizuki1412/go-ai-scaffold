package middleware

import (
	"time"

	"github.com/example/go-ai-scaffold/pkg/service/cachekit"
	"github.com/example/go-ai-scaffold/pkg/service/jwtkit"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/context"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/router"
)

// AuthJWT 登录态校验：JWT 本体有效 + 白名单 key 在，二者缺一即 401。
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
			// 滑动续期：鉴权通过即重置白名单 TTL（jwt.idle 空闲窗口），活跃用户不掉线；
			// JWT 自身 exp（jwt.expire）是绝对上限，由上方 ExpiresAt 校验兜底。
			if token != "" {
				cachekit.Renew("token:"+token, &cachekit.Param{Ttl: jwtkit.IdleTtl()})
			}
			ctx.Proxy.Next()
		}
	}
}
