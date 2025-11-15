package crawler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"flaremind/internal/cache"
	"flaremind/internal/models"
	"flaremind/internal/queue"
	"flaremind/pkg/utils"

	"golang.org/x/time/rate"
)

// CrawlManager 爬取管理器
type CrawlManager struct {
	renderer   *Renderer
	extractor  *Extractor
	converter  *Converter
	cache      *cache.Cache
	maxWorkers int
	timeout    time.Duration
	rateLimiter *rate.Limiter
	delay      time.Duration
}

// NewCrawlManager 创建新的爬取管理器
func NewCrawlManager(renderer *Renderer, extractor *Extractor, converter *Converter, cache *cache.Cache, maxWorkers int, timeout time.Duration) *CrawlManager {
	return &CrawlManager{
		renderer:    renderer,
		extractor:   extractor,
		converter:   converter,
		cache:       cache,
		maxWorkers:  maxWorkers,
		timeout:     timeout,
		rateLimiter: nil, // 将在 Crawl 方法中根据配置设置
		delay:       0,   // 将在 Crawl 方法中根据配置设置
	}
}

// Crawl 执行整站爬取
func (cm *CrawlManager) Crawl(ctx context.Context, startURL string, config models.CrawlConfig) ([]models.PageResult, error) {
	// 规范化起始 URL
	normalizedStartURL, err := utils.NormalizeURL(startURL)
	if err != nil {
		return nil, fmt.Errorf("invalid start URL: %w", err)
	}

	// 设置速率限制器
	if config.RateLimit > 0 {
		cm.rateLimiter = rate.NewLimiter(rate.Limit(config.RateLimit), int(config.RateLimit))
		log.Printf("Rate limit enabled: %.2f requests/second", config.RateLimit)
	} else {
		cm.rateLimiter = nil
	}

	// 设置请求延迟
	if config.Delay > 0 {
		cm.delay = time.Duration(config.Delay) * time.Millisecond
		log.Printf("Request delay enabled: %d ms between requests", config.Delay)
	} else {
		cm.delay = 0
	}

	log.Printf("Starting crawl from %s (maxDepth: %d, maxPages: %d)", normalizedStartURL, config.MaxDepth, config.MaxPages)

	// 初始化队列和深度映射
	q := queue.NewQueue()
	depthMap := make(map[string]int)
	var depthMu sync.RWMutex
	q.Add(normalizedStartURL)
	depthMap[normalizedStartURL] = 0

	// 结果存储
	var results []models.PageResult
	var resultsMu sync.Mutex

	// Worker 通道（包含 URL 和深度）
	type urlDepth struct {
		url   string
		depth int
	}
	urlChan := make(chan urlDepth, cm.maxWorkers)
	var wg sync.WaitGroup

	// 启动 workers
	for i := 0; i < cm.maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case ud, ok := <-urlChan:
					if !ok {
						return
					}

					// 检查深度和数量限制
					resultsMu.Lock()
					currentCount := len(results)
					resultsMu.Unlock()

					if currentCount >= config.MaxPages {
						return
					}

					// 应用速率限制
					if cm.rateLimiter != nil {
						if err := cm.rateLimiter.Wait(ctx); err != nil {
							log.Printf("Rate limiter error for %s: %v", ud.url, err)
							continue
						}
					}

					// 应用请求延迟
					if cm.delay > 0 {
						select {
						case <-ctx.Done():
							return
						case <-time.After(cm.delay):
						}
					}

					// 爬取页面
					html, err := cm.renderer.Render(ud.url)
					if err != nil {
						log.Printf("Failed to render %s: %v", ud.url, err)
						continue
					}

					// 提取主要内容
					content, err := cm.extractor.ExtractMainContent(html)
					if err != nil {
						log.Printf("Failed to extract content from %s: %v", ud.url, err)
						continue
					}

					// 转换为 Markdown
					markdown, err := cm.converter.HTMLToMarkdown(content)
					if err != nil {
						log.Printf("Failed to convert to markdown from %s: %v", ud.url, err)
						continue
					}

					// 缓存结果
					cm.cache.Set(ud.url, markdown, 24*time.Hour)

					// 保存结果
					result := models.PageResult{
						URL:      ud.url,
						Markdown: markdown,
						Depth:    ud.depth,
					}

					resultsMu.Lock()
					if len(results) < config.MaxPages {
						results = append(results, result)
						log.Printf("Successfully crawled %s (depth: %d, total: %d)", ud.url, ud.depth, len(results))
					}
					resultsMu.Unlock()

					// 提取链接并添加到队列
					if ud.depth < config.MaxDepth {
						linkExtractor := NewLinkExtractor(ud.url)
						links, err := linkExtractor.ExtractLinks(html, config.AllowedDomains)
						if err == nil {
							addedCount := 0
							for _, link := range links {
								if !q.IsVisited(link) {
									depthMu.Lock()
									depthMap[link] = ud.depth + 1
									depthMu.Unlock()
									if q.Add(link) {
										addedCount++
									}
								}
							}
							if addedCount > 0 {
								log.Printf("Added %d new links from %s (queue size: %d, visited: %d)", addedCount, ud.url, q.Size(), q.VisitedCount())
							}
						} else {
							log.Printf("Failed to extract links from %s: %v", ud.url, err)
						}
					}
				}
			}
		}()
	}

	// 主循环：分发任务
	go func() {
		defer close(urlChan)
		emptyCount := 0
		maxEmptyCount := 50 // 增加到 50 次（5秒），给 worker 更多时间处理
		lastQueueSize := 0
		stableCount := 0
		
		for {
			select {
			case <-ctx.Done():
				return
			default:
				resultsMu.Lock()
				currentCount := len(results)
				resultsMu.Unlock()

				if currentCount >= config.MaxPages {
					log.Printf("Reached max pages limit: %d", config.MaxPages)
					return
				}

				url, ok := q.Pop()
				if !ok {
					// 队列为空，等待一下
					time.Sleep(200 * time.Millisecond) // 增加等待时间到 200ms
					emptyCount++
					currentQueueSize := q.Size()
					
					// 检查队列大小是否稳定（没有新链接被添加）
					if currentQueueSize == lastQueueSize {
						stableCount++
					} else {
						stableCount = 0
						emptyCount = 0 // 有新链接，重置计数
						log.Printf("Queue size changed: %d -> %d, resetting counters", lastQueueSize, currentQueueSize)
					}
					lastQueueSize = currentQueueSize
					
					// 如果队列一直为空且已经访问过一些 URL，检查是否所有任务都完成
					if currentQueueSize == 0 && q.VisitedCount() > 0 {
						// 需要等待足够长时间，确保所有 worker 都完成处理
						// 并且队列大小稳定（没有新链接被添加）
						if emptyCount >= maxEmptyCount && stableCount >= 20 {
							// 等待足够长时间，确认没有新任务
							log.Printf("Queue empty and stable for %d checks, visited: %d, results: %d", emptyCount, q.VisitedCount(), currentCount)
							return
						}
						// 每 10 次检查输出一次状态
						if emptyCount%10 == 0 {
							log.Printf("Waiting for workers... (empty: %d, stable: %d, visited: %d, results: %d)", emptyCount, stableCount, q.VisitedCount(), currentCount)
						}
					} else if currentQueueSize == 0 && q.VisitedCount() == 0 {
						// 如果还没有访问任何 URL，继续等待
						emptyCount = 0
						stableCount = 0
					} else if currentQueueSize > 0 {
						// 队列有内容，重置计数
						emptyCount = 0
						stableCount = 0
					}
					continue
				}

				// 重置空计数和稳定计数
				emptyCount = 0
				stableCount = 0
				lastQueueSize = q.Size()

				// 检查是否已访问
				if q.IsVisited(url) {
					continue
				}

				// 获取深度
				depthMu.RLock()
				depth := depthMap[url]
				depthMu.RUnlock()

				// 标记为已访问（在发送到 worker 之前，避免重复处理）
				q.MarkVisited(url)

				// 发送到 worker
				select {
				case urlChan <- urlDepth{url: url, depth: depth}:
					log.Printf("Dispatched URL to worker: %s (depth: %d)", url, depth)
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	// 等待所有 workers 完成
	wg.Wait()

	log.Printf("Crawl completed: %d pages crawled, %d URLs visited", len(results), q.VisitedCount())

	return results, nil
}

