package stats

import (
	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
	"secsplit/cpprocs"
)

type cpProc struct {
	cpp  cpprocs.Proc
	proc aprocs.Proc
}

func NewCpProc(log *Log, name string, cpp cpprocs.Proc) cpprocs.Proc {
	return cpProc{
		cpp:  cpp,
		proc: NewProc(log, name, cpp),
	}
}

func (cp cpProc) Id() interface{} {
	return cp.cpp.Id()
}

func (cp cpProc) Ls() ([]checksum.Hash, error) {
	return cp.cpp.Ls()
}

func (cp cpProc) Process(c *ss.Chunk) <-chan aprocs.Res {
	return cp.proc.Process(c)
}

func (cp cpProc) Finish() error {
	return cp.proc.Finish()
}
