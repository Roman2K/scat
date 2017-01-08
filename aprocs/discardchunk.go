package aprocs

import "scat"

type discardChunks struct {
	proc Proc
}

func NewDiscardChunks(proc Proc) Proc {
	return discardChunks{proc: proc}
}

func (dc discardChunks) Process(c scat.Chunk) <-chan Res {
	out := make(chan Res)
	ch := dc.proc.Process(c)
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

func (dc discardChunks) Finish() error {
	return dc.proc.Finish()
}
