package mincopies

import (
	"math/rand"
	"sort"
	"sync"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/concur"
	"secsplit/cpprocs"
	"secsplit/cpprocs/copies"
)

type minCopies struct {
	min   int
	reg   *copies.Reg
	procs []cpprocs.Proc
}

var rand2 = func() int {
	return rand.Intn(2)
}

func New(min int, procs []cpprocs.Proc) (dynp aprocs.DynProcer, err error) {
	reg := copies.NewReg()
	err = reg.Add(procs)
	dynp = minCopies{
		min:   min,
		reg:   reg,
		procs: procs,
	}
	return
}

func (mc minCopies) Procs(c *ss.Chunk) ([]aprocs.Proc, error) {
	copies := mc.reg.List(c.Hash)
	copies.Mu.Lock()
	ncopies := copies.UnlockedLen()
	avail := make([]cpprocs.Proc, 0, len(mc.procs)-ncopies)
	for _, p := range mc.procs {
		if !copies.UnlockedContains(p) {
			avail = append(avail, p)
		}
	}
	missing := mc.min - ncopies
	if navail := len(avail); missing > navail {
		missing = navail
	}
	if missing < 0 {
		missing = 0
	}
	sort.Slice(avail, func(_, _ int) bool {
		return rand2() == 0
	})
	cpProcs := avail[:missing]
	wg := sync.WaitGroup{}
	wg.Add(len(cpProcs))
	go func() {
		defer copies.Mu.Unlock()
		wg.Wait()
	}()
	procs := make([]aprocs.Proc, len(cpProcs)+1)
	procs[0] = aprocs.Nop
	for i, p := range cpProcs {
		procs[i+1] = aprocs.NewOnEnd(aprocs.NewDiscardChunks(p), func(err error) {
			defer wg.Done()
			if err == nil {
				copies.UnlockedAdd(p)
			}
		})
	}
	return procs, nil
}

func (mc minCopies) Finish() error {
	return finishFuncs(mc.procs).FirstErr()
}

func finishFuncs(procs []cpprocs.Proc) (fns concur.Funcs) {
	fns = make(concur.Funcs, len(procs))
	for i, p := range procs {
		fns[i] = p.Finish
	}
	return
}
