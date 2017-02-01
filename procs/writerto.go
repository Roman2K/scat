package procs

import (
	"io"

	"github.com/Roman2K/scat"
)

type writerTo struct {
	w io.Writer
}

func NewWriterTo(w io.Writer) Proc {
	return InplaceFunc(writerTo{w: w}.process)
}

func (wt writerTo) process(c *scat.Chunk) (err error) {
	_, err = io.Copy(wt.w, c.Data().Reader())
	return
}
