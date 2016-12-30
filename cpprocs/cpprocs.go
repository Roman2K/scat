package cpprocs

import (
	"secsplit/aprocs"
	"secsplit/checksum"
)

type Proc interface {
	Id() interface{}
	Ls() ([]checksum.Hash, error)
	aprocs.Proc
}
