package gollum_test

import (
	"context"
	"testing"

	. "github.com/stillmatic/gollum"
	"github.com/stretchr/testify/assert"
	"gocloud.dev/blob/fileblob"
)

func TestMemoryDocStore(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryDocStore()
	doc := Document{ID: "1", Content: "test data"}
	doc2 := Document{ID: "2", Content: "test data 2"}

	// ensure store implements the DocStore interface
	var _ DocStore = store

	t.Run("Insert document", func(t *testing.T) {
		err := store.Insert(ctx, doc)
		assert.NoError(t, err)
		err = store.Insert(ctx, doc2)
		assert.NoError(t, err)
	})

	t.Run("Retrieve document", func(t *testing.T) {
		retrievedDoc, err := store.Retrieve(ctx, "1")
		assert.NoError(t, err)
		assert.Equal(t, doc, retrievedDoc)
	})

	t.Run("Retrieve non-existing document", func(t *testing.T) {
		_, err := store.Retrieve(ctx, "non-existing-id")
		assert.Error(t, err)
	})

	t.Run("Persist document store", func(t *testing.T) {
		// persist to testdata/docstore.json
		bucket, err := fileblob.OpenBucket("testdata", nil)
		assert.NoError(t, err)
		err = store.Persist(ctx, bucket, "docstore.json")
		assert.NoError(t, err)
	})

	t.Run("Load document store from disk", func(t *testing.T) {
		// load from testdata/docstore.json
		bucket, err := fileblob.OpenBucket("testdata", nil)
		assert.NoError(t, err)
		loadedStore, err := NewMemoryDocStoreFromDisk(ctx, bucket, "docstore.json")
		assert.NoError(t, err)
		assert.Equal(t, store, loadedStore)
	})

}
