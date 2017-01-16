package stats

import (
	"scat"
	"scat/procs"
)

type counterProc struct {
	statsd *Statsd
	id     Id
	proc   procs.Proc
}

func NewProc(d *Statsd, id Id, proc procs.Proc) procs.WrapperProc {
	return &counterProc{
		statsd: d,
		id:     id,
		proc:   proc,
	}
}

func (p *counterProc) Underlying() procs.Proc {
	return p.proc
}

func (p *counterProc) Process(c scat.Chunk) <-chan procs.Res {
	out := make(chan procs.Res)
	cnt := p.statsd.Counter(p.id)
	cnt.addInst(1)
	ch := p.proc.Process(c)
	go func() {
		defer cnt.addInst(-1)
		defer close(out)
		for res := range ch {
			if c := res.Chunk; c != nil {
				if sz, ok := c.Data().(scat.Sizer); ok {
					cnt.addOut(uint64(sz.Size()))
				}
			}
			out <- res
		}
	}()
	return out
}

func (p *counterProc) Finish() error {
	return p.proc.Finish()
}
