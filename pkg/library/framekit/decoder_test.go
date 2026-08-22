package framekit

import (
	"sync"
	"testing"
)

// P2-12 起步单测：Decoder 的 Take/Put/Recv 语义与 P2-1 并发修复。

// byteDecoder 每次取一个字节；收到 "E" 表示结束。
func byteDecoder(b []byte) ([]byte, []byte, bool) {
	if len(b) == 0 {
		return b, nil, false
	}
	first := b[:1]
	rest := b[1:]
	if first[0] == 'E' {
		return rest, first, true
	}
	return rest, first, false
}

func TestDecoderTake(t *testing.T) {
	d := NewDecoder(16, byteDecoder)
	d.Put([]byte("abc"))

	var got []byte
	for {
		r, over := d.Take()
		got = append(got, r...)
		if len(r) == 0 || over {
			break
		}
	}
	if string(got) != "abc" {
		t.Errorf("Take 拼接 = %q, want %q", got, "abc")
	}
	// 缓冲应已耗尽
	if r, _ := d.Take(); len(r) != 0 {
		t.Errorf("耗尽后 Take 应无数据，got %q", r)
	}
}

func TestDecoderRecv(t *testing.T) {
	d := NewDecoder(16, byteDecoder)

	var mu sync.Mutex
	var frames []string
	var overs int
	d.Recv(func(b []byte, over bool) {
		mu.Lock()
		defer mu.Unlock()
		frames = append(frames, string(b))
		if over {
			overs++
		}
	})
	// 先 Recv 后 Put：数据应逐帧送达回调
	d.Put([]byte("ab"))
	d.Put([]byte("cE")) // E 触发 over

	mu.Lock()
	defer mu.Unlock()
	want := []string{"a", "b", "c", "E"}
	if len(frames) != len(want) {
		t.Fatalf("收到帧 %v, want %v", frames, want)
	}
	for i, w := range want {
		if frames[i] != w {
			t.Errorf("frames[%d] = %q, want %q", i, frames[i], w)
		}
	}
	if overs != 1 {
		t.Errorf("over 次数 = %d, want 1", overs)
	}
}

func TestDecoderPutWithoutRecv(t *testing.T) {
	// 未注册 Recv 时 Put 只入队，不回调（不 panic）
	d := NewDecoder(16, byteDecoder)
	d.Put([]byte("xy"))
	if len(d.Bytes) != 2 {
		t.Errorf("缓冲长度 = %d, want 2", len(d.Bytes))
	}
}

func TestDecoderConcurrentPut(t *testing.T) {
	// 并发 Put（配合 -race 运行，验证 P2-1 的锁保护）
	d := NewDecoder(64, byteDecoder)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				d.Put([]byte("z"))
			}
		}()
	}
	wg.Wait()
	if len(d.Bytes) != 8*50 {
		t.Errorf("缓冲长度 = %d, want %d", len(d.Bytes), 8*50)
	}
}
