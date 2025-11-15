package crawler

import (
	"strings"
	"testing"
)

func TestExtractor_ExtractMainContent(t *testing.T) {
	extractor := NewExtractor()

	html := `
		<html>
			<head><title>Test</title></head>
			<body>
				<nav>Navigation</nav>
				<article>
					<h1>Main Title</h1>
					<p>This is the main content.</p>
					<p>Another paragraph.</p>
				</article>
				<aside>Sidebar</aside>
			</body>
		</html>
	`

	result, err := extractor.ExtractMainContent(html)
	if err != nil {
		t.Fatalf("ExtractMainContent() error = %v", err)
	}

	if !strings.Contains(result, "Main Title") {
		t.Error("Expected result to contain 'Main Title'")
	}

	if !strings.Contains(result, "main content") {
		t.Error("Expected result to contain 'main content'")
	}

	// 应该不包含导航和侧边栏
	if strings.Contains(result, "Navigation") {
		t.Error("Result should not contain navigation")
	}
	if strings.Contains(result, "Sidebar") {
		t.Error("Result should not contain sidebar")
	}
}

func TestExtractor_ExtractText(t *testing.T) {
	extractor := NewExtractor()

	html := `
		<html>
			<body>
				<h1>Title</h1>
				<p>Paragraph 1</p>
				<p>Paragraph 2</p>
			</body>
		</html>
	`

	result, err := extractor.ExtractText(html)
	if err != nil {
		t.Fatalf("ExtractText() error = %v", err)
	}

	if !strings.Contains(result, "Title") {
		t.Error("Expected result to contain 'Title'")
	}

	if !strings.Contains(result, "Paragraph 1") {
		t.Error("Expected result to contain 'Paragraph 1'")
	}
}


