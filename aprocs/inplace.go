package aprocs

import ss "secsplit"

type InplaceProcFunc func(*ss.Chunk) error

var _ Proc = InplaceProcFunc(func(*ss.Chunk) error { return nil })

func (fn InplaceProcFunc) Process(c *ss.Chunk) <-chan Res {
	ch := make(chan Res, 1)
	err := fn(c)
	ch <- Res{Chunk: c, Err: err}
	close(ch)
	return ch
}

func (InplaceProcFunc) Finish() error {
	return nil
}
