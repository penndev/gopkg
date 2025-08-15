package ttlmap

import (
	"container/heap"
	"context"
	"time"
)

type ttlEntry struct {
	key      string
	expireAt time.Time
}

// 小顶堆实现
type heapList []ttlEntry

func (h heapList) Len() int           { return len(h) }
func (h heapList) Less(i, j int) bool { return h[i].expireAt.Before(h[j].expireAt) }
func (h heapList) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *heapList) Push(x any)        { *h = append(*h, x.(ttlEntry)) }
func (h *heapList) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// GC 协程，精确调度过期
func (m *Map) gcLoop() {
	for {
		// 非空休眠
		m.heapCond.L.Lock()
		for m.heap.Len() == 0 {
			m.heapCond.Wait()
		}

		// 找到最近过期时间
		entry := m.heap[0]
		now := time.Now()
		if entry.expireAt.After(now) {
			// 还没到期，等待一会儿
			timeout := false
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				timer := time.NewTimer(entry.expireAt.Sub(now))
				defer timer.Stop()
				select {
				case <-timer.C:
					m.heapCond.Signal()
					timeout = true
				case <-ctx.Done():
					return
				}
			}()
			m.heapCond.Wait()
			cancel() // 放置go runtime泄露
			if !timeout {
				m.heapCond.L.Unlock() //新插入ttl数据
				continue
			}
		}
		// 过期了，从堆中移除
		heap.Pop(&m.heap)
		m.Delete(entry.key)
		m.heapCond.L.Unlock()
	}
}
