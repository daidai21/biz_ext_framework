package ext_model

import "testing"

type testObj struct {
	ID   string
	Name string
}

var _ ExtObj = testObj{}

func (o testObj) Key() string {
	return o.ID
}

func seedMap(values ...testObj) *ExtMap[testObj] {
	var m ExtMap[testObj]
	for _, value := range values {
		m.Set(value)
	}
	return &m
}

func TestExtMapZeroValueIsUsable(t *testing.T) {
	var m ExtMap[testObj]

	m.Set(testObj{ID: "alpha", Name: "Alpha"})

	value, ok := m.Get("alpha")
	if !ok {
		t.Fatal("expected alpha to exist")
	}
	if value.Name != "Alpha" {
		t.Fatalf("expected value Alpha, got %s", value.Name)
	}
}

func TestExtMapDelRemovesValue(t *testing.T) {
	m := seedMap(
		testObj{ID: "alpha", Name: "Alpha"},
		testObj{ID: "beta", Name: "Beta"},
	)

	value, ok := m.Del("alpha")
	if !ok {
		t.Fatal("expected delete to remove alpha")
	}
	if value.Name != "Alpha" {
		t.Fatalf("expected deleted value Alpha, got %s", value.Name)
	}

	_, ok = m.Get("alpha")
	if ok {
		t.Fatal("expected alpha to be removed")
	}
}

func TestExtMapForEachVisitsCurrentValues(t *testing.T) {
	m := seedMap(
		testObj{ID: "alpha", Name: "Alpha"},
		testObj{ID: "beta", Name: "Beta"},
	)

	visited := map[string]bool{}
	m.ForEach(func(value testObj) {
		visited[value.ID] = true
	})

	if len(visited) != 2 {
		t.Fatalf("expected to visit 2 values, got %d", len(visited))
	}
	if !visited["alpha"] || !visited["beta"] {
		t.Fatalf("expected alpha and beta to be visited, got %#v", visited)
	}
}

type userInfo struct {
	UserId int64
	ExtModel
}

var (
	_ ExtObj = userTaxInfo{}
	_ ExtObj = userPhdInfo{}
)

type userTaxInfo struct {
	TaxId string
}

func (u userTaxInfo) Key() string {
	return "userTaxInfo"
}

type userPhdInfo struct {
	PhdId string
}

func (u userPhdInfo) Key() string {
	return "userPhdInfo"
}

func TestReadmeAttachMultipleExtensionStructs(t *testing.T) {
	info := userInfo{
		UserId:   1,
		ExtModel: NewExtModel(),
	}
	info.Set(userTaxInfo{TaxId: "tax_2313"})
	info.Set(userPhdInfo{PhdId: "phd_6748392"})

	taxValue, ok := info.Get("userTaxInfo")
	if !ok {
		t.Fatal("expected userTaxInfo to exist")
	}
	tax, ok := taxValue.(userTaxInfo)
	if !ok {
		t.Fatalf("expected userTaxInfo type, got %T", taxValue)
	}
	if tax.TaxId != "tax_2313" {
		t.Fatalf("expected TaxId tax_2313, got %s", tax.TaxId)
	}

	phdValue, ok := info.Get("userPhdInfo")
	if !ok {
		t.Fatal("expected userPhdInfo to exist")
	}
	phd, ok := phdValue.(userPhdInfo)
	if !ok {
		t.Fatalf("expected userPhdInfo type, got %T", phdValue)
	}
	if phd.PhdId != "phd_6748392" {
		t.Fatalf("expected PhdId phd_6748392, got %s", phd.PhdId)
	}

	count := 0
	info.ForEach(func(value ExtObj) {
		count++
	})
	if count != 2 {
		t.Fatalf("expected 2 ext objects, got %d", count)
	}
}
