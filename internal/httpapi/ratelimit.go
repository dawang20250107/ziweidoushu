package httpapi

import (
	"sync"
	"time"
)

// ipLimiter 单 IP 令牌桶限流器。
//
// 32 路分片降低锁竞争;后台 janitor 定期回收闲置桶,防止内存随 IP 数无界增长。
type ipLimiter struct {
	rps    float64
	burst  float64
	shards [32]limiterShard
	stop   chan struct{}
}

type limiterShard struct {
	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	tokens float64
	last   time.Time
}

func newIPLimiter(rps float64, burst int) *ipLimiter {
	if rps <= 0 {
		return nil
	}
	if burst <= 0 {
		burst = int(rps)
	}
	l := &ipLimiter{rps: rps, burst: float64(burst), stop: make(chan struct{})}
	for i := range l.shards {
		l.shards[i].buckets = make(map[string]*bucket)
	}
	go l.janitor()
	return l
}

func (l *ipLimiter) shard(key string) *limiterShard {
	var h uint32 = 2166136261
	for i := 0; i < len(key); i++ {
		h ^= uint32(key[i])
		h *= 16777619
	}
	return &l.shards[h%32]
}

// allow 是否放行一次请求。
func (l *ipLimiter) allow(ip string) bool {
	now := time.Now()
	sh := l.shard(ip)
	sh.mu.Lock()
	defer sh.mu.Unlock()
	b, ok := sh.buckets[ip]
	if !ok {
		sh.buckets[ip] = &bucket{tokens: l.burst - 1, last: now}
		return true
	}
	elapsed := now.Sub(b.last).Seconds()
	b.tokens += elapsed * l.rps
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.last = now
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// janitor 每分钟回收 10 分钟未活动的桶。
func (l *ipLimiter) janitor() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-l.stop:
			return
		case <-ticker.C:
			cutoff := time.Now().Add(-10 * time.Minute)
			for i := range l.shards {
				sh := &l.shards[i]
				sh.mu.Lock()
				for ip, b := range sh.buckets {
					if b.last.Before(cutoff) {
						delete(sh.buckets, ip)
					}
				}
				sh.mu.Unlock()
			}
		}
	}
}

func (l *ipLimiter) close() {
	if l != nil {
		close(l.stop)
	}
}
