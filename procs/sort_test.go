package procs_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/procs"
)

func TestSortFinish(t *testing.T) {
	s := &procs.Sort{}

	// 0 missing
	// 1 ok
	res := s.Process(&ss.Chunk{Num: 1})
	assert.NoError(t, res.Err)
	err := s.Finish()
	assert.Equal(t, procs.ErrMissingFinalChunks, err)

	// idempotence
	err = s.Finish()
	assert.Equal(t, procs.ErrMissingFinalChunks, err)

	// 0 ok
	// 1 ok
	res = s.Process(&ss.Chunk{Num: 0})
	assert.NoError(t, res.Err)

	err = s.Finish()
	assert.NoError(t, err)

	// idempotence
	err = s.Finish()
	assert.NoError(t, err)
}
