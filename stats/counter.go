package stats

import (
	"sync"
	"time"
)

type counter struct {
	pos    uint
	start  time.Time
	inst   uint
	instMu sync.Mutex
	out    uint64
	dur    time.Duration
	outMu  sync.Mutex
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

func (cnt *counter) getInst() uint {
	cnt.instMu.Lock()
	defer cnt.instMu.Unlock()
	return cnt.inst
}

func (cnt *counter) addInstance() {
	cnt.instMu.Lock()
	defer cnt.instMu.Unlock()
	cnt.inst++
}

func (cnt *counter) removeInstance() {
	cnt.instMu.Lock()
	defer cnt.instMu.Unlock()
	cnt.inst--
}
