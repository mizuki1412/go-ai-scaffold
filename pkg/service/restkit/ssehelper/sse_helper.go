package ssehelper

import (
	"sync"

	"github.com/example/go-ai-scaffold/pkg/service/logkit"
	"github.com/example/go-ai-scaffold/pkg/service/restkit/context"
)

// 在线客户端注册表。
// P0 修复：原实现仅用 sync.Once 保证初始化，AddClient/RemoveClient/ToSend
// 对 map 的增删读全部无锁——HTTP 每请求一个 goroutine，并发即触发
// fatal error: concurrent map writes（不可 recover，直接打崩进程）。
// 现统一以 RWMutex 保护；channel 在 RemoveClient 中 close，
// 使 ServiceClient 的 for range 得以退出（原实现永不关闭，goroutine 泄漏）。
var (
	clientChannels map[string]chan string
	mux            sync.RWMutex
)

func init() {
	clientChannels = make(map[string]chan string)
}

// getChan 持读锁获取指定客户端的 channel。
func getChan(clientId string) (chan string, bool) {
	mux.RLock()
	defer mux.RUnlock()
	c, ok := clientChannels[clientId]
	return c, ok
}

func AddClient(clientId string, ctx *context.Context) {
	mux.Lock()
	if _, ok := clientChannels[clientId]; ok {
		mux.Unlock()
		logkit.Info("SSE Client already add: " + clientId)
		return
	}
	// 小缓冲 + ToSend 非阻塞发送：消费方短暂阻塞时不至于卡住发送方
	clientChannels[clientId] = make(chan string, 16)
	mux.Unlock()
	logkit.Info("SSE Client add: " + clientId)
	closeNotify := ctx.Proxy.Request.Context().Done()
	go func() {
		<-closeNotify
		RemoveClient(clientId)
	}()
}

func ServiceClient(clientId string, ctx *context.Context) {
	AddClient(clientId, ctx)
	ch, ok := getChan(clientId)
	if !ok {
		// AddClient 后立即被 RemoveClient（连接瞬断）的边界场景
		return
	}
	// RemoveClient 会 close(ch)，range 随之退出
	for msg := range ch {
		ctx.SendSSE(msg)
	}
}

func RemoveClient(clientId string) {
	mux.Lock()
	defer mux.Unlock()
	if c, ok := clientChannels[clientId]; ok {
		delete(clientChannels, clientId)
		close(c) // 通知 ServiceClient 的 for range 退出，回收 goroutine
		logkit.Info("SSE Client close: " + clientId)
	}
}

// ToSend 向指定客户端推送消息（非阻塞：缓冲满则丢弃并记日志）。
// 发送全程持读锁，与 RemoveClient 的写锁互斥，杜绝 send on closed channel。
func ToSend(clientId string, msg string) {
	mux.RLock()
	defer mux.RUnlock()
	if c, ok := clientChannels[clientId]; ok {
		select {
		case c <- msg:
		default:
			logkit.Error("SSE client buffer full, drop message: " + clientId)
		}
	}
}
