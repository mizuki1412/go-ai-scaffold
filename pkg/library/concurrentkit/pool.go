package concurrentkit

import (
	"container/list"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/example/go-ai-scaffold/pkg/library/c"
)

// Pool goroutine 池（移植自 GoFrame grpool，MIT）：
// 任务 FIFO 排队，按需扩容 worker 至上限，队列取空后 worker 自愈退出；
// 任务 panic 经 c.RecoverFuncWrapper 捕获记录，worker 不死，无需看护协程。
type Pool struct {
	limit  atomic.Int64 // 最大 worker 数，-1 不限
	count  atomic.Int64 // 当前存活 worker 数
	closed atomic.Bool  // 关闭标记，Cas 保证幂等
	mu     sync.Mutex   // 保护 jobs
	jobs   list.List    // FIFO 任务队列
}

var ErrPoolClosed = errors.New("concurrentkit: pool is closed")

func NewPool(limit ...int) *Pool {
	p := &Pool{}
	if len(limit) > 0 && limit[0] > 0 {
		p.limit.Store(int64(limit[0]))
	} else {
		p.limit.Store(-1)
	}
	return p
}

// Add 投递异步任务；池已关闭返回 ErrPoolClosed。
func (p *Pool) Add(f func()) error {
	if p.closed.Load() {
		return ErrPoolClosed
	}
	p.mu.Lock()
	p.jobs.PushBack(f)
	p.mu.Unlock()
	p.checkAndForkWorker()
	return nil
}

// Cap 返回最大 worker 数，-1 表示不限。
func (p *Pool) Cap() int {
	return int(p.limit.Load())
}

// SetCap 动态调整最大 worker 数并返回旧值；<=0 视为不限。
func (p *Pool) SetCap(cap int) int {
	if cap <= 0 {
		cap = -1
	}
	return int(p.limit.Swap(int64(cap)))
}

// Size 返回当前存活 worker 数。
func (p *Pool) Size() int {
	return int(p.count.Load())
}

// Jobs 返回排队中的任务数。
func (p *Pool) Jobs() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.jobs.Len()
}

// ClearJobs 清空排队任务并返回清除数量。
func (p *Pool) ClearJobs() int {
	p.mu.Lock()
	n := p.jobs.Len()
	p.jobs.Init()
	p.mu.Unlock()
	return n
}

// IsClosed 返回池是否已关闭。
func (p *Pool) IsClosed() bool {
	return p.closed.Load()
}

// Close 关闭池：不再接受新任务，存活 worker 随即退出（与 grpool 一致，
// 队列中未执行的任务将滞留；需排干请先等 Jobs()==0 再 Close）。幂等。
func (p *Pool) Close() {
	p.closed.CompareAndSwap(false, true)
}

// checkAndForkWorker CAS 自旋抢名额，抢到才 fork 新 worker。
func (p *Pool) checkAndForkWorker() {
	for {
		n := p.count.Load()
		if limit := p.limit.Load(); limit != -1 && n >= limit {
			return
		}
		if p.count.CompareAndSwap(n, n+1) {
			go p.worker()
			return
		}
	}
}

func (p *Pool) popJob() func() {
	p.mu.Lock()
	e := p.jobs.Front()
	if e == nil {
		p.mu.Unlock()
		return nil
	}
	p.jobs.Remove(e)
	p.mu.Unlock()
	return e.Value.(func())
}

func (p *Pool) worker() {
	dec := int64(-1) // 缩容路径已手动扣减，避免 defer 重复扣
	defer func() { p.count.Add(dec) }()
	for !p.closed.Load() {
		f := p.popJob()
		if f == nil {
			return
		}
		_ = c.RecoverFuncWrapper(f)
		// 缩容：超过上限时当前 worker 直接退役
		n := p.count.Load()
		if limit := p.limit.Load(); limit > 0 && n > limit && p.count.CompareAndSwap(n, n-1) {
			dec = 0
			return
		}
	}
}
