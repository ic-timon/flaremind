package crawler

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/chromedp/chromedp"
)

// Renderer 页面渲染器
type Renderer struct {
	timeout  time.Duration
	headless bool
}

// NewRenderer 创建新的渲染器
func NewRenderer(timeout time.Duration, headless bool) *Renderer {
	return &Renderer{
		timeout:  timeout,
		headless: headless,
	}
}

// Render 渲染页面并返回 HTML（带重试机制）
func (r *Renderer) Render(url string) (string, error) {
	ctx := context.Background()
	retryConfig := DefaultRetryConfig()

	var html string
	var lastErr error

	err := Retry(ctx, func() error {
		html, lastErr = r.renderPage(url)
		if lastErr != nil && IsRetryableError(lastErr) {
			return lastErr
		}
		return nil
	}, retryConfig)

	if err != nil {
		return "", err
	}

	if lastErr != nil {
		return "", lastErr
	}

	return html, nil
}

// findChromePath 查找 Chrome 可执行文件路径
func findChromePath() string {
	if runtime.GOOS == "windows" {
		// Windows 上常见的 Chrome 安装路径
		paths := []string{
			`C:\Program Files\Google\Chrome\Application\chrome.exe`,
			`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
			filepath.Join(os.Getenv("LOCALAPPDATA"), `Google\Chrome\Application\chrome.exe`),
			filepath.Join(os.Getenv("PROGRAMFILES"), `Google\Chrome\Application\chrome.exe`),
			filepath.Join(os.Getenv("PROGRAMFILES(X86)"), `Google\Chrome\Application\chrome.exe`),
		}
		for _, path := range paths {
			if _, err := os.Stat(path); err == nil {
				return path
			}
		}
		// 尝试从 PATH 中查找
		if path, err := exec.LookPath("chrome.exe"); err == nil {
			return path
		}
		if path, err := exec.LookPath("chromium.exe"); err == nil {
			return path
		}
	}
	// Linux/Mac 上 chromedp 会自动查找
	return ""
}

// renderPage 实际渲染页面的方法
func (r *Renderer) renderPage(url string) (string, error) {
	// 创建上下文
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// 设置超时
	ctx, cancel = context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// 创建选项
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", r.headless),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		// 设置 User-Agent 模拟真实浏览器
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
	)

	// 在 Windows 上指定 Chrome 路径
	if chromePath := findChromePath(); chromePath != "" {
		opts = append(opts, chromedp.ExecPath(chromePath))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	var html string

	// 执行任务
	err := chromedp.Run(ctx,
		// 隐藏 webdriver 特征
		chromedp.Evaluate(`Object.defineProperty(navigator, 'webdriver', {get: () => undefined})`, nil),
		chromedp.Navigate(url),
		chromedp.WaitVisible("body", chromedp.ByQuery),
		// 等待页面加载完成
		chromedp.Sleep(2*time.Second),
		// 滚动页面以触发懒加载
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight)`, nil),
		chromedp.Sleep(1*time.Second),
		chromedp.Evaluate(`window.scrollTo(0, 0)`, nil),
		chromedp.Sleep(1*time.Second),
		// 获取完整 HTML
		chromedp.OuterHTML("html", &html),
	)

	if err != nil {
		return "", err
	}

	return html, nil
}
