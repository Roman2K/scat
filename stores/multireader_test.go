package stores

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/procs"
	"scat/testutil"
)

func TestMultiReader(t *testing.T) {
	var (
		hash = testutil.Hash1.Hash
	)

	origShuffle := shuffle
	defer func() {
		shuffle = origShuffle
	}()
	shuffle = SortCopiersByIdString

	mem1 := NewMem()
	mem2 := NewMem()
	copiers := []Copier{
		Copier{"mem1", mem1, mem1.Unproc()},
		Copier{"mem2", mem2, mem2.Unproc()},
	}

	c := scat.NewChunk(0, nil)
	c.SetHash(hash)

	// none available
	mrd, err := NewMultiReader(copiers)
	assert.NoError(t, err)
	chunks, err := testutil.ReadChunks(mrd.Process(c))
	missErr, ok := err.(procs.MissingDataError)
	assert.True(t, ok)
	assert.Equal(t, errMultiReaderNoneAvail, missErr.Err)
	assert.Equal(t, []*scat.Chunk{c}, chunks)

	readData := func() string {
		chunks, err := testutil.ReadChunks(mrd.Process(c))
		assert.NoError(t, err)
		assert.Equal(t, 1, len(chunks))
		b, err := chunks[0].Data().Bytes()
		assert.NoError(t, err)
		return string(b)
	}

	// on mem2
	mem2.Set(hash, []byte("data2"))
	mrd, err = NewMultiReader(copiers)
	assert.NoError(t, err)
	assert.Equal(t, "data2", readData())

	// on mem2 and mem1
	mem1.Set(hash, []byte("data1"))
	mrd, err = NewMultiReader(copiers)
	assert.NoError(t, err)
	assert.Equal(t, "data1", readData())
}
