package procs

import (
	"io"
)

func NewJoin(w io.Writer) Proc {
	return NewMutex(Chain{NewSort(), NewWriterTo(w)})
}
