package aprocs

import (
	"io"

	ss "secsplit"
)

type writeTo struct {
	w io.Writer
}

func NewWriterTo(w io.Writer) Proc {
	return InplaceProcFunc(writeTo{w: w}.process)
}

func (wt writeTo) process(c *ss.Chunk) (err error) {
	_, err = wt.w.Write(c.Data)
	return
}
