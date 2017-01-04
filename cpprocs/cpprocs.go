package cpprocs

import (
	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
	"secsplit/concur"
	"secsplit/cpprocs/copies"
	"secsplit/cpprocs/quota"
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

type copier struct {
	id    interface{}
	quota uint64
	lsp   LsProc
}

type Copier interface {
	Identified
	Quota() uint64
	SetQuota(uint64)
	LsProc
}

func NewCopier(id interface{}, lsp LsProc) Copier {
	return &copier{
		id:    id,
		lsp:   lsp,
		quota: quota.Unlimited,
	}
}

func (cp *copier) Id() interface{}   { return cp.id }
func (cp *copier) Quota() uint64     { return cp.quota }
func (cp *copier) SetQuota(q uint64) { cp.quota = q }

func (cp *copier) Ls() ([]LsEntry, error) {
	return cp.lsp.Ls()
}

func (cp *copier) Process(c *ss.Chunk) <-chan aprocs.Res {
	return cp.lsp.Process(c)
}

func (cp *copier) Finish() error {
	return cp.lsp.Finish()
}

type LsProc interface {
	Lister
	aprocs.Proc
}

type lsProc struct {
	lister Lister
	proc   aprocs.Proc
}

func NewLsProc(lister Lister, proc aprocs.Proc) LsProc {
	return lsProc{lister: lister, proc: proc}
}

func (lsp lsProc) Process(c *ss.Chunk) <-chan aprocs.Res {
	return lsp.proc.Process(c)
}

func (lsp lsProc) Finish() error {
	return lsp.proc.Finish()
}

func (lsp lsProc) Ls() ([]LsEntry, error) {
	return lsp.lister.Ls()
}

type Reader interface {
	Identified
	LsProc
}

type reader struct {
	id  interface{}
	lsp LsProc
}

func NewReader(id interface{}, lsp LsProc) Reader {
	return reader{id: id, lsp: lsp}
}

func (r reader) Id() interface{} {
	return r.id
}

func (r reader) Ls() ([]LsEntry, error) {
	return r.lsp.Ls()
}

func (r reader) Process(c *ss.Chunk) <-chan aprocs.Res {
	return r.lsp.Process(c)
}

func (r reader) Finish() error {
	return r.lsp.Finish()
}

type LsProcer interface {
	LsProc() LsProc
}

type LsUnprocer interface {
	LsUnproc() LsProc
}

type LsProcUnprocer interface {
	LsProcer
	LsUnprocer
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
