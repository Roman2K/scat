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

func NewProc(proc procs.Proc, d *Statsd, id Id) procs.WrapperProc {
	return &counterProc{
		statsd: d,
		id:     id,
		proc:   proc,
	}
}

func (p *counterProc) Underlying() procs.Proc {
	return p.proc
}

func (p *counterProc) Process(c *scat.Chunk) <-chan procs.Res {
	ch := p.proc.Process(c)
	out := make(chan procs.Res)
	cnt := p.statsd.Counter(p.id)
	cnt.addInst(1)
	go func() {
		defer cnt.addInst(-1)
		defer close(out)
		for res := range ch {
			if c := res.Chunk; c != nil {
				if sizer, ok := c.Data().(scat.Sizer); ok {
					if sz := sizer.Size(); sz >= 0 {
						cnt.addOut(uint64(sz))
					}
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
