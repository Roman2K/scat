package mincopies

import (
	"math/rand"
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
	ncopies := copies.Len()
	all := make([]cpprocs.Copier, len(mc.copiers))
	copy(all, mc.copiers)
	sortCopiers(all)
	missing := mc.min - ncopies
	copiers := make([]cpprocs.Copier, 0, missing)
	for i, n := 0, len(all); i < n && i < missing; i++ {
		if copier := all[i]; !copies.Contains(copier) {
			copiers = append(copiers, copier)
		}
	}
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
				copies.Add(copier)
			}
		})
	}
	for i, copier := range copiers {
		procs[i+1] = copierProc(copier)
	}
	return procs, nil
}

var sortCopiers = func(copiers []cpprocs.Copier) {
	indexes := rand.Perm(len(copiers))
	dup := make([]cpprocs.Copier, len(indexes))
	for i, idx := range indexes {
		dup[i] = copiers[idx]
	}
	for i, c := range dup {
		copiers[i] = c
	}
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
