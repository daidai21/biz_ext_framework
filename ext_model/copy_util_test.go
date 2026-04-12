package ext_model

import (
	"strings"
	"testing"
)

type copyUser struct {
	ID   string
	Name string
}

var _ ExtObj = copyUser{}

func (u copyUser) Key() string {
	return u.ID
}

func seedUserModel(values ...copyUser) ExtModel {
	model := NewExtModel()
	for _, value := range values {
		model.Set(value)
	}
	return model
}

func TestReadmeCopyUtilityCopied(t *testing.T) {
	src := seedUserModel(
		copyUser{ID: "u1", Name: "Alice"},
		copyUser{ID: "u2", Name: "Bob"},
	)

	copied := CopyExtMap(src)

	v, ok := copied.Get("u1")
	if !ok {
		t.Fatal("expected key u1")
	}
	u, ok := v.(copyUser)
	if !ok {
		t.Fatalf("expected copyUser type, got %T", v)
	}
	if u.Name != "Alice" {
		t.Fatalf("expected Alice, got %s", u.Name)
	}
}

func TestReadmeCopyUtilityWithKeyFilter(t *testing.T) {
	src := seedUserModel(
		copyUser{ID: "u1", Name: "Alice"},
		copyUser{ID: "u2", Name: "Bob"},
	)

	filtered := CopyExtMap(src, WithKeyFilter(func(key string) bool {
		return key == "u1"
	}))

	_, ok := filtered.Get("u2")
	if ok {
		t.Fatal("expected u2 to be filtered out")
	}

	v, ok := filtered.Get("u1")
	if !ok {
		t.Fatal("expected key u1")
	}
	u, ok := v.(copyUser)
	if !ok {
		t.Fatalf("expected copyUser type, got %T", v)
	}
	if u.Name != "Alice" {
		t.Fatalf("expected Alice, got %s", u.Name)
	}
}

func TestReadmeCopyUtilityWithDeepCopy(t *testing.T) {
	src := seedUserModel(copyUser{ID: "u1", Name: "Alice"})

	deepCopied := CopyExtMap(src, WithDeepCopy(func(value ExtObj) ExtObj {
		u, ok := value.(copyUser)
		if !ok {
			t.Fatalf("expected copyUser type, got %T", value)
		}
		u.Name = strings.ToUpper(u.Name)
		return u
	}))

	v, ok := deepCopied.Get("u1")
	if !ok {
		t.Fatal("expected key u1")
	}
	u, ok := v.(copyUser)
	if !ok {
		t.Fatalf("expected copyUser type, got %T", v)
	}
	if u.Name != "ALICE" {
		t.Fatalf("expected ALICE, got %s", u.Name)
	}
}
