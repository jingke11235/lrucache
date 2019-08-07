// Package simplelru implements a not thread safe lru cache
package simplelru

import (
	"container/list"
	"time"
)


const (
	NoLimitSize = 0
	NoLimitTTL = 0
)

type EvictCallback func(k, v interface{})

type LRU struct {
	size int

	ttl time.Duration

	cache map[interface{}]*list.Element

	evictList *list.List

	onEvicted EvictCallback
}

type entry struct {
	key   interface{}
	value interface{}
	updatedAt time.Time
}

func NewLRU(size int, ttl time.Duration, onEvict EvictCallback) (*LRU, error) {

	if size <= NoLimitSize {
		size = NoLimitSize
	}
	if ttl <= NoLimitTTL {
		ttl = NoLimitTTL
	}

	return &LRU{
		size:      size,
		ttl: ttl,
		cache:     make(map[interface{}]*list.Element),
		evictList: list.New(),
		onEvicted: onEvict,
	}, nil
}

// Add if not exit - if exited update
func (c *LRU) Set(k,v interface{}) {

	if k == nil || v == nil {
		return
	}

	e := &entry{
		key: k,
		value:v,
		updatedAt:time.Now(),
	}

	if item, ok := c.cache[k]; ok {
		item.Value = e
		c.evictList.MoveToFront(item)
	} else {
		c.cache[k] = item
	}

	if c.size != NoLimitSize && c.evictList.Len() > c.size {
		c.removeOldest()
	}

	return
}

func (c *LRU) Get(k interface{}) (v interface{}, ok bool) {
	if item, ok := c.cache[k]; ok && !c.expired(k) {
		c.evictList.MoveToFront(item)
		if item.Value.(*entry).value == nil {
			return nil, false
		}
		return item.Value.(*entry).value, true
	}
	return
}

func (c *LRU) Contains(k interface{}) bool {
	_, ok := c.cache[k]
	return ok && !c.expired(k)
}

// Peek get a cache without move it to head
func (c *LRU) Peek(k interface{}) (v interface{}, ok bool) {

	var item *list.Element

	if item, ok = c.cache[k]; ok && !c.expired(k) {
		return item.Value.(*entry).value, true
	}

	return nil, ok
}

func (c *LRU) Remove(k interface{}) bool {
	if item, ok := c.cache[k]; ok {
		c.removeElement(item)
		return true
	}
	return false
}

func (c *LRU) RemoveOldest() (k, v interface{}, ok bool) {
	item := c.evictList.Back()
	if item != nil {
		c.removeElement(item)
		kv := item.Value.(*entry)
		return kv.key, kv.value, true
	}
	return nil,nil,false
}

func (c *LRU) Len() int {
	return c.evictList.Len()
}

// Keys returns keys that are not expired from oldest to newest
func (c *LRU) Keys() []interface{} {
	keys := make([]interface{}, 0)

	for item := c.evictList.Back(); item != nil && !c.expired(item.Value.(*entry).key); item = item.Prev() {
		keys = append(keys, item.Value.(*entry).key)
	}

	return keys
}

func (c *LRU) Purge() {
	for k, v := range c.cache {
		if c.onEvicted != nil {
			c.onEvicted(k, v.Value.(*entry).value)
		}
		delete(c.cache, k)
	}

	c.evictList.Init()
}

func (c *LRU) Resize(size int) int {
	diff := c.Len() - size
	if diff < 0 {
		diff = 0
	}
	for i := 0; i < diff; i++ {
		c.removeOldest()
	}
	c.size = size
	return diff
}

func (c *LRU) removeOldest() {
	item := c.evictList.Back()

	if item != nil {
		c.removeElement(item)
	}
}

func (c *LRU) removeElement(e *list.Element) {
	c.evictList.Remove(e)

	kv := e.Value.(*entry)

	delete(c.cache, kv.key)

	if c.onEvicted != nil {
		c.onEvicted(kv.key, kv.value)
	}
}

func (c *LRU) expired(k interface{}) bool {
	if c.ttl == NoLimitTTL {
		return false
	}

	if item, ok := c.cache[k]; ok {
		if time.Since(item.Value.(*entry).updatedAt) <= c.ttl {
			return false
		}
	}

	return true
}