package copies

import (
	"sync"

	"secsplit/checksum"
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
	m     map[interface{}]struct{}
	mapMu sync.Mutex
	Mu    sync.Mutex
}

type Owner interface {
	Id() interface{}
}

func (list *List) Add(o Owner) {
	list.mapMu.Lock()
	defer list.mapMu.Unlock()
	list.m[o.Id()] = struct{}{}
}

func (list *List) Contains(o Owner) (ok bool) {
	list.mapMu.Lock()
	defer list.mapMu.Unlock()
	_, ok = list.m[o.Id()]
	return
}

func (list *List) Len() int {
	list.mapMu.Lock()
	defer list.mapMu.Unlock()
	return len(list.m)
}
