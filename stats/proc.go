package stats

import (
	"time"

	ss "secsplit"
	"secsplit/aprocs"
)

type logProc struct {
	log  *Log
	name string
	proc aprocs.Proc
}

func NewProc(log *Log, name string, proc aprocs.Proc) aprocs.Proc {
	return &logProc{
		log:  log,
		name: name,
		proc: proc,
	}
}

func (p *logProc) Underlying() aprocs.Proc {
	return p.proc
}

func (p *logProc) Process(c *ss.Chunk) <-chan aprocs.Res {
	out := make(chan aprocs.Res)
	counters := p.log.Counter(p.name)
	counters.addInstance()
	lastOut := time.Now()
	ch := p.proc.Process(c)
	go func() {
		defer counters.removeInstance()
		defer close(out)
		for res := range ch {
			now := time.Now()
			dur := now.Sub(lastOut)
			lastOut = now
			if c := res.Chunk; c != nil {
				counters.addOut(uint64(len(c.Data)), dur)
			}
			out <- res
		}
	}()
	return out
}

func (p *logProc) Finish() (err error) {
	err = p.proc.Finish()
	if err != nil {
		return
	}
	return p.log.Finish()
}
