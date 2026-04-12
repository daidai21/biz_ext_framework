package biz_ctx

import "sync"

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
	Set(session BizSession, instance BizInstance)
	Get(sessionID string, instanceID string) (BizInstance, bool)
	Del(sessionID string, instanceID string) (BizInstance, bool)
	ForEach(sessionID string, fn func(instance BizInstance))
	List(sessionID string) []BizInstance
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

func (c *Ctx) Set(session BizSession, instance BizInstance) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.data == nil {
		c.data = make(map[string]map[string]BizInstance)
	}
	sessionID := session.BizSessionId()
	bucket := c.data[sessionID]
	if bucket == nil {
		bucket = make(map[string]BizInstance)
		c.data[sessionID] = bucket
	}
	bucket[instance.BizInstanceId()] = instance
}

func (c *Ctx) Get(sessionID string, instanceID string) (BizInstance, bool) {
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

func (c *Ctx) Del(sessionID string, instanceID string) (BizInstance, bool) {
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

func (c *Ctx) ForEach(sessionID string, fn func(instance BizInstance)) {
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

func (c *Ctx) List(sessionID string) []BizInstance {
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
