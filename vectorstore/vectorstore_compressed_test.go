package vectorstore_test

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	mathrand "math/rand"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stillmatic/gollum"
	"github.com/stillmatic/gollum/vectorstore"
	"github.com/stretchr/testify/assert"
)

func TestCompressedVectorStore(t *testing.T) {
	vs := vectorstore.NewGzipVectorStore()
	t.Run("implements interface", func(t *testing.T) {
		var vs2 vectorstore.VectorStore
		vs2 = vs
		assert.NotNil(t, vs2)
	})
	ctx := context.Background()
	testStrings := []string{
		"Japan's Seiko Epson Corp. has developed a 12-gram flying microrobot.",
		"The latest tiny flying robot has been unveiled in Japan.",
		"Michael Phelps won the gold medal in the 400 individual medley.",
	}
	t.Run("testInsert", func(t *testing.T) {
		for _, str := range testStrings {
			vs.Insert(ctx, gollum.Document{
				ID:      uuid.NewString(),
				Content: str,
			})
		}
		docs, err := vs.RetrieveAll(ctx)
		assert.NoError(t, err)
		assert.Equal(t, 5, len(docs))
	})
	t.Run("correctness", func(t *testing.T) {
		for _, str := range testStrings {
			vs.Insert(ctx, gollum.Document{
				ID:      uuid.NewString(),
				Content: str,
			})
		}
		docs, err := vs.Query(ctx, vectorstore.QueryRequest{
			Query: "Where was the new robot unveiled?",
			K:     5,
		})
		assert.NoError(t, err)
		assert.Equal(t, 3, len(docs))
		assert.Equal(t, "Japan's Seiko Epson Corp. has developed a 12-gram flying microrobot.", docs[0].Content)
		assert.Equal(t, "The latest tiny flying robot has been unveiled in Japan.", docs[1].Content)
	})
}

func BenchmarkCompressedVectorStore(b *testing.B) {
	ctx := context.Background()
	// Test different sizes
	sizes := []int{10, 100, 1000, 10_000, 100_000}
	// note that runtime doesn't really depend on K -
	ks := []int{1, 10, 100}
	// benchmark inserts
	stores := map[string]vectorstore.VectorStore{
		"DummyVectorStore":   vectorstore.NewDummyVectorStore(),
		"StdGzipVectorStore": vectorstore.NewStdGzipVectorStore(),
		"ZstdVectorStore":    vectorstore.NewZstdVectorStore(),
		"GzipVectorStore":    vectorstore.NewGzipVectorStore(),
	}

	for vsName, vs := range stores {
		// for _, size := range sizes {
		// 	b.Run(fmt.Sprintf("%s-Insert-%d", vsName, size), func(b *testing.B) {
		// 		// Create vector store using live compression
		// 		docs := make([]gollum.Document, size)
		// 		// Generate synthetic docs
		// 		for i := range docs {
		// 			docs[i] = syntheticDoc()
		// 		}
		// 		b.ReportAllocs()
		// 		b.ResetTimer()
		// 		for n := 0; n < b.N; n++ {
		// 			// Insert docs
		// 			for _, doc := range docs {
		// 				vs.Insert(ctx, doc)
		// 			}
		// 		}
		// 	})
		// }
		// // Concurrent writes to a slice are ok
		// for _, size := range sizes {
		// 	b.Run(fmt.Sprintf("%s-InsertConcurrent-%d", vsName, size), func(b *testing.B) {
		// 		// Create vector store using live compression
		// 		docs := make([]gollum.Document, size)
		// 		// Generate synthetic docs
		// 		for i := range docs {
		// 			docs[i] = syntheticDoc()
		// 		}
		// 		var wg sync.WaitGroup
		// 		sem := make(chan struct{}, 8)
		// 		b.ReportAllocs()
		// 		b.ResetTimer()
		// 		for n := 0; n < b.N; n++ {
		// 			// Insert docs
		// 			for _, doc := range docs {
		// 				wg.Add(1)
		// 				sem <- struct{}{}
		// 				go func(doc gollum.Document) {
		// 					defer wg.Done()
		// 					defer func() { <-sem }()
		// 					vs.Insert(ctx, doc)
		// 				}(doc)
		// 			}
		// 			wg.Wait()
		// 		}
		// 	})
		// }
		// benchmark queries
		for _, size := range sizes {
			f, err := os.Open("testdata/enwik8")
			if err != nil {
				panic(err)
			}
			defer f.Close()
			lines := make([]string, 0)
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			for _, k := range ks {
				if k <= size {
					b.Run(fmt.Sprintf("%s-Query-%d-%d", vsName, size, k), func(b *testing.B) {
						// Create vector store and insert docs
						for i := 0; i < size; i++ {
							vs.Insert(ctx, gollum.NewDocumentFromString(lines[i]))
						}
						query := vectorstore.QueryRequest{
							Query: lines[size+1],
						}
						b.ReportAllocs()
						b.ResetTimer()
						// Create query
						for n := 0; n < b.N; n++ {
							vs.Query(ctx, query)
						}
					})
				}
			}
		}
	}
}

