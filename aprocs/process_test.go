package aprocs_test

import (
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	ss "secsplit"
	"secsplit/aprocs"
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
	iter := &sliceIter{S: chunks}
	err = aprocs.Process(proc, iter)
	return
}

type sliceIter struct {
	S     []*ss.Chunk
	Delay time.Duration
	i     int
	chunk *ss.Chunk
}

var _ ss.ChunkIterator = &sliceIter{S: []*ss.Chunk{}}

func (it *sliceIter) Next() bool {
	if it.i < len(it.S) {
		it.chunk = it.S[it.i]
		it.i++
		return true
	}
	return false
}

func (it *sliceIter) Chunk() *ss.Chunk {
	return it.chunk
}

func (it *sliceIter) Err() error {
	return nil
}
