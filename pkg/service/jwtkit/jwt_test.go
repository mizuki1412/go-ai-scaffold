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
