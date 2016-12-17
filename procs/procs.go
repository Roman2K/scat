package procs

import ss "secsplit"

type Proc interface {
	Process(*ss.Chunk) Res
}

type ProcFinisher interface {
	Proc
	Finisher
}

type Finisher interface {
	Finish() error
}

type AsyncProc interface {
	Process(*ss.Chunk) <-chan Res
}

type AsyncProcFinisher interface {
	AsyncProc
	Finisher
}

type ender interface {
	end(*ss.Chunk, []*ss.Chunk)
}

type Res struct {
	Chunks []*ss.Chunk
	Err    error
}

type inplaceProcFunc func(*ss.Chunk) error

func (fn inplaceProcFunc) Process(c *ss.Chunk) Res {
	err := fn(c)
	return Res{Chunks: []*ss.Chunk{c}, Err: err}
}

type procFunc func(*ss.Chunk) Res

func (fn procFunc) Process(c *ss.Chunk) Res {
	return fn(c)
}
