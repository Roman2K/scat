package aprocs

import ss "secsplit"

type Proc interface {
	Process(*ss.Chunk) <-chan Res
	Finish() error
}

type EndProc interface {
	ProcessEnd(*ss.Chunk) error
}

type Res struct {
	Chunk *ss.Chunk
	Err   error
}
