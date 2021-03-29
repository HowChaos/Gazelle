package lru

import (
	"container/list"
)

//cache
type Cache struct {
	maxBytes int64                    //max value of memory we could use
	nBytes   int64                    //bytes we have used
	ll       *list.List               //two-direction list
	cache    map[string]*list.Element //real cache
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
}

type Entry struct {
	key   string
	value Value
}

type Value interface {
	Len() int
}

//initial
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

//len
// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}

//add/modify
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		//modify
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*Entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		var ele = c.ll.PushFront(&Entry{key, value})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

//delete, obey the LRU algorithm
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*Entry)
		delete(c.cache, kv.key)
		c.nBytes -= int64(kv.value.Len()) + int64(len(kv.key))
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

//query
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*Entry)
		return kv.value, true
	}
	return
}
