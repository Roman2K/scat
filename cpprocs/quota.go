package cpprocs

import (
	"sync"
)

var QuotaUnlimited = ^uint64(0)

type QuotaMan struct {
	usage   quotaUsage
	usageMu sync.Mutex
}

type quotaUsage map[CopierId]*usage

type usage struct {
	copier  Copier
	used    uint64
	deleted bool
}

var _ CopierAdder = &QuotaMan{}

func NewQuotaMan() *QuotaMan {
	return &QuotaMan{usage: make(quotaUsage)}
}

func (m *QuotaMan) AddCopier(cp Copier, entries []LsEntry) {
	used := uint64(0)
	for _, e := range entries {
		used += uint64(e.Size)
	}
	m.addUsed(cp, used)
}

func (m *QuotaMan) addUsed(cp Copier, used uint64) {
	m.usageMu.Lock()
	defer m.usageMu.Unlock()
	cpid := cp.Id()
	if _, ok := m.usage[cpid]; !ok {
		m.usage[cpid] = &usage{copier: cp}
	}
	u := m.usage[cpid]
	u.used += used
	if u.used >= u.copier.Quota() {
		u.deleted = true
	}
}

func (m *QuotaMan) Delete(cp Copier) {
	m.usageMu.Lock()
	defer m.usageMu.Unlock()
	u, ok := m.usage[cp.Id()]
	if !ok {
		return
	}
	u.deleted = true
}

func (m *QuotaMan) Copiers(use int64) (res []Copier) {
	m.usageMu.Lock()
	defer m.usageMu.Unlock()
	res = make([]Copier, 0, len(m.usage))
	for _, u := range m.usage {
		if u.deleted {
			continue
		}
		if u.used+uint64(use) >= u.copier.Quota() {
			continue
		}
		res = append(res, u.copier)
	}
	return
}
