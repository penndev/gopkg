package ttlmap

import (
	"container/list"
	"sync"
	"time"
)

type ttlDelay struct {
	key     any
	expired int64
}

type TTLMap struct {
	sync.Map               //数据真实存储map
	list     *list.List    //过期列表
	delay    time.Duration //存活时间
}

func (m *TTLMap) GetDelay() time.Duration {
	return m.delay
}

// 设置值并保存
func (m *TTLMap) Set(key, val any) {
	exp := time.Now().Add(m.delay)
	m.Store(key, val)
	m.list.PushBack(ttlDelay{
		key:     key,
		expired: exp.Unix(),
	})
}

// 开启时间轮
func (m *TTLMap) initTTL() {
	for {
		if m.list.Len() == 0 {
			time.Sleep(m.delay)
		} else {
			elements := m.list.Front()
			if td, ok := elements.Value.(ttlDelay); ok {
				// 判断是否到期。
				now := time.Now().Unix()
				if now < td.expired {
					time.Sleep(time.Duration(td.expired-now) * time.Second)
				}
				if _, ok := m.Load(td.key); ok {
					m.Delete(td.key)
				}
			}
			m.list.Remove(elements)
		}
	}
}

// delay time.Duration后删除
// > 最少1秒,低于1秒重置为1秒。
func New(delay time.Duration) *TTLMap {
	if delay < 1*time.Second {
		delay = 1 * time.Second
	}
	ttlMap := &TTLMap{
		delay: delay,
		list:  list.New(),
	}
	go ttlMap.initTTL()
	return ttlMap
}
