package aprocs

import (
	ss "secsplit"
	"secsplit/concur"
)

type chain struct {
	procs  []Proc
	enders []EndProc
	finish concur.Funcs
}

func NewChain(procs []Proc) Proc {
	enders := []EndProc{}
	for _, p := range procs {
		if e, ok := p.(EndProc); ok {
			enders = append(enders, e)
		}
	}
	return chain{
		procs:  procs,
		enders: enders,
		finish: finishFuncs(procs),
	}
}

func (chain chain) Process(c *ss.Chunk) <-chan Res {
	procs := chain.procs
	if len(chain.enders) > 0 {
		endProc := InplaceProcFunc(func(final *ss.Chunk) (err error) {
			for _, ender := range chain.enders {
				err = ender.ProcessEnd(c, final)
				if err != nil {
					return
				}
			}
			return
		})
		newProcs := make([]Proc, len(procs)+1)
		copy(newProcs, procs)
		newProcs[len(newProcs)-1] = endProc
		procs = newProcs
	}
	ch := make(chan Res)
	in := make(chan Res, 1)
	in <- Res{Chunk: c}
	close(in)
	var out chan Res
	for i, n := 0, len(procs); i < n; i++ {
		proc := procs[i]
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
		for res := range proc.Process(res.Chunk) {
			out <- res
		}
	}
}

func finishFuncs(procs []Proc) (fns concur.Funcs) {
	fns = make(concur.Funcs, len(procs))
	for i, p := range procs {
		fns[i] = p.Finish
	}
	return
}
