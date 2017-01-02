package testutil

import (
	ss "secsplit"
	"secsplit/aprocs"
)

func ReadChunks(ch <-chan aprocs.Res) (chunks []*ss.Chunk, err error) {
	for res := range ch {
		if e := res.Err; e != nil && err == nil {
			err = e
		}
		chunks = append(chunks, res.Chunk)
	}
	return
}

type SliceIter struct {
	S     []*ss.Chunk
	i     int
	chunk *ss.Chunk
}

var _ ss.ChunkIterator = &SliceIter{}

func (it *SliceIter) Next() bool {
	if it.i < len(it.S) {
		it.chunk = it.S[it.i]
		it.i++
		return true
	}
	return false
}

func (it *SliceIter) Chunk() *ss.Chunk {
	return it.chunk
}

func (it *SliceIter) Err() error {
	return nil
}
