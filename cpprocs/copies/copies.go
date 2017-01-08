package copies

import (
	"sync"

	"scat/checksum"
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
			m: make(map[interface{}]Owner),
		}
	}
	return r.m[h]
}

func (r *Reg) RemoveOwner(o Owner) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, list := range r.m {
		list.Remove(o)
	}
}

type List struct {
	m     map[interface{}]Owner
	mapMu sync.Mutex
	Mu    sync.Mutex
}

type Owner interface {
	Id() interface{}
}

func (list *List) Add(o Owner) {
	list.mapMu.Lock()
	defer list.mapMu.Unlock()
	list.m[o.Id()] = o
}

func (list *List) Remove(o Owner) {
	list.mapMu.Lock()
	defer list.mapMu.Unlock()
	delete(list.m, o.Id())
}

func (list *List) Contains(o Owner) (ok bool) {
	list.mapMu.Lock()
	defer list.mapMu.Unlock()
	_, ok = list.m[o.Id()]
	return
}

func (list *List) Owners() (owners []Owner) {
	list.mapMu.Lock()
	defer list.mapMu.Unlock()
	owners = make([]Owner, 0, len(list.m))
	for _, o := range list.m {
		owners = append(owners, o)
	}
	return
}

func (list *List) Len() int {
	list.mapMu.Lock()
	defer list.mapMu.Unlock()
	return len(list.m)
}
