package utils

import (
	"errors"
	"net/url"
	"strings"
)

// NormalizeURL 规范化 URL
func NormalizeURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	// 检查是否有有效的 scheme
	if u.Scheme == "" {
		return "", errors.New("URL must have a scheme (http or https)")
	}

	// 只接受 http 和 https
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errors.New("URL scheme must be http or https")
	}

	// 移除 fragment
	u.Fragment = ""

	// 转换为小写
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// 移除末尾的斜杠（除了根路径）
	if u.Path != "/" {
		u.Path = strings.TrimSuffix(u.Path, "/")
	}

	return u.String(), nil
}

// IsSameDomain 检查两个 URL 是否属于同一域名
func IsSameDomain(url1, url2 string) bool {
	u1, err1 := url.Parse(url1)
	u2, err2 := url.Parse(url2)

	if err1 != nil || err2 != nil {
		return false
	}

	return u1.Host == u2.Host
}

// ResolveURL 解析相对 URL 为绝对 URL
func ResolveURL(baseURL, relativeURL string) (string, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	rel, err := url.Parse(relativeURL)
	if err != nil {
		return "", err
	}

	return base.ResolveReference(rel).String(), nil
}

// IsValidURL 检查 URL 是否有效
func IsValidURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}


