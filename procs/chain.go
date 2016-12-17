package procs

import (
	"sync"

	ss "secsplit"
)

type chain struct {
	procs  []Proc
	enders []ender
}

func NewChain(procs []Proc) ProcFinisher {
	chain := &chain{procs: procs}
	for _, proc := range procs {
		if ender, ok := proc.(ender); ok {
			chain.enders = append(chain.enders, ender)
		}
	}
	return chain
}

func (chain *chain) Process(c *ss.Chunk) Res {
	chunks := []*ss.Chunk{c}
	for _, proc := range chain.procs {
		// TODO allocate len(chunks) * <max chunks output by this processor>
		out := make([]*ss.Chunk, 0, len(chunks))
		for _, c := range chunks {
			res := proc.Process(c)
			if res.Err != nil {
				return res
			}
			out = append(out, res.Chunks...)
		}
		chunks = out
	}
	for _, ender := range chain.enders {
		ender.end(c, chunks)
	}
	return Res{Chunks: chunks}
}

func (chain *chain) Finish() (err error) {
	results := make(chan error)
	wg := sync.WaitGroup{}
	for _, proc := range chain.procs {
		f, ok := proc.(Finisher)
		if !ok {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- f.Finish()
		}()
	}
	go func() {
		defer close(results)
		wg.Wait()
	}()
	for e := range results {
		if e != nil && err == nil {
			err = e
		}
	}
	return
}
