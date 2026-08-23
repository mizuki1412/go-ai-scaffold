package concurrentkit

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestKeyLockMutexExclusion(t *testing.T) {
	kl := NewKeyLock()
	var counter int64
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			kl.LockFunc("user:1", func() {
				counter++
			})
		}()
	}
	wg.Wait()
	if counter != 100 {
		t.Fatalf("LockFunc 应串行化同 key 并发, counter=%d", counter)
	}
	kl.Remove("user:1")
	kl.Clear()
}

func TestPoolCapAndCompletion(t *testing.T) {
	p := NewPool(3)
	var running, maxRunning, done int64
	var mu sync.Mutex
	for i := 0; i < 50; i++ {
		if err := p.Add(func() {
			cur := atomic.AddInt64(&running, 1)
			mu.Lock()
			if cur > maxRunning {
				maxRunning = cur
			}
			mu.Unlock()
			time.Sleep(time.Millisecond)
			atomic.AddInt64(&done, 1)
			atomic.AddInt64(&running, -1)
		}); err != nil {
			t.Fatalf("Add 不应失败: %v", err)
		}
	}
	for p.Jobs() > 0 || p.Size() > 0 {
		time.Sleep(time.Millisecond)
	}
	mu.Lock()
	max := maxRunning
	mu.Unlock()
	if done != 50 {
		t.Fatalf("全部任务应执行完成, done=%d", done)
	}
	if max > 3 {
		t.Fatalf("并发数不应超过 Cap=3, got %d", max)
	}

	p.Close()
	if err := p.Add(func() {}); err != ErrPoolClosed {
		t.Fatalf("关闭后 Add 应返回 ErrPoolClosed, got %v", err)
	}
}
