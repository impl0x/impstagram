package ttlcache // time to live cache

import (
	"sync"
	"time"
)

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

// Returns a new empty Cache and starts a cleaner goroutine which cleans the expired items in each interval
func New[K comparable, V any](interval time.Duration) *Cache[K, V] {
	c := &Cache[K, V]{
		items:  make(map[K]item[V]),
		Config: struct{ LazyDelete bool }{true},
	}
	go c.Cleaner(interval)
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
	i, ok := c.items[key]
	c.mu.RUnlock()
	// Passive delete
	if c.Config.LazyDelete && i.expiresAt.Before(time.Now()) {
		c.Delete(key)
	}
	return i.value, i.expiresAt, ok
}

// Deletes a value from the cache
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Cleaner goroutine which runs on the given interval and cleans up the expired items in the map
func (c *Cache[K, V]) Cleaner(interval time.Duration) {
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
