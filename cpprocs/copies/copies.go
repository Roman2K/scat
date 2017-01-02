package copies

import (
	"sync"

	"secsplit/checksum"
	"secsplit/concur"
	"secsplit/cpprocs"
)

type Reg struct {
	m  map[checksum.Hash]*List
	mu sync.Mutex
}

func NewReg() *Reg {
	return &Reg{
		m: make(map[checksum.Hash]*List),
	}
}

func (r *Reg) Add(copiers []cpprocs.Copier) error {
	fns := make(concur.Funcs, len(copiers))
	addCopierFunc := func(c cpprocs.Copier) func() error {
		return func() error { return r.addCopier(c) }
	}
	for i, c := range copiers {
		fns[i] = addCopierFunc(c)
	}
	return fns.FirstErr()
}

func (r *Reg) addCopier(c cpprocs.Copier) (err error) {
	hashes, err := c.Lister.Ls()
	if err != nil {
		return
	}
	for _, h := range hashes {
		r.List(h).Add(c)
	}
	return nil
}

func (r *Reg) List(h checksum.Hash) *List {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.m[h]; !ok {
		r.m[h] = &List{
			m: make(map[interface{}]struct{}),
		}
	}
	return r.m[h]
}

type List struct {
	m  map[interface{}]struct{}
	Mu sync.Mutex
}

func (list *List) Add(c cpprocs.Copier) {
	list.Mu.Lock()
	defer list.Mu.Unlock()
	list.UnlockedAdd(c)
}

func (list *List) UnlockedAdd(c cpprocs.Copier) {
	list.m[c.Id] = struct{}{}
}

func (list *List) UnlockedContains(c cpprocs.Copier) (ok bool) {
	_, ok = list.m[c.Id]
	return
}

func (list *List) UnlockedLen() int {
	return len(list.m)
}
