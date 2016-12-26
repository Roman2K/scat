package aprocs

import (
	"io"
	"sync"

	ss "secsplit"
	"secsplit/checksum"
	"secsplit/indexscan"
	"secsplit/seriessort"
)

type Index interface {
	Proc
	EndProc
}

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

func NewIndex(w io.Writer) Index {
	return &index{
		w:      w,
		order:  seriessort.Series{},
		finals: make(map[checksum.Hash][]indexEntry),
	}
}

func (idx *index) Process(c *ss.Chunk) <-chan Res {
	idx.setOrder(c)
	ch := make(chan Res, 1)
	idx.finalsMu.Lock()
	defer idx.finalsMu.Unlock()
	if _, ok := idx.finals[c.Hash]; !ok {
		idx.finals[c.Hash] = nil
		ch <- Res{Chunk: c}
	}
	close(ch)
	return ch
}

func (idx *index) ProcessEnd(c, final *ss.Chunk) error {
	idx.addFinal(c, final)
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

func (idx *index) addFinal(c *ss.Chunk, final *ss.Chunk) {
	entry := indexEntry{hash: &final.Hash, size: c.Size}
	idx.finalsMu.Lock()
	defer idx.finalsMu.Unlock()
	idx.finals[c.Hash] = append(idx.finals[c.Hash], entry)
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
