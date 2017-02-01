package procs

import (
	"sync"

	"github.com/Roman2K/scat"
)

type Chain []Proc

var _ Proc = Chain{}

func (chain Chain) Process(c *scat.Chunk) <-chan Res {
	procs := chain
	enders := chain.endProcs()
	if len(enders) > 0 {
		ecp := endCallProc{chunk: c, enders: enders}
		newProcs := make([]Proc, len(procs)+1)
		copy(newProcs, procs)
		newProcs[len(newProcs)-1] = ecp
		procs = newProcs
	}
	in := make(chan Res, 1)
	in <- Res{Chunk: c}
	close(in)
	var out chan Res
	for _, proc := range procs {
		out = make(chan Res)
		go process(out, in, proc)
		in = out
	}
	return out
}

func (procs Chain) endProcs() (enders []EndProc) {
	for _, p := range procs {
		if e, ok := underlying(p).(EndProc); ok {
			enders = append(enders, e)
		}
	}
	return
}

func (procs Chain) Finish() error {
	return finishFuncs(procs).FirstErr()
}

func process(out chan<- Res, in <-chan Res, proc Proc) {
	defer close(out)
	wg := sync.WaitGroup{}
	for res := range in {
		var ch <-chan Res
		if res.Err != nil {
			if errp, ok := underlying(proc).(ErrProc); ok && res.Chunk != nil {
				ch = errp.ProcessErr(res.Chunk, res.Err)
			} else {
				out <- res
				continue
			}
		} else {
			ch = proc.Process(res.Chunk)
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			for res := range ch {
				out <- res
			}
		}()
	}
	wg.Wait()
	if ecp, ok := proc.(endCallProc); ok {
		err := ecp.processEnd()
		if err != nil {
			out <- Res{Err: err}
		}
	}
}

type endCallProc struct {
	chunk  *scat.Chunk
	enders []EndProc
}

func (ecp endCallProc) Process(c *scat.Chunk) <-chan Res {
	return InplaceFunc(ecp.process).Process(c)
}

func (ecp endCallProc) process(final *scat.Chunk) (err error) {
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
