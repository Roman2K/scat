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
	proc := aprocs.ProcFunc(func(c *ss.Chunk) <-chan aprocs.Res {
		processed = append(processed, c.Num)
		err, _ := c.GetMeta("testErr").(error)
		ch := make(chan aprocs.Res, 1)
		ch <- aprocs.Res{Chunk: c, Err: err}
		close(ch)
		return ch
	})
	iter := &testhelp.SliceIter{S: chunks}
	err = aprocs.Process(proc, iter)
	return
}
