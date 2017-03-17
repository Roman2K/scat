package procs_test

import (
	"errors"
	"testing"

	"github.com/Roman2K/scat"
	"github.com/Roman2K/scat/procs"
	"github.com/Roman2K/scat/testutil"
	assert "github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	p := procs.Filter{
		Proc: procs.Nop,
		Filter: func(res procs.Res) procs.Res {
			return res
		},
	}
	c := scat.NewChunk(0, nil)
	chunks, err := testutil.ReadChunks(p.Process(c))
	assert.NoError(t, err)
	assert.Equal(t, []*scat.Chunk{c}, chunks)

	someErr := errors.New("some err")
	p = procs.Filter{
		Proc: procs.Nop,
		Filter: func(res procs.Res) procs.Res {
			res.Err = someErr
			return res
		},
	}
	c = scat.NewChunk(0, nil)
	chunks, err = testutil.ReadChunks(p.Process(c))
	assert.Equal(t, someErr, err)
	assert.Equal(t, []*scat.Chunk{c}, chunks)
}

func TestFilterFinish(t *testing.T) {
	testutil.TestFinishErrForward(t, func(proc procs.Proc) testutil.Finisher {
		return procs.Filter{
			Proc:   proc,
			Filter: func(res procs.Res) procs.Res { return res },
		}
	})
}
