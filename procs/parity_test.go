package procs_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/Roman2K/scat"
	"github.com/Roman2K/scat/procs"
)

func TestParityNonIntegrityError(t *testing.T) {
	parity, err := procs.NewParity(1, 1)
	assert.NoError(t, err)
	chunk := scat.NewChunk(0, nil)
	shardChunks := []*scat.Chunk{
		scat.NewChunk(0, nil),
		scat.NewChunk(0, nil),
	}
	chunk.Meta().Set("group", shardChunks)
	someErr := errors.New("some non-integrity err")

	shardChunks[1].Meta().Set("err", someErr)
	err = getErr(t, parity.Unproc().Process(chunk))
	assert.Equal(t, someErr, err)

	shardChunks[1].Meta().Set("err", procs.ErrIntegrityCheckFailed)
	err = getErr(t, parity.Unproc().Process(chunk))
	assert.Error(t, err)
	assert.NotEqual(t, procs.ErrIntegrityCheckFailed, err)

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
	parity, err := procs.NewParity(ndata, nparity)
	assert.NoError(t, err)
	testChunkNums(t, parity.Proc(), 2)
	testChunkNums(t, procs.Chain{
		procs.NewGroup(nshards),
		parity.Unproc(),
	}, nshards*2)
}
