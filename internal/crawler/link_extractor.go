package crawler

import (
	"strings"

	"flaremind/pkg/utils"

	"github.com/PuerkitoBio/goquery"
)

// LinkExtractor 链接提取器
type LinkExtractor struct {
	baseURL string
}

// NewLinkExtractor 创建新的链接提取器
func NewLinkExtractor(baseURL string) *LinkExtractor {
	return &LinkExtractor{
		baseURL: baseURL,
	}
}

// ExtractLinks 从 HTML 中提取所有链接
func (le *LinkExtractor) ExtractLinks(html string, allowedDomains []string) ([]string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	var links []string
	seen := make(map[string]bool)

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")
		if href == "" {
			return
		}

		// 解析为绝对 URL
		absoluteURL, err := utils.ResolveURL(le.baseURL, href)
		if err != nil {
			return
		}

		// 规范化 URL
		normalized, err := utils.NormalizeURL(absoluteURL)
		if err != nil {
			return
		}

		// 检查是否有效
		if !utils.IsValidURL(normalized) {
			return
		}

		// 检查域名限制
		if len(allowedDomains) > 0 {
			allowed := false
			for _, domain := range allowedDomains {
				// 处理带协议和不带协议的域名
				domainURL := domain
				if !strings.HasPrefix(domain, "http://") && !strings.HasPrefix(domain, "https://") {
					domainURL = "https://" + domain
				}
				if utils.IsSameDomain(normalized, domainURL) {
					allowed = true
					break
				}
			}
			if !allowed {
				return
			}
		} else {
			// 如果没有指定允许的域名，只允许同域名的链接
			if !utils.IsSameDomain(normalized, le.baseURL) {
				return
			}
		}

		// 去重
		if !seen[normalized] {
			seen[normalized] = true
			links = append(links, normalized)
		}
	})

	return links, nil
}

