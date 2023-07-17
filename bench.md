
```
goos: linux
goarch: amd64
pkg: github.com/stillmatic/gollum
cpu: AMD Ryzen 9 7950X 16-Core Processor            
BenchmarkMemoryVectorStore/BenchmarkInsert-n=10-32         	   14752	     83971 ns/op	   64912 B/op	      10 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=10-k=1-32      	  810657	      1256 ns/op	     288 B/op	       3 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=10-k=10-32     	  663574	      1639 ns/op	    2880 B/op	       6 allocs/op
BenchmarkMemoryVectorStore/BenchmarkInsert-n=100-32        	    1263	   1042804 ns/op	  646807 B/op	     190 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=100-k=1-32     	  113851	     10399 ns/op	     288 B/op	       3 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=100-k=10-32    	   93428	     12625 ns/op	    2880 B/op	       6 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=100-k=100-32   	   69256	     17569 ns/op	   25664 B/op	       9 allocs/op
BenchmarkMemoryVectorStore/BenchmarkInsert-n=1000-32       	     147	   8100071 ns/op	 6505065 B/op	    2734 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=1000-k=1-32    	   10000	    104921 ns/op	     288 B/op	       3 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=1000-k=10-32   	    9831	    123464 ns/op	    2880 B/op	       6 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=1000-k=100-32  	    7848	    152007 ns/op	   25664 B/op	       9 allocs/op
BenchmarkMemoryVectorStore/BenchmarkInsert-n=10000-32      	      13	  82907557 ns/op	64727761 B/op	   29740 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=10000-k=1-32   	     783	   1514925 ns/op	     288 B/op	       3 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=10000-k=10-32  	     679	   1692251 ns/op	    2880 B/op	       6 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=10000-k=100-32 	     650	   2095042 ns/op	   25664 B/op	       9 allocs/op
BenchmarkMemoryVectorStore/BenchmarkInsert-n=100000-32     	       2	 814309670 ns/op	648192728 B/op	  299774 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=100000-k=1-32  	      82	  16656264 ns/op	     288 B/op	       3 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=100000-k=10-32 	      68	  16402470 ns/op	    2880 B/op	       6 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=100000-k=100-32         	      64	  18205266 ns/op	   25664 B/op	       9 allocs/op
BenchmarkMemoryVectorStore/BenchmarkInsert-n=1000000-32             	       1	9578485965 ns/op	6552089784 B/op	 2999874 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=1000000-k=1-32          	       7	 161260588 ns/op	     288 B/op	       3 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=1000000-k=10-32         	       7	 212760511 ns/op	    2880 B/op	       6 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=1000000-k=100-32        	       4	 290261365 ns/op	   25664 B/op	       9 allocs/op
PASS
ok  	github.com/stillmatic/gollum	111.224s
```

post perf improvement - mac. stabilizes the allocations

```
goos: darwin
goarch: arm64
pkg: github.com/stillmatic/gollum
BenchmarkMemoryVectorStore/BenchmarkInsert-n=10-10         	    5341	    220817 ns/op	   65223 B/op	      10 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=10-k=1-10      	   60616	     19622 ns/op	     120 B/op	       4 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=10-k=10-10     	   60388	     20033 ns/op	     304 B/op	       3 allocs/op
BenchmarkMemoryVectorStore/BenchmarkInsert-n=100-10        	     536	   2202933 ns/op	  652278 B/op	     190 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=100-k=1-10     	    6152	    194476 ns/op	     120 B/op	       4 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=100-k=10-10    	    6094	    198124 ns/op	     624 B/op	       4 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=100-k=100-10   	    5946	    199925 ns/op	    2752 B/op	       3 allocs/op
BenchmarkMemoryVectorStore/BenchmarkInsert-n=1000-10       	      55	  22152592 ns/op	 6523947 B/op	    2735 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=1000-k=1-10    	     613	   1953824 ns/op	     120 B/op	       4 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=1000-k=10-10   	     610	   1987216 ns/op	     624 B/op	       4 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=1000-k=100-10  	     580	   2051436 ns/op	    5952 B/op	       4 allocs/op
BenchmarkMemoryVectorStore/BenchmarkInsert-n=10000-10      	       5	 222244750 ns/op	64782620 B/op	   29747 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=10000-k=1-10   	      61	  19383620 ns/op	     120 B/op	       4 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=10000-k=10-10  	      60	  19823898 ns/op	     624 B/op	       4 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=10000-k=100-10 	      57	  20027584 ns/op	    5952 B/op	       4 allocs/op
BenchmarkMemoryVectorStore/BenchmarkInsert-n=100000-10     	       1	2207505500 ns/op	648271208 B/op	  299808 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=100000-k=1-10  	       6	 196473680 ns/op	     120 B/op	       4 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=100000-k=10-10 	       6	 197389812 ns/op	     624 B/op	       4 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=100000-k=100-10         	       5	 200068883 ns/op	    5952 B/op	       4 allocs/op
BenchmarkMemoryVectorStore/BenchmarkInsert-n=1000000-10             	       1	22239769458 ns/op	6552038696 B/op	 2999849 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=1000000-k=1-10          	       1	1966544833 ns/op	     120 B/op	       4 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=1000000-k=10-10         	       1	1963972417 ns/op	     624 B/op	       4 allocs/op
BenchmarkMemoryVectorStore/BenchmarkQuery-n=1000000-k=100-10        	       1	1988149583 ns/op	    5952 B/op	       4 allocs/op
PASS
ok  	github.com/stillmatic/gollum	142.897s
```

mac is expected to be slower. however, post change, what we see is that our memury usage is much more stable - consistently 4 allocs per operation and much less memory usage too. the memory characteristics are proportional to `k`. the desktop chip is faster and has SIMD enhanced distance calculation, so not unexpected. the runtime is also consistently linear with `n` where `n` is the number of values in the db. 

