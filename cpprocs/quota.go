package cpprocs

var QuotaUnlimited = ^uint64(0)

type QuotaMan map[key]*usage

// Private type to prevent external adds like man[id] = nil
type key CopierId

type usage struct {
	copier Copier
	use    uint64
}

var _ CopierAdder = QuotaMan{}

func (m QuotaMan) AddCopier(cp Copier, entries []LsEntry) {
	use := uint64(0)
	for _, e := range entries {
		use += uint64(e.Size)
	}
	if _, ok := m[cp.Id()]; !ok {
		m[cp.Id()] = &usage{copier: cp}
	}
	m.AddUse(cp, use)
}

func (m QuotaMan) AddUse(cp Copier, use uint64) {
	u, ok := m[cp.Id()]
	if !ok {
		return
	}
	u.use += use
	if u.use >= u.copier.Quota() {
		delete(m, cp.Id())
	}
}

func (m QuotaMan) Delete(cp Copier) {
	delete(m, cp.Id())
}

func (m QuotaMan) Copiers(use uint64) (res []Copier) {
	res = make([]Copier, 0, len(m))
	for _, u := range m {
		if u.use+uint64(use) >= u.copier.Quota() {
			continue
		}
		res = append(res, u.copier)
	}
	return
}
