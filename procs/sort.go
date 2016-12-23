package procs

import (
	"sync"

	ss "secsplit"
	"secsplit/seriessort"
)

type Sort struct {
	order   seriessort.Series
	orderMu sync.Mutex
}

func (s *Sort) Process(c *ss.Chunk) Res {
	s.orderMu.Lock()
	s.order.Add(c.Num, c)
	sorted := s.order.Sorted()
	s.order.Drop(len(sorted))
	s.orderMu.Unlock()
	chunks := make([]*ss.Chunk, len(sorted))
	for i := range sorted {
		chunks[i] = sorted[i].(*ss.Chunk)
	}
	return Res{Chunks: chunks}
}

func (s *Sort) Finish() (err error) {
	s.orderMu.Lock()
	len := s.order.Len()
	s.orderMu.Unlock()
	if len > 0 {
		err = ErrShort
	}
	return
}
