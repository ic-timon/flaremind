package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"flaremind/internal/cache"
	"flaremind/internal/crawler"
	"flaremind/internal/models"
)

func main() {
	var url string
	var maxDepth int
	var maxPages int
	var timeout int
	var rateLimit float64
	var delay int
	var outputFile string

	flag.StringVar(&url, "url", "https://go.dev/", "URL to crawl")
	flag.IntVar(&maxDepth, "depth", 2, "Maximum crawl depth")
	flag.IntVar(&maxPages, "pages", 10, "Maximum number of pages to crawl")
	flag.IntVar(&timeout, "timeout", 60, "Timeout per page in seconds")
	flag.Float64Var(&rateLimit, "rate", 2.0, "Maximum requests per second (0 = unlimited)")
	flag.IntVar(&delay, "delay", 500, "Delay between requests in milliseconds")
	flag.StringVar(&outputFile, "o", "", "Output directory path (for multiple pages) or file path (for single page). If not specified, output JSON to stdout")
	flag.Parse()

	if url == "" {
		fmt.Println("Error: URL is required")
		flag.Usage()
		os.Exit(1)
	}

	log.Printf("Starting FlareMind CLI")
	log.Printf("Target URL: %s", url)
	log.Printf("Max Depth: %d", maxDepth)
	log.Printf("Max Pages: %d", maxPages)
	log.Printf("Timeout: %d seconds per page", timeout)
	log.Printf("Rate Limit: %.2f requests/second", rateLimit)
	log.Printf("Delay: %d ms between requests", delay)

	// 初始化组件
	renderer := crawler.NewRenderer(time.Duration(timeout)*time.Second, true)
	extractor := crawler.NewExtractor()
	converter := crawler.NewConverter()
	cacheInstance := cache.NewCache(24*time.Hour, 1*time.Hour)

	// 创建爬取管理器
	manager := crawler.NewCrawlManager(
		renderer,
		extractor,
		converter,
		cacheInstance,
		5,                                  // maxWorkers
		time.Duration(timeout)*time.Second, // timeout
	)

	// 解析 URL 获取域名
	parsedURL, err := parseURL(url)
	if err != nil {
		log.Fatalf("Invalid URL: %v", err)
	}

	// 创建配置
	config := models.CrawlConfig{
		MaxDepth:       maxDepth,
		MaxPages:       maxPages,
		AllowedDomains: []string{parsedURL.Host},
		MaxWorkers:     5,
		Timeout:        timeout,
		RateLimit:      rateLimit,
		Delay:          delay,
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout*maxPages+60)*time.Second)
	defer cancel()

	// 执行爬取
	log.Println("=" + strings.Repeat("=", 60) + "=")
	log.Println("Starting crawl...")
	log.Println("=" + strings.Repeat("=", 60) + "=")

	startTime := time.Now()
	pages, err := manager.Crawl(ctx, url, config)
	duration := time.Since(startTime)

	if err != nil {
		log.Fatalf("Crawl failed: %v", err)
	}

	// 输出结果
	log.Println("=" + strings.Repeat("=", 60) + "=")
	log.Printf("Crawl completed in %v", duration)
	log.Printf("Total pages crawled: %d", len(pages))
	log.Println("=" + strings.Repeat("=", 60) + "=")

	if len(pages) == 0 {
		log.Println("WARNING: No pages were crawled!")
		log.Println("This could be due to:")
		log.Println("  1. Rendering failure (check if Chrome/Chromium is installed)")
		log.Println("  2. Content extraction failure")
		log.Println("  3. Link extraction failure")
		log.Println("  4. Network timeout")
		os.Exit(1)
	}

	// 如果指定了输出路径，保存为 Markdown 格式
	if outputFile != "" {
		// 判断是目录还是文件
		// 如果只有一个页面且路径以 .md 结尾，保存为单个文件
		// 否则作为目录，每个页面保存为独立文件
		isSingleFile := len(pages) == 1 && strings.HasSuffix(strings.ToLower(outputFile), ".md")

		if isSingleFile {
			// 单个文件模式（向后兼容）
			dir := filepath.Dir(outputFile)
			if dir != "." && dir != "" {
				if err := os.MkdirAll(dir, 0755); err != nil {
					log.Fatalf("Failed to create output directory: %v", err)
				}
			}

			file, err := os.Create(outputFile)
			if err != nil {
				log.Fatalf("Failed to create output file: %v", err)
			}
			defer file.Close()

			page := pages[0]
			file.WriteString(fmt.Sprintf("# %s\n\n", page.URL))
			file.WriteString(fmt.Sprintf("**Source URL:** %s  \n", page.URL))
			file.WriteString(fmt.Sprintf("**Depth:** %d  \n\n", page.Depth))
			file.WriteString("---\n\n")
			file.WriteString(page.Markdown)
			file.WriteString("\n")

			log.Printf("Results saved to: %s (Markdown format)", outputFile)
		} else {
			// 目录模式：每个页面保存为独立文件
			outputDir := outputFile
			// 如果路径以 .md 结尾但不是单个文件，去掉扩展名作为目录名
			if strings.HasSuffix(strings.ToLower(outputDir), ".md") {
				outputDir = strings.TrimSuffix(outputDir, filepath.Ext(outputDir))
			}

			// 确保目录存在
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				log.Fatalf("Failed to create output directory: %v", err)
			}

			// 为每个页面创建文件
			savedCount := 0
			for i, page := range pages {
				// 生成安全的文件名
				filename := sanitizeFilename(page.URL, i)
				filePath := filepath.Join(outputDir, filename)

				file, err := os.Create(filePath)
				if err != nil {
					log.Printf("Failed to create file %s: %v", filePath, err)
					continue
				}

				// 写入 Markdown 内容
				file.WriteString(fmt.Sprintf("# %s\n\n", page.URL))
				file.WriteString(fmt.Sprintf("**Source URL:** %s  \n", page.URL))
				file.WriteString(fmt.Sprintf("**Depth:** %d  \n\n", page.Depth))
				file.WriteString("---\n\n")
				file.WriteString(page.Markdown)
				file.WriteString("\n")

				file.Close()
				savedCount++
			}

			log.Printf("Results saved to directory: %s (%d files)", outputDir, savedCount)
		}
	} else {
		// 输出到标准输出（JSON 格式）
		output := map[string]interface{}{
			"url":      url,
			"total":    len(pages),
			"duration": duration.String(),
			"pages":    pages,
		}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			log.Fatalf("Failed to encode output: %v", err)
		}
	}

	// 显示摘要信息
	fmt.Fprint(os.Stderr, "\nCrawl Summary:\n")
	separator := strings.Repeat("-", 80) + "\n"
	fmt.Fprint(os.Stderr, separator)
	fmt.Fprintf(os.Stderr, "Total pages: %d\n", len(pages))
	fmt.Fprintf(os.Stderr, "Duration: %v\n", duration)
	for i, page := range pages {
		fmt.Fprintf(os.Stderr, "[%d] %s (depth: %d)\n", i+1, page.URL, page.Depth)
	}
	fmt.Fprint(os.Stderr, separator)
}

