package ext_model

import "testing"

type copyTestObj struct {
	id     string
	name   string
	labels []string
}

func (o copyTestObj) Key() string {
	return o.id
}

var _ ExtObj = copyTestObj{}
var _ ExtModel[copyTestObj] = (*ExtMap[copyTestObj])(nil)

func seedCopyMap(values ...copyTestObj) *ExtMap[copyTestObj] {
	var m ExtMap[copyTestObj]
	for _, value := range values {
		m.Set(value)
	}
	return &m
}

func countCopyMap(m *ExtMap[copyTestObj]) int {
	count := 0
	m.ForEach(func(value copyTestObj) {
		count++
	})
	return count
}

func TestCopyExtMapCopiesAllValuesByDefault(t *testing.T) {
	src := seedCopyMap(
		copyTestObj{id: "alpha", name: "Alpha"},
		copyTestObj{id: "beta", name: "Beta"},
	)

	dst := CopyExtMap(src)

	if countCopyMap(dst) != 2 {
		t.Fatalf("expected copied map count 2, got %d", countCopyMap(dst))
	}

	value, ok := dst.Get("alpha")
	if !ok {
		t.Fatal("expected alpha to be copied")
	}
	if value.name != "Alpha" {
		t.Fatalf("expected Alpha, got %s", value.name)
	}
}

func TestCopyExtMapAppliesKeyFilter(t *testing.T) {
	src := seedCopyMap(
		copyTestObj{id: "alpha", name: "Alpha"},
		copyTestObj{id: "beta", name: "Beta"},
	)

	dst := CopyExtMap(src, WithKeyFilter[copyTestObj](func(key string) bool {
		return key == "beta"
	}))

	if countCopyMap(dst) != 1 {
		t.Fatalf("expected copied map count 1, got %d", countCopyMap(dst))
	}

	if _, ok := dst.Get("alpha"); ok {
		t.Fatal("expected alpha to be filtered out")
	}
	if _, ok := dst.Get("beta"); !ok {
		t.Fatal("expected beta to be copied")
	}
}

func TestCopyExtMapAppliesDeepCopy(t *testing.T) {
	src := seedCopyMap(
		copyTestObj{id: "alpha", name: "Alpha", labels: []string{"a", "b"}},
	)

	dst := CopyExtMap(src, WithDeepCopy[copyTestObj](func(value copyTestObj) copyTestObj {
		cloned := append([]string(nil), value.labels...)
		value.labels = cloned
		return value
	}))

	srcValue, _ := src.Get("alpha")
	srcValue.labels[0] = "changed"
	src.Set(srcValue)

	dstValue, ok := dst.Get("alpha")
	if !ok {
		t.Fatal("expected alpha to be copied")
	}
	if dstValue.labels[0] != "a" {
		t.Fatalf("expected deep-copied labels to stay independent, got %s", dstValue.labels[0])
	}
}

func TestCopyExtMapHandlesNilSource(t *testing.T) {
	dst := CopyExtMap[copyTestObj](nil)

	if countCopyMap(dst) != 0 {
		t.Fatalf("expected nil source to produce empty map, got %d", countCopyMap(dst))
	}
}
