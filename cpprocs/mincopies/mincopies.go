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
	min       int
	reg       *copies.Reg
	copiers   []cpprocs.Copier
	copiersMu sync.Mutex
}

func New(min int, copiers []cpprocs.Copier) (dynp aprocs.DynProcer, err error) {
	reg := copies.NewReg()
	err = reg.Add(copiers)
	dynp = &minCopies{
		min:     min,
		reg:     reg,
		copiers: copiers,
	}
	return
}

func (mc *minCopies) Procs(c *ss.Chunk) ([]aprocs.Proc, error) {
	copies := mc.reg.List(c.Hash)
	copies.Mu.Lock()
	ncopies := copies.Len()
	missing := mc.min - ncopies
	all := shuffle(mc.getCopiers())
	elected := make([]cpprocs.Copier, 0, missing)
	foCap := len(all) - ncopies - missing
	if foCap < 0 {
		foCap = 0
	}
	failover := make([]cpprocs.Copier, 0, foCap)
	for i, n := 0, len(all); i < n; i++ {
		cp := all[i]
		if copies.Contains(cp) {
			continue
		}
		if len(elected) < missing {
			elected = append(elected, cp)
		} else {
			failover = append(failover, cp)
		}
	}
	wg := sync.WaitGroup{}
	wg.Add(len(elected))
	go func() {
		defer copies.Mu.Unlock()
		wg.Wait()
	}()
	procs := make([]aprocs.Proc, len(elected)+1)
	procs[0] = aprocs.Nop
	copierProc := func(copier cpprocs.Copier) aprocs.Proc {
		copiers := append([]cpprocs.Copier{copier}, failover...)
		casc := make(aprocs.Cascade, len(copiers))
		for i := range copiers {
			cp := copiers[i]
			casc[i] = aprocs.NewOnEnd(cp.Proc, func(err error) {
				if err != nil {
					mc.deleteCopier(cp)
					return
				}
				copies.Add(cp)
			})
		}
		proc := aprocs.NewDiscardChunks(casc)
		return aprocs.NewOnEnd(proc, func(error) { wg.Done() })
	}
	for i, copier := range elected {
		procs[i+1] = copierProc(copier)
	}
	return procs, nil
}

var shuffle = func(copiers []cpprocs.Copier) (res []cpprocs.Copier) {
	indexes := rand.Perm(len(copiers))
	res = make([]cpprocs.Copier, len(indexes))
	for i, idx := range indexes {
		res[i] = copiers[idx]
	}
	return
}

func (mc *minCopies) getCopiers() (cps []cpprocs.Copier) {
	mc.copiersMu.Lock()
	defer mc.copiersMu.Unlock()
	cps = make([]cpprocs.Copier, len(mc.copiers))
	copy(cps, mc.copiers)
	return
}

func (mc *minCopies) deleteCopier(dcp cpprocs.Copier) {
	mc.copiersMu.Lock()
	defer mc.copiersMu.Unlock()
	cps := make([]cpprocs.Copier, 0, len(mc.copiers)-1)
	for _, cp := range mc.copiers {
		if cp.Id != dcp.Id {
			cps = append(cps, cp)
		}
	}
	mc.copiers = cps
}

func (mc *minCopies) Finish() error {
	return finishFuncs(mc.copiers).FirstErr()
}

func finishFuncs(copiers []cpprocs.Copier) (fns concur.Funcs) {
	fns = make(concur.Funcs, len(copiers))
	for i, c := range copiers {
		fns[i] = c.Proc.Finish
	}
	return
}
