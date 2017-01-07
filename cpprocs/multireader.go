package cpprocs

import (
	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/concur"
	"secsplit/cpprocs/copies"
)

type multireader struct {
	reg     *copies.Reg
	copiers []Copier
}

func NewMultiReader(copiers []Copier) (proc aprocs.Proc, err error) {
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

func (mrd multireader) Process(c *ss.Chunk) <-chan aprocs.Res {
	owners := mrd.reg.List(c.Hash).Owners()
	casc := make(aprocs.Cascade, len(owners))
	for i, o := range owners {
		proc := o.(aprocs.Proc)
		casc[i] = aprocs.NewOnEnd(proc, func(err error) {
			if err != nil {
				mrd.reg.RemoveOwner(o)
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
