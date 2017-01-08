package mincopies

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"

	"scat"
	"scat/aprocs"
	"scat/concur"
	"scat/cpprocs"
	"scat/cpprocs/copies"
	"scat/cpprocs/quota"
)

type minCopies struct {
	min    int
	qman   quota.Man
	qmanMu sync.Mutex
	reg    *copies.Reg
	finish func() error
}

func New(min int, qman quota.Man) (dynp aprocs.DynProcer, err error) {
	reg := copies.NewReg()
	ress := qman.Resources(0)
	ml := cpprocs.MultiLister(listers(ress))
	err = ml.AddEntriesTo([]cpprocs.LsEntryAdder{
		&cpprocs.QuotaEntryAdder{Qman: qman},
		cpprocs.CopiesEntryAdder{Reg: reg},
	})
	dynp = &minCopies{
		min:    min,
		qman:   qman,
		reg:    reg,
		finish: finishFuncs(ress).FirstErr,
	}
	return
}

func (mc *minCopies) Procs(c scat.Chunk) ([]aprocs.Proc, error) {
	copies := mc.reg.List(c.Hash())
	copies.Mu.Lock()
	dataUse := uint64(len(c.Data()))
	all := shuffle(mc.getCopiers(dataUse))
	ncopies := copies.Len()
	navail := len(all) - ncopies
	missing := mc.min - ncopies
	if missing < 0 {
		missing = 0
	}
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

func (mc *minCopies) getCopiers(use uint64) (cps []cpprocs.Copier) {
	mc.qmanMu.Lock()
	defer mc.qmanMu.Unlock()
	ress := mc.qman.Resources(use)
	cps = make([]cpprocs.Copier, len(ress))
	for i, res := range ress {
		cps[i] = res.(cpprocs.Copier)
	}
	return
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
	return mc.finish()
}

func listers(ress []quota.Res) (lsers []cpprocs.Lister) {
	lsers = make([]cpprocs.Lister, len(ress))
	for i, res := range ress {
		lsers[i] = res.(cpprocs.Lister)
	}
	return
}

func finishFuncs(ress []quota.Res) (fns concur.Funcs) {
	fns = make(concur.Funcs, len(ress))
	for i, res := range ress {
		fns[i] = res.(aprocs.Proc).Finish
	}
	return
}
