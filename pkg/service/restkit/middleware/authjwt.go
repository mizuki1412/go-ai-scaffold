package middleware

import (
	"slices"
	"time"

	"github.com/example/go-ai-scaffold/pkg/service/cachekit"
	"github.com/example/go-ai-scaffold/pkg/service/jwtkit"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/context"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/router"
)

// AuthJWT 登录态校验（不限定 token 域）：JWT 本体有效 + 白名单 key 在。
// B10: 用登录时签发的原始 token 字符串作 cache key，而非 jwt.Token() 重新签发。
func AuthJWT() router.Handler {
	return AuthJWTScope()
}

// AuthJWTScope 要求登录且 token 的 scope 命中给定集合之一。
// app 端与管理端共用同一 secretKey 与 token 缓存命名空间，必须靠 Claims.Scope
// 隔离两域 token，否则任意 app 用户的 token 可横向调用管理端接口（跨域越权）。
// 不传 scopes 表示不校验域（兼容既有 app 端路由）。
func AuthJWTScope(scopes ...string) router.Handler {
	return func(ctx *context.Context) {
		jwt := ctx.GetJwt()
		token := ctx.GetJwtToken()
		if !jwt.IsValid() || jwt.ExpiresAt.Before(time.Now()) || (token != "" && cachekit.Get("token:"+token) == "") {
			ctx.Json(context.RestRet{
				Result:  context.ResultAuthErr,
				Message: "登录失效",
			})
			ctx.Proxy.Abort()
			return
		}
		if len(scopes) > 0 && !slices.Contains(scopes, jwt.Scope) {
			ctx.Json(context.RestRet{
				Result:  context.ResultAuthErr,
				Message: "登录状态域不匹配，请从对应入口登录",
			})
			ctx.Proxy.Abort()
			return
		}
		// 滑动续期：鉴权通过即重置白名单 TTL（jwt.idle 空闲窗口），活跃用户不掉线；
		// JWT 自身 exp（jwt.expire）是绝对上限，由上方 ExpiresAt 校验兜底。
		if token != "" {
			cachekit.Renew("token:"+token, &cachekit.Param{Ttl: jwtkit.IdleTtl()})
		}
		ctx.Proxy.Next()
	}
}
