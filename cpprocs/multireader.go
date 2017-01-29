package cpprocs

import (
	"fmt"
	"os"
	"scat"
	"scat/concur"
	"scat/cpprocs/copies"
	"scat/procs"
)

type multireader struct {
	reg     *copies.Reg
	copiers []Copier
}

func NewMultiReader(copiers []Copier) (proc procs.Proc, err error) {
	ml := make(MultiLister, len(copiers))
	for i, cp := range copiers {
		ml[i] = cp
	}
	reg := copies.NewReg()
	proc = multireader{
		reg:     reg,
		copiers: copiers,
	}
	err = ml.AddEntriesTo([]LsEntryAdder{CopiesEntryAdder{Reg: reg}})
	return
}

func (mrd multireader) Process(c *scat.Chunk) <-chan procs.Res {
	owners := mrd.reg.List(c.Hash()).Owners()
	copiers := make([]Copier, len(owners))
	for i, o := range owners {
		copiers[i] = o.(Copier)
	}
	ShuffleCopiers(copiers)
	casc := make(procs.Cascade, len(copiers))
	for i, cp := range copiers {
		casc[i] = procs.NewOnEnd(cp, func(err error) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "multireader: copier error: %v\n", err)
				mrd.reg.RemoveOwner(cp)
			}
		})
	}
	return casc.Process(c)
}

func (mrd multireader) Finish() error {
	return finishFuncs(mrd.copiers).FirstErr()
}

func finishFuncs(copiers []Copier) (fns concur.Funcs) {
	fns = make(concur.Funcs, len(copiers))
	for i, cp := range copiers {
		fns[i] = cp.Finish
	}
	return
}
