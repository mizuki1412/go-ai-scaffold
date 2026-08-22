package cronkit

import (
	"sync"

	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/library/c"
	"github.com/example/go-ai-scaffold/pkg/library/timekit"
	"github.com/robfig/cron/v3"
)

var (
	scheduler     *cron.Cron
	schedulerOnce sync.Once
	// P1 修复：pool 原为 nil map（AddPool 一调即 assignment to entry in nil map panic），
	// 且增删读无锁。现在包初始化时创建，所有读写以互斥锁保护。
	pool    = map[string]*cron.Cron{}
	poolMux sync.Mutex
)

// Scheduler 返回默认调度器（懒初始化）。
// P1 修复：原 `if scheduler == nil` 懒初始化在并发调用下可能创建多个调度器，改 sync.Once。
func Scheduler() *cron.Cron {
	schedulerOnce.Do(func() {
		scheduler = NewScheduler()
	})
	return scheduler
}

func NewScheduler() *cron.Cron {
	return cron.New(cron.WithSeconds(), cron.WithLocation(timekit.GetLocation()))
}

func AddPool(key string, cr *cron.Cron) {
	poolMux.Lock()
	defer poolMux.Unlock()
	// 替换前停掉旧调度器（不能复用 RemovePool——锁不可重入会死锁）
	if old, ok := pool[key]; ok {
		old.Stop()
	}
	pool[key] = cr
}

func RemovePool(key string) {
	poolMux.Lock()
	defer poolMux.Unlock()
	if v, ok := pool[key]; ok {
		v.Stop()
		delete(pool, key)
	}
}

// AddFunc 给默认的scheduler add func， 封装上recover
func AddFunc(spec string, fun func()) {
	_, err := Scheduler().AddFunc(spec, func() {
		_ = c.RecoverFuncWrapper(fun)
	})
	if err != nil {
		panic(exception.New(err.Error()))
	}
}
