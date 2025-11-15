package crawler

import (
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Converter Markdown 转换器
type Converter struct {
}

// NewConverter 创建新的转换器
func NewConverter() *Converter {
	return &Converter{}
}

// HTMLToMarkdown 将 HTML 转换为 Markdown
func (c *Converter) HTMLToMarkdown(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	var markdown strings.Builder
	var lastWasTable bool

	// 遍历所有元素
	doc.Find("body").Children().Each(func(i int, s *goquery.Selection) {
		isTable := goquery.NodeName(s) == "table"
		elementContent := c.convertElement(s)
		
		// 如果是表格且前面不是表格，添加空行
		if isTable && !lastWasTable {
			markdown.WriteString("\n")
		}
		
		markdown.WriteString(elementContent)
		
		// 表格后不需要立即添加空行，会在下一个非表格元素时添加
		if !isTable {
			markdown.WriteString("\n\n")
		} else {
			markdown.WriteString("\n")
		}
		
		lastWasTable = isTable
	})

	result := markdown.String()
	
	// 确保表格前有空行（表格以 | 开头）
	// 匹配模式：非空行后紧跟表格行（以 | 开头），但表格行之间不应该有空行
	result = regexp.MustCompile(`([^\n\s])\n\n(\|)`).ReplaceAllString(result, "$1\n\n$2")
	// 移除表格行之间的空行
	result = regexp.MustCompile(`(\|.*\|)\n\n(\|)`).ReplaceAllString(result, "$1\n$2")
	
	// 清理多余的空行
	result = regexp.MustCompile(`\n{4,}`).ReplaceAllString(result, "\n\n\n")
	result = strings.TrimSpace(result)

	return result, nil
}

// convertElement 转换单个元素
func (c *Converter) convertElement(s *goquery.Selection) string {
	tagName := goquery.NodeName(s)
	var result strings.Builder

	switch tagName {
	case "h1":
		result.WriteString("# " + strings.TrimSpace(s.Text()) + "\n")
	case "h2":
		result.WriteString("## " + strings.TrimSpace(s.Text()) + "\n")
	case "h3":
		result.WriteString("### " + strings.TrimSpace(s.Text()) + "\n")
	case "h4":
		result.WriteString("#### " + strings.TrimSpace(s.Text()) + "\n")
	case "h5":
		result.WriteString("##### " + strings.TrimSpace(s.Text()) + "\n")
	case "h6":
		result.WriteString("###### " + strings.TrimSpace(s.Text()) + "\n")
	case "p":
		result.WriteString(c.convertInline(s) + "\n")
	case "ul", "ol":
		s.Find("li").Each(func(i int, li *goquery.Selection) {
			prefix := "- "
			if tagName == "ol" {
				prefix = strings.Repeat(" ", len(string(rune(i+1)))+1) + "1. "
			}
			result.WriteString(prefix + strings.TrimSpace(c.convertInline(li)) + "\n")
		})
	case "blockquote":
		lines := strings.Split(strings.TrimSpace(s.Text()), "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				result.WriteString("> " + strings.TrimSpace(line) + "\n")
			}
		}
	case "pre":
		code := s.Text()
		// 检查是否包含 code 标签
		if s.Find("code").Length() > 0 {
			lang := s.Find("code").First().AttrOr("class", "")
			if strings.HasPrefix(lang, "language-") {
				lang = strings.TrimPrefix(lang, "language-")
			}
			result.WriteString("```" + lang + "\n")
			result.WriteString(strings.TrimSpace(code) + "\n")
			result.WriteString("```\n")
		} else {
			result.WriteString("```\n")
			result.WriteString(strings.TrimSpace(code) + "\n")
			result.WriteString("```\n")
		}
	case "code":
		result.WriteString("`" + strings.TrimSpace(s.Text()) + "`")
	case "img":
		alt := s.AttrOr("alt", "")
		src := s.AttrOr("src", "")
		result.WriteString("![" + alt + "](" + src + ")\n")
	case "a":
		text := strings.TrimSpace(s.Text())
		href := s.AttrOr("href", "")
		result.WriteString("[" + text + "](" + href + ")")
	case "strong", "b":
		result.WriteString("**" + strings.TrimSpace(s.Text()) + "**")
	case "em", "i":
		result.WriteString("*" + strings.TrimSpace(s.Text()) + "*")
	case "hr":
		result.WriteString("---\n")
	case "table":
		result.WriteString(c.convertTable(s))
	default:
		// 递归处理子元素
		s.Contents().Each(func(i int, child *goquery.Selection) {
			if goquery.NodeName(child) == "#text" {
				text := strings.TrimSpace(child.Text())
				if text != "" {
					result.WriteString(text + " ")
				}
			} else {
				result.WriteString(c.convertElement(child))
			}
		})
	}

	return result.String()
}

// convertInline 转换内联元素
func (c *Converter) convertInline(s *goquery.Selection) string {
	var result strings.Builder

	s.Contents().Each(func(i int, child *goquery.Selection) {
		nodeName := goquery.NodeName(child)
		if nodeName == "#text" {
			result.WriteString(strings.TrimSpace(child.Text()) + " ")
		} else {
			result.WriteString(c.convertElement(child))
		}
	})

	return strings.TrimSpace(result.String())
}

// convertTable 转换表格
func (c *Converter) convertTable(s *goquery.Selection) string {
	var result strings.Builder

	// 在表格前添加空行，确保表格能正常渲染
	result.WriteString("\n")

	// 处理表头
	headers := []string{}
	s.Find("thead tr, tr").First().Find("th, td").Each(func(i int, th *goquery.Selection) {
		headers = append(headers, strings.TrimSpace(th.Text()))
	})

	if len(headers) > 0 {
		// 写入表头
		result.WriteString("| " + strings.Join(headers, " | ") + " |\n")
		// 写入分隔符
		separator := make([]string, len(headers))
		for i := range separator {
			separator[i] = "---"
		}
		result.WriteString("| " + strings.Join(separator, " | ") + " |\n")

		// 处理表格行
		s.Find("tbody tr, tr").Slice(1, goquery.ToEnd).Each(func(i int, tr *goquery.Selection) {
			cells := []string{}
			tr.Find("td").Each(func(j int, td *goquery.Selection) {
				cells = append(cells, strings.TrimSpace(td.Text()))
			})
			if len(cells) > 0 {
				result.WriteString("| " + strings.Join(cells, " | ") + " |\n")
			}
		})
	}

	return result.String()
}


