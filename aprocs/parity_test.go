package aprocs_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/aprocs"
)

func TestParityNonIntegrityError(t *testing.T) {
	parity, err := aprocs.NewParity(1, 1)
	assert.NoError(t, err)
	chunk := scat.NewChunk(0, nil)
	shardChunks := []scat.Chunk{
		scat.NewChunk(0, nil),
		scat.NewChunk(0, nil),
	}
	chunk.Meta().Set("group", shardChunks)
	someErr := errors.New("some non-integrity err")

	shardChunks[1].Meta().Set("err", someErr)
	err = getErr(t, parity.Unproc().Process(chunk))
	assert.Equal(t, someErr, err)

	shardChunks[1].Meta().Set("err", aprocs.ErrIntegrityCheckFailed)
	err = getErr(t, parity.Unproc().Process(chunk))
	assert.Error(t, err)
	assert.NotEqual(t, aprocs.ErrIntegrityCheckFailed, err)

	shardChunks[1].Meta().Set("err", nil)
	err = getErr(t, parity.Unproc().Process(chunk))
	assert.Error(t, err)
	assert.NotEqual(t, someErr, err)
}

func TestParityChunkNums(t *testing.T) {
	const (
		ndata   = 2
		nparity = 1
		nshards = ndata + nparity
	)
	parity, err := aprocs.NewParity(ndata, nparity)
	assert.NoError(t, err)
	testChunkNums(t, parity.Proc(), 2)
	testChunkNums(t, aprocs.Chain{
		aprocs.NewGroup(nshards),
		parity.Unproc(),
	}, nshards*2)
}
