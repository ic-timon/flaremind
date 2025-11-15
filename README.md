# FlareMind

一个使用 Golang 实现的命令行网页爬虫工具，支持单页抓取、整站爬取、JavaScript 渲染和 Markdown 转换。

## 功能特性

- **整站爬取（Crawl）**：递归爬取整个网站的所有页面
- **JavaScript 渲染**：使用 chromedp 处理动态内容
- **智能内容提取**：使用 Readability 类似算法识别主要内容，过滤广告和导航
- **Markdown 输出**：将爬取的内容转换为 Markdown 格式
- **JSON 导出**：支持将结果保存为 JSON 文件
- **缓存机制**：避免重复爬取，提高效率
- **错误重试**：自动重试机制，提高稳定性
- **并发控制**：支持多 worker 并发爬取
- **速率限制**：内置速率限制和请求延迟，防止 IP 被封

## 前置要求

- **Go 1.21 或更高版本**
- **Google Chrome 或 Chromium 浏览器**：FlareMind 使用 chromedp 进行页面渲染，需要系统已安装 Chrome 或 Chromium
  - Windows: 下载并安装 [Google Chrome](https://www.google.com/chrome/)
  - macOS: `brew install --cask google-chrome`
  - Linux: `sudo apt-get install google-chrome-stable` 或 `sudo yum install google-chrome-stable`

## 快速开始

### 安装依赖

```bash
go mod tidy
```

### 编译

```bash
go build -o flaremind.exe cmd/cli/main.go
```

### 使用示例

```bash
# 基本用法（输出到标准输出）
.\flaremind.exe -url https://go.dev/ -depth 2 -pages 10

# 保存结果到目录（多个页面，每个页面一个文件）
.\flaremind.exe -url https://go.dev/ -depth 2 -pages 10 -o output_dir

# 保存单个页面到文件（只有一个页面时）
.\flaremind.exe -url https://go.dev/ -depth 0 -pages 1 -o single_page.md

# 带速率限制（推荐，防止 IP 被封）
.\flaremind.exe -url https://go.dev/ -depth 2 -pages 10 -rate 1 -delay 1000 -o results_dir

# 参数说明
# -url: 要爬取的起始 URL（必需）
# -depth: 最大爬取深度（默认: 2）
# -pages: 最大页面数（默认: 10）
# -timeout: 每页超时时间，单位秒（默认: 60）
# -rate: 每秒最大请求数（默认: 2.0，0 表示无限制）
# -delay: 每个请求之间的延迟，单位毫秒（默认: 500）
# -o: 输出目录路径（多个页面）或文件路径（单个页面，需以 .md 结尾）
#     如果不指定则输出 JSON 到标准输出
```

## 输出格式

### Markdown 格式（使用 -o 参数）

当使用 `-o` 参数时，结果会保存为 Markdown 格式：

**多个页面（目录模式）**：
- 如果爬取多个页面，`-o` 指定的是目录路径
- 每个页面会保存为独立的 `.md` 文件
- 文件名基于 URL 自动生成（例如：`go_dev_index.md`、`go_dev_ref_spec.md`）

**单个页面（文件模式）**：
- 如果只有一个页面且 `-o` 以 `.md` 结尾，保存为单个文件
- 例如：`-o single_page.md`

**文件内容格式**：
```markdown
# https://go.dev/

**Source URL:** https://go.dev/  
**Depth:** 0  

---

# Build simple, secure, scalable systems with Go

An open-source programming language supported by Google...
```

### JSON 格式（标准输出）

如果不指定 `-o` 参数，结果会以 JSON 格式输出到标准输出，摘要信息输出到标准错误：

```json
{
  "url": "https://go.dev/",
  "total": 5,
  "duration": "23.7s",
  "pages": [
    {
      "url": "https://go.dev/",
      "markdown": "# Build simple, secure, scalable systems with Go\n\n...",
      "depth": 0
    }
  ]
}
```

## 项目结构

```
flaremind/
├── cmd/
│   └── cli/                 # CLI 工具入口
├── internal/
│   ├── crawler/         # 爬虫核心逻辑
│   │   ├── renderer.go      # 页面渲染器（chromedp）
│   │   ├── extractor.go     # 内容提取器
│   │   ├── converter.go     # Markdown 转换器
│   │   ├── link_extractor.go # 链接提取器
│   │   ├── manager.go        # 爬取管理器
│   │   └── retry.go         # 重试机制
│   ├── queue/            # URL 队列管理
│   ├── cache/            # 缓存管理
│   └── models/           # 数据模型
├── pkg/
│   └── utils/            # 工具函数
├── test_integration.ps1  # 集成测试脚本（PowerShell）
├── test_integration.sh   # 集成测试脚本（Bash）
├── go.mod
├── go.sum
└── README.md
```

## 技术栈

- **chromedp**: 无头浏览器控制，处理 JavaScript 渲染
- **goquery**: HTML 解析和 DOM 操作
- **go-cache**: 内存缓存，避免重复爬取
- **golang.org/x/time/rate**: 速率限制器

## 核心算法

### 内容提取算法

使用类似 Readability 的算法来识别主要内容：

1. **文本密度计算**：计算每个区域的文本密度
2. **链接密度过滤**：过滤链接密度过高的区域（通常是导航或广告）
3. **内容评分**：根据段落数、列表数、图片数等因素评分
4. **关键词匹配**：识别包含内容关键词的区域
5. **噪音过滤**：移除包含广告、导航等关键词的区域

### 爬取策略

- **BFS（广度优先搜索）**：按深度逐层爬取
- **并发控制**：使用 worker pool 模式控制并发数
- **去重机制**：URL 规范化后去重
- **深度限制**：防止无限爬取
- **域名限制**：只爬取指定域名的页面
- **速率限制**：使用令牌桶算法限制请求频率

## 速率限制和防封机制

为了防止请求过密集导致 IP 被封，FlareMind 提供了以下保护机制：

1. **速率限制（Rate Limiting）**：使用令牌桶算法限制每秒最大请求数
   - 默认值：2 请求/秒
   - 可通过 `-rate` 参数配置
   - 设置为 0 表示无限制（不推荐）

2. **请求延迟（Request Delay）**：在每个请求之间添加固定延迟
   - 默认值：500 毫秒
   - 可通过 `-delay` 参数配置
   - 设置为 0 表示无延迟（不推荐）

**推荐配置：**
- 保守配置：`-rate 1 -delay 1000`（每秒 1 个请求，每次间隔 1 秒）
- 标准配置：`-rate 2 -delay 500`（每秒 2 个请求，每次间隔 0.5 秒）
- 快速配置：`-rate 5 -delay 200`（每秒 5 个请求，每次间隔 0.2 秒）

**注意：** 过于激进的配置可能导致 IP 被封，请根据目标网站的实际情况调整。

## 开发

### 运行测试

```bash
# 运行单元测试
go test ./...

# 运行集成测试（需要较长时间，会实际爬取网站）
go test -v ./cmd/cli/... -timeout 10m

# 运行特定网站的集成测试
go test -v ./cmd/cli/... -run TestCrawlBaiduNews -timeout 5m
go test -v ./cmd/cli/... -run TestCrawlBilibili -timeout 5m
go test -v ./cmd/cli/... -run TestCrawlDouyin -timeout 5m
go test -v ./cmd/cli/... -run TestCrawlXiaohongshu -timeout 5m

# 使用测试脚本（PowerShell）
.\test_integration.ps1

# 使用测试脚本（Bash）
bash test_integration.sh
```

### 集成测试

项目包含针对以下网站的集成测试：

- **百度新闻** (`https://news.baidu.com/`)
- **Bilibili 热门** (`https://www.bilibili.com/v/popular`)
- **抖音** (`https://www.douyin.com/`)
- **小红书** (`https://www.xiaohongshu.com/explore`)

这些测试会实际爬取网站并保存结果到 `test_output/` 目录，用于验证爬虫在不同网站上的表现。

**注意**：
- 集成测试需要较长时间（每个测试约 30-60 秒）
- 测试会实际访问网站，请遵守网站的 robots.txt 和使用条款
- 建议使用 `-short` 标志跳过集成测试：`go test -short ./...`

### 代码检查

```bash
go vet ./...
```

## License

MIT
