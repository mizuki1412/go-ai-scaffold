package exception

import (
	"errors"
	"testing"
)

func TestNewCodeAndCodeOf(t *testing.T) {
	e := NewCode(403, "禁止访问")
	if e.Code != 403 || e.Msg != "禁止访问" {
		t.Fatalf("NewCode 字段不符: %+v", e)
	}
	if got := CodeOf(e); got != 403 {
		t.Fatalf("CodeOf(Exception) = %d, want 403", got)
	}
	if got := CodeOf(errors.New("plain")); got != CodeNone {
		t.Fatalf("CodeOf(普通error) 应为 0, got %d", got)
	}
}

func TestWrapInheritsCodeAndUnwrap(t *testing.T) {
	root := errors.New("root cause")
	wrapped := Wrap(root, "下单失败")
	if wrapped.Code != CodeNone {
		t.Fatalf("无码根因不应继承出非零码, got %d", wrapped.Code)
	}
	if !errors.Is(wrapped, root) {
		t.Fatal("Wrap 后应能通过 errors.Is 找到根因")
	}
	if wrapped.Msg != "下单失败: root cause" {
		t.Fatalf("Wrap 消息应拼接根因, got %q", wrapped.Msg)
	}

	// 带码 Exception 被包装后，链上取码
	inner := NewCode(30001, "库存不足")
	outer := Wrap(inner, "创建订单")
	if got := CodeOf(outer); got != 30001 {
		t.Fatalf("Wrap 应继承业务码, got %d", got)
	}
	var target Exception
	if !errors.As(outer, &target) || target.Code != 30001 {
		t.Fatal("errors.As 应能从包装链取出带码异常")
	}
}
