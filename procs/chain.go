package procs

import (
	"errors"
	"sync"

	ss "secsplit"
)

type chain struct {
	procs  []Proc
	enders []EndProc
}

func NewChain(procs []Proc) chain {
	chain := chain{procs: procs}
	for _, proc := range procs {
		if ender, ok := proc.(EndProc); ok {
			chain.enders = append(chain.enders, ender)
		}
	}
	return chain
}

type errProcLoc struct {
	i   int
	err error
}

func (chain chain) Process(c *ss.Chunk) Res {
	chunks, err := chain.process(c)
	return Res{Chunks: chunks, Err: err}
}

func (chain chain) process(c *ss.Chunk) (chunks []*ss.Chunk, err error) {
	chunks = []*ss.Chunk{c}
	for i := range chain.procs {
		chunks, err = chain.processAt(i, chunks)
		if err != nil {
			return
		}
	}
	// TODO parallel
	for _, ender := range chain.enders {
		err = ender.ProcessEnd(c, chunks)
		if err != nil {
			return
		}
	}
	return
}

func (chain chain) processAt(procIdx int, chunks []*ss.Chunk) (
	[]*ss.Chunk, error,
) {
	// TODO allocate len(chunks) * <max chunks output by this processor>
	out := make([]*ss.Chunk, 0, len(chunks))
	// TODO parallel
	for _, c := range chunks {
		spawned, err := chain.processChunkAt(procIdx, c)
		if err != nil {
			return nil, err
		}
		out = append(out, spawned...)
	}
	return out, nil
}

func (chain chain) processChunkAt(procIdx int, c *ss.Chunk) (
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

func (chain chain) Finish() (err error) {
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
