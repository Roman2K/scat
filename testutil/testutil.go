package testutil

import (
	"scat"
	"scat/checksum"
	"scat/procs"
	"scat/stores"
)

func ReadChunks(ch <-chan procs.Res) (chunks []*scat.Chunk, err error) {
	for res := range ch {
		if e := res.Err; e != nil && err == nil {
			err = e
		}
		chunks = append(chunks, res.Chunk)
	}
	return
}

type SliceIter struct {
	S     []*scat.Chunk
	i     int
	chunk *scat.Chunk
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

func (it *SliceIter) Chunk() *scat.Chunk {
	return it.chunk
}

func (it *SliceIter) Err() error {
	return nil
}

type SliceLister []stores.LsEntry

var _ stores.Lister = SliceLister{}

func (sl SliceLister) Ls() ([]stores.LsEntry, error) {
	return []stores.LsEntry(sl), nil
}

type FinishErrProc struct {
	Err error
}

var _ procs.Proc = FinishErrProc{}

func (p FinishErrProc) Process(*scat.Chunk) <-chan procs.Res {
	panic("Process() not implemented")
}

func (p FinishErrProc) Finish() error {
	return p.Err
}

var Hash1 = struct {
	Hash checksum.Hash
	Hex  string
}{
	Hash: checksum.Hash{
		44, 242, 77, 186, 95, 176, 163, 14, 38, 232, 59, 42, 197, 185, 226, 158,
		27, 22, 30, 92, 31, 167, 66, 94, 115, 4, 51, 98, 147, 139, 152, 36,
	},
	Hex: "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
}
