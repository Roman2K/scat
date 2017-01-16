package procs

import (
	"scat"
	"scat/split"
)

type splitProc struct {
	min, max uint
}

func NewSplit(min, max uint) Proc {
	return ChunkIterFunc(splitProc{min, max}.process)
}

func (s splitProc) process(c scat.Chunk) scat.ChunkIter {
	return split.NewSplitter(c.Num(), c.Data().Reader(), s.min, s.max)
}
