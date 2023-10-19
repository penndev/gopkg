package ttlmap

import (
	"container/list"
	"log"
	"sync"
	"time"
)

type ttldelay struct {
	key     any
	expried int64
}

type TTLMap struct {
	sync.Map
	expried *list.List
	delay   time.Duration
}

func (m *TTLMap) Set(key, val any) {
	exp := time.Now().Add(m.delay)
	m.Store(key, val)
	m.expried.PushBack(ttldelay{
		key:     key,
		expried: exp.Unix(),
	})
}

func (m *TTLMap) initTTL() {
	for {
		if m.expried.Len() == 0 {
			time.Sleep(m.delay)
		} else {
			elements := m.expried.Front()
			if td, ok := elements.Value.(ttldelay); ok {
				// 判断是否到期。
				now := time.Now().Unix()
				if now < td.expried {
					time.Sleep(time.Duration(td.expried-now) * time.Second)
				}
				if v, ok := m.Load(td.key); ok {
					log.Println("load -> " + v.(string))
				}
				m.Delete(td.key)
				log.Println("delete -> " + td.key.(string))
			}
			m.expried.Remove(elements)
		}
	}
}

func New(delay time.Duration) *TTLMap {
	ttlmap := &TTLMap{
		delay:   delay,
		expried: list.New(),
	}
	go ttlmap.initTTL()
	return ttlmap
}
