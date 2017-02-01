package procs_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/Roman2K/scat"
	"github.com/Roman2K/scat/procs"
	"github.com/Roman2K/scat/testutil"
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
