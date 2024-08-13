# Document Stores

Gollum provides several implementations of document stores. Document stores solve the problem of, we have lots of documents and want to find the most relevant documents to a particular query.

Names are TBD.

# docstore

Docstore is a simple document store which provides no indexing. It simply has insert and retrieve, you must know the ID. 

Think of it as essentially a key-value store. In the future, we will probably extend this to have a KV interface and provide an implementation backed by Redis or DragonflyDB.

# memory vector store

This is a simple document store that takes an embedding model and embeds documents on insert. At retrieval time, it embeds the search query and does a simple KNN lookup.

# xyz vector store

I haven't gotten around to actually writing any of these implementations but it should be simple to imagine clients for Weaviate or Pinecone following the interface. I don't actually use them though :) 



# compressed store

This is inspired by [link] - basically we can use gzip to compress the documents, then at query time, compute `enc(term) + enc(doc)` - in this case, `enc(doc)` is computed on insert. The idea is that the more similar your term and document are, the shorter the encoded representation is - because there is less entropy that needs to be included in the compressed representation. 


On an M1 Max

```
BenchmarkCompressedVectorStore/ZstdVectorStore-Query-10-1-10         	   25092	     49245 ns/op	     288 B/op	       3 allocs/op
BenchmarkCompressedVectorStore/ZstdVectorStore-Query-10-10-10        	   13317	    107361 ns/op	    2880 B/op	       6 allocs/op
BenchmarkCompressedVectorStore/ZstdVectorStore-Query-100-1-10        	    2582	    657344 ns/op	     300 B/op	       3 allocs/op
BenchmarkCompressedVectorStore/ZstdVectorStore-Query-100-10-10       	    1051	   1286167 ns/op	    2880 B/op	       6 allocs/op
BenchmarkCompressedVectorStore/ZstdVectorStore-Query-100-100-10      	     688	   1879785 ns/op	   25887 B/op	       9 allocs/op
BenchmarkCompressedVectorStore/ZstdVectorStore-Query-1000-1-10       	     202	   7782683 ns/op	     569 B/op	       3 allocs/op
BenchmarkCompressedVectorStore/ZstdVectorStore-Query-1000-10-10      	      92	  11442531 ns/op	    3461 B/op	       6 allocs/op
BenchmarkCompressedVectorStore/ZstdVectorStore-Query-1000-100-10     	      73	  15614635 ns/op	   25765 B/op	       9 allocs/op
BenchmarkCompressedVectorStore/ZstdVectorStore-Query-10000-1-10      	      32	  52570010 ns/op	     288 B/op	       3 allocs/op
BenchmarkCompressedVectorStore/ZstdVectorStore-Query-10000-10-10     	      15	  95620514 ns/op	    3161 B/op	       6 allocs/op
BenchmarkCompressedVectorStore/ZstdVectorStore-Query-10000-100-10    	       9	 138641384 ns/op	   28024 B/op	      10 allocs/op
BenchmarkCompressedVectorStore/ZstdVectorStore-Query-100000-1-10     	       2	 523147917 ns/op	    6624 B/op	       5 allocs/op
BenchmarkCompressedVectorStore/ZstdVectorStore-Query-100000-10-10    	       2	 985437625 ns/op	    9280 B/op	      10 allocs/op
BenchmarkCompressedVectorStore/ZstdVectorStore-Query-100000-100-10   	       1	1123811333 ns/op	   33600 B/op	      14 allocs/op
BenchmarkCompressedVectorStore/GzipVectorStore-Query-10-1-10         	   18867	    107243 ns/op	     288 B/op	       3 allocs/op
BenchmarkCompressedVectorStore/GzipVectorStore-Query-10-10-10        	    8036	    202150 ns/op	    2880 B/op	       6 allocs/op
BenchmarkCompressedVectorStore/GzipVectorStore-Query-100-1-10        	    1106	   1541100 ns/op	     288 B/op	       3 allocs/op
BenchmarkCompressedVectorStore/GzipVectorStore-Query-100-10-10       	     511	   2303053 ns/op	    2880 B/op	       6 allocs/op
BenchmarkCompressedVectorStore/GzipVectorStore-Query-100-100-10      	     327	   3784640 ns/op	   25664 B/op	       9 allocs/op
BenchmarkCompressedVectorStore/GzipVectorStore-Query-1000-1-10       	     100	  12470110 ns/op	     288 B/op	       3 allocs/op
BenchmarkCompressedVectorStore/GzipVectorStore-Query-1000-10-10      	      70	  21186270 ns/op	    2880 B/op	       6 allocs/op
BenchmarkCompressedVectorStore/GzipVectorStore-Query-1000-100-10     	      58	  26575966 ns/op	   25664 B/op	       9 allocs/op
BenchmarkCompressedVectorStore/GzipVectorStore-Query-10000-1-10      	      16	 102998362 ns/op	     288 B/op	       3 allocs/op
BenchmarkCompressedVectorStore/GzipVectorStore-Query-10000-10-10     	       7	 200548024 ns/op	    2880 B/op	       6 allocs/op
BenchmarkCompressedVectorStore/GzipVectorStore-Query-10000-100-10    	       4	 290161719 ns/op	   25664 B/op	       9 allocs/op
BenchmarkCompressedVectorStore/GzipVectorStore-Query-100000-1-10     	       2	1115429562 ns/op	     288 B/op	       3 allocs/op
BenchmarkCompressedVectorStore/GzipVectorStore-Query-100000-10-10    	       1	1561993375 ns/op	    2880 B/op	       6 allocs/op
BenchmarkCompressedVectorStore/GzipVectorStore-Query-100000-100-10   	       1	1751069333 ns/op	   25664 B/op	       9 allocs/op
PASS
```

I have done a fair amount of allocation chasing, I think there is a bit more work to do, but actually, it's pretty slow. 

I also think adding [Lempel-Ziv Jaccard Distance](https://arxiv.org/pdf/1708.03346.pdf) is quite promising. We would need to write it in Go or add Cgo bindings to the C version. 