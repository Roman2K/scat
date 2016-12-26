package aprocs

import (
	"errors"

	ss "secsplit"
)

var ErrShort = errors.New("missing final chunks")

type Proc interface {
	Process(*ss.Chunk) <-chan Res
	Finish() error
}

type EndProc interface {
	ProcessFinal(*ss.Chunk, *ss.Chunk) error
	ProcessEnd(*ss.Chunk) error
}

type Res struct {
	Chunk *ss.Chunk
	Err   error
}
