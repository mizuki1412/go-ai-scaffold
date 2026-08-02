package context

import (
	"net/http"
	"strings"

	"github.com/example/go-ai-scaffold/pkg/library/c"
	"github.com/example/go-ai-scaffold/pkg/service/jwtkit"
)

var HeaderTokenKey = "Authorization"
var CookieTokenKey = "token"

// BearerPrefix Swagger UI / OpenAPI http/bearer scheme 发送 "Bearer <token>"。
const BearerPrefix = "Bearer "

// JwtCookie 配置项。可在应用启动时通过 configkit 覆盖，或直接赋值。
// - Domain: cookie 作用域名。为空时用请求 Host 的域名。
// - Secure: 是否仅 HTTPS 传输。生产环境应为 true。
// - SameSite: SameSite 策略。跨域场景需 Lax/None（None 时 Secure 必须 true）。
var (
	JwtCookieDomain   = ""
	JwtCookieSecure   = false
	JwtCookieSameSite = http.SameSiteLaxMode
)

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
