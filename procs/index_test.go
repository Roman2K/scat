package procs_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"

	"github.com/Roman2K/scat"
	"github.com/Roman2K/scat/checksum"
	"github.com/Roman2K/scat/procs"
	"github.com/Roman2K/scat/testutil"
)

func TestIndex(t *testing.T) {
	buf := &bytes.Buffer{}
	idx := procs.NewIndexProc(buf)
	nlines := func() int {
		return strings.Count(buf.String(), "\n")
	}
	end := func(num int, contents string, targetSize, chCount int) {
		c := testIndexChunk(num, targetSize, sum(contents))
		ch := idx.Process(c)
		count := 0
		for range ch {
			count++
			err := idx.ProcessFinal(c, c)
			assert.NoError(t, err)
		}
		assert.Equal(t, chCount, count)
		err := idx.ProcessEnd(c)
		assert.NoError(t, err)
	}

	end(1, "b", 22, 1)
	assert.Equal(t, 0, nlines())
	end(0, "a", 11, 1)
	assert.Equal(t, 2, nlines())

	// short
	end(3, "c", 44, 1)
	err := idx.Finish()
	assert.Equal(t, procs.ErrShort, err)
	assert.Equal(t, 2, nlines())

	// idempotence
	err = idx.Finish()
	assert.Equal(t, procs.ErrShort, err)
	assert.Equal(t, 2, nlines())

	// dup
	end(2, "a", 33, 0)
	assert.Equal(t, 4, nlines())

	// output
	expectedIndex := "" +
		sumStr("a") + " 11\n" +
		sumStr("b") + " 22\n" +
		sumStr("a") + " 11\n" +
		sumStr("c") + " 44\n"
	assert.Equal(t, expectedIndex, buf.String())

	// fully flushed
	err = idx.Finish()
	assert.NoError(t, err)
	assert.Equal(t, 4, nlines())

	// idempotence
	err = idx.Finish()
	assert.NoError(t, err)
	assert.Equal(t, 4, nlines())
}

func TestIndexSameChunkNewData(t *testing.T) {
	buf := &bytes.Buffer{}
	idx := procs.NewIndexProc(buf)
	c := testIndexChunk(0, 123, sum("a"))
	_, err := testutil.ReadChunks(idx.Process(c))
	assert.NoError(t, err)
	newChunk := c.WithData(nil)
	newChunk.SetHash(sum("b"))
	err = idx.ProcessFinal(c, newChunk)
	assert.NoError(t, err)
	err = idx.ProcessEnd(c)
	assert.NoError(t, err)
	expectedIndex := sumStr("b") + " 123\n"
	assert.Equal(t, expectedIndex, buf.String())
}

func TestIndexProcessFinalDup(t *testing.T) {
	idx := procs.NewIndexProc(ioutil.Discard)
	c0 := testIndexChunk(0, 0, sum("c0"))
	c01 := testIndexChunk(1, 0, sum("c0"))

	chunks, err := testutil.ReadChunks(idx.Process(c0))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(chunks))

	chunks, err = testutil.ReadChunks(idx.Process(c01))
	assert.NoError(t, err)
	assert.Equal(t, 0, len(chunks))

	err = idx.ProcessEnd(c01)
	assert.NoError(t, err)

	err = idx.ProcessFinal(c0, c0)
	assert.NoError(t, err)

	err = idx.ProcessEnd(c0)
	assert.NoError(t, err)
}

func TestIndexFinalsOutOfOrder(t *testing.T) {
	buf := &bytes.Buffer{}
	idx := procs.NewIndexProc(buf)
	c0 := testIndexChunk(0, 0, sum("c0"))
	c1 := testIndexChunk(1, 1, sum("c1"))
	c2 := testIndexChunk(2, 2, sum("c2"))
	c3 := testIndexChunk(3, 3, sum("c3"))

	chunks, err := testutil.ReadChunks(idx.Process(c0))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(chunks))

	chunks, err = testutil.ReadChunks(idx.Process(c1))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(chunks))

	idx.ProcessFinal(c1, c1)
	idx.ProcessEnd(c1)
	idx.ProcessFinal(c0, c3)
	idx.ProcessFinal(c0, c2)
	idx.ProcessEnd(c0)

	expectedIndex := "" +
		sumStr("c2") + " 0\n" +
		sumStr("c3") + " 0\n" +
		sumStr("c1") + " 1\n"
	assert.Equal(t, expectedIndex, buf.String())
}

func TestIndexProcessFinalError(t *testing.T) {
	idx := procs.NewIndexProc(ioutil.Discard)
	c0 := testIndexChunk(0, 0, sum("c0"))
	c01 := testIndexChunk(1, 0, sum("c0"))

	// add final to unprocessed chunk
	err := idx.ProcessFinal(c0, c0)
	assert.Equal(t, procs.ErrIndexUnprocessedChunk, err)

	// ...
	idx.Process(c0)

	// add final to dup chunk
	err = idx.ProcessFinal(c01, c01)
	assert.Equal(t, procs.ErrIndexDup, err)

	// add final after end of process
	err = idx.ProcessFinal(c0, c0)
	assert.NoError(t, err)
	err = idx.ProcessEnd(c0)
	assert.NoError(t, err)
	err = idx.ProcessFinal(c0, c0)
	assert.Equal(t, procs.ErrIndexProcessEnded, err)
}

func TestIndexProcessEndError(t *testing.T) {
	idx := procs.NewIndexProc(ioutil.Discard)
	c0 := testIndexChunk(0, 0, sum("c0"))
	c01 := testIndexChunk(1, 0, sum("c0"))

	// end of unprocess chunk
	err := idx.ProcessEnd(c0)
	assert.Equal(t, procs.ErrIndexUnprocessedChunk, err)

	// ...
	idx.Process(c0)

	// end of dup chunk
	err = idx.ProcessEnd(c01)
	assert.NoError(t, err)
}

func sumStr(s string) string {
	h := checksum.SumBytes([]byte(s))
	return fmt.Sprintf("%x", h)
}

func sum(data string) checksum.Hash {
	return checksum.SumBytes([]byte(data))
}

func testIndexChunk(num, targetSize int, hash checksum.Hash) (c *scat.Chunk) {
	c = scat.NewChunk(num, nil)
	c.SetTargetSize(targetSize)
	c.SetHash(hash)
	return
}
