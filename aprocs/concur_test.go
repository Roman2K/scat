package aprocs_test

import (
	"errors"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
)

func TestConcur(t *testing.T) {
	a := aprocs.InplaceProcFunc(func(c *ss.Chunk) error {
		time.Sleep(20 * time.Millisecond)
		return nil
	})
	b := aprocs.InplaceProcFunc(func(c *ss.Chunk) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	// error
	someErr := errors.New("some err")
	dynp := testDynProcer{[]aprocs.Proc{a, a, b}, someErr}
	conc := aprocs.NewConcur(2, dynp)
	_, err := readChunks(conc.Process(&ss.Chunk{}))
	assert.Equal(t, someErr, err)

	// no error
	dynp = testDynProcer{[]aprocs.Proc{a, a, b}, nil}
	c := &ss.Chunk{}
	conc = aprocs.NewConcur(2, dynp)
	start := time.Now()
	chunks, err := readChunks(conc.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*ss.Chunk{c, c, c}, chunks)
	elapsed := time.Now().Sub(start)
	assert.True(t, elapsed > 30*time.Millisecond)
	assert.True(t, elapsed < 40*time.Millisecond)
}

type testDynProcer struct {
	procs []aprocs.Proc
	err   error
}

func (dynp testDynProcer) Procs(*ss.Chunk) ([]aprocs.Proc, error) {
	return dynp.procs, dynp.err
}

func (dynp testDynProcer) Finish() error {
	panic("Finish() not implemented")
}
