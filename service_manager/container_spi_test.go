package service_manager

import "testing"

func TestSPIContainer(t *testing.T) {
	container := NewSPIContainer[string]()

	if err := container.Register("audit", "impl-a"); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if err := container.Register("audit", "impl-b"); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	impls := container.Implementations("audit")
	if len(impls) != 2 || impls[0] != "impl-a" || impls[1] != "impl-b" {
		t.Fatalf("unexpected impls: %v", impls)
	}
}

func TestSPIContainerInvalidDefinition(t *testing.T) {
	container := NewSPIContainer[string]()
	if err := container.Register("", "impl-a"); err == nil {
		t.Fatalf("expected invalid definition error")
	}
}
