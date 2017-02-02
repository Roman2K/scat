package procs_test

import (
	"errors"
	"testing"
	"time"

	"gitlab.com/Roman2K/scat"
	"gitlab.com/Roman2K/scat/procs"
	"gitlab.com/Roman2K/scat/testutil"
	assert "github.com/stretchr/testify/require"
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
