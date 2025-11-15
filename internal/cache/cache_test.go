package cache

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	cache := NewCache(5*time.Minute, 10*time.Minute)

	// 测试 Set 和 Get
	key := "test-key"
	value := "test-value"
	cache.Set(key, value, time.Hour)

	retrieved, found := cache.Get(key)
	if !found {
		t.Error("Expected to find key in cache")
	}
	if retrieved != value {
		t.Errorf("Expected %v, got %v", value, retrieved)
	}

	// 测试不存在的 key
	_, found = cache.Get("non-existent")
	if found {
		t.Error("Expected not to find non-existent key")
	}

	// 测试 Delete
	cache.Delete(key)
	_, found = cache.Get(key)
	if found {
		t.Error("Expected key to be deleted")
	}

	// 测试 Clear
	cache.Set("key1", "value1", time.Hour)
	cache.Set("key2", "value2", time.Hour)
	cache.Clear()
	if cache.ItemCount() != 0 {
		t.Errorf("Expected cache to be empty, got %d items", cache.ItemCount())
	}
}


