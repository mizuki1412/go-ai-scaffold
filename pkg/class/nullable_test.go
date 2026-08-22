package class

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
)

// P2-12 起步单测：可空类型的 JSON 序列化 / Set 语义 / driver 接口。
// 表驱动 + 并发用例（配合 go test -race 可暴露 MapStringSync 的锁拷贝类问题）。

func TestStringJSON(t *testing.T) {
	tests := []struct {
		name string
		in   String
		want string
	}{
		{"有效值序列化", NewString("abc"), `"abc"`},
		{"无效值序列化为null", NewString(), `null`},
		{"空串也是有效值", NewString(""), `""`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.in)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}
			if string(b) != tt.want {
				t.Errorf("marshal = %s, want %s", b, tt.want)
			}
		})
	}
}

func TestStringUnmarshalJSON(t *testing.T) {
	var s String
	if err := json.Unmarshal([]byte(`"hello"`), &s); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !s.IsValid() || s.String != "hello" {
		t.Errorf("got valid=%v val=%q, want valid=true val=%q", s.IsValid(), s.String, "hello")
	}
	if err := json.Unmarshal([]byte(`null`), &s); err != nil {
		t.Fatalf("unmarshal null error: %v", err)
	}
	if s.IsValid() {
		t.Errorf("null 后应无效，got valid=true")
	}
}

func TestStringSet(t *testing.T) {
	s := NewString(int64(42))
	if !s.IsValid() || s.String != "42" {
		t.Errorf("Set(int64) got %q, want %q", s.String, "42")
	}
	// Set(String) 拷贝有效性语义（无效拷过去仍是无效）
	invalid := NewString()
	s.Set(invalid)
	if s.IsValid() {
		t.Errorf("Set(无效String) 应保持无效")
	}
	s.Remove()
	if s.IsValid() || s.String != "" {
		t.Errorf("Remove 后应无效且清空")
	}
}

func TestInt64JSON(t *testing.T) {
	// int64 序列化为字符串，防止 JS 数值溢出
	b, err := json.Marshal(NewInt64(9007199254740993))
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if string(b) != `"9007199254740993"` {
		t.Errorf("marshal = %s, want \"9007199254740993\"", b)
	}
	var i Int64
	if err := json.Unmarshal([]byte(`"123"`), &i); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !i.IsValid() || i.Int64 != 123 {
		t.Errorf("got valid=%v val=%d, want valid=true val=123", i.IsValid(), i.Int64)
	}
	if err := json.Unmarshal([]byte(`null`), &i); err != nil {
		t.Fatalf("unmarshal null error: %v", err)
	}
	if i.IsValid() {
		t.Errorf("null 后应无效")
	}
}

func TestMapStringSyncBasics(t *testing.T) {
	m := NewMapStringSync()
	if !m.IsEmpty() {
		t.Errorf("新构造应为空")
	}
	m.Put("a", 1)
	if m.IsEmpty() || !m.Contains("a") {
		t.Errorf("Put 后应为非空且 Contains")
	}
	if got := m.Get("a"); got != 1 {
		t.Errorf("Get(a) = %v, want 1", got)
	}
	m.PutIfAbsent("a", 2)
	if got := m.Get("a"); got != 1 {
		t.Errorf("PutIfAbsent 不应覆盖已有 key: got %v, want 1", got)
	}
	m.Delete("a")
	if m.Contains("a") {
		t.Errorf("Delete 后不应 Contains")
	}
	m.Remove()
	if m.IsValid() {
		t.Errorf("Remove 后应无效")
	}
}

func TestMapStringSyncJSONAndDriver(t *testing.T) {
	m := NewMapStringSync(map[string]any{"a": 1})
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatalf("回读 error: %v", err)
	}
	if v, ok := out["a"].(float64); !ok || v != 1 {
		t.Errorf("marshal 内容错误: %s", b)
	}

	// 无效值序列化为 null
	invalid := NewMapStringSync()
	invalid.Remove()
	b, err = json.Marshal(invalid)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	if string(b) != "null" {
		t.Errorf("无效值 marshal = %s, want null", b)
	}

	// driver.Valuer / sql.Scanner 往返
	m2 := NewMapStringSync()
	if err := m2.Scan(`{"k":"v"}`); err != nil {
		t.Fatalf("scan error: %v", err)
	}
	v, err := m2.Value()
	if err != nil {
		t.Fatalf("value error: %v", err)
	}
	s, ok := v.(string)
	if !ok || !strings.Contains(s, `"k"`) {
		t.Errorf("Value() = %v, want 含 k 的 JSON 串", v)
	}
}

func TestMapStringSyncConcurrent(t *testing.T) {
	// 并发读写（配合 -race 运行）
	m := NewMapStringSync()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				m.Put("key", j)
				_ = m.Get("key")
				_ = m.Contains("key")
			}
		}()
	}
	wg.Wait()
	if !m.Contains("key") {
		t.Errorf("并发写后应存在 key")
	}
}
