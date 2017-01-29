package cpprocs

import (
	"math/rand"

	"scat"
	"scat/checksum"
	"scat/concur"
	"scat/cpprocs/copies"
	"scat/cpprocs/quota"
	"scat/procs"
)

type Lister interface {
	Ls() ([]LsEntry, error)
}

type LsEntry struct {
	Hash checksum.Hash
	Size int64
}

type Identified interface {
	Id() interface{}
}

type Copier interface {
	Identified
	Lister
	procs.Proc
}

type copier struct {
	id   interface{}
	lser Lister
	proc procs.Proc
}

func NewCopier(id interface{}, lser Lister, proc procs.Proc) Copier {
	return &copier{id, lser, proc}
}

func (cp *copier) Id() interface{} {
	return cp.id
}

func (cp *copier) Ls() ([]LsEntry, error) {
	return cp.lser.Ls()
}

func (cp *copier) Process(c *scat.Chunk) <-chan procs.Res {
	return cp.proc.Process(c)
}

func (cp *copier) Finish() error {
	return cp.proc.Finish()
}

type LsProcUnprocer interface {
	Lister
	procs.ProcUnprocer
}

type LsEntryAdder interface {
	AddLsEntry(Lister, LsEntry)
}

type MultiLister []Lister

func (ml MultiLister) AddEntriesTo(adders []LsEntryAdder) error {
	fns := make(concur.Funcs, len(ml))
	for i := range ml {
		lser := ml[i]
		fns[i] = func() (err error) {
			ls, err := lser.Ls()
			if err != nil {
				return
			}
			for _, a := range adders {
				for _, e := range ls {
					a.AddLsEntry(lser, e)
				}
			}
			return
		}
	}
	return fns.FirstErr()
}

type CopiesEntryAdder struct {
	Reg *copies.Reg
}

func (a CopiesEntryAdder) AddLsEntry(lser Lister, e LsEntry) {
	owner := lser.(copies.Owner)
	a.Reg.List(e.Hash).Add(owner)
}

type QuotaEntryAdder struct {
	Qman *quota.Man
}

func (a QuotaEntryAdder) AddLsEntry(lser Lister, e LsEntry) {
	a.Qman.AddUse(lser.(quota.Res), uint64(e.Size))
}

func ShuffleCopiers(copiers []Copier) (res []Copier) {
	indexes := rand.Perm(len(copiers))
	res = make([]Copier, len(indexes))
	for i, idx := range indexes {
		res[i] = copiers[idx]
	}
	return
}
