package procs

import (
	ss "secsplit"
	"secsplit/aprocs"
)

type async struct {
	proc Proc
}

func A(proc Proc) aprocs.Proc {
	return async{proc}
}

func (ap async) Process(c *ss.Chunk) <-chan aprocs.Res {
	var ch chan aprocs.Res
	res := ap.proc.Process(c)
	if res.Err != nil {
		ch = make(chan aprocs.Res, 1)
		ch <- aprocs.Res{Err: res.Err}
	} else {
		ch = make(chan aprocs.Res, len(res.Chunks))
		for _, c := range res.Chunks {
			ch <- aprocs.Res{Chunk: c}
		}
	}
	close(ch)
	return ch
}

func (ap async) Finish() error {
	if f, ok := ap.proc.(Finisher); ok {
		return f.Finish()
	}
	return nil
}
