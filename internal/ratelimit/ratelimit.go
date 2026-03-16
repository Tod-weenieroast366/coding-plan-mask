// Package ratelimit 提供并发安全的速率限制
package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Limiter 速率限制器
type Limiter struct {
	mu       sync.RWMutex
	limiters map[string]*rate.Limiter
	rate     int // 每秒请求数
	burst    int // 突发请求数
}

// NewLimiter 创建新的速率限制器
func NewLimiter(requestsPerWindow int, window time.Duration) *Limiter {
	// 将窗口内的请求数转换为每秒请求数
	ratePerSecond := float64(requestsPerWindow) / window.Seconds()
	if ratePerSecond < 1 {
		ratePerSecond = 1
	}

	burst := requestsPerWindow / 10 // 突发为窗口的 1/10
	if burst < 1 {
		burst = 1
	}

	return &Limiter{
		limiters: make(map[string]*rate.Limiter),
		rate:     int(ratePerSecond),
		burst:    burst,
	}
}

// Allow 检查指定 key 是否允许请求
func (l *Limiter) Allow(key string) bool {
	l.mu.RLock()
	limiter, exists := l.limiters[key]
	l.mu.RUnlock()

	if !exists {
		l.mu.Lock()
		// 双重检查
		if limiter, exists = l.limiters[key]; !exists {
			limiter = rate.NewLimiter(rate.Limit(l.rate), l.burst)
			l.limiters[key] = limiter
		}
		l.mu.Unlock()
	}

	return limiter.Allow()
}

// Wait 阻塞等待直到允许请求
func (l *Limiter) Wait(key string) {
	l.mu.RLock()
	limiter, exists := l.limiters[key]
	l.mu.RUnlock()

	if !exists {
		l.mu.Lock()
		if limiter, exists = l.limiters[key]; !exists {
			limiter = rate.NewLimiter(rate.Limit(l.rate), l.burst)
			l.limiters[key] = limiter
		}
		l.mu.Unlock()
	}

	limiter.Wait(nil)
}

// GlobalLimiter 全局速率限制器
type GlobalLimiter struct {
	mu           sync.RWMutex
	count        int
	windowStart  time.Time
	windowSize   time.Duration
	maxRequests  int
}

// NewGlobalLimiter 创建全局速率限制器
func NewGlobalLimiter(maxRequests int, windowSize time.Duration) *GlobalLimiter {
	return &GlobalLimiter{
		windowStart: time.Now(),
		windowSize:  windowSize,
		maxRequests: maxRequests,
	}
}

// Allow 检查是否允许请求（全局）
func (l *GlobalLimiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// 重置窗口
	if now.Sub(l.windowStart) > l.windowSize {
		l.count = 0
		l.windowStart = now
	}

	if l.count >= l.maxRequests {
		return false
	}

	l.count++
	return true
}

// Reset 重置计数器
func (l *GlobalLimiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.count = 0
	l.windowStart = time.Now()
}

// Stats 返回统计信息
func (l *GlobalLimiter) Stats() (count int, maxRequests int, remaining time.Duration) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	elapsed := time.Since(l.windowStart)
	if elapsed > l.windowSize {
		remaining = 0
	} else {
		remaining = l.windowSize - elapsed
	}

	return l.count, l.maxRequests, remaining
}
