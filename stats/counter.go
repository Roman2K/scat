package stats

import (
	"sync"
	"time"
)

type Counter struct {
	pos                uint32
	inst               uint32
	since0             time.Time
	last               time.Time
	instMu             sync.Mutex
	out                uint64
	dur                time.Duration
	outMu              sync.Mutex
	QuotaUse, QuotaMax uint64
}

func (cnt *Counter) addOut(nbytes uint64) {
	cnt.outMu.Lock()
	defer cnt.outMu.Unlock()
	cnt.out += nbytes
}

func (cnt *Counter) getOut() (uint64, time.Duration) {
	cnt.outMu.Lock()
	cnt.instMu.Lock()
	ninst := cnt.inst
	since0 := cnt.since0
	out := cnt.out
	dur := cnt.dur
	cnt.instMu.Unlock()
	cnt.outMu.Unlock()
	now := time.Now()
	if ninst > 0 {
		dur += now.Sub(since0)
	}
	return out, dur
}

func (cnt *Counter) getInst() uint32 {
	cnt.instMu.Lock()
	defer cnt.instMu.Unlock()
	return cnt.inst
}

func (cnt *Counter) addInst(delta int) {
	cnt.instMu.Lock()
	defer cnt.instMu.Unlock()
	now := time.Now()
	was0 := cnt.inst == 0
	if delta > 0 {
		cnt.inst += uint32(delta)
	} else {
		cnt.inst -= uint32(-delta)
	}
	cnt.last = now
	if was0 {
		cnt.since0 = now
	} else if cnt.inst == 0 {
		cnt.dur += now.Sub(cnt.since0)
	}
}
