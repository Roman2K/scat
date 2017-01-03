package aprocs_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/testutil"
)

func TestCascade(t *testing.T) {
	nop := aprocs.Nop
	someErr := errors.New("some err")
	errp := aprocs.InplaceProcFunc(func(*ss.Chunk) error {
		return someErr
	})
	c := &ss.Chunk{}

	casc := aprocs.Cascade{}
	chunks, err := readChunks(casc.Process(&ss.Chunk{}))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(chunks))

	casc = aprocs.Cascade{nop}
	chunks, err = readChunks(casc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*ss.Chunk{c}, chunks)

	casc = aprocs.Cascade{nop, nop}
	chunks, err = readChunks(casc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*ss.Chunk{c}, chunks)

	casc = aprocs.Cascade{nop, errp}
	chunks, err = readChunks(casc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*ss.Chunk{c}, chunks)

	casc = aprocs.Cascade{errp}
	chunks, err = readChunks(casc.Process(c))
	assert.Equal(t, someErr, err)
	assert.Equal(t, []*ss.Chunk{c}, chunks)

	casc = aprocs.Cascade{errp, nop}
	chunks, err = readChunks(casc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*ss.Chunk{c}, chunks)
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
