package aprocs

import (
	"io"
	"sync"

	ss "secsplit"
	"secsplit/seriessort"
)

type writeTo struct {
	w       io.Writer
	order   seriessort.Series
	orderMu sync.Mutex
}

func NewWriterTo(w io.Writer) Proc {
	return &writeTo{w: w}
}

func (wt *writeTo) Process(c *ss.Chunk) <-chan Res {
	return InplaceProcFunc(wt.process).Process(c)
}

func (wt *writeTo) process(c *ss.Chunk) (err error) {
	wt.orderMu.Lock()
	wt.order.Add(c.Num, c)
	sorted := wt.order.Sorted()
	wt.order.Drop(len(sorted))
	wt.orderMu.Unlock()
	for _, val := range sorted {
		data := val.(*ss.Chunk).Data
		_, err = wt.w.Write(data)
		if err != nil {
			return
		}
	}
	return
}

func (wt *writeTo) Finish() error {
	wt.orderMu.Lock()
	len := wt.order.Len()
	wt.orderMu.Unlock()
	if len > 0 {
		return ErrShort
	}
	return nil
}
