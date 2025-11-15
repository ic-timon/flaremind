package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"flaremind/internal/cache"
	"flaremind/internal/crawler"
	"flaremind/internal/models"
)

// TestCrawlBaiduNews 测试爬取百度新闻
func TestCrawlBaiduNews(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 初始化组件
	renderer := crawler.NewRenderer(30*time.Second, true)
	extractor := crawler.NewExtractor()
	converter := crawler.NewConverter()
	cacheInstance := cache.NewCache(24*time.Hour, 1*time.Hour)

	manager := crawler.NewCrawlManager(
		renderer,
		extractor,
		converter,
		cacheInstance,
		3,
		30*time.Second,
	)

	// 创建配置
	config := models.CrawlConfig{
		MaxDepth:       1,
		MaxPages:       3,
		AllowedDomains: []string{"news.baidu.com", "baidu.com"},
		MaxWorkers:     3,
		Timeout:        30,
		RateLimit:      1.0, // 保守的速率限制
		Delay:          2000, // 2秒延迟
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 执行爬取
	pages, err := manager.Crawl(ctx, "https://news.baidu.com/", config)
	if err != nil {
		t.Fatalf("Failed to crawl Baidu News: %v", err)
	}

	if len(pages) == 0 {
		t.Error("No pages crawled from Baidu News")
	}

	// 保存结果到测试目录
	outputDir := "test_output/baidu_news"
	os.MkdirAll(outputDir, 0755)
	
	for i, page := range pages {
		filename := filepath.Join(outputDir, sanitizeFilename(page.URL, i))
		file, err := os.Create(filename)
		if err != nil {
			t.Logf("Failed to create file %s: %v", filename, err)
			continue
		}
		file.WriteString(fmt.Sprintf("# %s\n\n**Source URL:** %s  \n**Depth:** %d  \n\n---\n\n%s\n", 
			page.URL, page.URL, page.Depth, page.Markdown))
		file.Close()
	}

	t.Logf("Successfully crawled %d pages from Baidu News", len(pages))
}

// TestCrawlBilibili 测试爬取 Bilibili 热门
func TestCrawlBilibili(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 初始化组件
	renderer := crawler.NewRenderer(30*time.Second, true)
	extractor := crawler.NewExtractor()
	converter := crawler.NewConverter()
	cacheInstance := cache.NewCache(24*time.Hour, 1*time.Hour)

	manager := crawler.NewCrawlManager(
		renderer,
		extractor,
		converter,
		cacheInstance,
		3,
		30*time.Second,
	)

	// 创建配置
	config := models.CrawlConfig{
		MaxDepth:       1,
		MaxPages:       3,
		AllowedDomains: []string{"www.bilibili.com", "bilibili.com"},
		MaxWorkers:     3,
		Timeout:        30,
		RateLimit:      1.0,
		Delay:          2000,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 执行爬取
	pages, err := manager.Crawl(ctx, "https://www.bilibili.com/v/popular", config)
	if err != nil {
		t.Fatalf("Failed to crawl Bilibili: %v", err)
	}

	if len(pages) == 0 {
		t.Error("No pages crawled from Bilibili")
	}

	// 保存结果
	outputDir := "test_output/bilibili"
	os.MkdirAll(outputDir, 0755)
	
	for i, page := range pages {
		filename := filepath.Join(outputDir, sanitizeFilename(page.URL, i))
		file, err := os.Create(filename)
		if err != nil {
			t.Logf("Failed to create file %s: %v", filename, err)
			continue
		}
		file.WriteString(fmt.Sprintf("# %s\n\n**Source URL:** %s  \n**Depth:** %d  \n\n---\n\n%s\n", 
			page.URL, page.URL, page.Depth, page.Markdown))
		file.Close()
	}

	t.Logf("Successfully crawled %d pages from Bilibili", len(pages))
}

// TestCrawlDouyin 测试爬取抖音
func TestCrawlDouyin(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 初始化组件
	renderer := crawler.NewRenderer(30*time.Second, true)
	extractor := crawler.NewExtractor()
	converter := crawler.NewConverter()
	cacheInstance := cache.NewCache(24*time.Hour, 1*time.Hour)

	manager := crawler.NewCrawlManager(
		renderer,
		extractor,
		converter,
		cacheInstance,
		3,
		30*time.Second,
	)

	// 创建配置
	config := models.CrawlConfig{
		MaxDepth:       1,
		MaxPages:       3,
		AllowedDomains: []string{"www.douyin.com", "douyin.com"},
		MaxWorkers:     3,
		Timeout:        30,
		RateLimit:      1.0,
		Delay:          2000,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 执行爬取
	pages, err := manager.Crawl(ctx, "https://www.douyin.com/", config)
	if err != nil {
		t.Fatalf("Failed to crawl Douyin: %v", err)
	}

	if len(pages) == 0 {
		t.Error("No pages crawled from Douyin")
	}

	// 保存结果
	outputDir := "test_output/douyin"
	os.MkdirAll(outputDir, 0755)
	
	for i, page := range pages {
		filename := filepath.Join(outputDir, sanitizeFilename(page.URL, i))
		file, err := os.Create(filename)
		if err != nil {
			t.Logf("Failed to create file %s: %v", filename, err)
			continue
		}
		file.WriteString(fmt.Sprintf("# %s\n\n**Source URL:** %s  \n**Depth:** %d  \n\n---\n\n%s\n", 
			page.URL, page.URL, page.Depth, page.Markdown))
		file.Close()
	}

	t.Logf("Successfully crawled %d pages from Douyin", len(pages))
}

// TestCrawlXiaohongshu 测试爬取小红书
func TestCrawlXiaohongshu(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// 初始化组件
	renderer := crawler.NewRenderer(30*time.Second, true)
	extractor := crawler.NewExtractor()
	converter := crawler.NewConverter()
	cacheInstance := cache.NewCache(24*time.Hour, 1*time.Hour)

	manager := crawler.NewCrawlManager(
		renderer,
		extractor,
		converter,
		cacheInstance,
		3,
		30*time.Second,
	)

	// 创建配置
	config := models.CrawlConfig{
		MaxDepth:       1,
		MaxPages:       3,
		AllowedDomains: []string{"www.xiaohongshu.com", "xiaohongshu.com"},
		MaxWorkers:     3,
		Timeout:        30,
		RateLimit:      1.0,
		Delay:          2000,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// 执行爬取
	pages, err := manager.Crawl(ctx, "https://www.xiaohongshu.com/explore", config)
	if err != nil {
		t.Fatalf("Failed to crawl Xiaohongshu: %v", err)
	}

	if len(pages) == 0 {
		t.Error("No pages crawled from Xiaohongshu")
	}

	// 保存结果
	outputDir := "test_output/xiaohongshu"
	os.MkdirAll(outputDir, 0755)
	
	for i, page := range pages {
		filename := filepath.Join(outputDir, sanitizeFilename(page.URL, i))
		file, err := os.Create(filename)
		if err != nil {
			t.Logf("Failed to create file %s: %v", filename, err)
			continue
		}
		file.WriteString(fmt.Sprintf("# %s\n\n**Source URL:** %s  \n**Depth:** %d  \n\n---\n\n%s\n", 
			page.URL, page.URL, page.Depth, page.Markdown))
		file.Close()
	}

	t.Logf("Successfully crawled %d pages from Xiaohongshu", len(pages))
}

