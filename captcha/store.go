package captcha

import (
	"time"

	"github.com/penndev/gopkg/ttlmap"
)

type ttlStore struct {
	m *ttlmap.Map
}

// NewTTLStore 基于 ttlmap 的单机存储。
func NewTTLStore() Storage {
	return &ttlStore{m: ttlmap.New()}
}

func (s *ttlStore) Set(id string, value any, ttl time.Duration) {
	s.m.Set(id, value, ttl)
}

func (s *ttlStore) Get(id string) (any, bool) {
	return s.m.Get(id)
}

func (s *ttlStore) Delete(id string) {
	s.m.Delete(id)
}
