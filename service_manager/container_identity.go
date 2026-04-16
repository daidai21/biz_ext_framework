package service_manager

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/daidai21/biz_ext_framework/biz_identity"
)

var (
	ErrInvalidIdentityScope = fmt.Errorf("invalid identity scope")
)

// IdentityContainer manages whitelisted business identity scopes.
// A scope matches itself and all descendant identities under the same prefix.
type IdentityContainer struct {
	mu     sync.RWMutex
	scopes map[string]struct{}
}

func NewIdentityContainer(scopes ...string) (*IdentityContainer, error) {
	container := &IdentityContainer{}
	for _, scope := range scopes {
		if err := container.AllowScope(scope); err != nil {
			return nil, err
		}
	}
	return container, nil
}

func (c *IdentityContainer) AllowScope(scope string) error {
	if err := biz_identity.ValidateIdentityID(scope); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidIdentityScope, err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.scopes == nil {
		c.scopes = make(map[string]struct{})
	}
	c.scopes[scope] = struct{}{}
	return nil
}

func (c *IdentityContainer) RevokeScope(scope string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.scopes, scope)
}

func (c *IdentityContainer) IsAllowed(identityID string) bool {
	if err := biz_identity.ValidateIdentityID(identityID); err != nil {
		return false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	for scope := range c.scopes {
		if scope == identityID || strings.HasPrefix(identityID, scope+".") {
			return true
		}
	}
	return false
}

func (c *IdentityContainer) IsIdentityAllowed(identity biz_identity.BizIdentity) bool {
	if identity == nil {
		return false
	}
	return c.IsAllowed(identity.IdentityId())
}

func (c *IdentityContainer) Scopes() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	scopes := make([]string, 0, len(c.scopes))
	for scope := range c.scopes {
		scopes = append(scopes, scope)
	}
	slices.Sort(scopes)
	return scopes
}
