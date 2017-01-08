package aprocs

import "scat"

type onEnd struct {
	proc Proc
	cb   func(error)
}

func NewOnEnd(proc Proc, cb func(error)) Proc {
	return onEnd{proc: proc, cb: cb}
}

func (oe onEnd) Process(c scat.Chunk) <-chan Res {
	out := make(chan Res)
	ch := oe.proc.Process(c)
	go func() {
		defer close(out)
		var err error
		defer func() { oe.cb(err) }()
		for res := range ch {
			if e := res.Err; e != nil && err == nil {
				err = e
			}
			out <- res
		}
	}()
	return out
}

func (oe onEnd) Finish() error {
	return oe.proc.Finish()
}
