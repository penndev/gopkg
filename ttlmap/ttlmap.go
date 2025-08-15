package ttlmap

import (
	"container/heap"
	"sync"
	"time"
)

type item struct {
	value    any
	expireAt time.Time
}

// Map 带 TTL 和精细 GC 调度
type Map struct {
	sync.Map          // 数据存储
	heap     heapList // ttl堆排序
	heapCond *sync.Cond
}

// 设置键值和 TTL
func (m *Map) Set(key string, value any, ttl time.Duration) {
	if ttl <= 0 { // 持久化存储
		ttl = time.Hour * 24
	}
	expireAt := time.Now().Add(ttl)
	m.Store(key, item{value, expireAt})
	m.heapCond.L.Lock()
	heap.Push(&m.heap, ttlEntry{key, expireAt})
	m.heapCond.L.Unlock()
	m.heapCond.Signal()
}

func (m *Map) Get(key string) (any, bool) {
	v, ok := m.Load(key)
	if !ok {
		return nil, false
	}
	it := v.(item)
	if time.Now().After(it.expireAt) {
		m.Delete(key)
		return nil, false
	}
	return it.value, true
}

// 构造器
func New() *Map {
	ttlMap := &Map{}
	ttlMap.heapCond = sync.NewCond(&sync.Mutex{})
	ttlMap.heap = make(heapList, 0)
	go ttlMap.gcLoop()
	return ttlMap
}
