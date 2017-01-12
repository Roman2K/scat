package stats

import (
	"scat"
	"scat/aprocs"
)

type counterProc struct {
	statsd *Statsd
	id     Id
	proc   aprocs.Proc
}

func NewProc(d *Statsd, id Id, proc aprocs.Proc) aprocs.WrapperProc {
	return &counterProc{
		statsd: d,
		id:     id,
		proc:   proc,
	}
}

func (p *counterProc) Underlying() aprocs.Proc {
	return p.proc
}

func (p *counterProc) Process(c scat.Chunk) <-chan aprocs.Res {
	out := make(chan aprocs.Res)
	cnt := p.statsd.Counter(p.id)
	cnt.addInst(1)
	ch := p.proc.Process(c)
	go func() {
		defer cnt.addInst(-1)
		defer close(out)
		for res := range ch {
			if c := res.Chunk; c != nil {
				cnt.addOut(uint64(len(c.Data())))
			}
			out <- res
		}
	}()
	return out
}

func (p *counterProc) Finish() error {
	return p.proc.Finish()
}
