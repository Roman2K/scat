package procs

import (
	"io"

	"scat"
)

type writeTo struct {
	w io.Writer
}

func NewWriterTo(w io.Writer) Proc {
	return InplaceFunc(writeTo{w: w}.process)
}

func (wt writeTo) process(c scat.Chunk) (err error) {
	_, err = io.Copy(wt.w, c.Data().Reader())
	return
}
