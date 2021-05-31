package code

import (
	"container/list"
	"errors"
)

type CacheModel struct {
	size   int
	values *list.List
	items  map[uint32]*list.Element
}

type entry struct {
	key   uint32
	value string
}

func NewCache(size int) (*CacheModel, error) {
	if size <= 0 {
		return nil, errors.New("negative size")
	}
	return &CacheModel{
		size:   size,
		values: list.New(),
		items:  make(map[uint32]*list.Element),
	}, nil
}

func (c *CacheModel) Get(key uint32) (string, bool) {
	val, ok := c.items[key]
	if ok {
		c.values.MoveToFront(val)
		result := val.Value.(*entry).value
		return result, true
	}

	return "", false
}

func (c *CacheModel) Put(key uint32, value string) {
	val, ok := c.items[key]
	if ok {
		c.values.MoveToFront(val)
		val.Value.(*entry).value = value
		return
	}

	if c.values.Len() >= c.size {
		old := c.values.Back()
		if old != nil {
			c.values.Remove(old)
			delete(c.items, old.Value.(*entry).key)
		}
	}

	element := &entry{
		key:   key,
		value: value,
	}
	c.values.PushFront(element)
	c.items[key] = c.values.Front()
}
