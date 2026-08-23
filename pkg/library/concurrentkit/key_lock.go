package concurrentkit

import "sync"

// KeyLock 按 key 加锁的内存锁（移植自 GoFrame gmlock，MIT）。
// 注意：锁不自动过期，长期不用的 key 需手动 Remove 释放。
type KeyLock struct {
	mu sync.RWMutex
	m  map[string]*sync.RWMutex
}

func NewKeyLock() *KeyLock {
	return &KeyLock{m: make(map[string]*sync.RWMutex)}
}

// Lock 对 key 加写锁，已被占则阻塞。
func (l *KeyLock) Lock(key string) {
	l.getOrNewMutex(key).Lock()
}

// TryLock 尝试对 key 加写锁，成功返回 true。
func (l *KeyLock) TryLock(key string) bool {
	return l.getOrNewMutex(key).TryLock()
}

// Unlock 释放 key 的写锁。
func (l *KeyLock) Unlock(key string) {
	if v := l.getMutex(key); v != nil {
		v.Unlock()
	}
}

// RLock 对 key 加读锁，存在写锁时阻塞。
func (l *KeyLock) RLock(key string) {
	l.getOrNewMutex(key).RLock()
}

// TryRLock 尝试对 key 加读锁，存在写锁时返回 false。
func (l *KeyLock) TryRLock(key string) bool {
	return l.getOrNewMutex(key).TryRLock()
}

// RUnlock 释放 key 的读锁。
func (l *KeyLock) RUnlock(key string) {
	if v := l.getMutex(key); v != nil {
		v.RUnlock()
	}
}

// LockFunc 写锁执行 f，结束自动释放（单飞：同 key 并发串行化）。
func (l *KeyLock) LockFunc(key string, f func()) {
	l.Lock(key)
	defer l.Unlock(key)
	f()
}

// RLockFunc 读锁执行 f，结束自动释放。
func (l *KeyLock) RLockFunc(key string, f func()) {
	l.RLock(key)
	defer l.RUnlock(key)
	f()
}

// TryLockFunc 尝试写锁执行 f；拿不到锁不执行并返回 false。
func (l *KeyLock) TryLockFunc(key string, f func()) bool {
	if l.TryLock(key) {
		defer l.Unlock(key)
		f()
		return true
	}
	return false
}

// TryRLockFunc 尝试读锁执行 f；拿不到锁不执行并返回 false。
func (l *KeyLock) TryRLockFunc(key string, f func()) bool {
	if l.TryRLock(key) {
		defer l.RUnlock(key)
		f()
		return true
	}
	return false
}

// Remove 移除 key 对应的锁（需确认无人在用时才可调用）。
func (l *KeyLock) Remove(key string) {
	l.mu.Lock()
	delete(l.m, key)
	l.mu.Unlock()
}

// Clear 清空所有 key 的锁。
func (l *KeyLock) Clear() {
	l.mu.Lock()
	l.m = make(map[string]*sync.RWMutex)
	l.mu.Unlock()
}

// getOrNewMutex 双检查获取或创建 key 的锁，创建路径仅一个 goroutine 执行。
func (l *KeyLock) getOrNewMutex(key string) *sync.RWMutex {
	if v := l.getMutex(key); v != nil {
		return v
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	v, ok := l.m[key]
	if !ok {
		v = &sync.RWMutex{}
		l.m[key] = v
	}
	return v
}

func (l *KeyLock) getMutex(key string) *sync.RWMutex {
	l.mu.RLock()
	v := l.m[key]
	l.mu.RUnlock()
	return v
}
