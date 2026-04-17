package cache

import (
	"sync"
	"time"
)

type entry struct {
	value     any
	expiresAt time.Time
}

type Cache struct {
	mu          sync.RWMutex
	items       map[string]entry
	defaultTTL  time.Duration
	stopCleanup chan struct{}
}

func New(defaultTTL, cleanupInterval time.Duration) *Cache {
	c := &Cache{
		items:       make(map[string]entry),
		defaultTTL:  defaultTTL,
		stopCleanup: make(chan struct{}),
	}
	go c.runCleanup(cleanupInterval)
	return c
}

func (c *Cache) Set(key string, value any) {
	c.SetTTL(key, value, c.defaultTTL)
}

func (c *Cache) SetTTL(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = entry{value: value, expiresAt: time.Now().Add(ttl)}
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *Cache) DeletePrefix(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.items {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(c.items, k)
		}
	}
}

func (c *Cache) Close() {
	close(c.stopCleanup)
}

func (c *Cache) runCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.evict()
		case <-c.stopCleanup:
			return
		}
	}
}

func (c *Cache) evict() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	for k, e := range c.items {
		if now.After(e.expiresAt) {
			delete(c.items, k)
		}
	}
}
