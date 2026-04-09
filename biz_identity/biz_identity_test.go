package biz_identity

import (
	"fmt"
	"testing"
)

type testIdentity struct {
	id       string
	priority int
}

func (i testIdentity) IdentityId() string {
	return i.id
}

func (i testIdentity) Priority() int {
	return i.priority
}

var _ BizIdentity = testIdentity{}
var _ Parser[testIdentity] = ParseFunc[testIdentity](nil)
var _ Validator[testIdentity] = ValidateFunc[testIdentity](nil)
var _ Validator[testIdentity] = DefaultValidator[testIdentity]{}

func TestBizIdentityInterface(t *testing.T) {
	identity := testIdentity{
		id:       "SELLER.SHOP.OPERATOR",
		priority: 10,
	}

	if identity.IdentityId() != "SELLER.SHOP.OPERATOR" {
		t.Fatalf("unexpected identity id: %s", identity.IdentityId())
	}
	if identity.Priority() != 10 {
		t.Fatalf("unexpected priority: %d", identity.Priority())
	}
}

func TestParserInterface(t *testing.T) {
	parser := ParseFunc[testIdentity](func(info map[string]string) (testIdentity, error) {
		id := info["identity_id"]
		if id == "" {
			return testIdentity{}, fmt.Errorf("identity_id is required")
		}
		return testIdentity{
			id:       id,
			priority: 5,
		}, nil
	})

	identity, err := parser.Parser(map[string]string{
		"identity_id": "SELLER.SHOP.OPERATOR",
	})
	if err != nil {
		t.Fatalf("expected parser to build identity, got %v", err)
	}
	if identity.IdentityId() != "SELLER.SHOP.OPERATOR" {
		t.Fatalf("unexpected identity id: %s", identity.IdentityId())
	}
	if identity.Priority() != 5 {
		t.Fatalf("unexpected priority: %d", identity.Priority())
	}
}

func TestValidatorInterface(t *testing.T) {
	validator := ValidateFunc[testIdentity](func(identity testIdentity) error {
		if identity.IdentityId() == "" {
			return fmt.Errorf("identity id is required")
		}
		if identity.Priority() < 0 {
			return fmt.Errorf("priority must be non-negative")
		}
		return nil
	})

	if err := validator.Validate(testIdentity{id: "SELLER.SHOP.OPERATOR", priority: 1}); err != nil {
		t.Fatalf("expected validator to accept identity, got %v", err)
	}

	if err := validator.Validate(testIdentity{id: "", priority: 1}); err == nil {
		t.Fatal("expected validator to reject empty identity id")
	}
}

func TestDefaultValidator(t *testing.T) {
	validator := DefaultValidator[testIdentity]{}

	validCases := []string{
		"SELLER",
		"SELLER.SHOP",
		"SELLER.SHOP.OPERATOR",
		"A.B.C.D.E.F.G.H.I.J",
	}

	for _, identityID := range validCases {
		if err := validator.Validate(testIdentity{id: identityID}); err != nil {
			t.Fatalf("expected %q to be valid, got %v", identityID, err)
		}
	}

	invalidCases := []string{
		"",
		"seller.shop.operator",
		"SELLER_SHOP",
		".SELLER",
		"SELLER.",
		"SELLER..SHOP",
		"SELLER.SHOP.123",
		"A.B.C.D.E.F.G.H.I.J.K",
	}

	for _, identityID := range invalidCases {
		if err := validator.Validate(testIdentity{id: identityID}); err == nil {
			t.Fatalf("expected %q to be invalid", identityID)
		}
	}
}
