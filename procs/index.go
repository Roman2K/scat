package procs

import (
	"io"
	"sync"

	ss "secsplit"
	"secsplit/checksum"
	"secsplit/indexscan"
	"secsplit/seriessort"
)

type index struct {
	w        io.Writer
	order    seriessort.Series
	orderMu  sync.Mutex
	finals   map[checksum.Hash][]indexEntry
	finalsMu sync.Mutex
}

type indexEntry struct {
	hash *checksum.Hash
	size int
}

func NewIndex(w io.Writer) *index {
	return &index{
		w:      w,
		order:  seriessort.Series{},
		finals: make(map[checksum.Hash][]indexEntry),
	}
}

func (idx *index) Process(c *ss.Chunk) Res {
	idx.finalsMu.Lock()
	defer idx.finalsMu.Unlock()
	chunks := make([]*ss.Chunk, 0, 1)
	if _, ok := idx.finals[c.Hash]; !ok {
		idx.finals[c.Hash] = nil
		chunks = append(chunks, c)
	}
	return Res{Chunks: chunks}
}

func (idx *index) ProcessEnd(c *ss.Chunk, finals []*ss.Chunk) (err error) {
	idx.setOrder(c)
	idx.setFinals(c, finals)
	return idx.flush()
}

func (idx *index) Finish() (err error) {
	err = idx.flush()
	if err != nil {
		return
	}
	idx.orderMu.Lock()
	defer idx.orderMu.Unlock()
	if idx.order.Len() > 0 {
		return ErrShort
	}
	return
}

func (idx *index) flush() (err error) {
	idx.orderMu.Lock()
	defer idx.orderMu.Unlock()
	sorted := idx.order.Sorted()
	i := 0
	defer func() {
		idx.order.Drop(i)
	}()
	for n := len(sorted); i < n; i++ {
		hash := sorted[i].(*checksum.Hash)
		finals := idx.getFinals(*hash)
		if len(finals) == 0 {
			return
		}
		err = writeEntries(idx.w, finals)
		if err != nil {
			return
		}
	}
	return
}

func (idx *index) setOrder(c *ss.Chunk) {
	idx.orderMu.Lock()
	defer idx.orderMu.Unlock()
	idx.order.Add(c.Num, &c.Hash)
}

func (idx *index) setFinals(c *ss.Chunk, finals []*ss.Chunk) {
	idx.finalsMu.Lock()
	defer idx.finalsMu.Unlock()
	if len(finals) == 0 {
		return
	}
	entries := make([]indexEntry, len(finals))
	for i, fc := range finals {
		entries[i] = indexEntry{hash: &fc.Hash, size: c.Size}
	}
	idx.finals[c.Hash] = entries
}

func (idx *index) getFinals(hash checksum.Hash) []indexEntry {
	idx.finalsMu.Lock()
	defer idx.finalsMu.Unlock()
	return idx.finals[hash]
}

func writeEntries(w io.Writer, entries []indexEntry) (err error) {
	for _, entry := range entries {
		_, err = indexscan.Write(w, *entry.hash, entry.size)
		if err != nil {
			return
		}
	}
	return
}
