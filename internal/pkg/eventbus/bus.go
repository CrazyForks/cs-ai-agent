package eventbus

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

type Handler[T any] func(ctx context.Context, event T)

type ErrorHandler func(ctx context.Context, err error)

type Bus[T any] struct {
	mu       sync.RWMutex
	handlers map[uint64]Handler[T]
	nextID   uint64

	onError ErrorHandler

	// 用于限制 PublishAsync 的 goroutine 并发数；nil 表示不限制
	asyncSem chan struct{}
}

func New[T any](opts ...Option[T]) *Bus[T] {
	b := &Bus[T]{
		handlers: make(map[uint64]Handler[T]),
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

type Option[T any] func(*Bus[T])

func WithErrorHandler[T any](fn ErrorHandler) Option[T] {
	return func(b *Bus[T]) {
		b.onError = fn
	}
}

func WithAsyncConcurrency[T any](n int) Option[T] {
	return func(b *Bus[T]) {
		if n > 0 {
			b.asyncSem = make(chan struct{}, n)
		}
	}
}

// Subscribe 返回 handlerID 和取消订阅函数
func (b *Bus[T]) Subscribe(h Handler[T]) (uint64, func()) {
	id := atomic.AddUint64(&b.nextID, 1)

	b.mu.Lock()
	b.handlers[id] = h
	b.mu.Unlock()

	return id, func() {
		b.Unsubscribe(id)
	}
}

func (b *Bus[T]) SubscribeOnce(h Handler[T]) (uint64, func()) {
	var id uint64

	wrapper := func(ctx context.Context, event T) {
		b.Unsubscribe(id)
		h(ctx, event)
	}

	id = atomic.AddUint64(&b.nextID, 1)

	b.mu.Lock()
	b.handlers[id] = wrapper
	b.mu.Unlock()

	return id, func() {
		b.Unsubscribe(id)
	}
}

func (b *Bus[T]) Unsubscribe(id uint64) {
	b.mu.Lock()
	delete(b.handlers, id)
	b.mu.Unlock()
}

func (b *Bus[T]) Publish(ctx context.Context, event T) {
	handlers := b.snapshotHandlers()
	for _, h := range handlers {
		b.callHandler(ctx, h, event)
	}
}

func (b *Bus[T]) PublishAsync(ctx context.Context, event T) {
	handlers := b.snapshotHandlers()
	for _, h := range handlers {
		if b.asyncSem != nil {
			b.asyncSem <- struct{}{}
			go func() {
				defer func() { <-b.asyncSem }()
				b.callHandler(ctx, h, event)
			}()
			continue
		}

		go b.callHandler(ctx, h, event)
	}
}

func (b *Bus[T]) HandlerCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.handlers)
}

func (b *Bus[T]) snapshotHandlers() []Handler[T] {
	b.mu.RLock()
	defer b.mu.RUnlock()

	handlers := make([]Handler[T], 0, len(b.handlers))
	for _, h := range b.handlers {
		handlers = append(handlers, h)
	}
	return handlers
}

func (b *Bus[T]) callHandler(ctx context.Context, h Handler[T], event T) {
	defer func() {
		if r := recover(); r != nil {
			if b.onError != nil {
				b.onError(ctx, fmt.Errorf("event handler panic: %v\n%s", r, debug.Stack()))
			}
		}
	}()
	h(ctx, event)
}
