package procs_test

import (
	"errors"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
	"gitlab.com/Roman2K/scat"
	"gitlab.com/Roman2K/scat/procs"
	"gitlab.com/Roman2K/scat/testutil"
)

func TestConcur(t *testing.T) {
	a := procs.InplaceFunc(func(c *scat.Chunk) error {
		time.Sleep(20 * time.Millisecond)
		return nil
	})
	b := procs.InplaceFunc(func(c *scat.Chunk) error {
		time.Sleep(30 * time.Millisecond)
		return nil
	})

	// error
	someErr := errors.New("some err")
	dynp := testDynProcer{[]procs.Proc{a, a, b}, someErr}
	conc := procs.NewConcur(2, dynp)
	_, err := testutil.ReadChunks(conc.Process(scat.NewChunk(0, nil)))
	assert.Equal(t, someErr, err)

	// no error
	dynp = testDynProcer{[]procs.Proc{a, a, b}, nil}
	c := scat.NewChunk(0, nil)
	conc = procs.NewConcur(2, dynp)
	start := time.Now()
	chunks, err := testutil.ReadChunks(conc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*scat.Chunk{c, c, c}, chunks)
	elapsed := time.Now().Sub(start)
	assert.True(t, elapsed > 20*time.Millisecond)
	assert.True(t, elapsed < 65*time.Millisecond)
}

func TestConcurFinish(t *testing.T) {
	testutil.TestFinishErrForward(t, func(proc procs.Proc) testutil.Finisher {
		dynp := testProcDynProcer{proc}
		return procs.NewConcur(0, dynp)
	})
}

type testDynProcer struct {
	procs []procs.Proc
	err   error
}

func (dynp testDynProcer) Procs(*scat.Chunk) ([]procs.Proc, error) {
	return dynp.procs, dynp.err
}

func (dynp testDynProcer) Finish() error {
	panic("Finish() not implemented")
}

type testProcDynProcer struct {
	procs.Proc
}

func (dynp testProcDynProcer) Procs(*scat.Chunk) ([]procs.Proc, error) {
	return []procs.Proc{dynp.Proc}, nil
}
