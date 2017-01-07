package stats

import (
	"sync"
	"time"
)

type counter struct {
	pos    uint32
	start  time.Time
	inst   uint32
	instMu sync.Mutex
	out    uint64
	dur    time.Duration
	outMu  sync.Mutex
	last   time.Time
}

func (cnt *counter) addOut(nbytes uint64, dur time.Duration) {
	cnt.outMu.Lock()
	defer cnt.outMu.Unlock()
	cnt.out += nbytes
	cnt.dur += dur
}

func (cnt *counter) getOut() (uint64, time.Duration) {
	cnt.outMu.Lock()
	defer cnt.outMu.Unlock()
	return cnt.out, cnt.dur
}

func (cnt *counter) getInst() uint32 {
	cnt.instMu.Lock()
	defer cnt.instMu.Unlock()
	return cnt.inst
}

func (cnt *counter) addInst(delta int) {
	cnt.instMu.Lock()
	defer cnt.instMu.Unlock()
	if delta > 0 {
		cnt.inst += uint32(delta)
	} else {
		cnt.inst -= uint32(-delta)
	}
	cnt.last = time.Now()
}