func parseURL(rawURL string) (*url.URL, error) {
	return url.Parse(rawURL)
}

// sanitizeFilename 从 URL 生成安全的文件名
func sanitizeFilename(urlStr string, index int) string {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		// 如果解析失败，使用索引作为文件名
		return fmt.Sprintf("page_%03d.md", index+1)
	}

	// 构建文件名：域名 + 路径
	var parts []string

	// 添加域名（去除端口）
	host := parsed.Hostname()
	if host != "" {
		parts = append(parts, strings.ReplaceAll(host, ".", "_"))
	}

	// 添加路径部分
	path := parsed.Path
	if path == "" || path == "/" {
		path = "index"
	} else {
		// 清理路径
		path = strings.TrimPrefix(path, "/")
		path = strings.TrimSuffix(path, "/")
		path = strings.ReplaceAll(path, "/", "_")
		path = strings.ReplaceAll(path, "\\", "_")
	}

	// 限制路径长度
	if len(path) > 100 {
		path = path[:100]
	}

	if path != "" {
		parts = append(parts, path)
	}

	// 如果文件名为空，使用索引
	filename := strings.Join(parts, "_")
	if filename == "" {
		filename = fmt.Sprintf("page_%03d", index+1)
	}

	// 移除不安全字符
	filename = strings.ReplaceAll(filename, ":", "_")
	filename = strings.ReplaceAll(filename, "*", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "<", "_")
	filename = strings.ReplaceAll(filename, ">", "_")
	filename = strings.ReplaceAll(filename, "|", "_")
	filename = strings.ReplaceAll(filename, "\"", "_")

	// 确保文件名不超过 200 个字符
	if len(filename) > 200 {
		filename = filename[:200]
	}

	return filename + ".md"
}
