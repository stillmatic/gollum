package gollum

import (
	"bytes"
	stdgzip "compress/gzip"
	"context"
	"io"
	"sync"

	gzip "github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
	"github.com/stillmatic/gollum/syncpool"
)

// Compressor is a single method interface that returns a compressed representation of an object.
type Compressor interface {
	Compress(src []byte) []byte
}

// GzipCompressor uses the klauspost/compress gzip compressor.
// We generally suggest using this optimized implementation over the stdlib.
type GzipCompressor struct {
	pool    syncpool.Pool[*gzip.Writer]
	bufPool syncpool.Pool[*bytes.Buffer]
}

// ZstdCompressor uses the klauspost/compress zstd compressor.
type ZstdCompressor struct {
	pool    syncpool.Pool[*zstd.Encoder]
	bufPool syncpool.Pool[*bytes.Buffer]
	enc     *zstd.Encoder
}

// StdGzipCompressor uses the std gzip compressor.
type StdGzipCompressor struct {
	pool    syncpool.Pool[*stdgzip.Writer]
	bufPool syncpool.Pool[*bytes.Buffer]
}

type DummyCompressor struct {
	bufPool syncpool.Pool[*bytes.Buffer]
}

func (g *GzipCompressor) Compress(src []byte) []byte {
	gz := g.pool.Get()
	b := g.bufPool.Get()
	b.Reset()
	defer g.pool.Put(gz)
	defer g.bufPool.Put(b)

	gz.Reset(b)
	if _, err := gz.Write(src); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	return b.Bytes()
}

func (g *ZstdCompressor) Compress(src []byte) []byte {
	// return g.enc.EncodeAll(src, make([]byte, 0, len(src)))
	b := g.bufPool.Get()
	defer g.bufPool.Put(b)
	b.Reset()
	enc := g.enc
	// zstd := g.pool.Get()
	enc.Reset(b)
	if _, err := enc.Write(src); err != nil {
		panic(err)
	}
	if err := enc.Flush(); err != nil {
		panic(err)
	}

	// g.pool.Put(zstd)
	return b.Bytes()
}

func (g *DummyCompressor) Compress(src []byte) []byte {
	return src
}

func (g *StdGzipCompressor) Compress(src []byte) []byte {
	b := g.bufPool.Get()
	defer g.bufPool.Put(b)
	b.Reset()
	gz := g.pool.Get()
	defer g.pool.Put(gz)
	gz.Reset(b)

	if _, err := gz.Write(src); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	return b.Bytes()
}

type CompressedDocument struct {
	*Document
	Encoded   []byte
	Unencoded []byte
}

type CompressedVectorStore struct {
	Data       []CompressedDocument
	Compressor Compressor
}

// Insert compresses the document and inserts it into the store.
// An alternative implementation would ONLY store the compressed representation and decompress as necessary.
func (ts *CompressedVectorStore) Insert(ctx context.Context, d Document) error {
	bb := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(bb)
	bb.Reset()
	bb.WriteString(d.Content)
	docBytes := bb.Bytes()
	encoded := ts.Compressor.Compress(docBytes)
	ts.Data = append(ts.Data, CompressedDocument{Document: &d, Encoded: encoded, Unencoded: docBytes})
	return nil
}

func minMax(val1, val2 float64) (float64, float64) {
	if val1 < val2 {
		return val1, val2
	}
	return val2, val1
}

var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

var nodePool = syncpool.New[NodeSimilarity](func() NodeSimilarity {
	return NodeSimilarity{}
})

var spaceBytes = []byte(" ")

func (cvs *CompressedVectorStore) Query(ctx context.Context, qb QueryRequest) ([]*Document, error) {
	bb := bufPool.Get().(*bytes.Buffer)
	defer bufPool.Put(bb)
	bb.Reset()
	queryBytes := make([]byte, len(qb.Query))
	copy(queryBytes, qb.Query)
	searchTermEncoded := cvs.Compressor.Compress(queryBytes)

	h := Heap{}
	h.Init()
	k := qb.K
	if k > len(cvs.Data) {
		k = len(cvs.Data)
	}

	for _, doc := range cvs.Data {
		Cx1 := float64(len(searchTermEncoded))
		Cx2 := float64(len(doc.Encoded))
		bb.Write(queryBytes)
		bb.Write(spaceBytes)
		bb.Write(doc.Unencoded)
		x1x2 := cvs.Compressor.Compress(bb.Bytes())
		Cx1x2 := float64(len(x1x2))
		min, max := minMax(Cx1, Cx2)
		ncd := (Cx1x2 - min) / (max)
		// ncd := 0.5

		node := NodeSimilarity{
			Document:   doc.Document,
			Similarity: float32(ncd),
		}
		// node := nodePool.Get()
		// node.Document = doc.Document
		// node.Similarity = float32(ncd)

		h.Push(node)
		if h.Len() > k {
			h.Pop()
		}
		// nodePool.Put(node)
		bb.Reset()
	}

	var docs []*Document
	// docs := make([]*Document, k)
	// for i := range docs {
	// 	docs[k-i-1] = h.Pop().Document
	// }

	return docs, nil
}

func (cvs *CompressedVectorStore) RetrieveAll(ctx context.Context) ([]Document, error) {
	docs := make([]Document, len(cvs.Data))
	for i, doc := range cvs.Data {
		docs[i] = *doc.Document
	}
	return docs, nil
}

func NewStdGzipVectorStore() *CompressedVectorStore {
	w := io.Discard
	return &CompressedVectorStore{
		Compressor: &StdGzipCompressor{
			pool: syncpool.New[*stdgzip.Writer](func() *stdgzip.Writer {
				return stdgzip.NewWriter(w)
			}),
		},
	}
}

func NewGzipVectorStore() *CompressedVectorStore {
	w := io.Discard
	return &CompressedVectorStore{
		Compressor: &GzipCompressor{
			pool: syncpool.New[*gzip.Writer](func() *gzip.Writer {
				return gzip.NewWriter(w)
			}),
		},
	}
}

func NewZstdVectorStore() *CompressedVectorStore {
	w := io.Discard
	enc, err := zstd.NewWriter(w, zstd.WithEncoderCRC(false))
	if err != nil {
		panic(err)
	}
	return &CompressedVectorStore{
		Compressor: &ZstdCompressor{
			enc: enc,
			pool: syncpool.New[*zstd.Encoder](func() *zstd.Encoder {
				enc, err := zstd.NewWriter(w, zstd.WithEncoderCRC(false))
				if err != nil {
					panic(err)
				}
				return enc
			}),
		},
	}
}

func NewDummyVectorStore() *CompressedVectorStore {
	return &CompressedVectorStore{
		Compressor: &DummyCompressor{
			bufPool: syncpool.New[*bytes.Buffer](func() *bytes.Buffer {
				return new(bytes.Buffer)
			}),
		},
	}
}
