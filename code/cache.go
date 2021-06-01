package code

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

type CacheModel struct {
	size      int
	values    *list.List
	items     map[uint32]*list.Element
	mutex     sync.Mutex
	lifetime  time.Duration
	cleartime time.Duration
}

// структура занимает
type entry struct {
	key       uint32
	value     string
	timeStart time.Time
}

func NewCache(size int, lifetime time.Duration) (*CacheModel, error) {
	if size <= 0 {
		return nil, errors.New("negative size")
	}
	cache := &CacheModel{
		size:      size,
		values:    list.New(),
		items:     make(map[uint32]*list.Element),
		lifetime:  lifetime,
		cleartime: 1 * time.Second,
	}

	go cache.garbageCleaning()

	return cache, nil
}

func (c *CacheModel) Get(key uint32) (string, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	val, ok := c.items[key]
	if !ok {
		return "", false
	}

	if c.lifetime < time.Since(val.Value.(*entry).timeStart) {
		c.removeEntry(val)
		return "", false
	}
	c.values.MoveToFront(val)
	result := val.Value.(*entry).value
	return result, true
}

func (c *CacheModel) Put(key uint32, value string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	val, ok := c.items[key]
	if ok {
		c.values.MoveToFront(val)
		val.Value.(*entry).value = value
		val.Value.(*entry).timeStart = time.Now()
		return
	}

	if c.values.Len() >= c.size {
		old := c.values.Back()
		if old != nil {
			c.removeEntry(old)
		}
	}

	element := &entry{
		key:       key,
		value:     value,
		timeStart: time.Now(),
	}
	c.values.PushFront(element)
	c.items[key] = c.values.Front()
}

func (c *CacheModel) removeEntry(e *list.Element) {
	c.values.Remove(e)
	delete(c.items, e.Value.(*entry).key)
}

func (c *CacheModel) garbageCleaning() {
	for {
		<-time.After(c.cleartime)

		if c.values.Len() == 0 {
			return
		}

		c.clearItems()
	}
}

func (c *CacheModel) clearItems() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	timeNow := time.Now()
	for _, element := range c.items {
		timePassed := timeNow.Sub(element.Value.(*entry).timeStart)

		if timePassed > c.lifetime {
			c.removeEntry(element)
		}
	}
}
