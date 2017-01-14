package testutil

import (
	"scat"
	"scat/procs"
	"scat/cpprocs"
)

func ReadChunks(ch <-chan procs.Res) (chunks []scat.Chunk, err error) {
	for res := range ch {
		if e := res.Err; e != nil && err == nil {
			err = e
		}
		chunks = append(chunks, res.Chunk)
	}
	return
}

type SliceIter struct {
	S     []scat.Chunk
	i     int
	chunk scat.Chunk
}

var _ scat.ChunkIter = &SliceIter{}

func (it *SliceIter) Next() bool {
	if it.i < len(it.S) {
		it.chunk = it.S[it.i]
		it.i++
		return true
	}
	return false
}

func (it *SliceIter) Chunk() scat.Chunk {
	return it.chunk
}

func (it *SliceIter) Err() error {
	return nil
}

type SliceLister []cpprocs.LsEntry

var _ cpprocs.Lister = SliceLister{}

func (sl SliceLister) Ls() ([]cpprocs.LsEntry, error) {
	return []cpprocs.LsEntry(sl), nil
}

type FinishErrProc struct {
	Err error
}

var _ procs.Proc = FinishErrProc{}

func (p FinishErrProc) Process(scat.Chunk) <-chan procs.Res {
	panic("Process() not implemented")
}

func (p FinishErrProc) Finish() error {
	return p.Err
}
