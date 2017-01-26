package quota

import "sync"

var Unlimited = ^uint64(0)

type Man struct {
	m     m
	mu    sync.RWMutex
	OnUse useCb
}

type m map[interface{}]*usage

type useCb func(res Res, use, max uint64)

type usage struct {
	res      Res
	use, max uint64
}

type Res interface {
	Id() interface{}
}

func NewMan() *Man {
	return &Man{m: make(m)}
}

func (man *Man) AddRes(res Res) {
	man.AddResQuota(res, Unlimited)
}

func (man *Man) AddResQuota(res Res, max uint64) {
	id := res.Id()
	man.mu.Lock()
	defer man.mu.Unlock()
	if _, ok := man.m[id]; !ok {
		man.m[id] = &usage{res: res}
	}
	man.m[id].max = max
}

func (man *Man) AddUse(res Res, use uint64) {
	u, ok := man.addUse(res, use)
	if ok && man.OnUse != nil {
		man.OnUse(res, u.use, u.max)
	}
}

func (man *Man) addUse(res Res, use uint64) (u *usage, ok bool) {
	id := res.Id()
	man.mu.Lock()
	defer man.mu.Unlock()
	u, ok = man.m[id]
	if !ok {
		return
	}
	u.use += use
	if u.use >= u.max {
		delete(man.m, res.Id())
	}
	return
}

func (man *Man) Delete(res Res) {
	man.mu.Lock()
	defer man.mu.Unlock()
	delete(man.m, res.Id())
}

func (man *Man) Resources(use uint64) (ress []Res) {
	ress = make([]Res, 0, len(man.m))
	man.mu.RLock()
	defer man.mu.RUnlock()
	for _, u := range man.m {
		if u.use+use >= u.max {
			continue
		}
		ress = append(ress, u.res)
	}
	return
}
