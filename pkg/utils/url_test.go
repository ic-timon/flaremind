package utils

import "testing"

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"https://example.com/page#fragment", "https://example.com/page", false},
		{"https://Example.com/Page/", "https://example.com/Page", false},
		{"https://example.com", "https://example.com", false},
		{"invalid-url", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := NormalizeURL(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NormalizeURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("NormalizeURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsSameDomain(t *testing.T) {
	tests := []struct {
		url1     string
		url2     string
		expected bool
	}{
		{"https://example.com/page1", "https://example.com/page2", true},
		{"https://example.com", "https://other.com", false},
		{"http://example.com", "https://example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.url1+" vs "+tt.url2, func(t *testing.T) {
			result := IsSameDomain(tt.url1, tt.url2)
			if result != tt.expected {
				t.Errorf("IsSameDomain() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestResolveURL(t *testing.T) {
	tests := []struct {
		base     string
		relative string
		expected string
		wantErr  bool
	}{
		{"https://example.com/page1", "/page2", "https://example.com/page2", false},
		{"https://example.com/", "page2", "https://example.com/page2", false},
		{"https://example.com", "https://other.com/page", "https://other.com/page", false},
	}

	for _, tt := range tests {
		t.Run(tt.base+" + "+tt.relative, func(t *testing.T) {
			result, err := ResolveURL(tt.base, tt.relative)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("ResolveURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"ftp://example.com", false},
		{"not-a-url", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := IsValidURL(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidURL() = %v, want %v", result, tt.expected)
			}
		})
	}
}


