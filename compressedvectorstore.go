package gollum

import (
	"bytes"
	stdgzip "compress/gzip"
	"context"
	// gzip "github.com/klauspost/compress/gzip"
)

// Compressor is a single method interface that returns a compressed representation of an object.
type Compressor interface {
	Compress(src []byte) []byte
}

// GzipCompressor uses the klauspost/compress gzip compressor.
// We generally suggest using this optimized implementation over the stdlib.
type GzipCompressor struct{}

// GzipCompressor uses the std gzip compressor.
type StdGzipCompressor struct{}

// func (g GzipCompressor) Compress(src []byte) []byte {
// 	var b bytes.Buffer
// 	gz := gzip.NewWriter(&b)

// 	if _, err := gz.Write(src); err != nil {
// 		panic(err)
// 	}
// 	if err := gz.Close(); err != nil {
// 		panic(err)
// 	}
// 	return b.Bytes()
// }

func (g StdGzipCompressor) Compress(src []byte) []byte {
	var b bytes.Buffer
	gz := stdgzip.NewWriter(&b)

	if _, err := gz.Write(src); err != nil {
		panic(err)
	}
	if err := gz.Close(); err != nil {
		panic(err)
	}
	return b.Bytes()
}

type CompressedDocument struct {
	Document
	Encoded []byte
}

type CompressedVectorStore struct {
	Data       []CompressedDocument
	Compressor Compressor
}

func (ts *CompressedVectorStore) Insert(ctx context.Context, d Document) error {
	encoded := ts.Compressor.Compress([]byte(d.Content))
	ts.Data = append(ts.Data, CompressedDocument{Document: d, Encoded: encoded})
	return nil
}

func minMax(val1, val2 int) (int, int) {
	if val1 < val2 {
		return val1, val2
	}
	return val2, val1
}

func (cvs *CompressedVectorStore) Query(ctx context.Context, qb QueryRequest) ([]Document, error) {
	searchTermEncoded := cvs.Compressor.Compress([]byte(qb.Query))

	h := Heap{}
	h.Init()
	k := qb.K

	for _, doc := range cvs.Data {
		Cx1x2 := len(cvs.Compressor.Compress(append(searchTermEncoded, doc.Encoded...)))
		min, max := minMax(len(searchTermEncoded), len(doc.Encoded))
		ncd := float32(Cx1x2-min) / float32(max)

		// We want a max heap, so we take the negative of ncd
		node := nodeSimilarity{
			Document:   doc.Document,
			Similarity: -ncd,
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

func (cvs *CompressedVectorStore) RetrieveAll(ctx context.Context) ([]Document, error) {
	docs := make([]Document, len(cvs.Data))
	for i, doc := range cvs.Data {
		docs[i] = doc.Document
	}
	return docs, nil
}

func NewStdGzipVectorStore() *CompressedVectorStore {
	return &CompressedVectorStore{
		Compressor: StdGzipCompressor{},
	}
}
