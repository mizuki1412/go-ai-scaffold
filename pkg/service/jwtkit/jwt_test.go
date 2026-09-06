package jwtkit

import (
	"testing"
	"time"

	"github.com/example/go-ai-scaffold/pkg/cli/configkey"
	"github.com/example/go-ai-scaffold/pkg/service/configkit"
)

// TestIdleTtl 空闲窗口读取与退化语义：jwt.idle<=0（未配置或显式禁用）时回退 jwt.expire（不滑动）。
func TestIdleTtl(t *testing.T) {
	configkit.Set(configkey.JwtIdle, 2)
	if got := IdleTtl(); got != 2*time.Hour {
		t.Errorf("IdleTtl() = %v, want 2h", got)
	}

	configkit.Set(configkey.JwtIdle, 0)
	configkit.Set(configkey.JwtExpire, 6)
	if got := IdleTtl(); got != 6*time.Hour {
		t.Errorf("IdleTtl() = %v, want 6h（idle<=0 时退化到 expire）", got)
	}
}

// TestTokenRoundtrip 签发与解析闭环（需先配置密钥与有效期）。
func TestTokenRoundtrip(t *testing.T) {
	configkit.Set(configkey.JwtSecretKey, "test-secret")
	configkit.Set(configkey.JwtExpire, 1)
	claims := New(42)
	token := claims.Token()
	parsed := Parse(token)
	if parsed.IdInt() != 42 {
		t.Errorf("Parse(Token(New(42))).IdInt() = %d, want 42", parsed.IdInt())
	}
	if !parsed.IsValid() {
		t.Error("解析出的 claims 应在有效期内")
	}
}

// TestNewScope token 域标记随签发/解析往返；不传 scope 时为空（兼容存量 token）。
func TestNewScope(t *testing.T) {
	configkit.Set(configkey.JwtSecretKey, "test-secret")
	configkit.Set(configkey.JwtExpire, 1)
	if got := Parse(New(7, "admin").Token()).Scope; got != "admin" {
		t.Errorf("scope = %q, want admin", got)
	}
	if got := Parse(New(8).Token()).Scope; got != "" {
		t.Errorf("scope = %q, want 空", got)
	}
}

// TestParseErr 非法 token 返回 error 而非 panic（ReadToken 可选登录链路依赖此语义：
// 客户端携带历史遗留随机串/垃圾 token 应视为未登录，不产生 ERROR 级异常日志）。
func TestParseErr(t *testing.T) {
	configkit.Set(configkey.JwtSecretKey, "test-secret")
	configkit.Set(configkey.JwtExpire, 1)
	for _, bad := range []string{"", "not-a-jwt", "a.b.c", "Bearer abc.def.ghi"} {
		if _, err := ParseErr(bad); err == nil {
			t.Errorf("ParseErr(%q) 应返回 error", bad)
		}
	}
	parsed, err := ParseErr(New(9).Token())
	if err != nil || parsed.IdInt() != 9 {
		t.Errorf("ParseErr(合法token) = %v, %v; want id 9, nil", parsed, err)
	}
}
