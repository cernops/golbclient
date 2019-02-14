package metrics

import (
	"sync"
)

type MetricsCache struct {
	mutex sync.RWMutex
	cacheMap map[string]int
}

func NewMetricsCache() *MetricsCache{
	return &MetricsCache{sync.RWMutex{},make(map[string]int)}
}

func (mc *MetricsCache) Contains(key string) bool {
	defer mc.mutex.RUnlock()
	mc.mutex.RLock()
	_, found := mc.cacheMap[key]
	return found
}

func (mc *MetricsCache) Put(key string, metric int) {
	defer mc.mutex.Unlock()
	mc.mutex.Lock()
	mc.cacheMap[key] = metric
}

func (mc *MetricsCache) Get(key string) int {
	defer mc.mutex.RUnlock()
	mc.mutex.RLock()
	value, _ := mc.cacheMap[key]
	return value
}