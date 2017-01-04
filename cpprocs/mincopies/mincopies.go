package mincopies

import (
	"errors"
	"fmt"
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
	copiers []cpprocs.Copier
	reg     *copies.Reg
	qman    cpprocs.QuotaMan
	qmanMu  sync.Mutex
}

func New(min int, copiers []cpprocs.Copier) (dynp aprocs.DynProcer, err error) {
	reg := copies.NewReg()
	qman := make(cpprocs.QuotaMan)
	err = addCopiers([]cpprocs.CopierAdder{
		reg,
		&lockedCopierAdder{adder: qman},
	}, copiers)
	dynp = &minCopies{
		min:     min,
		copiers: copiers,
		reg:     reg,
		qman:    qman,
	}
	return
}

type lockedCopierAdder struct {
	adder cpprocs.CopierAdder
	mu    sync.Mutex
}

func (a *lockedCopierAdder) AddCopier(
	cp cpprocs.Copier, entries []cpprocs.LsEntry,
) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.adder.AddCopier(cp, entries)
}

func addCopiers(adders []cpprocs.CopierAdder, copiers []cpprocs.Copier) error {
	fns := make(concur.Funcs, len(copiers))
	for i := range copiers {
		cp := copiers[i]
		fns[i] = func() (err error) {
			ls, err := cp.Ls()
			if err != nil {
				return
			}
			for _, a := range adders {
				a.AddCopier(cp, ls)
			}
			return
		}
	}
	return fns.FirstErr()
}

func (mc *minCopies) Procs(c *ss.Chunk) ([]aprocs.Proc, error) {
	copies := mc.reg.List(c.Hash)
	copies.Mu.Lock()
	dataUse := uint64(len(c.Data))
	all := shuffle(mc.getCopiers(dataUse))
	ncopies := copies.Len()
	missing := mc.min - ncopies
	navail := len(all) - ncopies
	if missing > navail {
		return nil, errors.New(fmt.Sprintf(
			"missing copiers to meet min requirement:"+
				" min=%d copies=%d missing=%d avail=%d",
			mc.min, ncopies, missing, navail,
		))
	}
	elected := make([]cpprocs.Copier, 0, missing)
	failover := make([]cpprocs.Copier, 0, navail-missing)
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
			casc[i] = aprocs.NewOnEnd(cp, func(err error) {
				if err != nil {
					mc.deleteCopier(cp)
					return
				}
				copies.Add(cp)
				mc.addUse(cp, dataUse)
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

func (mc *minCopies) getCopiers(use uint64) []cpprocs.Copier {
	mc.qmanMu.Lock()
	defer mc.qmanMu.Unlock()
	return mc.qman.Copiers(use)
}

func (mc *minCopies) deleteCopier(cp cpprocs.Copier) {
	mc.qmanMu.Lock()
	defer mc.qmanMu.Unlock()
	mc.qman.Delete(cp)
}

func (mc *minCopies) addUse(cp cpprocs.Copier, use uint64) {
	mc.qmanMu.Lock()
	defer mc.qmanMu.Unlock()
	mc.qman.AddUse(cp, use)
}

var shuffle = func(copiers []cpprocs.Copier) (res []cpprocs.Copier) {
	indexes := rand.Perm(len(copiers))
	res = make([]cpprocs.Copier, len(indexes))
	for i, idx := range indexes {
		res[i] = copiers[idx]
	}
	return
}

func (mc *minCopies) Finish() error {
	return finishFuncs(mc.copiers).FirstErr()
}

func finishFuncs(copiers []cpprocs.Copier) (fns concur.Funcs) {
	fns = make(concur.Funcs, len(copiers))
	for i, cp := range copiers {
		fns[i] = cp.Finish
	}
	return
}
