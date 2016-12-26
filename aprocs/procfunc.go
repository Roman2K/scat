package aprocs

import ss "secsplit"

type ProcFunc func(*ss.Chunk) <-chan Res

var _ Proc = ProcFunc(func(*ss.Chunk) <-chan Res { return nil })

func (fn ProcFunc) Process(c *ss.Chunk) <-chan Res {
	return fn(c)
}

func (fn ProcFunc) Finish() error {
	return nil
}
