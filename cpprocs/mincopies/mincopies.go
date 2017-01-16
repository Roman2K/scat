package mincopies

import (
	"errors"
	"fmt"
	"sync"

	"scat"
	"scat/concur"
	"scat/cpprocs"
	"scat/cpprocs/copies"
	"scat/cpprocs/quota"
	"scat/procs"
)

type minCopies struct {
	min    int
	qman   quota.Man
	qmanMu sync.Mutex
	reg    *copies.Reg
	finish func() error
}

func New(min int, qman quota.Man) (dynp procs.DynProcer, err error) {
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

func calcDataUse(d scat.Data) (uint64, error) {
	sz, ok := d.(scat.Sizer)
	if !ok {
		return 0, errors.New("sized-data required for calculating data use")
	}
	return uint64(sz.Size()), nil
}

func (mc *minCopies) Procs(c scat.Chunk) ([]procs.Proc, error) {
	copies := mc.reg.List(c.Hash())
	copies.Mu.Lock()
	dataUse, err := calcDataUse(c.Data())
	if err != nil {
		return nil, err
	}
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
	cpProcs := make([]procs.Proc, len(elected)+1)
	cpProcs[0] = procs.Nop
	copierProc := func(copier cpprocs.Copier) procs.Proc {
		copiers := append([]cpprocs.Copier{copier}, failover...)
		casc := make(procs.Cascade, len(copiers))
		for i := range copiers {
			cp := copiers[i]
			casc[i] = procs.NewOnEnd(cp, func(err error) {
				if err != nil {
					mc.deleteCopier(cp)
					return
				}
				copies.Add(cp)
				mc.addUse(cp, dataUse)
			})
		}
		proc := procs.NewDiscardChunks(casc)
		return procs.NewOnEnd(proc, func(error) { wg.Done() })
	}
	for i, copier := range elected {
		cpProcs[i+1] = copierProc(copier)
	}
	return cpProcs, nil
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

var shuffle = cpprocs.ShuffleCopiers

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
		fns[i] = res.(procs.Proc).Finish
	}
	return
}
