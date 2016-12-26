package testhelp

import ss "secsplit"

type SliceIter struct {
	S     []*ss.Chunk
	i     int
	chunk *ss.Chunk
}

var _ ss.ChunkIterator = &SliceIter{S: []*ss.Chunk{}}

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
