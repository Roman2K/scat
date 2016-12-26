package aprocs_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
	"secsplit/procs"
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
			Hash: checksum.Sum([]byte(contents)),
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

	// fully flushed
	err = idx.Finish()
	assert.NoError(t, err)
	assert.Equal(t, 4, nlines())

	// idempotence
	err = idx.Finish()
	assert.NoError(t, err)
	assert.Equal(t, 4, nlines())

	expectedIndex := sumStr("a") + " 11\n" +
		sumStr("b") + " 22\n" +
		sumStr("a") + " 11\n" +
		sumStr("c") + " 44\n"
	assert.Equal(t, expectedIndex, buf.String())
}

func sumStr(s string) string {
	h := checksum.Sum([]byte(s))
	return fmt.Sprintf("%x", h)
}
