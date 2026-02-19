package ctxcache

import (
	"context"
	"errors"
	"sync"

	"google.golang.org/grpc"
)

var ErrContextCacheKeyNotExists = errors.New("context cache key correct init")

// 定义缓存数据的类型
type contextCacheKey struct{}

// contextCacheData 代表你想要缓存的数据
type contextCacheData struct {
	// 存储你的缓存数据
	m sync.Map
}

// NewContextWithCache Creates a new child context with cache
func NewContextWithCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextCacheKey{}, &contextCacheData{})
}

// Get Try to get a value from context
func Get(ctx context.Context, key any) (interface{}, bool) {
	cache, ok := ctx.Value(contextCacheKey{}).(*contextCacheData)
	if !ok {
		return nil, false
	}
	return cache.m.Load(key)
}

// Set Add a new value into context cache
func Set(ctx context.Context, key, val any) error {
	cache, ok := ctx.Value(contextCacheKey{}).(*contextCacheData)
	if !ok {
		return ErrContextCacheKeyNotExists
	}
	cache.m.Store(key, val)
	return nil
}

// wrapServerStream server stream wrap
type wrapServerStream struct {
	grpc.ServerStream
	WrappedContext context.Context
}

// Context Return a warped stream context
func (w *wrapServerStream) Context() context.Context {
	return w.WrappedContext
}

// ContextCacheUnaryInterceptor unary interceptor to init context cache
func ContextCacheUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx = NewContextWithCache(ctx)
		resp, err := handler(ctx, req)
		return resp, err
	}
}

// ContextCacheStreamInterceptor stream interceptor to init context cache
func ContextCacheStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := NewContextWithCache(ss.Context())
		wrappedStream := &wrapServerStream{
			ServerStream: ss,
		}
		wrappedStream.WrappedContext = ctx
		return handler(srv, wrappedStream)
	}
}
