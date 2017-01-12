package stats

import (
	"fmt"
	"io"
	"runtime"
	"sort"
	"sync"
	"time"

	humanize "github.com/dustin/go-humanize"

	"scat/cpprocs/quota"
)

const aliveThreshold = 1 * time.Second

type Statsd struct {
	counters   map[Id]*Counter
	countersMu sync.Mutex
	nextPos    uint32
}

type Id interface{}

func New() *Statsd {
	return &Statsd{
		counters: make(map[Id]*Counter),
	}
}

func (st *Statsd) Counter(id Id) *Counter {
	st.countersMu.Lock()
	defer st.countersMu.Unlock()
	if _, ok := st.counters[id]; !ok {
		st.counters[id] = &Counter{pos: st.nextPos}
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
		ninst := cnt.getInst()
		out, dur := cnt.getOut()
		line := fmt.Sprintf("%15s\tx%d\t%10s/s\t%10s\t%10s\t%7s\n",
			scnt.id,
			ninst,
			rateStr(out, dur),
			quotaStr(cnt.QuotaUse),
			quotaStr(cnt.QuotaMax),
			quotaFillStr(cnt.QuotaUse, cnt.QuotaMax),
		)
		if ninst == 0 && now.Sub(cnt.last) > aliveThreshold {
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

func rateStr(n uint64, d time.Duration) string {
	rate := uint64(float64(n) / d.Seconds())
	return humanize.IBytes(rate)
}

func quotaStr(n uint64) string {
	switch n {
	case quota.Unlimited:
		return "\u221E"
	case 0:
		return ""
	}
	return humanize.IBytes(n)
}

func quotaFillStr(used, max uint64) string {
	switch max {
	case quota.Unlimited:
		return "\u221E"
	case 0:
		return ""
	}
	pct := float64(used) / float64(max) * 100
	return fmt.Sprintf("%.2f%%", pct)
}
