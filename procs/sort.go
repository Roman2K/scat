package procs

import (
	"sync"

	"gitlab.com/Roman2K/scat"
	"gitlab.com/Roman2K/scat/seriessort"
)

type Sort struct {
	series   seriessort.Series
	seriesMu sync.Mutex
}

func (s *Sort) Process(c *scat.Chunk) <-chan Res {
	s.seriesMu.Lock()
	s.series.Add(c.Num(), c)
	sorted := s.series.Sorted()
	s.series.Drop(len(sorted))
	s.seriesMu.Unlock()
	ch := make(chan Res, len(sorted))
	defer close(ch)
	for _, val := range sorted {
		ch <- Res{Chunk: val.(*scat.Chunk)}
	}
	return ch
}

func (s *Sort) Finish() error {
	s.seriesMu.Lock()
	len := s.series.Len()
	s.seriesMu.Unlock()
	if len > 0 {
		return ErrShort
	}
	return nil
}
