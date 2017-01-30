package multireader

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/procs"
	"scat/stores"
	"scat/testutil"
)

func TestMultiReader(t *testing.T) {
	var (
		hash = testutil.Hash1.Hash
	)

	shuffleOrig := shuffle
	defer func() {
		shuffle = shuffleOrig
	}()
	shuffle = testutil.SortCopiersByIdString

	mem1 := stores.NewMem()
	mem2 := stores.NewMem()
	copiers := []stores.Copier{
		stores.NewCopier("mem1", mem1, mem1.Unproc()),
		stores.NewCopier("mem2", mem2, mem2.Unproc()),
	}

	c := scat.NewChunk(0, nil)
	c.SetHash(hash)

	// none available
	mrd, err := New(copiers)
	assert.NoError(t, err)
	_, err = testutil.ReadChunks(mrd.Process(c))
	assert.Equal(t, procs.ErrMissingData, err)

	readData := func() string {
		chunks, err := testutil.ReadChunks(mrd.Process(c))
		assert.NoError(t, err)
		assert.Equal(t, 1, len(chunks))
		b, err := chunks[0].Data().Bytes()
		assert.NoError(t, err)
		return string(b)
	}

	// on mem2
	mem2.SetData(hash, []byte("data2"))
	mrd, err = New(copiers)
	assert.NoError(t, err)
	assert.Equal(t, "data2", readData())

	// on mem2 and mem1
	mem1.SetData(hash, []byte("data1"))
	mrd, err = New(copiers)
	assert.NoError(t, err)
	assert.Equal(t, "data1", readData())
}
