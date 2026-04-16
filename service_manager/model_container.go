package service_manager

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/daidai21/biz_ext_framework/ext_model"
)

var (
	ErrInvalidRPCMethod = errors.New("invalid rpc method")
)

// ModelContainer manages outbound RPC model whitelist policies.
// The RPC method key format is "PSM#Method".
type ModelContainer struct {
	mu         sync.RWMutex
	whitelists map[string]map[string]struct{}
}

func NewModelContainer() *ModelContainer {
	return &ModelContainer{}
}

func (c *ModelContainer) SetWhitelist(rpcMethod string, allowedKeys []string) error {
	if err := validateRPCMethod(rpcMethod); err != nil {
		return err
	}

	whitelist := make(map[string]struct{}, len(allowedKeys))
	for _, key := range allowedKeys {
		if key == "" {
			continue
		}
		whitelist[key] = struct{}{}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.whitelists == nil {
		c.whitelists = make(map[string]map[string]struct{})
	}
	c.whitelists[rpcMethod] = whitelist
	return nil
}

func (c *ModelContainer) RemoveWhitelist(rpcMethod string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.whitelists, rpcMethod)
}

func (c *ModelContainer) Whitelist(rpcMethod string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	whitelist := c.whitelists[rpcMethod]
	keys := make([]string, 0, len(whitelist))
	for key := range whitelist {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}

// FilterForRPC copies src and only keeps ext_model keys that are explicitly allowed.
// If rpcMethod has no configured whitelist, an empty model is returned by default.
func (c *ModelContainer) FilterForRPC(rpcMethod string, src ext_model.ExtModel) (ext_model.ExtModel, error) {
	if err := validateRPCMethod(rpcMethod); err != nil {
		return nil, err
	}

	dst := ext_model.NewExtModel()
	if src == nil {
		return dst, nil
	}

	c.mu.RLock()
	whitelist := c.whitelists[rpcMethod]
	c.mu.RUnlock()

	src.ForEach(func(value ext_model.ExtObj) {
		if _, ok := whitelist[value.Key()]; ok {
			dst.Set(value)
		}
	})
	return dst, nil
}

func validateRPCMethod(rpcMethod string) error {
	psm, method, ok := strings.Cut(rpcMethod, "#")
	if !ok || psm == "" || method == "" || strings.Contains(method, "#") {
		return fmt.Errorf("%w: %q", ErrInvalidRPCMethod, rpcMethod)
	}
	return nil
}
