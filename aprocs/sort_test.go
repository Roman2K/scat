package aprocs_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
)

func TestSortFinish(t *testing.T) {
	sortp := aprocs.NewSort()

	// 0 missing
	// 1 ok
	_, err := readChunks(sortp.Process(&ss.Chunk{Num: 1}))
	assert.NoError(t, err)
	err = sortp.Finish()
	assert.Equal(t, aprocs.ErrShort, err)

	// idempotence
	err = sortp.Finish()
	assert.Equal(t, aprocs.ErrShort, err)

	// 0 ok
	// 1 ok
	_, err = readChunks(sortp.Process(&ss.Chunk{Num: 0}))
	assert.NoError(t, err)
	err = sortp.Finish()
	assert.NoError(t, err)

	// idempotence
	err = sortp.Finish()
	assert.NoError(t, err)
}
