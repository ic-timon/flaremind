package models

// PageResult 单个页面结果
type PageResult struct {
	URL      string `json:"url"`
	Markdown string `json:"markdown"`
	Depth    int    `json:"depth"`
}

// CrawlConfig 爬取配置
type CrawlConfig struct {
	MaxDepth       int
	MaxPages       int
	AllowedDomains []string
	MaxWorkers     int
	Timeout        int     // 超时时间（秒）
	RateLimit      float64 // 每秒最大请求数（0 表示无限制）
	Delay          int     // 每个请求之间的延迟（毫秒）
}