// Helper functions
func syntheticString() string {
	// Random length between 8 and 32
	randLength := mathrand.Intn(32-8+1) + 8

	// Generate random bytes
	randBytes := make([]byte, randLength)
	rand.Read(randBytes)

	// Format as hex string
	return fmt.Sprintf("%x", randBytes)
}

// syntheticQuery return query request with random embedding
func syntheticQuery(k int) vectorstore.QueryRequest {
	return vectorstore.QueryRequest{
		Query: syntheticString(),
		K:     k,
	}
}

func BenchmarkStringToBytes(b *testing.B) {
	st := syntheticString()
	b.ResetTimer()
	b.Run("byteSlice", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			_ = []byte(st)
		}
	})
	b.Run("byteSliceCopy", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			bts := make([]byte, len(st))
			copy(bts, st)
		}
	})
	b.Run("byteSliceCopyAppend", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			bts := make([]byte, 0)
			bts = append(bts, st...)
			_ = bts
		}
	})
	b.Run("bytesBuffer", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			bb := bytes.NewBufferString(st)
			_ = bb.Bytes()
		}
	})
	b.Run("bytesBufferEmpty", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			var bb bytes.Buffer
			bb.WriteString(st)
			_ = bb.Bytes()
		}
	})
}

func dummyCompress(src []byte) []byte {
	return src
}

func minMax(val1, val2 float64) (float64, float64) {
	if val1 < val2 {
		return val1, val2
	}
	return val2, val1
}

func BenchmarkE2E(b *testing.B) {
	f, err := os.Open("testdata/enwik8")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	lines := make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	// st1 := syntheticString()
	// st2 := syntheticString()
	st1 := lines[1]
	st2 := lines[2]
	b.ResetTimer()
	b.Run("minMax", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			Cx1 := float64(len(st1))
			Cx2 := float64(len(st2))
			min, max := minMax(Cx1, Cx2)
			_ = min
			_ = max
		}
	})
	var bb bytes.Buffer
	b.Run("resetBytesBufferBytes", func(b *testing.B) {
		st1b := []byte(st1)
		st2b := []byte(st2)
		spb := []byte(" ")
		for n := 0; n < b.N; n++ {
			Cx1 := float64(len(st1b))
			Cx2 := float64(len(st2b))
			bb.Reset()
			bb.Write(st1b)
			bb.Write(spb)
			bb.Write(st2b)
			b_ := bb.Bytes()
			x1x2 := dummyCompress(b_)
			Cx1x2 := float64(len(x1x2))
			min, max := minMax(Cx1, Cx2)
			ncd := (Cx1x2 - min) / (max)
			_ = ncd
		}
	})
}

func BenchmarkConcatenateStrings(b *testing.B) {
	f, err := os.Open("testdata/enwik8")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	lines := make([]string, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	// st1 := syntheticString()
	// st2 := syntheticString()
	st1 := lines[1]
	st2 := lines[2]
	b.ResetTimer()
	b.Run("minMax", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			Cx1 := float64(len(st1))
			Cx2 := float64(len(st2))
			min, max := minMax(Cx1, Cx2)
			_ = min
			_ = max
		}
	})
	b.Run("concatenate", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			_ = []byte(st1 + " " + st2)
		}
	})
	b.Run("bytesBuffer", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			bb := bytes.NewBufferString(st1)
			bb.WriteString(" ")
			bb.WriteString(st2)
			_ = bb.Bytes()
		}
	})
	b.Run("bytesBufferEmpty", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			var bb bytes.Buffer
			bb.WriteString(st1)
			bb.WriteString(" ")
			bb.WriteString(st2)
			_ = bb.Bytes()
		}
	})
	var bb bytes.Buffer
	b.Run("resetBytesBuffer", func(b *testing.B) {
		for n := 0; n < b.N; n++ {
			bb.Reset()
			bb.WriteString(st1)
			bb.WriteString(" ")
			bb.WriteString(st2)
			_ = bb.Bytes()
		}
	})

}

func BenchmarkCompress(b *testing.B) {
	compressors := map[string]vectorstore.Compressor{
		"DummyCompressor":   vectorstore.NewDummyVectorStore().Compressor,
		"StdGzipCompressor": vectorstore.NewStdGzipVectorStore().Compressor,
		"ZstdCompressor":    vectorstore.NewZstdVectorStore().Compressor,
		"GzipCompressor":    vectorstore.NewGzipVectorStore().Compressor,
	}
	str := syntheticString()
	b.ResetTimer()
	for name, compressor := range compressors {
		b.Run(name, func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				_ = compressor.Compress([]byte(str))
			}
		})
	}
}
