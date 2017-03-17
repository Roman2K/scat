package procs_test

import (
	"errors"
	"testing"

	"github.com/Roman2K/scat"
	"github.com/Roman2K/scat/procs"
	"github.com/Roman2K/scat/testutil"
	assert "github.com/stretchr/testify/require"
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
	testutil.TestFinishErrForward(t, func(proc procs.Proc) testutil.Finisher {
		return procs.Cascade{procs.Nop, proc}
	})
}
