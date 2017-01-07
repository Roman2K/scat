package stats

import (
	"time"

	ss "secsplit"
	"secsplit/aprocs"
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

func (p *counterProc) Process(c *ss.Chunk) <-chan aprocs.Res {
	out := make(chan aprocs.Res)
	cnt := p.statsd.Counter(p.id)
	cnt.addInst(1)
	lastOut := time.Now()
	ch := p.proc.Process(c)
	go func() {
		defer cnt.addInst(-1)
		defer close(out)
		for res := range ch {
			now := time.Now()
			dur := now.Sub(lastOut)
			lastOut = now
			if c := res.Chunk; c != nil {
				cnt.addOut(uint64(len(c.Data)), dur)
			}
			out <- res
		}
	}()
	return out
}

func (p *counterProc) Finish() error {
	return p.proc.Finish()
}
