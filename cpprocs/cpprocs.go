package cpprocs

import (
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
	Id() CopierId
	Lister() Lister
	Proc() aprocs.Proc
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

func (c *copier) Id() CopierId      { return c.id }
func (c *copier) Lister() Lister    { return c.lister }
func (c *copier) Proc() aprocs.Proc { return c.proc }
func (c *copier) Quota() uint64     { return c.quota }
func (c *copier) SetQuota(q uint64) { c.quota = q }

type CopierAdder interface {
	AddCopier(Copier, []LsEntry)
}
