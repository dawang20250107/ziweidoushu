package httpapi

import (
	"container/list"
	"sync"
)

// lruCache 分片 LRU 缓存(命盘结果)。
//
// 排盘为纯函数,同一生辰重复请求(名人盘、分享回看、小程序回访)直接命中,
// 16 路分片将锁竞争摊薄到可忽略。
type lruCache struct {
	shards [16]lruShard
}

type lruShard struct {
	mu    sync.Mutex
	cap   int
	items map[string]*list.Element
	order *list.List // 头部最新
}

type lruEntry struct {
	key string
	val any
}

func newLRUCache(capacity int) *lruCache {
	if capacity <= 0 {
		return nil
	}
	per := capacity / 16
	if per < 8 {
		per = 8
	}
	c := &lruCache{}
	for i := range c.shards {
		c.shards[i] = lruShard{cap: per, items: make(map[string]*list.Element), order: list.New()}
	}
	return c
}

func (c *lruCache) shard(key string) *lruShard {
	var h uint32 = 2166136261
	for i := 0; i < len(key); i++ {
		h ^= uint32(key[i])
		h *= 16777619
	}
	return &c.shards[h%16]
}

// Get 命中返回值与 true。
func (c *lruCache) Get(key string) (any, bool) {
	if c == nil {
		return nil, false
	}
	sh := c.shard(key)
	sh.mu.Lock()
	defer sh.mu.Unlock()
	if el, ok := sh.items[key]; ok {
		sh.order.MoveToFront(el)
		return el.Value.(*lruEntry).val, true
	}
	return nil, false
}

// Set 写入并按容量淘汰最久未用。
func (c *lruCache) Set(key string, val any) {
	if c == nil {
		return
	}
	sh := c.shard(key)
	sh.mu.Lock()
	defer sh.mu.Unlock()
	if el, ok := sh.items[key]; ok {
		el.Value.(*lruEntry).val = val
		sh.order.MoveToFront(el)
		return
	}
	el := sh.order.PushFront(&lruEntry{key: key, val: val})
	sh.items[key] = el
	if sh.order.Len() > sh.cap {
		oldest := sh.order.Back()
		if oldest != nil {
			sh.order.Remove(oldest)
			delete(sh.items, oldest.Value.(*lruEntry).key)
		}
	}
}
