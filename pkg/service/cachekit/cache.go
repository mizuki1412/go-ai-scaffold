package cachekit

import (
	"context"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/library/c"
	"github.com/example/go-ai-scaffold/pkg/service/rediskit"
)

var _cache *ristretto.Cache[string, string]
var once sync.Once

func _getCache() {
	if _cache == nil {
		once.Do(func() {
			cache, err := ristretto.NewCache(&ristretto.Config[string, string]{
				NumCounters: 1e7,     // number of keys to track frequency of (10M).
				MaxCost:     1 << 30, // maximum cost of cache (1GB).
				BufferItems: 64,      // number of keys per Get buffer.
			})
			if err != nil {
				panic(exception.New(err.Error()))
			}
			_cache = cache
		})
	}
}

type Param struct {
	Ttl  time.Duration
	Cost int64
	// 如果存在redis配置，将从redis操作，本地cache只是第二顺序处理
	// 控制忽略redis
	IgnoreRedis bool
}

func _handleParam(ps []*Param) *Param {
	_getCache()
	var p *Param
	if len(ps) == 0 {
		p = nil
	} else {
		p = ps[0]
	}
	if p == nil {
		p = &Param{}
	}
	return p
}

func Set(key string, value string, ps ...*Param) {
	p := _handleParam(ps)
	if rediskit.HasConfig() && !p.IgnoreRedis {
		_ = c.RecoverFuncWrapper(func() {
			rediskit.Set(context.Background(), rediskit.GetKeyWithPrefix(key), value, p.Ttl)
		})
	}
	// 同时也存入cache
	var res bool
	if p.Ttl > 0 {
		res = _cache.SetWithTTL(key, value, p.Cost, p.Ttl)
	} else {
		res = _cache.Set(key, value, p.Cost)
	}
	if !res {
		panic(exception.New("cache failed: " + key))
	}
}

// Get 读取缓存。
// P0 修复：配置了 redis 且未忽略时以 redis 为准（多实例部署下，登录态等共享数据
// 必须由 redis 判定存在性——本地副本既看不到其他实例的写入，也看不到删除/登出）；
// 未配置 redis 或 IgnoreRedis 时只读本地 ristretto。
// 原实现 `r, _ = _cache.Get(key)` 会无条件覆盖 redis 命中值，导致本地 miss 时
// 恒返回空串（多实例登录态校验必然失败），redis 层形同虚设。
func Get(key string, ps ...*Param) string {
	p := _handleParam(ps)
	if rediskit.HasConfig() && !p.IgnoreRedis {
		return rediskit.Get(context.Background(), rediskit.GetKeyWithPrefix(key), "")
	}
	r, _ := _cache.Get(key)
	return r
}

func Del(key string, ps ...*Param) {
	p := _handleParam(ps)
	if rediskit.HasConfig() && !p.IgnoreRedis {
		rediskit.Del(context.Background(), rediskit.GetKeyWithPrefix(key))
	}
	_cache.Del(key)
}

func Renew(key string, ps ...*Param) {
	p := _handleParam(ps)
	if p.Ttl <= 0 {
		return
	}
	if rediskit.HasConfig() && !p.IgnoreRedis {
		rediskit.Expire(context.Background(), rediskit.GetKeyWithPrefix(key), p.Ttl)
	}
	v, ok := _cache.Get(key)
	if ok {
		_cache.SetWithTTL(key, v, p.Cost, p.Ttl)
	}
}
