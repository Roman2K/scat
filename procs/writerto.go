package procs

import (
	"io"

	"gitlab.com/Roman2K/scat"
)

type WriterTo struct {
	W io.Writer
}

func (wt WriterTo) Process(c *scat.Chunk) <-chan Res {
	ch := make(chan Res, 1)
	defer close(ch)
	_, err := io.Copy(wt.W, c.Data().Reader())
	ch <- Res{Chunk: c, Err: err}
	return ch
}

func (wt WriterTo) Finish() error {
	return nil
}
