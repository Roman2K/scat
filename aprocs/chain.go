package aprocs

import (
	ss "secsplit"
	"secsplit/concur"
)

type chain struct {
	procs  []Proc
	finish concur.Funcs
}

func NewChain(procs []Proc) Proc {
	return chain{
		procs:  procs,
		finish: finishFuncs(procs),
	}
}

func (chain chain) Process(c *ss.Chunk) <-chan Res {
	ch := make(chan Res)
	in := make(chan Res, 1)
	in <- Res{Chunk: c}
	close(in)
	var out chan Res
	for i, n := 0, len(chain.procs); i < n; i++ {
		proc := chain.procs[i]
		if i < n-1 {
			out = make(chan Res)
		} else {
			out = ch
		}
		go process(out, in, proc)
		in = out
	}
	return ch
}

func (chain chain) Finish() error {
	return chain.finish.FirstErr()
}

func process(out chan<- Res, in <-chan Res, proc Proc) {
	defer close(out)
	for res := range in {
		if res.Err != nil {
			out <- res
			continue
		}
		out <- <-proc.Process(res.Chunk)
	}
}

func finishFuncs(procs []Proc) (fns concur.Funcs) {
	fns = make(concur.Funcs, len(procs))
	for i, p := range procs {
		fns[i] = p.Finish
	}
	return
}
