package procs_test

import (
	"errors"
	"testing"

	"github.com/Roman2K/scat"
	"github.com/Roman2K/scat/procs"
	"github.com/Roman2K/scat/testutil"
	assert "github.com/stretchr/testify/require"
)

func TestOnEnd(t *testing.T) {
	received := []error{}
	proc := procs.InplaceFunc(func(*scat.Chunk) error {
		return nil
	})
	oe := procs.OnEnd{proc, func(err error) {
		received = append(received, err)
	}}
	c := scat.NewChunk(0, nil)
	chunks, err := testutil.ReadChunks(oe.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*scat.Chunk{c}, chunks)
	assert.Equal(t, []error{nil}, received)
}

func TestOnEndError(t *testing.T) {
	received := []error{}
	someErr := errors.New("some err")
	proc := procs.InplaceFunc(func(*scat.Chunk) error {
		return someErr
	})
	oe := procs.OnEnd{proc, func(err error) {
		received = append(received, err)
	}}
	c := scat.NewChunk(0, nil)
	chunks, err := testutil.ReadChunks(oe.Process(c))
	assert.Equal(t, someErr, err)
	assert.Equal(t, []*scat.Chunk{c}, chunks)
	assert.Equal(t, []error{someErr}, received)
}

func TestOnEndFinish(t *testing.T) {
	testutil.TestFinishErrForward(t, func(proc procs.Proc) testutil.Finisher {
		return procs.OnEnd{proc, func(error) {}}
	})
}
