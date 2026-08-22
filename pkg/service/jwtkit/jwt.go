package jwtkit

import (
	"github.com/example/go-ai-scaffold/pkg/class"
	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/cli/configkey"
	"github.com/example/go-ai-scaffold/pkg/service/configkit"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cast"
	"time"
)

type Claims struct {
	Id    any             `json:"id"`
	Ext   class.MapString `json:"ext"`
	Scope string          `json:"scope,omitempty"`
	Valid bool            `json:"valid"`
	jwt.RegisteredClaims
}

// New id: string or int。可选 scope 标记 token 所属域（如 "app"/"admin"），
// 供 middleware.AuthJWTScope 在同一 secretKey 下隔离 app 端与管理端 token。
func New(id any, scope ...string) Claims {
	c := Claims{
		Id:    id,
		Valid: true,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(configkit.GetInt(configkey.JwtExpire)) * time.Hour)), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),                                                                       // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),                                                                       // 生效时间
		},
	}
	if len(scope) > 0 {
		c.Scope = scope[0]
	}
	return c
}

// IdleTtl 返回会话空闲窗口时长（jwt.idle 小时）。
// SetJwtCookie 写白名单与 AuthJWT 滑动续期共用；jwt.idle<=0（未配置或显式禁用）时
// 退化为旧语义：窗口=过期时间（jwt.expire），即不滑动。
func IdleTtl() time.Duration {
	idle := configkit.GetInt(configkey.JwtIdle)
	if idle <= 0 {
		idle = configkit.GetInt(configkey.JwtExpire)
	}
	return time.Duration(idle) * time.Hour
}

// secretKey 返回 JWT 签名密钥。未配置时直接 panic：
// 空密钥/可预测默认密钥意味着任何人都能伪造任意用户的 token（P0 安全修复）。
func secretKey() []byte {
	s := configkit.GetString(configkey.JwtSecretKey)
	if s == "" {
		panic(exception.New("jwt 密钥未配置（jwt.secretKey），禁止以空密钥签发/解析 token"))
	}
	return []byte(s)
}

func (c Claims) Token() string {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	s, err := t.SignedString(secretKey())
	if err != nil {
		panic(exception.New("jwt token err: " + err.Error()))
	}
	return s
}

func (c Claims) IdInt() int {
	return cast.ToInt(c.Id)
}
func (c Claims) IdInt32() int32 {
	return cast.ToInt32(c.Id)
}
func (c Claims) IdInt64() int64 {
	return cast.ToInt64(c.Id)
}
func (c Claims) IdStr() string {
	return cast.ToString(c.Id)
}

func (c Claims) IsValid() bool {
	if !c.Valid {
		return false
	}
	if c.ExpiresAt.Unix() < time.Now().Unix() {
		return false
	}
	return true
}

func Parse(token string) Claims {
	t, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (any, error) {
		return secretKey(), nil
	})

	// P2 修复：ParseWithClaims 出错时返回的 t 为 nil，原实现直接访问 t.Claims
	// 会 nil deref 而非给出清晰报错；先判 err
	if err != nil {
		panic(exception.New("jwt parse err: " + err.Error()))
	}
	if claims, ok := t.Claims.(*Claims); ok && t.Valid {
		return *claims
	}
	panic(exception.New("jwt parse err: invalid token"))
}
