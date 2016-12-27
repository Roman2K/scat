package aprocs

import (
	"errors"

	ss "secsplit"
)

var ErrShort = errors.New("missing final chunks")

var Nop Proc

func init() {
	Nop = InplaceProcFunc(func(*ss.Chunk) error { return nil })
}

type Proc interface {
	Process(*ss.Chunk) <-chan Res
	Finish() error
}

type EndProc interface {
	ProcessFinal(*ss.Chunk, *ss.Chunk) error
	ProcessEnd(*ss.Chunk) error
}

type ErrProc interface {
	ProcessErr(*ss.Chunk, error) <-chan Res
}

type Res struct {
	Chunk *ss.Chunk
	Err   error
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
