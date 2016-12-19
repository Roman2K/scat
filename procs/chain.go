package procs

import (
	"errors"
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

type errProcLoc struct {
	i   int
	err error
}

func (chain *chain) Process(c *ss.Chunk) Res {
	chunks := []*ss.Chunk{c}
	for i := range chain.procs {
		// TODO allocate len(chunks) * <max chunks output by this processor>
		out := make([]*ss.Chunk, 0, len(chunks))
		for _, c := range chunks {
			spawned, err := chain.process(c, i)
			if err != nil {
				return Res{Err: err}
			}
			out = append(out, spawned...)
		}
		chunks = out
	}
	for _, ender := range chain.enders {
		ender.end(c, chunks)
	}
	return Res{Chunks: chunks}
}

func (chain *chain) process(c *ss.Chunk, procIdx int) (
	spawned []*ss.Chunk, err error,
) {
	proc := chain.procs[procIdx]

	// Initial value for spawned is passing through c for error handling by next
	// Procs
	spawned = []*ss.Chunk{c}

	// Process() or ProcessErr()
	process := func() Res {
		return proc.Process(c)
	}
	if errLoc, ok := c.GetMeta("errProcLoc").(errProcLoc); ok {
		switch {
		case procIdx < errLoc.i:
			return
		case procIdx == errLoc.i:
			c.SetMeta("errProcLoc", nil)
			process = func() Res {
				return proc.(ErrProc).ProcessErr(c, errLoc.err)
			}
		default:
			err = errors.New("ErrProc skipped")
			return
		}
	}

	// Do the processing and prepare for error handling by another Proc
	res := process()
	if res.Err != nil {
		// findErrProc() could be cached but errors shouldn't be the normal
		// case. So, for the sake of code simplicity, call it for every error.
		j := findErrProc(chain.procs, procIdx+1)
		if j < 0 {
			err = res.Err
			return
		}
		c.SetMeta("errProcLoc", errProcLoc{i: j, err: res.Err})
		return
	}

	// If all went well, return chunks spawned by the Proc
	spawned = res.Chunks
	return
}

func findErrProc(procs []Proc, i int) int {
	for ; i < len(procs); i++ {
		if _, ok := procs[i].(ErrProc); ok {
			return i
		}
	}
	return -1
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
