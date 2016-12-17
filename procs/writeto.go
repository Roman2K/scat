package procs

import (
	"io"

	ss "secsplit"
)

type writerTo struct {
	w io.Writer
}

func WriteTo(w io.Writer) Proc {
	return inplaceProcFunc(writerTo{w}.process)
}

func (wt writerTo) process(c *ss.Chunk) (err error) {
	_, err = wt.w.Write(c.Data)
	return
}
