package procs_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"
	"gitlab.com/Roman2K/scat"
	"gitlab.com/Roman2K/scat/procs"
	"gitlab.com/Roman2K/scat/testutil"
)

func TestParityTargetSize(t *testing.T) {
	var (
		data = scat.BytesData("abc")
	)
	parity, err := procs.NewParity(1, 1)
	assert.NoError(t, err)
	c := scat.NewChunk(0, data)
	c.SetTargetSize(99)
	chunks, err := testutil.ReadChunks(parity.Proc().Process(c))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(chunks))
	assert.Equal(t, len(data), chunks[0].TargetSize())
	assert.Equal(t, len(data), chunks[1].TargetSize())
}

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

	testutil.SetGroupErr(shardChunks[1], someErr)
	err = getErr(t, parity.Unproc().Process(chunk))
	assert.Equal(t, someErr, err)

	testutil.SetGroupErr(shardChunks[1], procs.ErrIntegrityCheckFailed)
	err = getErr(t, parity.Unproc().Process(chunk))
	assert.Error(t, err)
	assert.NotEqual(t, procs.ErrIntegrityCheckFailed, err)

	testutil.SetGroupErr(shardChunks[1], nil)
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
