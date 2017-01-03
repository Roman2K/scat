package copies

import (
	"sync"

	"secsplit/checksum"
	"secsplit/cpprocs"
)

type Reg struct {
	m  map[checksum.Hash]*List
	mu sync.Mutex
}

var _ cpprocs.CopierAdder = &Reg{}

func NewReg() *Reg {
	return &Reg{
		m: make(map[checksum.Hash]*List),
	}
}

func (r *Reg) AddCopier(cp cpprocs.Copier, entries []cpprocs.LsEntry) {
	for _, e := range entries {
		r.List(e.Hash).Add(cp)
	}
}

func (r *Reg) List(h checksum.Hash) *List {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.m[h]; !ok {
		r.m[h] = &List{
			m: make(map[cpprocs.CopierId]struct{}),
		}
	}
	return r.m[h]
}

type List struct {
	m     map[cpprocs.CopierId]struct{}
	mapMu sync.Mutex
	Mu    sync.Mutex
}

func (list *List) Add(c cpprocs.Copier) {
	list.mapMu.Lock()
	defer list.mapMu.Unlock()
	list.m[c.Id()] = struct{}{}
}

func (list *List) Contains(c cpprocs.Copier) (ok bool) {
	list.mapMu.Lock()
	defer list.mapMu.Unlock()
	_, ok = list.m[c.Id()]
	return
}

func (list *List) Len() int {
	list.mapMu.Lock()
	defer list.mapMu.Unlock()
	return len(list.m)
}
