package code

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPut(t *testing.T) {
	cache, err := NewCache(3, 4*time.Second)
	assert.Equal(t, err, nil, "error with create")

	cache.Put(1, "str1")
	assert.Equal(t, "str1", cache.values.Front().Value.(*entry).value, "value not added")
}

func TestGet(t *testing.T) {
	cache, err := NewCache(3, 4*time.Second)
	assert.Equal(t, err, nil, "error with create")

	cache.Put(1, "str1")
	cache.Put(2, "str2")
	cache.Put(3, "str3")

	result, ok := cache.Get(1)
	assert.Equal(t, true, ok, "value not found")
	assert.Equal(t, "str1", result, "incorrect value")
}

func TestSimple(t *testing.T) {
	size := 3
	cache, err := NewCache(size, 4*time.Second)
	assert.Equal(t, err, nil, "error with create")

	cache.Put(1, "str1")
	cache.Put(2, "str2")
	cache.Put(3, "str3")

	cache.Get(3)
	cache.Get(2)
	cache.Get(1)
	cache.Get(3)

	cache.Put(4, "str4")

	assert.Equal(t, size, cache.values.Len(), "the size does not match")

	_, ok := cache.Get(2)
	assert.Equal(t, false, ok, "value found")

	result, ok := cache.Get(1)
	assert.Equal(t, true, ok, "value found")
	assert.Equal(t, "str1", result, "incorrect value")

	result, ok = cache.Get(3)
	assert.Equal(t, true, ok, "value found")
	assert.Equal(t, "str3", result, "incorrect value")

	result, ok = cache.Get(4)
	assert.Equal(t, true, ok, "value found")
	assert.Equal(t, "str4", result, "incorrect value")
}

func TestTimeSimple(t *testing.T) {
	lifetime := 4 * time.Second
	cache, err := NewCache(3, lifetime)
	assert.Equal(t, err, nil, "error with create")

	cache.Put(1, "str1")

	<-time.After(lifetime)

	_, ok := cache.Get(1)
	assert.Equal(t, false, ok, "value found")
}

func TestTimeHard(t *testing.T) {
	lifetime := 4 * time.Second
	cache, err := NewCache(3, lifetime)
	assert.Equal(t, err, nil, "error with create")

	cache.Put(1, "str1")
	cache.Put(2, "str2")
	cache.Put(3, "str3")

	<-time.After(lifetime + 2*cache.cleartime)

	assert.Equal(t, 0, cache.values.Len(), "incorrect size")
}
