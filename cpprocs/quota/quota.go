package quota

var Unlimited = ^uint64(0)

type man map[interface{}]*usage

type Res interface {
	Id() interface{}
}

type Man interface {
	AddRes(Res)
	AddResQuota(Res, uint64)
	AddUse(Res, uint64)
	Delete(Res)
	Resources(uint64) []Res
}

func NewMan() Man {
	return man{}
}

type usage struct {
	res      Res
	max, use uint64
}

func (m man) AddRes(res Res) {
	m.AddResQuota(res, Unlimited)
}

func (m man) AddResQuota(res Res, max uint64) {
	id := res.Id()
	if _, ok := m[id]; !ok {
		m[id] = &usage{res: res}
	}
	m[id].max = max
}

func (m man) AddUse(res Res, use uint64) {
	id := res.Id()
	u, ok := m[id]
	if !ok {
		return
	}
	u.use += use
	if u.use >= u.max {
		delete(m, id)
	}
}

func (m man) Delete(res Res) {
	delete(m, res.Id())
}

func (m man) Resources(use uint64) (ress []Res) {
	ress = make([]Res, 0, len(m))
	for _, u := range m {
		if u.use+use >= u.max {
			continue
		}
		ress = append(ress, u.res)
	}
	return
}
