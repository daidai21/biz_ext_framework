package service_manager

import "testing"

type testIdentity struct {
	id string
}

func (i testIdentity) IdentityId() string {
	return i.id
}

func (i testIdentity) Priority() int {
	return 0
}

func TestIdentityContainer(t *testing.T) {
	container, err := NewIdentityContainer("SELLER.SHOP")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if !container.IsAllowed("SELLER.SHOP") {
		t.Fatalf("expected exact scope allowed")
	}
	if !container.IsAllowed("SELLER.SHOP.OPERATOR") {
		t.Fatalf("expected descendant scope allowed")
	}
	if container.IsAllowed("BUYER.SHOP") {
		t.Fatalf("expected unrelated scope denied")
	}
	if !container.IsIdentityAllowed(testIdentity{id: "SELLER.SHOP.CASHIER"}) {
		t.Fatalf("expected identity allowed")
	}

	container.RevokeScope("SELLER.SHOP")
	if container.IsAllowed("SELLER.SHOP.OPERATOR") {
		t.Fatalf("expected scope revoked")
	}
}

func TestIdentityContainerInvalidScope(t *testing.T) {
	container := &IdentityContainer{}
	if err := container.AllowScope("seller.shop"); err == nil {
		t.Fatalf("expected invalid scope error")
	}
}
