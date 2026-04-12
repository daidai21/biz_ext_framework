package biz_ctx

import (
	"context"
	"sync"
)

// BizInstance is the technical identity for one business instance under a session.
type BizInstance interface {
	BizInstanceId() string
}

// BizSession stores business instances directly (1:N) and is intended to live in context.Context.
type BizSession interface {
	BizSessionId() string
	Set(instance BizInstance)
	Get(instanceID string) (BizInstance, bool)
	Del(instanceID string) (BizInstance, bool)
	ForEach(fn func(instance BizInstance))
	List() []BizInstance
}

// Session is the default concurrency-safe BizSession implementation.
// The zero value is not recommended; use NewBizSession.
type Session struct {
	id string

	mu   sync.RWMutex
	data map[string]BizInstance
}

var _ BizSession = (*Session)(nil)

func NewBizSession(sessionID string) BizSession {
	return &Session{
		id:   sessionID,
		data: make(map[string]BizInstance),
	}
}

func (s *Session) BizSessionId() string {
	return s.id
}

func (s *Session) Set(instance BizInstance) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.data == nil {
		s.data = make(map[string]BizInstance)
	}
	s.data[instance.BizInstanceId()] = instance
}

func (s *Session) Get(instanceID string) (BizInstance, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.data == nil {
		return nil, false
	}
	instance, ok := s.data[instanceID]
	return instance, ok
}

func (s *Session) Del(instanceID string) (BizInstance, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.data == nil {
		return nil, false
	}

	instance, ok := s.data[instanceID]
	if !ok {
		return nil, false
	}
	delete(s.data, instanceID)
	return instance, true
}

func (s *Session) ForEach(fn func(instance BizInstance)) {
	s.mu.RLock()
	if len(s.data) == 0 {
		s.mu.RUnlock()
		return
	}

	snapshot := make([]BizInstance, 0, len(s.data))
	for _, instance := range s.data {
		snapshot = append(snapshot, instance)
	}
	s.mu.RUnlock()

	for _, instance := range snapshot {
		fn(instance)
	}
}

func (s *Session) List() []BizInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.data) == 0 {
		return nil
	}

	result := make([]BizInstance, 0, len(s.data))
	for _, instance := range s.data {
		result = append(result, instance)
	}
	return result
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
	if session.BizSessionId() == "" {
		return nil, false
	}
	return session, true
}
