package aprocs_test

import (
	"errors"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"

	"scat"
	"scat/aprocs"
	"scat/testutil"
)

func TestProcess(t *testing.T) {
	chunks := []scat.Chunk{
		scat.NewChunk(0, nil),
	}
	processed, err := process(chunks)
	assert.NoError(t, err)
	assert.Equal(t, []int{0}, processed)
}

func TestProcessErr(t *testing.T) {
	testErr := errors.New("test err")
	errChunk := scat.NewChunk(1, nil)
	errChunk.Meta().Set("testErr", testErr)
	chunks := []scat.Chunk{
		scat.NewChunk(0, nil),
		errChunk,
		scat.NewChunk(2, nil),
	}
	processed, err := process(chunks)
	assert.Equal(t, testErr, err)
	sort.Ints(processed)
	assert.Equal(t, []int{0}, processed[:1])
}

func process(chunks []scat.Chunk) (processed []int, err error) {
	proc := aprocs.ProcFunc(func(c scat.Chunk) <-chan aprocs.Res {
		processed = append(processed, c.Num())
		err, _ := c.Meta().Get("testErr").(error)
		ch := make(chan aprocs.Res, 1)
		ch <- aprocs.Res{Chunk: c, Err: err}
		close(ch)
		return ch
	})
	iter := &testutil.SliceIter{S: chunks}
	err = aprocs.Process(proc, iter)
	return
}
