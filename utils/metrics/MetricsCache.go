package metrics

import (
	"fmt"
	"sync"
)

// Cache : Caching struct using for the memoization of the metrics evaluation
// @see LBAlias
type Cache struct {
	mutex sync.RWMutex
	cacheMap map[string]int
}

// NewMetricsCache : Constructs a ready to use instance of the Cache struct
func NewMetricsCache() *Cache {
	return &Cache{sync.RWMutex{},make(map[string]int)}
}

// Contains : Checks if the Cache has the given key in memory
func (mc *Cache) Contains(key string) bool {
	defer mc.mutex.RUnlock()
	mc.mutex.RLock()
	_, found := mc.cacheMap[key]
	return found
}

// Put : Puts a new key:value pair into the Cache struct
func (mc *Cache) Put(key string, metric int) {
	defer mc.mutex.Unlock()
	mc.mutex.Lock()
	mc.cacheMap[key] = metric
}

// Get : Given a key, find the corresponding value within the Cache. An error is returned instead
// if no value was found
func (mc *Cache) Get(key string) (int, error) {
	defer mc.mutex.RUnlock()
	mc.mutex.RLock()
	value, found := mc.cacheMap[key]
	if !found {
		return -1, fmt.Errorf("no value found for thr given key [%s]", key)
	}
	return value, nil
}