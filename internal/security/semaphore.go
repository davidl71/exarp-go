// Package security provides semaphore-based concurrency limiting for tool execution.
package security

import (
	"sync"
)

type SemaphoreLimiter struct {
	acquire func() bool
	release func()
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

func (s *SemaphoreLimiter) Release() {
	s.release()
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
