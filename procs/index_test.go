package procs_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
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

	idx := procs.NewIndex(buf)
	end := func(num int, contents string, size int, setFinals bool) {
		c := &ss.Chunk{
			Num:  num,
			Size: size,
			Hash: checksum.Sum([]byte(contents)),
		}
		finals := []*ss.Chunk{}
		if setFinals {
			finals = append(finals, c)
		}
		err := idx.ProcessEnd(c, finals)
		assert.NoError(t, err)
	}

	end(1, "b", 22, true)
	assert.Equal(t, 0, nlines())
	end(0, "a", 11, false)
	assert.Equal(t, 0, nlines())

	err := idx.Finish()
	assert.Equal(t, procs.ErrMissingFinalChunks, err)
	assert.Equal(t, 0, nlines())

	end(2, "a", 33, true)
	assert.Equal(t, 3, nlines())

	// idempotence
	err = idx.Finish()
	assert.NoError(t, err)
	assert.Equal(t, 3, nlines())

	expectedIndex := sumStr("a") + " 33\n" +
		sumStr("b") + " 22\n" +
		sumStr("a") + " 33\n"
	assert.Equal(t, expectedIndex, buf.String())
}
