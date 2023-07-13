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


On an M1 Max - using stdgzip 

```
BenchmarkCompressedVectorStore/Insert-10-10         	   14868	     82872 ns/op	    5921 B/op	      10 allocs/op
BenchmarkCompressedVectorStore/Insert-100-10        	    1333	    959247 ns/op	   65141 B/op	     100 allocs/op
BenchmarkCompressedVectorStore/Insert-1000-10       	     121	   9476993 ns/op	  582882 B/op	    1001 allocs/op
BenchmarkCompressedVectorStore/Insert-10000-10      	      12	  95115309 ns/op	 5678265 B/op	   10010 allocs/op
BenchmarkCompressedVectorStore/InsertConcurrent-10-10         	   17211	     68893 ns/op	   24545 B/op	      30 allocs/op
BenchmarkCompressedVectorStore/InsertConcurrent-100-10        	    3165	    444302 ns/op	  376192 B/op	     300 allocs/op
BenchmarkCompressedVectorStore/InsertConcurrent-1000-10       	     270	   4889688 ns/op	 4088148 B/op	    3005 allocs/op
BenchmarkCompressedVectorStore/InsertConcurrent-10000-10      	      32	  41150789 ns/op	31635464 B/op	   30055 allocs/op
BenchmarkCompressedVectorStore/Query-10-1-10                  	   10000	    123291 ns/op	     336 B/op	       4 allocs/op
BenchmarkCompressedVectorStore/Query-10-5-10                  	   14800	     77127 ns/op	    1504 B/op	       6 allocs/op
BenchmarkCompressedVectorStore/Query-100-1-10                 	    1792	    771284 ns/op	     336 B/op	       4 allocs/op
BenchmarkCompressedVectorStore/Query-100-5-10                 	    2720	    759757 ns/op	    1820 B/op	       6 allocs/op
BenchmarkCompressedVectorStore/Query-100-25-10                	    2886	    635924 ns/op	    6669 B/op	       8 allocs/op
BenchmarkCompressedVectorStore/Query-100-100-10               	    2006	    757195 ns/op	   26972 B/op	      10 allocs/op
BenchmarkCompressedVectorStore/Query-1000-1-10                	     207	   6323184 ns/op	     312 B/op	       4 allocs/op
BenchmarkCompressedVectorStore/Query-1000-5-10                	     237	   8359605 ns/op	    1488 B/op	       6 allocs/op
BenchmarkCompressedVectorStore/Query-1000-25-10               	     188	   6447841 ns/op	    6407 B/op	       8 allocs/op
BenchmarkCompressedVectorStore/Query-1000-100-10              	     184	   6311957 ns/op	   30214 B/op	      10 allocs/op
BenchmarkCompressedVectorStore/Query-10000-1-10               	      30	  60109108 ns/op	     423 B/op	       4 allocs/op
BenchmarkCompressedVectorStore/Query-10000-5-10               	      30	  62862942 ns/op	    1459 B/op	       6 allocs/op
BenchmarkCompressedVectorStore/Query-10000-25-10              	      20	  66013446 ns/op	    6423 B/op	       8 allocs/op
BenchmarkCompressedVectorStore/Query-10000-100-10             	      27	  66908579 ns/op	   25763 B/op	      10 allocs/op
```

We expect `klauspost/compress` should improve gzip throughput by 5-10x. That should be pretty solid. Concurrent inserts are not carefully tested -- the basic intuition is that stdlib gzip is singlethreaded and we are internally writing to a slice, so doing multiple goroutines should speed up inserts, at the cost of memory overhead per routine.  

Future improvements here also include `zstd` - in practice the pure Go implementation is quite a bit faster than gzip, but the CGo library is miles faster. I do not really want to add Cgo to this library, but would consider it in a `contrib` package or similar.

I also think adding [Lempel-Ziv Jaccard Distance](https://arxiv.org/pdf/1708.03346.pdf) is quite promising. We would need to write it in Go or add Cgo bindings to the C version. 