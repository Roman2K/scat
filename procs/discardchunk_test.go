package procs_test

import (
	"errors"
	"testing"

	"gitlab.com/Roman2K/scat"
	"gitlab.com/Roman2K/scat/procs"
	"gitlab.com/Roman2K/scat/testutil"
	assert "github.com/stretchr/testify/require"
)

func TestDiscardChunks(t *testing.T) {
	proc := procs.InplaceFunc(func(*scat.Chunk) error {
		return nil
	})
	dc := procs.DiscardChunks{proc}
	c := scat.NewChunk(0, nil)
	chunks, err := testutil.ReadChunks(dc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(chunks))
}

func TestDiscardChunksError(t *testing.T) {
	someErr := errors.New("some err")
	proc := procs.InplaceFunc(func(*scat.Chunk) error {
		return someErr
	})
	dc := procs.DiscardChunks{proc}
	c := scat.NewChunk(0, nil)
	chunks, err := testutil.ReadChunks(dc.Process(c))
	assert.Equal(t, []*scat.Chunk{c}, chunks)
	assert.Equal(t, someErr, err)
}

func TestDiscardChunksFinish(t *testing.T) {
	testutil.TestFinishErrForward(t, func(proc procs.Proc) testutil.Finisher {
		return procs.DiscardChunks{proc}
	})
}
