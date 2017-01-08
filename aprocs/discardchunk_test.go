package aprocs_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/aprocs"
	"scat/testutil"
)

func TestDiscardChunks(t *testing.T) {
	proc := aprocs.InplaceFunc(func(scat.Chunk) error {
		return nil
	})
	dc := aprocs.NewDiscardChunks(proc)
	c := scat.NewChunk(0, nil)
	chunks, err := readChunks(dc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(chunks))
}

func TestDiscardChunksError(t *testing.T) {
	someErr := errors.New("some err")
	proc := aprocs.InplaceFunc(func(scat.Chunk) error {
		return someErr
	})
	dc := aprocs.NewDiscardChunks(proc)
	c := scat.NewChunk(0, nil)
	chunks, err := readChunks(dc.Process(c))
	assert.Equal(t, []scat.Chunk{c}, chunks)
	assert.Equal(t, someErr, err)
}

func TestDiscardChunksFinish(t *testing.T) {
	proc := testutil.FinishErrProc{Err: nil}
	dc := aprocs.NewDiscardChunks(proc)
	err := dc.Finish()
	assert.NoError(t, err)
}

func TestDiscardChunksFinishError(t *testing.T) {
	someErr := errors.New("some err")
	proc := testutil.FinishErrProc{Err: someErr}
	dc := aprocs.NewDiscardChunks(proc)
	err := dc.Finish()
	assert.Equal(t, someErr, err)
}
