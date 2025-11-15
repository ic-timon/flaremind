package crawler

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Extractor 内容提取器
type Extractor struct {
}

// NewExtractor 创建新的提取器
func NewExtractor() *Extractor {
	return &Extractor{}
}

// ExtractMainContent 提取主要内容（使用 Readability 类似算法）
func (e *Extractor) ExtractMainContent(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	// 移除不需要的元素
	doc.Find("script, style, nav, header, footer, aside, .ad, .advertisement, .sidebar, .navigation, .menu, .social, .share, .comments, .comment, iframe, embed, object").Remove()

	// 尝试找到主要内容区域
	var content *goquery.Selection

	// 优先级：article > main > .content > .post > body
	if selection := doc.Find("article"); selection.Length() > 0 {
		content = selection.First()
	} else if selection := doc.Find("main"); selection.Length() > 0 {
		content = selection.First()
	} else if selection := doc.Find(".content, .post, .article, .entry, .post-content"); selection.Length() > 0 {
		// 使用 Readability 算法选择最佳内容区域
		content = e.selectBestContent(selection)
	} else {
		// 使用 Readability 算法从 body 中选择最佳内容
		content = e.selectBestContent(doc.Find("body"))
	}

	// 清理内容
	e.cleanContent(content)

	// 提取 HTML
	result, err := content.Html()
	if err != nil {
		return "", err
	}

	return result, nil
}

// selectBestContent 使用 Readability 算法选择最佳内容区域
func (e *Extractor) selectBestContent(selections *goquery.Selection) *goquery.Selection {
	var bestScore float64
	var bestSelection *goquery.Selection

	selections.Each(func(i int, s *goquery.Selection) {
		score := e.calculateContentScore(s)
		if score > bestScore {
			bestScore = score
			bestSelection = s
		}
	})

	if bestSelection == nil {
		return selections.First()
	}

	return bestSelection
}

// calculateContentScore 计算内容分数（Readability 算法）
func (e *Extractor) calculateContentScore(s *goquery.Selection) float64 {
	// 获取文本内容
	text := s.Text()
	textLength := len(strings.TrimSpace(text))

	// 如果文本太短，分数很低
	if textLength < 100 {
		return 0
	}

	// 计算链接密度
	linkText := ""
	s.Find("a").Each(func(i int, a *goquery.Selection) {
		linkText += a.Text() + " "
	})
	linkDensity := float64(len(strings.TrimSpace(linkText))) / float64(textLength)
	if linkDensity > 0.5 {
		return 0 // 链接密度太高，可能是导航或广告
	}

	// 计算段落数量
	paragraphCount := s.Find("p").Length()

	// 计算列表数量
	listCount := s.Find("ul, ol").Length()

	// 计算图片数量
	imageCount := s.Find("img").Length()

	// 基础分数：文本长度
	score := float64(textLength)

	// 加分项
	score += float64(paragraphCount) * 10  // 段落加分
	score += float64(listCount) * 5         // 列表加分
	score += float64(imageCount) * 2        // 图片加分

	// 检查是否包含常见的内容标识符
	class := s.AttrOr("class", "")
	id := s.AttrOr("id", "")
	contentKeywords := []string{"content", "article", "post", "entry", "main", "body"}
	for _, keyword := range contentKeywords {
		if strings.Contains(strings.ToLower(class), keyword) || strings.Contains(strings.ToLower(id), keyword) {
			score += 50
		}
	}

	// 减分项：包含广告或导航关键词
	negativeKeywords := []string{"ad", "advertisement", "sidebar", "nav", "menu", "footer", "header"}
	for _, keyword := range negativeKeywords {
		if strings.Contains(strings.ToLower(class), keyword) || strings.Contains(strings.ToLower(id), keyword) {
			score -= 100
		}
	}

	return score
}

// cleanContent 清理内容
func (e *Extractor) cleanContent(s *goquery.Selection) {
	// 移除空的段落和 div
	s.Find("p, div").Each(func(i int, elem *goquery.Selection) {
		text := strings.TrimSpace(elem.Text())
		if text == "" {
			elem.Remove()
		}
	})

	// 移除只有空白字符的元素
	s.Find("*").Each(func(i int, elem *goquery.Selection) {
		text := strings.TrimSpace(elem.Text())
		html, _ := elem.Html()
		// 如果元素只包含空白或换行符
		if text == "" && regexp.MustCompile(`^\s*$`).MatchString(html) {
			elem.Remove()
		}
	})

	// 移除属性（保留必要的）
	s.Find("*").Each(func(i int, elem *goquery.Selection) {
		// 保留图片的 src 和 alt
		if goquery.NodeName(elem) == "img" {
			return
		}
		// 保留链接的 href
		if goquery.NodeName(elem) == "a" {
			return
		}
		// 移除其他属性
		elem.RemoveAttr("class")
		elem.RemoveAttr("id")
		elem.RemoveAttr("style")
		elem.RemoveAttr("onclick")
	})
}

// ExtractText 提取纯文本内容
func (e *Extractor) ExtractText(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	// 移除脚本和样式
	doc.Find("script, style").Remove()

	// 获取文本
	text := doc.Find("body").Text()
	// 清理空白字符
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n\n\n", "\n\n")

	return text, nil
}

