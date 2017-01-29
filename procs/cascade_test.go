package procs_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/procs"
	"scat/testutil"
)

func TestCascade(t *testing.T) {
	nop := procs.Nop
	someErr := errors.New("some err")
	errp := procs.ProcFunc(func(c *scat.Chunk) <-chan procs.Res {
		ch := make(chan procs.Res, 2)
		defer close(ch)
		ch <- procs.Res{Err: someErr, Chunk: c}
		ch <- procs.Res{Chunk: c}
		return ch
	})
	c := scat.NewChunk(0, nil)

	casc := procs.Cascade{}
	chunks, err := testutil.ReadChunks(casc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(chunks))

	casc = procs.Cascade{nop}
	chunks, err = testutil.ReadChunks(casc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*scat.Chunk{c}, chunks)

	casc = procs.Cascade{nop, nop}
	chunks, err = testutil.ReadChunks(casc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*scat.Chunk{c}, chunks)

	casc = procs.Cascade{nop, errp}
	chunks, err = testutil.ReadChunks(casc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*scat.Chunk{c}, chunks)

	casc = procs.Cascade{errp}
	chunks, err = testutil.ReadChunks(casc.Process(c))
	assert.Equal(t, someErr, err)
	assert.Equal(t, []*scat.Chunk{c, c}, chunks)

	casc = procs.Cascade{errp, nop}
	chunks, err = testutil.ReadChunks(casc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*scat.Chunk{c}, chunks)
}

func TestCascadeFinish(t *testing.T) {
	casc := procs.Cascade{testutil.FinishErrProc{Err: nil}}
	err := casc.Finish()
	assert.NoError(t, err)
}

func TestCascadeFinishError(t *testing.T) {
	someErr := errors.New("some err")
	casc := procs.Cascade{testutil.FinishErrProc{Err: someErr}}
	err := casc.Finish()
	assert.Equal(t, someErr, err)
}
