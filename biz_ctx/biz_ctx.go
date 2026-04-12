package biz_ctx

import (
	"context"
	"sync"
)

// BizSession is the technical identity for one business session.
type BizSession interface {
	BizSessionId() string
}

// BizInstance is the technical identity for one business instance under a session.
type BizInstance interface {
	BizInstanceId() string
}

// BizCtx describes a 1:N relationship from biz_session to biz_instance.
type BizCtx interface {
	Set(ctx context.Context, instance BizInstance) bool
	Get(ctx context.Context, instanceID string) (BizInstance, bool)
	Del(ctx context.Context, instanceID string) (BizInstance, bool)
	ForEach(ctx context.Context, fn func(instance BizInstance))
	List(ctx context.Context) []BizInstance
}

type bizSessionContextKey struct{}

// WithBizSession writes BizSession into context.Context.
func WithBizSession(ctx context.Context, session BizSession) context.Context {
	return context.WithValue(ctx, bizSessionContextKey{}, session)
}

// BizSessionFromContext reads BizSession from context.Context.
func BizSessionFromContext(ctx context.Context) (BizSession, bool) {
	if ctx == nil {
		return nil, false
	}
	session, ok := ctx.Value(bizSessionContextKey{}).(BizSession)
	if !ok || session == nil {
		return nil, false
	}
	return session, true
}

// Ctx is the default concurrency-safe BizCtx implementation.
// The zero value is ready to use.
type Ctx struct {
	mu   sync.RWMutex
	data map[string]map[string]BizInstance
}

var _ BizCtx = (*Ctx)(nil)

func NewBizCtx() BizCtx {
	return &Ctx{}
}

func (c *Ctx) Set(ctx context.Context, instance BizInstance) bool {
	sessionID, ok := bizSessionIDFromContext(ctx)
	if !ok {
		return false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.data == nil {
		c.data = make(map[string]map[string]BizInstance)
	}
	bucket := c.data[sessionID]
	if bucket == nil {
		bucket = make(map[string]BizInstance)
		c.data[sessionID] = bucket
	}
	bucket[instance.BizInstanceId()] = instance
	return true
}

func (c *Ctx) Get(ctx context.Context, instanceID string) (BizInstance, bool) {
	sessionID, ok := bizSessionIDFromContext(ctx)
	if !ok {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.data == nil {
		return nil, false
	}
	bucket := c.data[sessionID]
	if bucket == nil {
		return nil, false
	}
	instance, ok := bucket[instanceID]
	return instance, ok
}

func (c *Ctx) Del(ctx context.Context, instanceID string) (BizInstance, bool) {
	sessionID, ok := bizSessionIDFromContext(ctx)
	if !ok {
		return nil, false
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.data == nil {
		return nil, false
	}
	bucket := c.data[sessionID]
	if bucket == nil {
		return nil, false
	}

	instance, ok := bucket[instanceID]
	if !ok {
		return nil, false
	}

	delete(bucket, instanceID)
	if len(bucket) == 0 {
		delete(c.data, sessionID)
	}
	return instance, true
}

func (c *Ctx) ForEach(ctx context.Context, fn func(instance BizInstance)) {
	sessionID, ok := bizSessionIDFromContext(ctx)
	if !ok {
		return
	}

	c.mu.RLock()
	bucket := c.data[sessionID]
	if len(bucket) == 0 {
		c.mu.RUnlock()
		return
	}

	snapshot := make([]BizInstance, 0, len(bucket))
	for _, instance := range bucket {
		snapshot = append(snapshot, instance)
	}
	c.mu.RUnlock()

	for _, instance := range snapshot {
		fn(instance)
	}
}

func (c *Ctx) List(ctx context.Context) []BizInstance {
	sessionID, ok := bizSessionIDFromContext(ctx)
	if !ok {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	bucket := c.data[sessionID]
	if len(bucket) == 0 {
		return nil
	}

	result := make([]BizInstance, 0, len(bucket))
	for _, instance := range bucket {
		result = append(result, instance)
	}
	return result
}

func bizSessionIDFromContext(ctx context.Context) (string, bool) {
	session, ok := BizSessionFromContext(ctx)
	if !ok {
		return "", false
	}
	sessionID := session.BizSessionId()
	if sessionID == "" {
		return "", false
	}
	return sessionID, true
}
