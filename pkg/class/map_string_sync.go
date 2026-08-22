package class

import (
	"database/sql/driver"
	"github.com/example/go-ai-scaffold/pkg/class/exception"
	"github.com/example/go-ai-scaffold/pkg/class/utils"
	"github.com/example/go-ai-scaffold/pkg/library/jsonkit"
	"github.com/example/go-ai-scaffold/pkg/library/mapkit"
	"sync"
)

// MapStringSync 同时继承scan和value方法。
// 注意：内嵌 sync.RWMutex，因此本类型的值不可拷贝（go vet copylocks）——
// model 字段请使用 *MapStringSync 指针形式，构造请用 NewMapStringSync/NMapStringSync。
type MapStringSync struct {
	sync.RWMutex
	Map   map[string]any
	Valid bool
}

// MarshalJSON 序列化时持读锁，避免与并发写冲突（原 todo 已解决）。
func (th *MapStringSync) MarshalJSON() ([]byte, error) {
	th.RLock()
	defer th.RUnlock()
	if th.Valid {
		return jsonkit.Marshal(th.Map)
	}
	return []byte("null"), nil
}

func (th *MapStringSync) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		th.Lock()
		defer th.Unlock()
		th.Valid = false
		return nil
	}
	var s map[string]any
	if err := jsonkit.Unmarshal(data, &s); err != nil {
		return err
	}
	th.Lock()
	defer th.Unlock()
	th.Valid = true
	th.Map = s
	return nil
}

// Scan implements the Scanner interface.
func (th *MapStringSync) Scan(value any) error {
	if value == nil {
		th.Lock()
		defer th.Unlock()
		th.Map, th.Valid = nil, false
		return nil
	}
	val := utils.TransScanValue2String(value)
	m := jsonkit.ParseMap(val)
	th.Lock()
	defer th.Unlock()
	th.Valid = true
	th.Map = m
	return nil
}

// Value implements the driver Valuer interface.
// P0 修复：原为值接收者，每次调用都会拷贝内嵌的 sync.RWMutex（go vet:
// Value passes lock by value），锁状态作废且与并发 Lock 构成数据竞争。
func (th *MapStringSync) Value() (driver.Value, error) {
	th.RLock()
	defer th.RUnlock()
	if !th.Valid || th.Map == nil {
		return nil, nil
	}
	return jsonkit.ToString(th.Map), nil
}

// IsValid P0 修复：同 Value()，改为指针接收者并持读锁。
func (th *MapStringSync) IsValid() bool {
	th.RLock()
	defer th.RUnlock()
	return th.Valid
}

// NewMapStringSync 构造并返回指针。
// P0 修复：原实现按值返回，把内嵌的 sync.RWMutex 一并拷出（go vet:
// return copies lock value）。需要值语义请改用不带锁的 MapString。
func NewMapStringSync(val ...any) *MapStringSync {
	th := &MapStringSync{}
	if len(val) > 0 {
		th.Set(val[0])
	} else {
		th.Set(map[string]any{})
	}
	return th
}

// NMapStringSync 与 NewMapStringSync 等价（均返回指针），保留以兼容既有调用方。
func NMapStringSync(val ...any) *MapStringSync {
	return NewMapStringSync(val...)
}

// Set 设置 map 内容。
// P0 修复：移除 MapStringSync 值形式入参——值断言同样会拷贝内嵌锁；
// 仅接受指针形式与其他无锁类型。
func (th *MapStringSync) Set(val any) {
	th.Lock()
	defer th.Unlock()
	switch v := val.(type) {
	case *MapStringSync:
		if v == nil {
			th.Map = map[string]any{}
			th.Valid = false
		} else {
			th.copyFrom(v.Map, v.Valid)
		}
	case MapString:
		th.copyFrom(v.Map, v.Valid)
	case *MapString:
		if v == nil {
			th.Map = map[string]any{}
			th.Valid = false
		} else {
			th.copyFrom(v.Map, v.Valid)
		}
	case map[string]any:
		th.Map = v
		th.Valid = true
	default:
		panic(exception.New("class.MapStringSync set error"))
	}
}

// copyFrom 在已持写锁的前提下填充字段。
func (th *MapStringSync) copyFrom(m map[string]any, valid bool) {
	if m == nil {
		th.Map = map[string]any{}
	} else {
		th.Map = m
	}
	th.Valid = valid
}

func (th *MapStringSync) PutAll(val map[string]any) {
	th.Lock()
	defer th.Unlock()
	if th.Map == nil {
		th.Map = map[string]any{}
	}
	mapkit.PutAll(th.Map, val)
	th.Valid = true
}

func (th *MapStringSync) PutIfAbsent(key string, val any) {
	th.Lock()
	defer th.Unlock()
	if th.Map == nil {
		th.Map = map[string]any{}
	}
	if _, ok := th.Map[key]; !ok {
		th.Map[key] = val
	}
	th.Valid = true
}

func (th *MapStringSync) Put(key string, val any) {
	th.Lock()
	defer th.Unlock()
	if th.Map == nil {
		th.Map = map[string]any{}
	}
	th.Map[key] = val
	th.Valid = true
}

func (th *MapStringSync) Remove() {
	th.Lock()
	defer th.Unlock()
	th.Valid = false
	clear(th.Map)
}

func (th *MapStringSync) Delete(key string) {
	th.Lock()
	defer th.Unlock()
	delete(th.Map, key)
}

func (th *MapStringSync) IsEmpty() bool {
	if !th.Valid {
		return true
	}
	if len(th.Map) == 0 {
		return true
	}
	return false
}

func (th *MapStringSync) Contains(key string) bool {
	th.RLock()
	defer th.RUnlock()
	v, ok := th.Map[key]
	if ok {
		return v != nil
	}
	return ok
}

func (th *MapStringSync) Get(key string) any {
	th.RLock()
	defer th.RUnlock()
	v, _ := th.Map[key]
	return v
}

func (th *MapStringSync) Entries() map[string]any {
	th.RLock()
	defer th.RUnlock()
	m := map[string]any{}
	for k, v := range th.Map {
		if vv, ok := v.(*MapStringSync); ok {
			m[k] = vv.Entries()
		} else {
			m[k] = v
		}
	}
	return m
}
