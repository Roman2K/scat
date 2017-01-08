package aprocs_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/aprocs"
	"scat/testutil"
)

func TestCascade(t *testing.T) {
	nop := aprocs.Nop
	someErr := errors.New("some err")
	errp := aprocs.ProcFunc(func(c scat.Chunk) <-chan aprocs.Res {
		ch := make(chan aprocs.Res, 2)
		ch <- aprocs.Res{Err: someErr, Chunk: c}
		ch <- aprocs.Res{Chunk: c}
		close(ch)
		return ch
	})
	c := scat.NewChunk(0, nil)

	casc := aprocs.Cascade{}
	chunks, err := readChunks(casc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(chunks))

	casc = aprocs.Cascade{nop}
	chunks, err = readChunks(casc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []scat.Chunk{c}, chunks)

	casc = aprocs.Cascade{nop, nop}
	chunks, err = readChunks(casc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []scat.Chunk{c}, chunks)

	casc = aprocs.Cascade{nop, errp}
	chunks, err = readChunks(casc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []scat.Chunk{c}, chunks)

	casc = aprocs.Cascade{errp}
	chunks, err = readChunks(casc.Process(c))
	assert.Equal(t, someErr, err)
	assert.Equal(t, []scat.Chunk{c, c}, chunks)

	casc = aprocs.Cascade{errp, nop}
	chunks, err = readChunks(casc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []scat.Chunk{c}, chunks)
}

func TestCascadeFinish(t *testing.T) {
	casc := aprocs.Cascade{testutil.FinishErrProc{Err: nil}}
	err := casc.Finish()
	assert.NoError(t, err)
}

func TestCascadeFinishError(t *testing.T) {
	someErr := errors.New("some err")
	casc := aprocs.Cascade{testutil.FinishErrProc{Err: someErr}}
	err := casc.Finish()
	assert.Equal(t, someErr, err)
}
