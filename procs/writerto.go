package procs

import (
	"io"

	"github.com/Roman2K/scat"
)

type WriterTo struct {
	W io.Writer
}

func (wt WriterTo) Process(c *scat.Chunk) <-chan Res {
	_, err := io.Copy(wt.W, c.Data().Reader())
	return SingleRes(c, err)
}

func (wt WriterTo) Finish() error {
	return nil
}
