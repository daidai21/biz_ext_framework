package ext_model

import "testing"

type copyTestObj struct {
	id     string
	name   string
	labels []string
}

type objectKeySet map[string]struct{}

func (s objectKeySet) Contains(key string) bool {
	_, ok := s[key]
	return ok
}

func (o copyTestObj) Key() string {
	return o.id
}

var _ ExtObj = copyTestObj{}
var _ ExtModel[copyTestObj] = (*ExtMap[copyTestObj])(nil)

var extObjectKeyBlackSet = objectKeySet{
	"blocked": {},
	"skip":    {},
}

var objectKeyFilter = WithKeyFilter[copyTestObj](func(objectKey string) bool {
	return !extObjectKeyBlackSet.Contains(objectKey)
})

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

func resolveCopyExtMapOptions[V ExtObj](opts ...CopyExtMapOption[V]) copyExtMapOptions[V] {
	options := copyExtMapOptions[V]{}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}
	return options
}

func TestObjectKeyFilter(t *testing.T) {
	options := resolveCopyExtMapOptions(objectKeyFilter)
	if options.keyFilter == nil {
		t.Fatal("expected keyFilter to be configured")
	}

	testCases := []struct {
		key  string
		want bool
	}{
		{key: "alpha", want: true},
		{key: "blocked", want: false},
		{key: "skip", want: false},
	}

	for _, tc := range testCases {
		if got := options.keyFilter(tc.key); got != tc.want {
			t.Fatalf("expected key %s => %t, got %t", tc.key, tc.want, got)
		}
	}
}

func TestWithDeepCopyOption(t *testing.T) {
	options := resolveCopyExtMapOptions(WithDeepCopy[copyTestObj](func(value copyTestObj) copyTestObj {
		value.labels = append([]string(nil), value.labels...)
		return value
	}))
	if options.deepCopy == nil {
		t.Fatal("expected deepCopy to be configured")
	}

	original := copyTestObj{id: "alpha", labels: []string{"a", "b"}}
	cloned := options.deepCopy(original)
	original.labels[0] = "changed"

	if cloned.labels[0] != "a" {
		t.Fatalf("expected cloned labels to stay independent, got %s", cloned.labels[0])
	}
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
