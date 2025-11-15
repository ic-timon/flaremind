package queue

import (
	"sync"

	"flaremind/pkg/utils"
)

// Queue URL 队列管理器
type Queue struct {
	mu       sync.RWMutex
	urls     []string
	visited  map[string]bool
	normalized map[string]string // 原始 URL -> 规范化 URL 的映射
}

// NewQueue 创建新的队列
func NewQueue() *Queue {
	return &Queue{
		urls:       make([]string, 0),
		visited:    make(map[string]bool),
		normalized: make(map[string]string),
	}
}

// Add 添加 URL 到队列（如果未访问过）
func (q *Queue) Add(url string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	// 规范化 URL
	normalized, err := utils.NormalizeURL(url)
	if err != nil {
		return false
	}

	// 检查是否已访问
	if q.visited[normalized] {
		return false
	}

	// 检查是否已在队列中
	for _, u := range q.urls {
		if u == normalized {
			return false
		}
	}

	// 添加到队列
	q.urls = append(q.urls, normalized)
	q.normalized[url] = normalized

	return true
}

// Pop 从队列中取出一个 URL
func (q *Queue) Pop() (string, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.urls) == 0 {
		return "", false
	}

	url := q.urls[0]
	q.urls = q.urls[1:]

	return url, true
}

// MarkVisited 标记 URL 为已访问
func (q *Queue) MarkVisited(url string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	normalized, err := utils.NormalizeURL(url)
	if err != nil {
		return
	}

	q.visited[normalized] = true
}

// IsVisited 检查 URL 是否已访问
func (q *Queue) IsVisited(url string) bool {
	q.mu.RLock()
	defer q.mu.RUnlock()

	normalized, err := utils.NormalizeURL(url)
	if err != nil {
		return false
	}

	return q.visited[normalized]
}

// Size 返回队列大小
func (q *Queue) Size() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.urls)
}

// VisitedCount 返回已访问的 URL 数量
func (q *Queue) VisitedCount() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return len(q.visited)
}



