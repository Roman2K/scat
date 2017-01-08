package aprocs_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/aprocs"
)

func TestSortFinish(t *testing.T) {
	sortp := aprocs.NewSort()

	// 0 missing
	// 1 ok
	_, err := readChunks(sortp.Process(scat.NewChunk(1, nil)))
	assert.NoError(t, err)
	err = sortp.Finish()
	assert.Equal(t, aprocs.ErrShort, err)

	// idempotence
	err = sortp.Finish()
	assert.Equal(t, aprocs.ErrShort, err)

	// 0 ok
	// 1 ok
	_, err = readChunks(sortp.Process(scat.NewChunk(0, nil)))
	assert.NoError(t, err)
	err = sortp.Finish()
	assert.NoError(t, err)

	// idempotence
	err = sortp.Finish()
	assert.NoError(t, err)
}
