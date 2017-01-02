package aprocs_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/testutil"
)

func TestOnEnd(t *testing.T) {
	received := []error{}
	proc := aprocs.InplaceProcFunc(func(*ss.Chunk) error {
		return nil
	})
	oe := aprocs.NewOnEnd(proc, func(err error) {
		received = append(received, err)
	})
	c := &ss.Chunk{}
	chunks, err := readChunks(oe.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*ss.Chunk{c}, chunks)
	assert.Equal(t, []error{nil}, received)
}

func TestOnEndError(t *testing.T) {
	received := []error{}
	someErr := errors.New("some err")
	proc := aprocs.InplaceProcFunc(func(*ss.Chunk) error {
		return someErr
	})
	oe := aprocs.NewOnEnd(proc, func(err error) {
		received = append(received, err)
	})
	c := &ss.Chunk{}
	chunks, err := readChunks(oe.Process(c))
	assert.Equal(t, someErr, err)
	assert.Equal(t, []*ss.Chunk{c}, chunks)
	assert.Equal(t, []error{someErr}, received)
}

func TestOnEndFinish(t *testing.T) {
	proc := testutil.FinishErrProc{Err: nil}
	oe := aprocs.NewOnEnd(proc, func(error) {})
	err := oe.Finish()
	assert.NoError(t, err)
}

func TestOnEndFinishError(t *testing.T) {
	someErr := errors.New("some err")
	proc := testutil.FinishErrProc{Err: someErr}
	oe := aprocs.NewOnEnd(proc, func(error) {})
	err := oe.Finish()
	assert.Equal(t, someErr, err)
}
