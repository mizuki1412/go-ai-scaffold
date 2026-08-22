package framekit

import "sync"

type Decoder struct {
	Bytes []byte // 待处理数据
	sync.RWMutex
	TakeHandler func([]byte) ([]byte, []byte, bool) // return 未处理的bytes、成功出库的bytes、是否结束
	enableRecv  bool
	recvHandler func([]byte, bool)
}

func NewDecoder(initCapacity int, takeHandler func([]byte) ([]byte, []byte, bool)) *Decoder {
	return &Decoder{
		Bytes:       make([]byte, 0, initCapacity),
		TakeHandler: takeHandler,
	}
}

// Take 取一次
func (th *Decoder) Take() ([]byte, bool) {
	th.Lock()
	defer th.Unlock()
	o, r, over := th.TakeHandler(th.Bytes)
	th.Bytes = o
	return r, over
}

// Recv 注册接收模式回调。此后每次 Put 追加数据都会循环 Take，
// 将取出的数据交给 f。
//
// P2 修复：enableRecv/recvHandler 的写读已纳入 Decoder 内嵌锁保护
// （原写读均无锁，与 Put 并发构成数据竞争）。
// 使用约束：必须在首次 Put 之前调用，否则后注册的回调收不到此前已入队的数据。
func (th *Decoder) Recv(f func([]byte, bool)) {
	th.Lock()
	defer th.Unlock()
	th.enableRecv = true
	th.recvHandler = f
}

func (th *Decoder) Put(data []byte) {
	if len(data) > 0 {
		th.Lock()
		th.Bytes = append(th.Bytes, data...)
		th.Unlock()
	}
	// 触发接收模式（P2：快照后调用，避免与 Recv 注册并发时的数据竞争）
	th.RLock()
	recvOn := th.enableRecv
	handler := th.recvHandler
	th.RUnlock()
	if recvOn && handler != nil {
		for {
			d, over := th.Take()
			if over {
				handler(d, over)
				break
			}
			if len(d) == 0 {
				break
			} else {
				handler(d, over)
			}
		}
	}
}
