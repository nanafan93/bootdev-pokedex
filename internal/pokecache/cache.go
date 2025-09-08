package pokecache

import (
	"fmt"
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	data      []byte
}

type Cache struct {
	Entries  map[string]cacheEntry
	interval time.Duration
	mu       sync.RWMutex
}

func NewCache(interval time.Duration) *Cache {
	c := &Cache{
		Entries:  make(map[string]cacheEntry),
		interval: interval,
	}
	go c.ReapLoop()
	return c
}

func (c *Cache) Add(key string, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Entries[key] = cacheEntry{
		createdAt: time.Now(),
		data:      data,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, exists := c.Entries[key]
	if !exists {
		fmt.Println("Cache miss !")
		return nil, false
	} else {
		fmt.Println("Cache hit !")
		return entry.data, true
	}
}

func (c *Cache) ReapLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.Entries {
			if time.Since(entry.createdAt) > c.interval {
				fmt.Printf("Reaping cache entry for key: %s\n", key)
				delete(c.Entries, key)
			}
		}
		c.mu.Unlock()
	}
}
