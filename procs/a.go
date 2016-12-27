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
	if res.Err != nil && len(res.Chunks) == 0 {
		panic("procs.A: won't send err associated with no chunk")
	}
	ch = make(chan aprocs.Res, len(res.Chunks))
	for _, c := range res.Chunks {
		ch <- aprocs.Res{Chunk: c, Err: res.Err}
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
