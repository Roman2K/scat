package quota

var Unlimited = ^uint64(0)

type Man struct {
	m     m
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

func NewMan() Man {
	return Man{m: make(m)}
}

func (man Man) AddRes(res Res) {
	man.AddResQuota(res, Unlimited)
}

func (man Man) AddResQuota(res Res, max uint64) {
	id := res.Id()
	if _, ok := man.m[id]; !ok {
		man.m[id] = &usage{res: res}
	}
	man.m[id].max = max
}

func (man Man) AddUse(res Res, use uint64) {
	id := res.Id()
	u, ok := man.m[id]
	if !ok {
		return
	}
	u.use += use
	if u.use >= u.max {
		delete(man.m, id)
	}
	if man.OnUse != nil {
		man.OnUse(res, u.use, u.max)
	}
}

func (man Man) Delete(res Res) {
	delete(man.m, res.Id())
}

func (man Man) Resources(use uint64) (ress []Res) {
	ress = make([]Res, 0, len(man.m))
	for _, u := range man.m {
		if u.use+use >= u.max {
			continue
		}
		ress = append(ress, u.res)
	}
	return
}
