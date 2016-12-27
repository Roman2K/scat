package aprocs_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
)

func TestParityNonIntegrityError(t *testing.T) {
	parity, err := aprocs.NewParity(1, 1)
	assert.NoError(t, err)
	chunk := &ss.Chunk{}
	shardChunks := []*ss.Chunk{
		&ss.Chunk{},
		&ss.Chunk{},
	}
	chunk.SetMeta("group", shardChunks)
	someErr := errors.New("some non-integrity err")

	shardChunks[1].SetMeta("err", someErr)
	err = getErr(t, parity.Unproc().Process(chunk))
	assert.Equal(t, someErr, err)

	shardChunks[1].SetMeta("err", aprocs.ErrIntegrityCheckFailed)
	err = getErr(t, parity.Unproc().Process(chunk))
	assert.Error(t, err)
	assert.NotEqual(t, aprocs.ErrIntegrityCheckFailed, err)

	shardChunks[1].SetMeta("err", nil)
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
	testChunkNums(t, aprocs.NewChain([]aprocs.Proc{
		aprocs.NewGroup(nshards),
		parity.Unproc(),
	}), nshards*2)
}
