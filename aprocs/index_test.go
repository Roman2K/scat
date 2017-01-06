package aprocs_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
	"secsplit/procs"
	"secsplit/testutil"
)

func TestIndex(t *testing.T) {
	buf := &bytes.Buffer{}
	idx := aprocs.NewIndex(buf)
	nlines := func() int {
		return strings.Count(buf.String(), "\n")
	}
	end := func(num int, contents string, size int, chCount int) {
		c := &ss.Chunk{
			Num:  num,
			Size: size,
			Hash: sum(contents),
		}
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

func TestIndexFinalsOutOfOrder(t *testing.T) {
	buf := &bytes.Buffer{}
	idx := aprocs.NewIndex(buf)
	c0 := &ss.Chunk{Num: 0, Size: 0, Hash: sum("c0")}
	c1 := &ss.Chunk{Num: 1, Size: 1, Hash: sum("c1")}
	c2 := &ss.Chunk{Num: 2, Size: 2, Hash: sum("c2")}
	c3 := &ss.Chunk{Num: 3, Size: 3, Hash: sum("c3")}

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
	idx := aprocs.NewIndex(ioutil.Discard)
	c0 := &ss.Chunk{Num: 0, Size: 0, Hash: sum("c0")}

	// add final to unprocessed chunk
	err := idx.ProcessFinal(c0, c0)
	assert.Equal(t, aprocs.ErrIndexUnprocessedChunk, err)

	// add final after end of process
	idx.Process(c0)
	err = idx.ProcessFinal(c0, c0)
	assert.NoError(t, err)
	err = idx.ProcessEnd(c0)
	assert.NoError(t, err)
	err = idx.ProcessFinal(c0, c0)
	assert.Equal(t, aprocs.ErrIndexProcessEnded, err)
}

func TestIndexProcessEndError(t *testing.T) {
	idx := aprocs.NewIndex(ioutil.Discard)
	c0 := &ss.Chunk{Num: 0, Size: 0, Hash: sum("c0")}

	// end of unprocess chunk
	err := idx.ProcessEnd(c0)
	assert.Equal(t, aprocs.ErrIndexUnprocessedChunk, err)

	// ok
	idx.Process(c0)
	err = idx.ProcessEnd(c0)
	assert.NoError(t, err)
}

func sumStr(s string) string {
	h := checksum.Sum([]byte(s))
	return fmt.Sprintf("%x", h)
}

func sum(data string) checksum.Hash {
	return checksum.Sum([]byte(data))
}
