package ext_model

import "testing"

type testObj struct {
	ID   string
	Name string
}

var _ ExtObj = testObj{}
var _ ExtModel[testObj] = (*ExtMap[testObj])(nil)

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

func countMap(m *ExtMap[testObj]) int {
	count := 0
	m.ForEach(func(value testObj) {
		count++
	})
	return count
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

func TestExtMapUsesValueKeyForOverwrite(t *testing.T) {
	m := seedMap(
		testObj{ID: "alpha", Name: "Alpha"},
		testObj{ID: "alpha", Name: "Override"},
	)

	if countMap(m) != 1 {
		t.Fatalf("expected count 1 when keys collide, got %d", countMap(m))
	}

	value, ok := m.Get("alpha")
	if !ok {
		t.Fatal("expected alpha to exist")
	}
	if value.Name != "Override" {
		t.Fatalf("expected value Override, got %s", value.Name)
	}
}
