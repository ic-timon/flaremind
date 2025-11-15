package crawler

import (
	"strings"
	"testing"
)

func TestConverter_HTMLToMarkdown(t *testing.T) {
	converter := NewConverter()

	tests := []struct {
		name     string
		html     string
		contains []string
	}{
		{
			name:     "Heading",
			html:     "<h1>Title</h1>",
			contains: []string{"# Title"},
		},
		{
			name:     "Paragraph",
			html:     "<p>This is a paragraph.</p>",
			contains: []string{"This is a paragraph."},
		},
		{
			name:     "Link",
			html:     `<a href="https://example.com">Link</a>`,
			contains: []string{"[Link](https://example.com)"},
		},
		{
			name:     "Image",
			html:     `<img src="image.jpg" alt="Image">`,
			contains: []string{"![Image](image.jpg)"},
		},
		{
			name:     "List",
			html:     "<ul><li>Item 1</li><li>Item 2</li></ul>",
			contains: []string{"- Item 1", "- Item 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := converter.HTMLToMarkdown(tt.html)
			if err != nil {
				t.Fatalf("HTMLToMarkdown() error = %v", err)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', got: %s", expected, result)
				}
			}
		})
	}
}


