package aprocs_test

import (
	"io/ioutil"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
)

func TestWriterToFinish(t *testing.T) {
	wt := aprocs.NewWriterTo(ioutil.Discard)

	// 0 missing
	// 1 ok
	_, err := readChunks(wt.Process(&ss.Chunk{Num: 1}))
	assert.NoError(t, err)
	err = wt.Finish()
	assert.Equal(t, aprocs.ErrShort, err)

	// idempotence
	err = wt.Finish()
	assert.Equal(t, aprocs.ErrShort, err)

	// 0 ok
	// 1 ok
	_, err = readChunks(wt.Process(&ss.Chunk{Num: 0}))
	assert.NoError(t, err)
	err = wt.Finish()
	assert.NoError(t, err)

	// idempotence
	err = wt.Finish()
	assert.NoError(t, err)
}
