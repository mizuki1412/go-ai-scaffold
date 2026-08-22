package cryptokit

import (
	"strings"
	"testing"
)

// P2-12 起步单测：密码哈希三件套（P1-9 引入）的回归测试。
// 覆盖：bcrypt 往返、MD5 存量兼容、NeedUpgrade 判定、错误密码拒绝、盐随机性。

func TestHashPwdAndCheckPwd(t *testing.T) {
	pwd := "s3cret-密码"
	hashed := HashPwd(pwd)

	if !strings.HasPrefix(hashed, "$2") {
		t.Fatalf("HashPwd 应生成 bcrypt 哈希（$2 前缀），got %q", hashed[:8])
	}
	if !CheckPwd(pwd, hashed) {
		t.Errorf("正确密码校验失败")
	}
	if CheckPwd("wrong", hashed) {
		t.Errorf("错误密码不应通过")
	}
	if NeedUpgrade(hashed) {
		t.Errorf("bcrypt 哈希不应需要升级")
	}
}

func TestCheckPwdLegacyMD5(t *testing.T) {
	pwd := "legacy-pass"
	stored := MD5(pwd) // 存量 MD5 哈希

	if !CheckPwd(pwd, stored) {
		t.Errorf("存量 MD5 应能通过校验（兼容窗口）")
	}
	if CheckPwd("wrong", stored) {
		t.Errorf("错误密码不应通过 MD5 校验")
	}
	if !NeedUpgrade(stored) {
		t.Errorf("MD5 哈希应判定为待升级")
	}
}

func TestHashPwdSaltRandom(t *testing.T) {
	// bcrypt 带随机盐：同一密码两次哈希结果不同（等值匹配不可行的根因）
	h1 := HashPwd("same")
	h2 := HashPwd("same")
	if h1 == h2 {
		t.Errorf("同一密码两次 HashPwd 结果不应相同")
	}
	// 但两者都能通过校验
	if !CheckPwd("same", h1) || !CheckPwd("same", h2) {
		t.Errorf("两个哈希都应校验通过")
	}
}
