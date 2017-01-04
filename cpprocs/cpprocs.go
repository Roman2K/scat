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
	id     CopierId
	lister Lister
	proc   aprocs.Proc
	quota  uint64
}

type Copier interface {
	Lister
	aprocs.Proc
	Id() CopierId
	Quota() uint64
	SetQuota(uint64)
}

func NewCopier(id CopierId, lister Lister, proc aprocs.Proc) Copier {
	return &copier{
		id:     id,
		lister: lister,
		proc:   proc,
		quota:  QuotaUnlimited,
	}
}

func (cp *copier) Id() CopierId      { return cp.id }
func (cp *copier) Quota() uint64     { return cp.quota }
func (cp *copier) SetQuota(q uint64) { cp.quota = q }

func (cp *copier) Ls() ([]LsEntry, error) {
	return cp.lister.Ls()
}

func (cp *copier) Process(c *ss.Chunk) <-chan aprocs.Res {
	return cp.proc.Process(c)
}

func (cp *copier) Finish() error {
	return cp.proc.Finish()
}

type CopierAdder interface {
	AddCopier(Copier, []LsEntry)
}
