package procs

import (
	"sync"

	ss "secsplit"
)

type Chain []Proc

var _ ProcFinisher = Chain{}

func (chain Chain) Process(c *ss.Chunk) Res {
	chunks := []*ss.Chunk{c}
	for _, proc := range chain {
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
	return Res{Chunks: chunks}
}

func (chain Chain) Finish() (err error) {
	results := make(chan error)
	wg := sync.WaitGroup{}
	for _, proc := range chain {
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
