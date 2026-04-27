# Benchmark

Generated at: 2026-04-27 12:23:32 +0800

Command: `go test -run '^$' -bench . -benchmem -count=1 ./...`

```
goos: linux
goarch: amd64
pkg: github.com/daidai21/biz_ext_framework/ext_model
cpu: Intel(R) Xeon(R) Platinum 8260 CPU @ 2.40GHz
BenchmarkExtModelSet-64                	 8391818	       137.4 ns/op	      32 B/op	       1 allocs/op
BenchmarkExtModelGetHit-64             	24112221	        43.84 ns/op	       0 B/op	       0 allocs/op
BenchmarkExtModelGetMiss-64            	47777900	        23.23 ns/op	       0 B/op	       0 allocs/op
BenchmarkExtModelDelHitRestore-64      	 4774592	       246.2 ns/op	      32 B/op	       1 allocs/op
BenchmarkExtModelForEach-64            	   79663	     18331 ns/op	       0 B/op	       0 allocs/op
BenchmarkCopyExtMapNoOptions-64        	    4510	    237576 ns/op	  160402 B/op	      25 allocs/op
BenchmarkCopyExtMapWithKeyFilter-64    	    8475	    141172 ns/op	   78400 B/op	      20 allocs/op
BenchmarkCopyExtMapWithDeepCopy-64     	    3897	    317064 ns/op	  193170 B/op	    1049 allocs/op
BenchmarkGetAsHit-64                   	26696937	        41.58 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetAsTypeMismatch-64          	30130128	        42.80 ns/op	       0 B/op	       0 allocs/op
BenchmarkGetAsMiss-64                  	40864126	        28.53 ns/op	       0 B/op	       0 allocs/op
PASS
ok  	github.com/daidai21/biz_ext_framework/ext_model	13.896s
```
