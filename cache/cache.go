package cache

import (
	"time"

	patrickmn "github.com/patrickmn/go-cache"
)

// A wrapper around patrickmn/go-cache that implements a simple Cache interface.
type Cache struct {
	patrickmn.Cache
}

func New(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return &Cache{Cache: *patrickmn.New(defaultExpiration, cleanupInterval)}
}

func (c Cache) Get(k string) interface{} {
	v, ok := c.Cache.Get(k)
	if !ok {
		return nil
	}
	return v
}

func (c Cache) Set(k string, v interface{}) {
	c.Cache.SetDefault(k, v)
}
