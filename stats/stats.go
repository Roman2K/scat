package stats

import (
	"fmt"
	"io"
	"runtime"
	"scat/slidecnt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	humanize "github.com/dustin/go-humanize"
)

const (
	aliveThreshold = 1 * time.Second
)

type Statsd struct {
	counters   map[Id]*Counter
	countersMu sync.Mutex
	nextPos    uint32
}

type Id interface{}

type Counter struct {
	pos      uint32
	last     time.Time
	inst     int32
	out      slidecnt.Counter
	outMu    sync.Mutex
	QuotaUse uint64
	QuotaMax uint64
}

const unlimited = ^uint64(0)

func (cnt *Counter) addInst(delta int32) {
	atomic.AddInt32(&cnt.inst, delta)
	cnt.last = time.Now()
}

func (cnt *Counter) addOut(delta uint64) {
	cnt.outMu.Lock()
	defer cnt.outMu.Unlock()
	cnt.out.Add(delta)
}

func (cnt *Counter) outAvgRate(unit time.Duration) uint64 {
	cnt.outMu.Lock()
	defer cnt.outMu.Unlock()
	return cnt.out.AvgRate(unit)
}

func New() *Statsd {
	return &Statsd{
		counters: make(map[Id]*Counter),
	}
}

func (st *Statsd) Counter(id Id) *Counter {
	st.countersMu.Lock()
	defer st.countersMu.Unlock()
	if _, ok := st.counters[id]; !ok {
		st.counters[id] = &Counter{
			pos: st.nextPos,
			out: slidecnt.New(5 * time.Second),
		}
		st.nextPos++
	}
	return st.counters[id]
}

type sortedCounter struct {
	id  Id
	cnt *Counter
}

func (st *Statsd) sortedCounters() (scnts []sortedCounter) {
	st.countersMu.Lock()
	defer st.countersMu.Unlock()
	ids := make([]Id, 0, len(st.counters))
	for id := range st.counters {
		ids = append(ids, id)
	}
	pos := func(i int) uint32 {
		return st.counters[ids[i]].pos
	}
	sort.Slice(ids, func(i, j int) bool {
		return pos(i) < pos(j)
	})
	scnts = make([]sortedCounter, len(ids))
	for i, id := range ids {
		scnts[i] = sortedCounter{id: id, cnt: st.counters[id]}
	}
	return
}

func (st *Statsd) WriteTo(w io.Writer) (written int64, err error) {
	write := func(str string) error {
		n, err := w.Write([]byte(str))
		written += int64(n)
		return err
	}

	// Headers
	err = write(fmt.Sprintf("%15s\t%s\t%12s\t%10s\t%10s\t%7s\n",
		"PROC", "INST", "RATE", "USE", "QUOTA", "FILL",
	))
	if err != nil {
		return
	}

	// Procs
	now := time.Now()
	for _, scnt := range st.sortedCounters() {
		cnt := scnt.cnt
		inst := cnt.inst
		dead := inst == 0 && now.Sub(cnt.last) > aliveThreshold
		out := ""
		if !dead {
			out = humanize.IBytes(cnt.outAvgRate(time.Second)) + "/s"
		}
		line := fmt.Sprintf("%15s\tx%d\t%12s\t%10s\t%10s\t%7s\n",
			scnt.id,
			inst,
			out,
			quotaStr(cnt.QuotaUse),
			quotaStr(cnt.QuotaMax),
			quotaFillStr(cnt.QuotaUse, cnt.QuotaMax),
		)
		if dead {
			line = fmt.Sprintf("\x1b[90m%s\x1b[0m", line)
		}
		err = write(line)
		if err != nil {
			return
		}
	}

	// Goroutines
	err = write(fmt.Sprintf("%15s\tx%d\n",
		"(goroutines)", runtime.NumGoroutine(),
	))
	return
}

func quotaStr(n uint64) string {
	switch n {
	case unlimited:
		return "\u221E"
	case 0:
		return ""
	}
	return humanize.IBytes(n)
}

func quotaFillStr(used, max uint64) string {
	switch max {
	case unlimited:
		return "\u221E"
	case 0:
		return ""
	}
	pct := float64(used) / float64(max) * 100
	return fmt.Sprintf("%.2f%%", pct)
}
