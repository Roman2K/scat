package cpprocs

import (
	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/checksum"
)

type Lister interface {
	Ls() ([]LsEntry, error)
}

type LsEntry struct {
	Hash checksum.Hash
	Size int64
}

type CopierId interface{}

type copier struct {
	id    CopierId
	quota uint64
	lsp   LsProc
}

type Copier interface {
	Id() CopierId
	Quota() uint64
	SetQuota(uint64)
	LsProc
}

func NewCopier(id CopierId, lsp LsProc) Copier {
	return &copier{
		id:    id,
		lsp:   lsp,
		quota: QuotaUnlimited,
	}
}

func (cp *copier) Id() CopierId      { return cp.id }
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

type CopierAdder interface {
	AddCopier(Copier, []LsEntry)
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
