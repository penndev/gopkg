package captcha

import (
	"time"

	"github.com/penndev/gopkg/ttlmap"
)

// Store 默认单机存储；集群请自行替换。
var Store ttlmap.Map = *ttlmap.New()

// StoreAlive 验证码默认存活时间。
var StoreAlive = 5 * time.Minute
