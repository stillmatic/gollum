
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