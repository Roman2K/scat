package procs

import "scat"

type OnEnd struct {
	Proc
	OnEnd func(error)
}

func (oe OnEnd) Process(c *scat.Chunk) <-chan Res {
	ch := oe.Proc.Process(c)
	out := make(chan Res)
	go func() {
		defer close(out)
		var err error
		defer func() { oe.OnEnd(err) }()
		for res := range ch {
			if e := res.Err; e != nil && err == nil {
				err = e
			}
			out <- res
		}
	}()
	return out
}
