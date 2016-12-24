package procs

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
)

func TestParityNonIntegrityError(t *testing.T) {
	parity, err := Parity(1, 1)
	assert.NoError(t, err)
	chunk := &ss.Chunk{}
	shardChunks := []*ss.Chunk{
		&ss.Chunk{},
		&ss.Chunk{},
	}
	chunk.SetMeta("group", shardChunks)
	someErr := errors.New("some non-integrity err")

	shardChunks[1].SetMeta("err", someErr)
	res := parity.Unproc().Process(chunk)
	assert.Equal(t, someErr, res.Err)

	shardChunks[1].SetMeta("err", errIntegrityCheckFailed)
	res = parity.Unproc().Process(chunk)
	assert.Error(t, res.Err)
	assert.NotEqual(t, errIntegrityCheckFailed, res.Err)

	shardChunks[1].SetMeta("err", nil)
	res = parity.Unproc().Process(chunk)
	assert.Error(t, res.Err)
	assert.NotEqual(t, someErr, res.Err)
}

func TestParityChunkNum(t *testing.T) {
	const (
		ndata   = 2
		nparity = 1
		nshards = ndata + nparity
	)
	parity, err := Parity(ndata, nparity)
	assert.NoError(t, err)
	testChunkNums(t, parity.Proc(), 2)
	testChunkNums(t, NewChain([]Proc{
		Group(nshards),
		parity.Unproc(),
	}), nshards*2)
}
