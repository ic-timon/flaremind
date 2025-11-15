package crawler

import (
	"context"
	"errors"
	"strings"
	"time"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries      int
	InitialDelay    time.Duration
	MaxDelay        time.Duration
	BackoffMultiplier float64
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:       3,
		InitialDelay:     1 * time.Second,
		MaxDelay:         10 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// RetryableFunc 可重试的函数类型
type RetryableFunc func() error

// Retry 执行带重试的操作
func Retry(ctx context.Context, fn RetryableFunc, config RetryConfig) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 执行操作
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// 如果是最后一次尝试，不再等待
		if attempt == config.MaxRetries {
			break
		}

		// 等待后重试
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// 指数退避
			delay = time.Duration(float64(delay) * config.BackoffMultiplier)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}

	return lastErr
}

// IsRetryableError 判断错误是否可重试
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	// 网络超时错误
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	// 可以添加更多可重试的错误类型
	errStr := err.Error()
	retryablePatterns := []string{
		"timeout",
		"connection",
		"network",
		"temporary",
		"temporarily unavailable",
	}

	for _, pattern := range retryablePatterns {
		if contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// contains 检查字符串是否包含子字符串（不区分大小写）
func contains(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}

