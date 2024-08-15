with `go test -bench=. -cpu 1 -benchmem -benchtime=1s`

with n=32

```
goos: linux
goarch: amd64
pkg: github.com/stillmatic/gollum/internal/hash
cpu: AMD Ryzen 9 7950X 16-Core Processor
BenchmarkFNV32             48111             25479 ns/op               8 B/op          1 allocs/op
BenchmarkFNV64             47475             25640 ns/op               8 B/op          1 allocs/op
BenchmarkFNV128            38230             31021 ns/op              16 B/op          1 allocs/op
BenchmarkMD5               44103             27324 ns/op              16 B/op          1 allocs/op
BenchmarkSHA1              69037             17428 ns/op              24 B/op          1 allocs/op
BenchmarkSHA224            98686             12203 ns/op              32 B/op          1 allocs/op
BenchmarkSHA256            97653             12591 ns/op              32 B/op          1 allocs/op
BenchmarkSHA512            39342             29469 ns/op              64 B/op          1 allocs/op
BenchmarkMurmur3          170241              6990 ns/op               8 B/op          1 allocs/op
BenchmarkXxhash           756992              1552 ns/op               8 B/op          1 allocs/op
PASS
ok      github.com/stillmatic/gollum/internal/hash      13.927s
```

with n=512 

```
goos: linux
goarch: amd64
pkg: github.com/stillmatic/gollum/internal/hash
cpu: AMD Ryzen 9 7950X 16-Core Processor
BenchmarkFNV32              3097            407335 ns/op               8 B/op          1 allocs/op
BenchmarkFNV64              3090            389704 ns/op               8 B/op          1 allocs/op
BenchmarkFNV128             2314            499748 ns/op              16 B/op          1 allocs/op
BenchmarkMD5                2682            443034 ns/op              16 B/op          1 allocs/op
BenchmarkSHA1               4224            280997 ns/op              24 B/op          1 allocs/op
BenchmarkSHA224             5968            198692 ns/op              32 B/op          1 allocs/op
BenchmarkSHA256             5952            201876 ns/op              32 B/op          1 allocs/op
BenchmarkSHA512             2544            489953 ns/op              64 B/op          1 allocs/op
BenchmarkMurmur3            9975            119104 ns/op               8 B/op          1 allocs/op
BenchmarkXxhash            46592             24758 ns/op               8 B/op          1 allocs/op
PASS
ok      github.com/stillmatic/gollum/internal/hash      12.580s
```

Conclusion: `xxhash` is an order of magnitude improvement on large strings. `murmur3` is good for medium size but doesn't seem to scale as well.