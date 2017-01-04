package cpprocs

import (
	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/concur"
	"secsplit/cpprocs/copies"
)

type multireader struct {
	reg     *copies.Reg
	readers []Reader
}

func NewMultiReader(readers []Reader) (proc aprocs.Proc, err error) {
	ml := make(MultiLister, len(readers))
	for i, rd := range readers {
		ml[i] = rd
	}
	reg := copies.NewReg()
	proc = multireader{
		reg:     reg,
		readers: readers,
	}
	err = ml.AddEntriesTo([]LsEntryAdder{CopiesEntryAdder{Reg: reg}})
	return
}

func (mrd multireader) Process(c *ss.Chunk) <-chan aprocs.Res {
	owners := mrd.reg.List(c.Hash).Owners()
	casc := make(aprocs.Cascade, len(owners))
	for i, o := range owners {
		casc[i] = aprocs.NewOnEnd(o.(aprocs.Proc), func(err error) {
			if err != nil {
				mrd.reg.RemoveOwner(o)
			}
		})
	}
	return casc.Process(c)
}

func (mrd multireader) Finish() error {
	return finishFuncs(mrd.readers).FirstErr()
}

func finishFuncs(readers []Reader) (fns concur.Funcs) {
	fns = make(concur.Funcs, len(readers))
	for i, rd := range readers {
		fns[i] = rd.Finish
	}
	return
}
