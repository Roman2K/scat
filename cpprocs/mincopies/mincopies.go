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
	min     int
	reg     *copies.Reg
	copiers []cpprocs.Copier
}

var rand2 = func() int {
	return rand.Intn(2)
}

func New(min int, copiers []cpprocs.Copier) (dynp aprocs.DynProcer, err error) {
	reg := copies.NewReg()
	err = reg.Add(copiers)
	dynp = minCopies{
		min:     min,
		reg:     reg,
		copiers: copiers,
	}
	return
}

func (mc minCopies) Procs(c *ss.Chunk) ([]aprocs.Proc, error) {
	copies := mc.reg.List(c.Hash)
	copies.Mu.Lock()
	ncopies := copies.UnlockedLen()
	avail := make([]cpprocs.Copier, 0, len(mc.copiers)-ncopies)
	for _, copier := range mc.copiers {
		if !copies.UnlockedContains(copier) {
			avail = append(avail, copier)
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
	copiers := avail[:missing]
	wg := sync.WaitGroup{}
	wg.Add(len(copiers))
	go func() {
		defer copies.Mu.Unlock()
		wg.Wait()
	}()
	procs := make([]aprocs.Proc, len(copiers)+1)
	procs[0] = aprocs.Nop
	copierProc := func(copier cpprocs.Copier) aprocs.Proc {
		proc := aprocs.NewDiscardChunks(copier.Proc)
		return aprocs.NewOnEnd(proc, func(err error) {
			defer wg.Done()
			if err == nil {
				copies.UnlockedAdd(copier)
			}
		})
	}
	for i, copier := range copiers {
		procs[i+1] = copierProc(copier)
	}
	return procs, nil
}

func (mc minCopies) Finish() error {
	return finishFuncs(mc.copiers).FirstErr()
}

func finishFuncs(copiers []cpprocs.Copier) (fns concur.Funcs) {
	fns = make(concur.Funcs, len(copiers))
	for i, c := range copiers {
		fns[i] = c.Proc.Finish
	}
	return
}
