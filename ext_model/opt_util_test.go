package ext_model

import "testing"

type optUserTaxInfo struct {
	TaxId string
}

func (u optUserTaxInfo) Key() string {
	return "userTaxInfo"
}

type optUserPhdInfo struct {
	PhdId string
}

func (u optUserPhdInfo) Key() string {
	return "userPhdInfo"
}

func TestGetAsSuccess(t *testing.T) {
	model := NewExtModel()
	model.Set(optUserTaxInfo{TaxId: "tax_2313"})

	tax, ok := GetAs[optUserTaxInfo](model, "userTaxInfo")
	if !ok {
		t.Fatal("expected generic get success")
	}
	if tax.TaxId != "tax_2313" {
		t.Fatalf("expected tax_2313, got %s", tax.TaxId)
	}
}

func TestGetAsKeyNotFound(t *testing.T) {
	model := NewExtModel()

	_, ok := GetAs[optUserTaxInfo](model, "missing")
	if ok {
		t.Fatal("expected generic get fail when key missing")
	}
}

func TestGetAsTypeMismatch(t *testing.T) {
	model := NewExtModel()
	model.Set(optUserTaxInfo{TaxId: "tax_2313"})

	_, ok := GetAs[optUserPhdInfo](model, "userTaxInfo")
	if ok {
		t.Fatal("expected generic get fail when type mismatch")
	}
}

func TestGetAsNilModel(t *testing.T) {
	_, ok := GetAs[optUserTaxInfo](nil, "userTaxInfo")
	if ok {
		t.Fatal("expected generic get fail when model is nil")
	}
}
