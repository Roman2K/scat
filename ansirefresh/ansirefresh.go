package ansirefresh

import (
	"bytes"
	"io"
	"sync"
)

var (
	lf          = []byte{'\n'}
	clear       = []byte("\x1b[2K\r")
	moveUpClear = []byte("\x1b[0A\x1b[2K\r")
)

type writer struct {
	w       io.Writer
	nlines  int
	flushed bool
	mu      sync.Mutex
}

type WriteFlusher interface {
	io.Writer
	Flush() error
}

func NewWriter(w io.Writer) WriteFlusher {
	return &writer{w: w}
}

func (w *writer) Write(b []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.flushed {
		err = w.clear()
		if err != nil {
			return
		}
		w.nlines = 0
		w.flushed = false
	}
	w.nlines += bytes.Count(b, lf)
	return w.w.Write(b)
}

func (w *writer) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.flushed = true
	return nil
}

func (w *writer) clear() (err error) {
	_, err = w.w.Write(clear)
	if err != nil {
		return
	}
	for i := 0; i < w.nlines; i++ {
		_, err = w.w.Write(moveUpClear)
		if err != nil {
			return
		}
	}
	return
}
