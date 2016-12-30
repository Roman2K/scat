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

func (r *Reg) Add(procs []cpprocs.Proc) error {
	fns := make(concur.Funcs, len(procs))
	addProcFunc := func(p cpprocs.Proc) func() error {
		return func() error { return r.addProc(p) }
	}
	for i, p := range procs {
		fns[i] = addProcFunc(p)
	}
	return fns.FirstErr()
}

func (r *Reg) addProc(p cpprocs.Proc) (err error) {
	hashes, err := p.Ls()
	if err != nil {
		return
	}
	for _, h := range hashes {
		r.List(h).Add(p)
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

func (list *List) Add(p cpprocs.Proc) {
	list.Mu.Lock()
	defer list.Mu.Unlock()
	list.UnlockedAdd(p)
}

func (list *List) UnlockedAdd(p cpprocs.Proc) {
	list.m[p.Id()] = struct{}{}
}

func (list *List) UnlockedContains(p cpprocs.Proc) (ok bool) {
	_, ok = list.m[p.Id()]
	return
}

func (list *List) UnlockedLen() int {
	return len(list.m)
}
