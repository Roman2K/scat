package cpprocs

import (
	"secsplit/aprocs"
	"secsplit/checksum"
)

type Lister interface {
	Ls() ([]checksum.Hash, error)
}

type Copier struct {
	Id     interface{}
	Lister Lister
	Proc   aprocs.Proc
}
