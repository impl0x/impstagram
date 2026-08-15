package ttlcache

import (
	"backend/internal/config"
	"sync"
	"time"
)

// Starts the global cleaner if global cleaning is selected
func init(){
	if CleaningMethod==GlobalCleaner{
		go globalCleaner(config.TTLCacheCleanInterval)
	}
}

// item wraps the value with its specific expiration time
type item[V any] struct {
	value     V
	expiresAt time.Time
}

// Cache is a generic, thread-safe map with an automatic cleaner
type Cache[K comparable, V any] struct {
	items map[K]item[V]
	mu    sync.RWMutex

	Config struct {
		LazyDelete bool // Default: True, removes an item from the map on the Get call if the item has expired, before the cleaner does it
	}
}

// Returns a new empty Cache
func New[K comparable, V any]() *Cache[K, V] {
	c:=&Cache[K, V]{
		items: make(map[K]item[V]),
	}
	if CleaningMethod==PerCleaner{
		go perCleaner(c, config.TTLCacheCleanInterval)
	} else if CleaningMethod==GlobalCleaner{
		globalCacheList = append(globalCacheList, c) // todo: fix the generic issue
	}
	return c
}

// Adds a new item to the cache
func (c *Cache[K, V]) Add(key K, value V, expiresAt time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = item[V]{value, expiresAt}
}

// Gets a value from the cache
func (c *Cache[K, V]) Get(key K) (V, time.Time, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	i, ok := c.items[key]
	return i.value, i.expiresAt, ok
}

// Deletes a value from the cache
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// perCleaner goroutine which runs on the given interval and cleans up the expired items in the map
func perCleaner[K comparable, V any](c *Cache[K, V], interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		c.mu.Lock()
		for k, v := range c.items {
			if v.expiresAt.Before(now) {
				delete(c.items, k)
			}
		}
		c.mu.Unlock()
	}
}

// an idea for the cleaner:
// instead of each cache launching their own cleaner,
// we could have one global cleaning goroutine which checks on a global interval
// and cleans all the cache objects in the global list
// some sort of implementation:

var globalCacheList = []*Cache[any, any]{}

func globalCleaner(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		for _, c := range globalCacheList {
			c.mu.Lock()
			now := time.Now()
			for k, v := range c.items {
				if v.expiresAt.Before(now) {
					delete(c.items, k)
				}
			}
			c.mu.Unlock()
		}
	}
}
