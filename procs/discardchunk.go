package procs

import "github.com/Roman2K/scat"

type DiscardChunks struct {
	Proc
}

func (dc DiscardChunks) Process(c *scat.Chunk) <-chan Res {
	ch := dc.Proc.Process(c)
	out := make(chan Res)
	go func() {
		defer close(out)
		for res := range ch {
			if res.Err == nil {
				continue
			}
			out <- res
		}
	}()
	return out
}
