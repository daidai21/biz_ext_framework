package service_manager

import (
	"errors"
	"testing"

	"github.com/daidai21/biz_ext_framework/ext_model"
)

type testExtObj struct {
	key string
}

func (o testExtObj) Key() string {
	return o.key
}

func TestModelContainerFilterForRPC(t *testing.T) {
	container := NewModelContainer()
	if err := container.SetWhitelist("psm.order#CreateOrder", []string{"user", "risk"}); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	src := ext_model.NewExtModel()
	src.Set(testExtObj{key: "user"})
	src.Set(testExtObj{key: "risk"})
	src.Set(testExtObj{key: "secret"})

	filtered, err := container.FilterForRPC("psm.order#CreateOrder", src)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if _, ok := filtered.Get("user"); !ok {
		t.Fatalf("expected user kept")
	}
	if _, ok := filtered.Get("risk"); !ok {
		t.Fatalf("expected risk kept")
	}
	if _, ok := filtered.Get("secret"); ok {
		t.Fatalf("expected secret removed")
	}
	if _, ok := src.Get("secret"); !ok {
		t.Fatalf("expected source remain unchanged")
	}
}

func TestModelContainerFilterForRPCWithoutPolicy(t *testing.T) {
	container := NewModelContainer()

	src := ext_model.NewExtModel()
	src.Set(testExtObj{key: "user"})

	filtered, err := container.FilterForRPC("psm.order#CreateOrder", src)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if _, ok := filtered.Get("user"); ok {
		t.Fatalf("expected default deny without policy")
	}
}

func TestModelContainerInvalidRPCMethod(t *testing.T) {
	container := NewModelContainer()
	err := container.SetWhitelist("psm.order.CreateOrder", []string{"user"})
	if !errors.Is(err, ErrInvalidRPCMethod) {
		t.Fatalf("expected ErrInvalidRPCMethod, got %v", err)
	}
}
