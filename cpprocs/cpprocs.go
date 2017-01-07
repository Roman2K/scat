package cpprocs

import (
	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
	"secsplit/concur"
	"secsplit/cpprocs/copies"
	"secsplit/cpprocs/quota"
	"sync"
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
	aprocs.Proc
}

type copier struct {
	id   interface{}
	lser Lister
	proc aprocs.Proc
}

func NewCopier(id interface{}, lser Lister, proc aprocs.Proc) Copier {
	return &copier{id, lser, proc}
}

func (cp *copier) Id() interface{} {
	return cp.id
}

func (cp *copier) Ls() ([]LsEntry, error) {
	return cp.lser.Ls()
}

func (cp *copier) Process(c *ss.Chunk) <-chan aprocs.Res {
	return cp.proc.Process(c)
}

func (cp *copier) Finish() error {
	return cp.proc.Finish()
}

type Reader interface {
	Identified
	Lister
	aprocs.Proc
}

type reader struct {
	id   interface{}
	lser Lister
	proc aprocs.Proc
}

func NewReader(id interface{}, lser Lister, proc aprocs.Proc) Reader {
	return reader{id, lser, proc}
}

func (r reader) Id() interface{} {
	return r.id
}

func (r reader) Ls() ([]LsEntry, error) {
	return r.lser.Ls()
}

func (r reader) Process(c *ss.Chunk) <-chan aprocs.Res {
	return r.proc.Process(c)
}

func (r reader) Finish() error {
	return r.proc.Finish()
}

type LsProcUnprocer interface {
	Lister
	aprocs.ProcUnprocer
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
	Qman quota.Man
	mu   sync.Mutex
}

func (a *QuotaEntryAdder) AddLsEntry(lser Lister, e LsEntry) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.Qman.AddUse(lser.(quota.Res), uint64(e.Size))
}
