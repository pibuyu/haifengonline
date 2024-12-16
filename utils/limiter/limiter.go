package limiter

import (
	"golang.org/x/time/rate"
	"sync"
	"time"
)

type Limiters struct {
	limiters *sync.Map //并发安全的map
}
type Limiter struct {
	limiter *rate.Limiter
	lastGet time.Time
	key     string
}

var GlobalLimiters = &Limiters{
	limiters: &sync.Map{},
}
var once = sync.Once{}

// NewLimiter 获取limiter对象
func NewLimiter(r rate.Limit, b int, key string) *Limiter {
	once.Do(func() { //启动定时清理过期的限流器
		go GlobalLimiters.clearLimiter()
	})
	keyLimiter := GlobalLimiters.getLimiter(r, b, key)
	return keyLimiter
}
func (l *Limiter) Allow() bool {
	l.lastGet = time.Now()
	return l.limiter.Allow()
}

// r:向桶中放token的速率；b：令牌桶的大小；key：可以对服务的id/ip地址进行限制
func (ls *Limiters) getLimiter(r rate.Limit, b int, key string) *Limiter {
	limiter, ok := ls.limiters.Load(key)
	if ok {
		return limiter.(*Limiter)
	}

	l := &Limiter{
		limiter: rate.NewLimiter(r, b),
		lastGet: time.Now(),
		key:     key,
	}
	ls.limiters.Store(key, l)
	return l
}

// 清除过期的限流器
func (ls *Limiters) clearLimiter() {
	for {
		time.Sleep(1 * time.Minute)
		ls.limiters.Range(func(key, value interface{}) bool {
			//超过1分钟
			if time.Now().Unix()-value.(*Limiter).lastGet.Unix() > 60 {
				ls.limiters.Delete(key)
			}
			return true
		})
	}
}
