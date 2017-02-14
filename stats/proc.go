package stats

import (
	"gitlab.com/Roman2K/scat"
	"gitlab.com/Roman2K/scat/procs"
)

type Proc struct {
	D  *Statsd
	Id id
	procs.Proc
}

var _ procs.WrapperProc = Proc{}

func (p Proc) Underlying() procs.Proc {
	return p.Proc
}

func (p Proc) Process(c *scat.Chunk) <-chan procs.Res {
	ch := p.Proc.Process(c)
	out := make(chan procs.Res)
	cnt := p.D.Counter(p.Id)
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
