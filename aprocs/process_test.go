package aprocs_test

import (
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/testhelp"
)

func TestProcess(t *testing.T) {
	chunks := []*ss.Chunk{
		&ss.Chunk{Num: 0},
	}
	processed, err := process(chunks)
	assert.NoError(t, err)
	assert.Equal(t, []int{0}, processed)
}

func TestProcessErr(t *testing.T) {
	testErr := errors.New("test err")
	errChunk := &ss.Chunk{Num: 1}
	errChunk.SetMeta("testErr", testErr)
	chunks := []*ss.Chunk{
		&ss.Chunk{Num: 0},
		errChunk,
		&ss.Chunk{Num: 2},
	}
	processed, err := process(chunks)
	assert.Equal(t, testErr, err)
	sort.Ints(processed)
	assert.Equal(t, []int{0}, processed[:1])
}

func process(chunks []*ss.Chunk) (processed []int, err error) {
	proc := aprocs.InplaceProcFunc(func(c *ss.Chunk) error {
		if val := c.GetMeta("testErr"); val != nil {
			return val.(error)
		}
		processed = append(processed, c.Num)
		return nil
	})
	iter := &testhelp.SliceIter{S: chunks}
	err = aprocs.Process(proc, iter)
	return
}
