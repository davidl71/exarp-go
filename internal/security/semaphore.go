// semaphore.go — Semaphore-based concurrency limiting for tool execution.
package security

import (
	"context"
	"sync"
	"time"
)

type SemaphoreLimiter struct {
	acquire func() bool
	release func()
	mu      sync.Mutex
}

func NewSemaphoreLimiter(permits int) *SemaphoreLimiter {
	if permits <= 0 {
		permits = 10
	}
	ch := make(chan struct{}, permits)

	return &SemaphoreLimiter{
		acquire: func() bool {
			select {
			case ch <- struct{}{}:
				return true
			default:
				return false
			}
		},
		release: func() {
			select {
			case <-ch:
			default:
			}
		},
	}
}

func (s *SemaphoreLimiter) TryAcquire() bool {
	return s.acquire()
}

func (s *SemaphoreLimiter) Acquire(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	s.mu.Lock()
	ch := make(chan struct{}, 1)
	s.mu.Unlock()

	select {
	case ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *SemaphoreLimiter) Release() {
	s.release()
}

type SemaphoreLimiterPool struct {
	limiters    map[string]*SemaphoreLimiter
	mu          sync.RWMutex
	defaultLimi int
}

func NewSemaphoreLimiterPool(defaultLimit int) *SemaphoreLimiterPool {
	if defaultLimit <= 0 {
		defaultLimit = 10
	}
	return &SemaphoreLimiterPool{
		limiters:    make(map[string]*SemaphoreLimiter),
		defaultLimi: defaultLimit,
	}
}

func (p *SemaphoreLimiterPool) Get(name string) *SemaphoreLimiter {
	p.mu.RLock()
	limiter, exists := p.limiters[name]
	p.mu.RUnlock()

	if exists {
		return limiter
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if limiter, exists = p.limiters[name]; exists {
		return limiter
	}

	limiter = NewSemaphoreLimiter(p.defaultLimi)
	p.limiters[name] = limiter
	return limiter
}

var (
	globalToolSemaphore *SemaphoreLimiter
	toolSemaphoreOnce   sync.Once
)

func GetToolSemaphore(permits int) *SemaphoreLimiter {
	toolSemaphoreOnce.Do(func() {
		globalToolSemaphore = NewSemaphoreLimiter(permits)
	})
	return globalToolSemaphore
}

type SemaphoreContextKey string

const SemaphoreKey SemaphoreContextKey = "tool_semaphore"

func ContextWithSemaphore(ctx context.Context, limiter *SemaphoreLimiter) context.Context {
	return context.WithValue(ctx, SemaphoreKey, limiter)
}

func SemaphoreFromContext(ctx context.Context) *SemaphoreLimiter {
	if limiter, ok := ctx.Value(SemaphoreKey).(*SemaphoreLimiter); ok {
		return limiter
	}
	return nil
}

func WaitForSemaphore(ctx context.Context, limiter *SemaphoreLimiter, timeout time.Duration) error {
	if limiter == nil {
		return nil
	}

	acquired := make(chan error, 1)

	go func() {
		err := limiter.Acquire(ctx)
		acquired <- err
	}()

	select {
	case err := <-acquired:
		return err
	case <-time.After(timeout):
		limiter.Release()
		return context.DeadlineExceeded
	}
}
