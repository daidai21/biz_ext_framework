package ext_model

import (
	"strconv"
	"testing"
)

var (
	benchSinkObj  ExtObj
	benchSinkBool bool
	benchSinkInt  int
	benchSinkUser copyUser
)

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

func BenchmarkExtModelSet(b *testing.B) {
	model := NewExtModel()
	values := make([]testObj, 1024)
	for i := range values {
		values[i] = testObj{ID: "k" + strconv.Itoa(i), Name: "n"}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.Set(values[i%len(values)])
	}
}

func BenchmarkExtModelGetHit(b *testing.B) {
	model := NewExtModel()
	keys := make([]string, 1024)
	for i := range keys {
		key := "k" + strconv.Itoa(i)
		keys[i] = key
		model.Set(testObj{ID: key, Name: "n"})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v, ok := model.Get(keys[i%len(keys)])
		benchSinkObj, benchSinkBool = v, ok
	}
}

func BenchmarkExtModelGetMiss(b *testing.B) {
	model := NewExtModel()
	missKeys := make([]string, 1024)
	for i := range missKeys {
		missKeys[i] = "missing_" + strconv.Itoa(i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v, ok := model.Get(missKeys[i%len(missKeys)])
		benchSinkObj, benchSinkBool = v, ok
	}
}

func BenchmarkExtModelDelHitRestore(b *testing.B) {
	model := NewExtModel()
	values := make([]testObj, 1024)
	for i := range values {
		values[i] = testObj{ID: "k" + strconv.Itoa(i), Name: "n"}
		model.Set(values[i])
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v := values[i%len(values)]
		obj, ok := model.Del(v.ID)
		model.Set(v)
		benchSinkObj, benchSinkBool = obj, ok
	}
}

func BenchmarkExtModelForEach(b *testing.B) {
	model := NewExtModel()
	for i := 0; i < 1024; i++ {
		model.Set(testObj{ID: "k" + strconv.Itoa(i), Name: "n"})
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		model.ForEach(func(ExtObj) {
			count++
		})
		benchSinkInt = count
	}
}

func BenchmarkCopyExtMapNoOptions(b *testing.B) {
	src := seedUserModel(makeCopyUsers(1024)...)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copied := CopyExtMap(src)
		v, _ := copied.Get("u0")
		benchSinkObj = v
	}
}

func BenchmarkCopyExtMapWithKeyFilter(b *testing.B) {
	src := seedUserModel(makeCopyUsers(1024)...)
	filter := WithKeyFilter(func(key string) bool {
		// 过滤一半 key，避免字符串解析开销
		return len(key) > 1 && (key[len(key)-1]-'0')%2 == 0
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copied := CopyExtMap(src, filter)
		v, _ := copied.Get("u0")
		benchSinkObj = v
	}
}

func BenchmarkCopyExtMapWithDeepCopy(b *testing.B) {
	src := seedUserModel(makeCopyUsers(1024)...)
	deepCopy := WithDeepCopy(func(value ExtObj) ExtObj {
		u, ok := value.(copyUser)
		if !ok {
			return value
		}
		u.Name = "X"
		return u
	})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copied := CopyExtMap(src, deepCopy)
		v, _ := copied.Get("u0")
		benchSinkObj = v
	}
}

func BenchmarkGetAsHit(b *testing.B) {
	model := NewExtModel()
	model.Set(copyUser{ID: "u0", Name: "Alice"})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u, ok := GetAs[copyUser](model, "u0")
		benchSinkUser, benchSinkBool = u, ok
	}
}

func BenchmarkGetAsTypeMismatch(b *testing.B) {
	model := NewExtModel()
	model.Set(testObj{ID: "u0", Name: "not-copyUser"})

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u, ok := GetAs[copyUser](model, "u0")
		benchSinkUser, benchSinkBool = u, ok
	}
}

func BenchmarkGetAsMiss(b *testing.B) {
	model := NewExtModel()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u, ok := GetAs[copyUser](model, "missing")
		benchSinkUser, benchSinkBool = u, ok
	}
}

func makeCopyUsers(n int) []copyUser {
	users := make([]copyUser, n)
	for i := 0; i < n; i++ {
		users[i] = copyUser{ID: "u" + strconv.Itoa(i), Name: "n"}
	}
	return users
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
