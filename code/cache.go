package code

import (
	"container/list"
	"errors"
	"sync"
	"time"
)

const (
	initSize    uint32 = 80  // занимает первоначальная структура
	mapElemSize uint32 = 12  // 4 байта ключ + 8 байт указатель
	elemSize    uint32 = 156 // 4 байта ключ + 24 байта time.Time + 128 значение
)

type CacheModel struct {
	size          int
	memory        uint32
	values        *list.List
	valuesHash    map[uint32]*list.Element
	mutex         sync.Mutex
	lifetime      time.Duration
	cleartime     time.Duration
	currentMemory uint32
}

type entry struct {
	key       uint32
	value     string
	timeStart time.Time
}

func NewCache(size int, memory uint32, lifetime time.Duration) (*CacheModel, error) {
	if size <= 0 {
		return nil, errors.New("negative size")
	}
	cache := &CacheModel{
		size:          size,
		values:        list.New(),
		valuesHash:    make(map[uint32]*list.Element),
		lifetime:      lifetime,
		memory:        memory,
		cleartime:     1 * time.Second,
		currentMemory: initSize,
	}

	// запускаем переодическую очистку протухших ключей
	go cache.garbageCleaning()

	return cache, nil
}

func (c *CacheModel) Get(key uint32) (string, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// если элемента нет в кэше
	val, ok := c.valuesHash[key]
	if !ok {
		return "", false
	}

	// если протух элемент
	if c.lifetime < time.Since(val.Value.(*entry).timeStart) {
		c.removeEntry(val)
		return "", false
	}

	// убираем в начало очереди элемент
	c.values.MoveToFront(val)
	result := val.Value.(*entry).value
	return result, true
}

func (c *CacheModel) Put(key uint32, value string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// если элементов больше чем может быть, убираем из очереди последний
	if c.values.Len() >= int(c.size) {
		old := c.values.Back()
		if old != nil {
			c.removeEntry(old)
		}
	}

	// проверка на переполнение памяти
	if c.currentMemory+mapElemSize+elemSize > c.memory {
		return errors.New("memory overflow")
	}
	c.currentMemory += mapElemSize + elemSize

	element := &entry{
		key:       key,
		value:     value,
		timeStart: time.Now(),
	}
	c.values.PushFront(element)
	c.valuesHash[key] = c.values.Front()

	return nil
}

// необходимо удалить элемент из map и из list
func (c *CacheModel) removeEntry(e *list.Element) {
	c.values.Remove(e)
	delete(c.valuesHash, e.Value.(*entry).key)
	c.currentMemory -= (mapElemSize + elemSize)
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
	for _, element := range c.valuesHash {
		timePassed := timeNow.Sub(element.Value.(*entry).timeStart)

		if timePassed > c.lifetime {
			c.removeEntry(element)
		}
	}
}
