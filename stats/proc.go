package stats

import (
	"time"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/cpprocs"
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

type logLsProc struct {
	lister cpprocs.Lister
	proc   aprocs.Proc
}

func NewLsProc(log *Log, name string, lsp cpprocs.LsProc) cpprocs.LsProc {
	return logLsProc{
		lister: lsp,
		proc:   NewProc(log, name, lsp),
	}
}

func (lsp logLsProc) Ls() ([]cpprocs.LsEntry, error) {
	return lsp.lister.Ls()
}

func (lsp logLsProc) Process(c *ss.Chunk) <-chan aprocs.Res {
	return lsp.proc.Process(c)
}

func (lsp logLsProc) Finish() error {
	return lsp.proc.Finish()
}
