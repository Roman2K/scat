package stats

import (
	"sync"
	"time"
)

type Counter struct {
	pos                uint32
	start              time.Time
	inst               uint32
	last               time.Time
	instMu             sync.Mutex
	out                uint64
	dur                time.Duration
	outMu              sync.Mutex
	QuotaUse, QuotaMax uint64
}

func (cnt *Counter) addOut(nbytes uint64, dur time.Duration) {
	cnt.outMu.Lock()
	defer cnt.outMu.Unlock()
	cnt.out += nbytes
	cnt.dur += dur
}

func (cnt *Counter) getOut() (uint64, time.Duration) {
	cnt.outMu.Lock()
	defer cnt.outMu.Unlock()
	return cnt.out, cnt.dur
}

func (cnt *Counter) getInst() uint32 {
	cnt.instMu.Lock()
	defer cnt.instMu.Unlock()
	return cnt.inst
}

func (cnt *Counter) addInst(delta int) {
	cnt.instMu.Lock()
	defer cnt.instMu.Unlock()
	if delta > 0 {
		cnt.inst += uint32(delta)
	} else {
		cnt.inst -= uint32(-delta)
	}
	cnt.last = time.Now()
}
