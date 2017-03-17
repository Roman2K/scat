package procs_test

import (
	"errors"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
	"github.com/Roman2K/scat"
	"github.com/Roman2K/scat/procs"
	"github.com/Roman2K/scat/testutil"
)

func TestConcur(t *testing.T) {
	const wait = 40 * time.Millisecond

	done := make(chan struct{})
	go func() {
		defer close(done)
		time.Sleep(wait)
	}()
	proc := procs.InplaceFunc(func(*scat.Chunk) error {
		<-done
		time.Sleep(wait)
		return nil
	})

	// error
	someErr := errors.New("some err")
	dynp := testDynProcer{[]procs.Proc{proc, proc}, someErr}
	conc := procs.NewConcur(len(dynp.procs), dynp)
	_, err := testutil.ReadChunks(conc.Process(scat.NewChunk(0, nil)))
	assert.Equal(t, someErr, err)

	// ok
	dynp = testDynProcer{[]procs.Proc{proc, proc}, nil}
	c := scat.NewChunk(0, nil)
	conc = procs.NewConcur(len(dynp.procs), dynp)
	start := time.Now()
	chunks, err := testutil.ReadChunks(conc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*scat.Chunk{c, c}, chunks)
	elapsed := time.Now().Sub(start)
	assert.True(t, elapsed > wait)
	assert.True(t, elapsed < wait*(time.Duration(1+len(dynp.procs))))
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
