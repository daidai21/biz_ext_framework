package ext_process

import "testing"

func TestNormalizeDefinitionActionDefault(t *testing.T) {
	merged, err := MergeImplementations([]string{"a"}, []string{"b"}, "")
	if err != nil {
		t.Fatalf("expected append default, got %v", err)
	}
	if len(merged) != 2 || merged[0] != "a" || merged[1] != "b" {
		t.Fatalf("unexpected merged result: %#v", merged)
	}
}
