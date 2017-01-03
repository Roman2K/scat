package stats

import (
	"fmt"
	"io"
	"sort"
	"sync"
	"time"

	humanize "github.com/dustin/go-humanize"

	"secsplit/ansirefresh"
)

type Log struct {
	w          ansirefresh.WriteFlusher
	counters   map[string]*counter
	countersMu sync.Mutex
	nextPos    uint
	done       chan struct{}
	closed     bool
	closeMu    sync.Mutex
}

func NewLog(w io.Writer, update time.Duration) *Log {
	done := make(chan struct{})
	log := &Log{
		w:        ansirefresh.NewWriter(w),
		counters: make(map[string]*counter),
		done:     done,
	}
	ticker := time.NewTicker(update)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.write()
			case <-done:
				return
			}
		}
	}()
	return log
}

func (log *Log) Counter(name string) *counter {
	log.countersMu.Lock()
	defer log.countersMu.Unlock()
	if _, ok := log.counters[name]; !ok {
		log.counters[name] = &counter{pos: log.nextPos, start: time.Now()}
		log.nextPos++
	}
	return log.counters[name]
}

func (log *Log) Finish() error {
	log.closeDone()
	return log.write()
}

func (log *Log) closeDone() {
	log.closeMu.Lock()
	defer log.closeMu.Unlock()
	if log.closed {
		return
	}
	close(log.done)
	log.closed = true
}

func (log *Log) write() error {
	log.countersMu.Lock()
	defer log.countersMu.Unlock()
	names := make([]string, 0, len(log.counters))
	for name := range log.counters {
		names = append(names, name)
	}
	sortable := func(i int) uint {
		return log.counters[names[i]].pos
	}
	sort.Slice(names, func(i, j int) bool {
		return sortable(i) < sortable(j)
	})
	now := time.Now()
	for _, name := range names {
		cnt := log.counters[name]
		out, dur := cnt.getOut()
		outRate := rateStr(out, now.Sub(cnt.start))
		ownRate := rateStr(out, dur)
		_, err := fmt.Fprintf(log.w,
			"%15s x%d:\t%9s/s\t\x1b[90m%9s/s\x1b[0m\n",
			name, cnt.getInst(), ownRate, outRate,
		)
		if err != nil {
			return err
		}
	}
	return log.w.Flush()
}

func rateStr(n uint64, d time.Duration) string {
	rate := uint64(float64(n) / d.Seconds())
	return humanize.IBytes(rate)
}
