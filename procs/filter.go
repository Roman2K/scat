package procs

import "github.com/Roman2K/scat"

type Filter struct {
	Proc
	Filter func(Res) Res
}

func (p Filter) Process(c *scat.Chunk) <-chan Res {
	ch := p.Proc.Process(c)
	out := make(chan Res)
	go func() {
		defer close(out)
		for res := range ch {
			out <- p.Filter(res)
		}
	}()
	return out
}
