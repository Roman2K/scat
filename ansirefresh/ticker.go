package ansirefresh

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

func NewWriteTicker(w WriteFlusher, wt io.WriterTo, d time.Duration) Ticker {
	write := func() {
		err := writeFlush(w, wt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "writeTicker error: %v\n", err)
		}
	}
	return NewTicker(write, d)
}

type Ticker interface {
	Stop()
}

type ticker struct {
	stop     chan struct{}
	done     chan struct{}
	stopOnce sync.Once
}

func NewTicker(fn func(), d time.Duration) Ticker {
	timeTicker := time.NewTicker(d)
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		defer timeTicker.Stop()
		for {
			select {
			case <-timeTicker.C:
				fn()
			case <-stop:
				fn()
				return
			}
		}
	}()
	return &ticker{stop: stop, done: done}
}

func (t *ticker) Stop() {
	t.stopOnce.Do(func() {
		close(t.stop)
	})
	<-t.done
}

func writeFlush(w WriteFlusher, wt io.WriterTo) (err error) {
	_, err = wt.WriteTo(w)
	if err != nil {
		return
	}
	return w.Flush()
}
