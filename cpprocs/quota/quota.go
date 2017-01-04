package quota

var Unlimited = ^uint64(0)

type Man map[interface{}]*usage

type usage struct {
	res Res
	use uint64
}

type Res interface {
	Id() interface{}
	Quota() uint64
}

func (m Man) AddRes(res Res) {
	if _, ok := m[res.Id()]; !ok {
		m[res.Id()] = &usage{res: res}
	}
}

func (m Man) AddUse(res Res, use uint64) {
	id := res.Id()
	u, ok := m[id]
	if !ok {
		return
	}
	u.use += use
	if u.use >= u.res.Quota() {
		delete(m, id)
	}
}

func (m Man) Delete(res Res) {
	delete(m, res.Id())
}

func (m Man) Resources(use uint64) (ress []Res) {
	ress = make([]Res, 0, len(m))
	for _, u := range m {
		if u.use+uint64(use) >= u.res.Quota() {
			continue
		}
		ress = append(ress, u.res)
	}
	return
}
