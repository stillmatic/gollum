package gollum

import (
	"bytes"
	"compress/gzip"
	"context"
	"sort"
)

// Compressor is a single method interface that returns a compressed representation of an object.
type Compressor interface {
	Compress(src []byte) []byte
}

type GzipCompressor struct{}

func (g GzipCompressor) Compress(src []byte) []byte {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

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
	distances := make([]float32, len(cvs.Data))

	for i, doc := range cvs.Data {
		Cx1x2 := len(cvs.Compressor.Compress(append(searchTermEncoded, doc.Encoded...)))
		min, max := minMax(len(searchTermEncoded), len(doc.Encoded))
		ncd := float32(Cx1x2-min) / float32(max)
		distances[i] = ncd
	}

	type kv struct {
		Key   int
		Value float32
	}

	var ss []kv
	for k, v := range distances {
		ss = append(ss, kv{k, v})
	}

	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value < ss[j].Value
	})

	topKDocs := make([]Document, qb.K)
	for i := 0; i < qb.K; i++ {
		topKDocs[i] = cvs.Data[ss[i].Key].Document
	}

	return topKDocs, nil
}

func (cvs *CompressedVectorStore) RetrieveAll(ctx context.Context) ([]Document, error) {
	docs := make([]Document, len(cvs.Data))
	for i, doc := range cvs.Data {
		docs[i] = doc.Document
	}
	return docs, nil
}

func NewGzipVectorStore() *CompressedVectorStore {
	return &CompressedVectorStore{
		Compressor: GzipCompressor{},
	}
}
