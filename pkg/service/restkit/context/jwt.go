package context

import (
	"strings"
	"time"

	"github.com/example/go-ai-scaffold/pkg/service/cachekit"
	"github.com/example/go-ai-scaffold/pkg/service/jwtkit"
	"github.com/spf13/cast"
)

// SetJwtCookie 存入cookie和cache。
// token 字符串是登录时签发的原始字符串，后续 DestroyJwt/AuthJWT 必须用同一字符串作 cache key。
func (ctx *Context) SetJwtCookie(c jwtkit.Claims, token string) {
	// B3: 从配置读取 cookie domain/secure/samesite，避免从 Origin 头推断导致跨域 cookie 失效
	domain := JwtCookieDomain
	secure := JwtCookieSecure
	sameSite := JwtCookieSameSite
	origin := ctx.Proxy.GetHeader("origin")
	origins := strings.Split(origin, "//")
	if len(origins) > 1 {
		origin = origins[1]
	}
	// 剥离端口：cookie Domain 不允许含端口
	if idx := strings.Index(origin, ":"); idx > 0 {
		origin = origin[:idx]
	}
	// 配置未指定 domain 时，用请求 Host 的域名（剥端口）
	if domain == "" {
		domain = origin
	}
	maxAge := 0
	if c.ExpiresAt != nil {
		maxAge = cast.ToInt(c.ExpiresAt.Unix() - time.Now().Unix())
		if maxAge < 0 {
			maxAge = 0
		}
		// 缓存 token，用于 AuthJWT 校验和 DestroyJwt 注销
		cachekit.Set("token:"+token, "1", &cachekit.Param{Ttl: time.Duration(maxAge) * time.Second})
	} else {
		cachekit.Set("token:"+token, "1")
	}
	// B3: 显式构造 http.Cookie，支持 SameSite/Secure
	ctx.Proxy.SetSameSite(sameSite)
	ctx.Proxy.SetCookie("token", token, maxAge, "/", domain, secure, true)
}

// GetJwt 在authjwt拦截器中进行jwt的过期校验。
// 内部已缓存解析结果，不会重复调用 jwtkit.Parse。
func (ctx *Context) GetJwt() jwtkit.Claims {
	if ctx.Get("jwt") == nil {
		ctx.ReadToken()
	}
	if c := ctx.Get("jwt"); c != nil {
		if cc, ok := c.(jwtkit.Claims); ok {
			return cc
		}
	}
	return jwtkit.Claims{}
}

// GetJwtToken 返回原始 token 字符串（登录时签发）。
// B2/B10: 鉴权缓存 key 和注销缓存 key 必须用此原始字符串，不能用 jwt.Token() 重新签发。
func (ctx *Context) GetJwtToken() string {
	if t, ok := ctx.Get("jwt-token").(string); ok {
		return t
	}
	return ""
}

// DestroyJwt 销毁jwt。
// B2: 用登录时签发的原始 token 字符串作 cache key，而非 jwt.Token() 重新签发。
func (ctx *Context) DestroyJwt() {
	token := ctx.GetJwtToken()
	if token != "" {
		cachekit.Del("token:" + token)
	}
}
