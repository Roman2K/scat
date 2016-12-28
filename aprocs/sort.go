package aprocs

import (
	"sync"

	ss "secsplit"
	"secsplit/seriessort"
)

type sortProc struct {
	series   seriessort.Series
	seriesMu sync.Mutex
}

func NewSort() Proc {
	return &sortProc{}
}

func (s *sortProc) Process(c *ss.Chunk) <-chan Res {
	s.seriesMu.Lock()
	s.series.Add(c.Num, c)
	sorted := s.series.Sorted()
	s.series.Drop(len(sorted))
	s.seriesMu.Unlock()
	ch := make(chan Res, len(sorted))
	for _, val := range sorted {
		ch <- Res{Chunk: val.(*ss.Chunk)}
	}
	close(ch)
	return ch
}

func (s *sortProc) Finish() error {
	s.seriesMu.Lock()
	len := s.series.Len()
	s.seriesMu.Unlock()
	if len > 0 {
		return ErrShort
	}
	return nil
}
