package ext_model

import "testing"

func TestExtMapGetAndDelMissing(t *testing.T) {
	var typed ExtMap[testObj]
	if _, ok := typed.Get("missing"); ok {
		t.Fatal("expected missing typed value")
	}
	if _, ok := typed.Del("missing"); ok {
		t.Fatal("expected missing typed delete")
	}

	model := NewExtModel()
	if _, ok := model.Get("missing"); ok {
		t.Fatal("expected missing ext model value")
	}
	if _, ok := model.Del("missing"); ok {
		t.Fatal("expected missing ext model delete")
	}
}

func TestCopyExtMapNilOption(t *testing.T) {
	src := NewExtModel()
	src.Set(testObj{ID: "alpha", Name: "Alpha"})

	dst := CopyExtMap(src, nil)
	value, ok := dst.Get("alpha")
	if !ok || value.(testObj).Name != "Alpha" {
		t.Fatalf("unexpected copied value: %v %v", value, ok)
	}
}
