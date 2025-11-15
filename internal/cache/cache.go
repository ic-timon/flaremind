package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
)

// Cache 缓存管理器
type Cache struct {
	c *cache.Cache
}

// NewCache 创建新的缓存
func NewCache(defaultExpiration, cleanupInterval time.Duration) *Cache {
	return &Cache{
		c: cache.New(defaultExpiration, cleanupInterval),
	}
}

// Get 获取缓存值
func (c *Cache) Get(key string) (string, bool) {
	value, found := c.c.Get(key)
	if !found {
		return "", false
	}

	str, ok := value.(string)
	if !ok {
		return "", false
	}

	return str, true
}

// Set 设置缓存值
func (c *Cache) Set(key string, value string, expiration time.Duration) {
	c.c.Set(key, value, expiration)
}

// Delete 删除缓存值
func (c *Cache) Delete(key string) {
	c.c.Delete(key)
}

// Clear 清空所有缓存
func (c *Cache) Clear() {
	c.c.Flush()
}

// ItemCount 返回缓存项数量
func (c *Cache) ItemCount() int {
	return c.c.ItemCount()
}


