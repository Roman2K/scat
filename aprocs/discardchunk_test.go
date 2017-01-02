package aprocs_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/testutil"
)

func TestDiscardChunks(t *testing.T) {
	proc := aprocs.InplaceProcFunc(func(*ss.Chunk) error {
		return nil
	})
	dc := aprocs.NewDiscardChunks(proc)
	c := &ss.Chunk{}
	chunks, err := readChunks(dc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(chunks))
}

func TestDiscardChunksError(t *testing.T) {
	someErr := errors.New("some err")
	proc := aprocs.InplaceProcFunc(func(*ss.Chunk) error {
		return someErr
	})
	dc := aprocs.NewDiscardChunks(proc)
	c := &ss.Chunk{}
	chunks, err := readChunks(dc.Process(c))
	assert.Equal(t, []*ss.Chunk{c}, chunks)
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
