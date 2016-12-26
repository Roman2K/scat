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
	nlines := func() int {
		return strings.Count(buf.String(), "\n")
	}

	sumStr := func(s string) string {
		h := checksum.Sum([]byte(s))
		return fmt.Sprintf("%x", h)
	}

	idx := aprocs.NewIndex(buf)
	end := func(num int, contents string, size int, chCount int, setFinal bool) {
		c := &ss.Chunk{
			Num:  num,
			Size: size,
			Hash: checksum.Sum([]byte(contents)),
		}
		ch := idx.Process(c)
		count := 0
		for range ch {
			count++
		}
		assert.Equal(t, chCount, count)
		if setFinal {
			err := idx.ProcessEnd(c, c)
			assert.NoError(t, err)
		}
	}

	end(1, "b", 22, 1, true)
	assert.Equal(t, 0, nlines())
	end(0, "a", 11, 1, false)
	assert.Equal(t, 0, nlines())

	err := idx.Finish()
	assert.Equal(t, procs.ErrShort, err)
	assert.Equal(t, 0, nlines())

	end(2, "a", 33, 0, true)
	assert.Equal(t, 3, nlines())

	err = idx.Finish()
	assert.NoError(t, err)

	// idempotence
	err = idx.Finish()
	assert.NoError(t, err)

	assert.Equal(t, 3, nlines())

	expectedIndex := sumStr("a") + " 33\n" +
		sumStr("b") + " 22\n" +
		sumStr("a") + " 33\n"
	assert.Equal(t, expectedIndex, buf.String())
}
