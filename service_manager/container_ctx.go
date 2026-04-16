package service_manager

import (
	"context"
	"errors"
	"slices"
	"sync"

	"github.com/daidai21/biz_ext_framework/biz_ctx"
)

var (
	ErrInvalidBizSessionID = errors.New("invalid biz session id")
	ErrNilBizSession       = errors.New("nil biz session")
)

// CtxContainer manages biz_ctx sessions and context injection.
type CtxContainer struct {
	mu       sync.RWMutex
	sessions map[string]biz_ctx.BizSession
}

func NewCtxContainer() *CtxContainer {
	return &CtxContainer{}
}

func (c *CtxContainer) Create(sessionID string) (biz_ctx.BizSession, error) {
	if sessionID == "" {
		return nil, ErrInvalidBizSessionID
	}

	session := biz_ctx.NewBizSession(sessionID)
	if err := c.Register(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (c *CtxContainer) Register(session biz_ctx.BizSession) error {
	if session == nil {
		return ErrNilBizSession
	}
	if session.BizSessionId() == "" {
		return ErrInvalidBizSessionID
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sessions == nil {
		c.sessions = make(map[string]biz_ctx.BizSession)
	}
	c.sessions[session.BizSessionId()] = session
	return nil
}

func (c *CtxContainer) Remove(sessionID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.sessions, sessionID)
}

func (c *CtxContainer) Get(sessionID string) (biz_ctx.BizSession, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	session, ok := c.sessions[sessionID]
	return session, ok
}

func (c *CtxContainer) SessionIDs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ids := make([]string, 0, len(c.sessions))
	for sessionID := range c.sessions {
		ids = append(ids, sessionID)
	}
	slices.Sort(ids)
	return ids
}

func (c *CtxContainer) WithSession(ctx context.Context, sessionID string) (context.Context, error) {
	session, ok := c.Get(sessionID)
	if !ok {
		return ctx, ErrContainerNotFound
	}
	return biz_ctx.WithBizSession(ctx, session), nil
}

func (c *CtxContainer) SessionFromContext(ctx context.Context) (biz_ctx.BizSession, bool) {
	return biz_ctx.BizSessionFromContext(ctx)
}
