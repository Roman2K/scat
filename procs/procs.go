package procs

import (
	"errors"

	ss "secsplit"
)

var ErrMissingFinalChunks = errors.New("missing final chunks")

var nop Proc

func init() {
	nop = inplaceProcFunc(func(*ss.Chunk) error { return nil })
}

type Proc interface {
	Process(*ss.Chunk) Res
}

type ErrProc interface {
	ProcessErr(*ss.Chunk, error) Res
}

type Finisher interface {
	Finish() error
}

type AsyncProc interface {
	Process(*ss.Chunk) <-chan Res
}

type Procer interface {
	Proc() Proc
}

type Unprocer interface {
	Unproc() Proc
}

type ProcUnprocer interface {
	Procer
	Unprocer
}

type EndProc interface {
	ProcessEnd(*ss.Chunk, []*ss.Chunk) error
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
