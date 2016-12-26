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
		ecp := endCallProc{chunk: c, enders: chain.enders}
		newProcs := make([]Proc, len(procs)+1)
		copy(newProcs, procs)
		newProcs[len(newProcs)-1] = ecp
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
	if ecp, ok := proc.(endCallProc); ok {
		err := ecp.processEnd()
		if err != nil {
			out <- Res{Err: err}
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

type endCallProc struct {
	chunk  *ss.Chunk
	enders []EndProc
}

func (ecp endCallProc) Process(c *ss.Chunk) <-chan Res {
	return InplaceProcFunc(ecp.process).Process(c)
}

func (ecp endCallProc) process(final *ss.Chunk) (err error) {
	for _, ender := range ecp.enders {
		err = ender.ProcessFinal(ecp.chunk, final)
		if err != nil {
			return
		}
	}
	return
}

func (ecp endCallProc) processEnd() (err error) {
	for _, ender := range ecp.enders {
		err = ender.ProcessEnd(ecp.chunk)
		if err != nil {
			return
		}
	}
	return
}

func (ecp endCallProc) Finish() error {
	return nil
}
