package testutil

import (
	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
	"secsplit/cpprocs"
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

type SliceLister []checksum.Hash

var _ cpprocs.Lister = SliceLister{}

func (sl SliceLister) Ls() ([]checksum.Hash, error) {
	return []checksum.Hash(sl), nil
}

type FinishErrProc struct {
	Err error
}

func (p FinishErrProc) Process(*ss.Chunk) <-chan aprocs.Res {
	panic("Process() not implemented")
}

func (p FinishErrProc) Finish() error {
	return p.Err
}
