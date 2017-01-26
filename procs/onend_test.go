package procs_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/procs"
	"scat/testutil"
)

func TestOnEnd(t *testing.T) {
	received := []error{}
	proc := procs.InplaceFunc(func(scat.Chunk) error {
		return nil
	})
	oe := procs.NewOnEnd(proc, func(err error) {
		received = append(received, err)
	})
	c := scat.NewChunk(0, nil)
	chunks, err := testutil.ReadChunks(oe.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []scat.Chunk{c}, chunks)
	assert.Equal(t, []error{nil}, received)
}

func TestOnEndError(t *testing.T) {
	received := []error{}
	someErr := errors.New("some err")
	proc := procs.InplaceFunc(func(scat.Chunk) error {
		return someErr
	})
	oe := procs.NewOnEnd(proc, func(err error) {
		received = append(received, err)
	})
	c := scat.NewChunk(0, nil)
	chunks, err := testutil.ReadChunks(oe.Process(c))
	assert.Equal(t, someErr, err)
	assert.Equal(t, []scat.Chunk{c}, chunks)
	assert.Equal(t, []error{someErr}, received)
}

func TestOnEndFinish(t *testing.T) {
	proc := testutil.FinishErrProc{Err: nil}
	oe := procs.NewOnEnd(proc, func(error) {})
	err := oe.Finish()
	assert.NoError(t, err)
}

func TestOnEndFinishError(t *testing.T) {
	someErr := errors.New("some err")
	proc := testutil.FinishErrProc{Err: someErr}
	oe := procs.NewOnEnd(proc, func(error) {})
	err := oe.Finish()
	assert.Equal(t, someErr, err)
}
