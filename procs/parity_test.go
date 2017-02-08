package procs_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"
	"gitlab.com/Roman2K/scat"
	"gitlab.com/Roman2K/scat/procs"
	"gitlab.com/Roman2K/scat/testutil"
)

func TestParityNonIntegrityError(t *testing.T) {
	parity, err := procs.NewParity(1, 1)
	assert.NoError(t, err)
	shardChunks := []*scat.Chunk{
		scat.NewChunk(0, nil),
		scat.NewChunk(1, nil),
	}
	chunk, err := testutil.Group(shardChunks)
	assert.NoError(t, err)
	someErr := errors.New("some non-integrity err")

	setGroupErr(shardChunks[1], someErr)
	err = getErr(t, parity.Unproc().Process(chunk))
	assert.Equal(t, someErr, err)

	setGroupErr(shardChunks[1], procs.ErrIntegrityCheckFailed)
	err = getErr(t, parity.Unproc().Process(chunk))
	assert.Error(t, err)
	assert.NotEqual(t, procs.ErrIntegrityCheckFailed, err)

	setGroupErr(shardChunks[1], nil)
	err = getErr(t, parity.Unproc().Process(chunk))
	assert.Error(t, err)
	assert.NotEqual(t, someErr, err)
}

func setGroupErr(c *scat.Chunk, err error) {
	ch := procs.NewGroup(1).ProcessErr(c, err)
	for range ch {
	}
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
