package gollum

import (
	"bytes"
	stdgzip "compress/gzip"
	"context"
	"io"

	gzip "github.com/klauspost/compress/gzip"
	"github.com/klauspost/compress/zstd"
	"github.com/stillmatic/gollum/syncpool"
)

// Compressor is a single method interface that returns a compressed representation of an object.
type Compressor interface {
	Compress(src []byte) []byte
	CompressIO(in io.Reader, out io.Writer) error
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
	pool syncpool.Pool[*stdgzip.Writer]
}

func (g *GzipCompressor) Compress(src []byte) []byte {
	var b bytes.Buffer
	gz := g.pool.Get()
	defer g.pool.Put(gz)

	gz.Reset(&b)
	if _, err := gz.Write(src); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	return b.Bytes()
}

func (g *GzipCompressor) CompressIO(in io.Reader, out io.Writer) error {
	enc := g.pool.Get()
	defer g.pool.Put(enc)
	var b bytes.Buffer
	enc.Reset(&b)

	if _, err := io.Copy(enc, in); err != nil {
		return err
	}
	if err := enc.Flush(); err != nil {
		return err
	}
	return enc.Close()
}

func (g *ZstdCompressor) Compress(src []byte) []byte {
	return g.enc.EncodeAll(src, make([]byte, 0, len(src)))

	var b bytes.Buffer
	zstd := g.enc
	// zstd := g.pool.Get()
	zstd.Reset(&b)
	if _, err := zstd.Write(src); err != nil {
		panic(err)
	}
	if err := zstd.Flush(); err != nil {
		panic(err)
	}

	// g.pool.Put(zstd)
	return b.Bytes()
}

func (g *ZstdCompressor) CompressIO(in io.Reader, out io.Writer) error {
	enc := g.enc
	// enc := g.pool.Get()
	// defer g.pool.Put(enc)
	var b bytes.Buffer
	enc.Reset(&b)

	if _, err := io.Copy(enc, in); err != nil {
		return err
	}
	if err := enc.Flush(); err != nil {
		return err
	}
	return enc.Close()
}

func (g *StdGzipCompressor) Compress(src []byte) []byte {
	w := io.Discard
	var b bytes.Buffer
	gz := g.pool.Get()
	gz.Reset(w)

	if _, err := gz.Write(src); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	g.pool.Put(gz)
	return b.Bytes()
}

func (g *StdGzipCompressor) CompressIO(in io.Reader, out io.Writer) error {
	enc := g.pool.Get()
	defer g.pool.Put(enc)
	var b bytes.Buffer
	enc.Reset(&b)

	if _, err := io.Copy(enc, in); err != nil {
		return err
	}
	if err := enc.Flush(); err != nil {
		return err
	}
	return enc.Close()
}

type CompressedDocument struct {
	Document
	Encoded []byte
}

type CompressedVectorStore struct {
	Data       []CompressedDocument
	Compressor Compressor
}

// Insert compresses the document and inserts it into the store.
// An alternative implementation would ONLY store the compressed representation and decompress as necessary.
func (ts *CompressedVectorStore) Insert(ctx context.Context, d Document) error {
	encoded := ts.Compressor.Compress([]byte(d.Content))
	ts.Data = append(ts.Data, CompressedDocument{Document: d, Encoded: encoded})
	return nil
}

func minMax(val1, val2 float64) (float64, float64) {
	if val1 < val2 {
		return val1, val2
	}
	return val2, val1
}

func (cvs *CompressedVectorStore) Query(ctx context.Context, qb QueryRequest) ([]Document, error) {
	// singlethreaded approach
	// out := &bytes.Buffer{}
	// qbytes := bytes.NewReader([]byte(qb.Query))
	// err := cvs.Compressor.CompressIO(qbytes, out)
	// if err != nil {
	// 	return nil, err
	// }
	// searchTermEncoded := out.Bytes()
	searchTermEncoded := cvs.Compressor.Compress([]byte(qb.Query))

	h := Heap{}
	h.Init()
	k := qb.K
	if k > len(cvs.Data) {
		k = len(cvs.Data)
	}

	for _, doc := range cvs.Data {
		Cx1 := float64(len(searchTermEncoded))
		Cx2 := float64(len(doc.Encoded))
		// x1x2 := bytes.NewReader([]byte(qb.Query + " " + doc.Content))
		x1x2 := cvs.Compressor.Compress([]byte(qb.Query + " " + doc.Content))
		// out.Reset()
		// err = cvs.Compressor.CompressIO(x1x2, out)
		// if err != nil {
		// 	return nil, err
		// }
		// Cx1x2 := float64(len(out.Bytes()))
		Cx1x2 := float64(len(x1x2))
		min, max := minMax(Cx1, Cx2)
		ncd := (Cx1x2 - min) / (max)

		node := nodeSimilarity{
			Document:   doc.Document,
			Similarity: float32(ncd),
		}

		h.Push(node)
		if h.Len() > k {
			h.Pop()
		}
	}

	docs := make([]Document, k)
	for i := range docs {
		docs[k-i-1] = h.Pop().Document
	}

	return docs, nil
}

// func (cvs *CompressedVectorStore) Query(ctx context.Context, qb QueryRequest) ([]Document, error) {
// 	// multithreaded approach
// 	searchTermEncoded := cvs.Compressor.Compress([]byte(qb.Query))

// 	distances := make([]nodeSimilarity, len(cvs.Data))
// 	k := qb.K

// 	var wg sync.WaitGroup
// 	sem := make(chan struct{}, 8)

// 	for i, doc := range cvs.Data {
// 		wg.Add(1)
// 		sem <- struct{}{}

// 		go func(i int, doc CompressedDocument) {
// 			defer wg.Done()
// 			defer func() { <-sem }()

// 			Cx1x2 := len(cvs.Compressor.Compress(append(searchTermEncoded, doc.Encoded...)))
// 			min, max := minMax(len(searchTermEncoded), len(doc.Encoded))
// 			ncd := float32(Cx1x2-min) / float32(max)

// 			node := nodeSimilarity{
// 				Document:   doc.Document,
// 				Similarity: ncd,
// 			}

// 			distances[i] = node
// 		}(i, doc)
// 	}
// 	wg.Wait()

// 	sort.Slice(distances, func(i, j int) bool {
// 		return distances[i].Similarity < distances[j].Similarity
// 	})

// 	docs := make([]Document, k)
// 	for i := range docs {
// 		docs[i] = distances[i].Document
// 	}

// 	return docs, nil
// }

func (cvs *CompressedVectorStore) RetrieveAll(ctx context.Context) ([]Document, error) {
	docs := make([]Document, len(cvs.Data))
	for i, doc := range cvs.Data {
		docs[i] = doc.Document
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
