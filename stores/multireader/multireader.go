package multireader

import (
	"fmt"
	"os"

	"scat"
	"scat/concur"
	"scat/procs"
	"scat/stores"
	"scat/stores/copies"
)

type mrd struct {
	reg     *copies.Reg
	copiers []stores.Copier
}

func New(copiers []stores.Copier) (proc procs.Proc, err error) {
	ml := make(stores.MultiLister, len(copiers))
	for i, cp := range copiers {
		ml[i] = cp
	}
	reg := copies.NewReg()
	proc = mrd{
		reg:     reg,
		copiers: copiers,
	}
	err = ml.AddEntriesTo([]stores.LsEntryAdder{
		stores.CopiesEntryAdder{Reg: reg},
	})
	return
}

var shuffle = stores.ShuffleCopiers // var for tests

func (mrd mrd) Process(c *scat.Chunk) <-chan procs.Res {
	owners := mrd.reg.List(c.Hash()).Owners()
	copiers := make([]stores.Copier, len(owners))
	for i, o := range owners {
		copiers[i] = o.(stores.Copier)
	}
	copiers = shuffle(copiers)
	casc := make(procs.Cascade, len(copiers))
	for i, cp := range copiers {
		casc[i] = procs.NewOnEnd(cp, func(err error) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "multireader: copier error: %v\n", err)
				mrd.reg.RemoveOwner(cp)
			}
		})
	}
	if len(casc) == 0 {
		ch := make(chan procs.Res, 1)
		defer close(ch)
		ch <- procs.Res{Err: procs.ErrMissingData}
		return ch
	}
	return casc.Process(c)
}

func (mrd mrd) Finish() error {
	return finishFuncs(mrd.copiers).FirstErr()
}

func finishFuncs(copiers []stores.Copier) (fns concur.Funcs) {
	fns = make(concur.Funcs, len(copiers))
	for i, cp := range copiers {
		fns[i] = cp.Finish
	}
	return
}
