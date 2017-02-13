package procs

import (
	"errors"
	"io"
	"sort"
	"sync"

	"gitlab.com/Roman2K/scat"
	"gitlab.com/Roman2K/scat/checksum"
	"gitlab.com/Roman2K/scat/index"
	"gitlab.com/Roman2K/scat/seriessort"
)

type IndexProc interface {
	Proc
	EndProc
}

type indexProc struct {
	w        io.Writer
	order    seriessort.Series
	orderMu  sync.Mutex
	finals   map[checksum.Hash]*finals
	finalsMu sync.RWMutex
}

type finals struct {
	num      int
	entries  []indexEntry
	complete bool
	mu       sync.Mutex
}

type indexEntry struct {
	num        int
	hash       checksum.Hash
	targetSize int
}

var (
	ErrIndexUnprocessedChunk = errors.New("unprocessed chunk")
	ErrIndexProcessEnded     = errors.New("process already ended")
	ErrIndexDup              = errors.New("won't process dup chunk")
)

func NewIndexProc(w io.Writer) IndexProc {
	return &indexProc{
		w:      w,
		order:  seriessort.Series{},
		finals: make(map[checksum.Hash]*finals),
	}
}

func (idx *indexProc) Process(c *scat.Chunk) <-chan Res {
	idx.setOrder(c)
	ch := make(chan Res, 1)
	defer close(ch)
	idx.finalsMu.Lock()
	defer idx.finalsMu.Unlock()
	if _, ok := idx.finals[c.Hash()]; !ok {
		idx.finals[c.Hash()] = &finals{
			num:     c.Num(),
			entries: make([]indexEntry, 0, 1),
		}
		ch <- Res{Chunk: c}
	}
	return ch
}

func (idx *indexProc) ProcessFinal(c, final *scat.Chunk) error {
	entry := indexEntry{
		num:        final.Num(),
		hash:       final.Hash(),
		targetSize: final.TargetSize(),
	}
	finals, ok := idx.getFinals(c.Hash())
	if !ok {
		return ErrIndexUnprocessedChunk
	}
	finals.mu.Lock()
	defer finals.mu.Unlock()
	if finals.complete {
		return ErrIndexProcessEnded
	}
	if finals.num != c.Num() {
		return ErrIndexDup
	}
	finals.entries = append(finals.entries, entry)
	return nil
}

func (idx *indexProc) ProcessEnd(c *scat.Chunk) (err error) {
	err = idx.setFinalsComplete(c)
	if err != nil {
		return
	}
	return idx.flush()
}

func (idx *indexProc) getFinals(hash checksum.Hash) (f *finals, ok bool) {
	idx.finalsMu.RLock()
	defer idx.finalsMu.RUnlock()
	f, ok = idx.finals[hash]
	return
}

func (idx *indexProc) setFinalsComplete(c *scat.Chunk) error {
	finals, ok := idx.getFinals(c.Hash())
	if !ok {
		return ErrIndexUnprocessedChunk
	}
	finals.mu.Lock()
	defer finals.mu.Unlock()
	if finals.num != c.Num() {
		return nil
	}
	finals.complete = true
	return nil
}

func (idx *indexProc) Finish() error {
	idx.orderMu.Lock()
	len := idx.order.Len()
	idx.orderMu.Unlock()
	if len > 0 {
		return ErrShort
	}
	return nil
}

func (idx *indexProc) flush() (err error) {
	idx.orderMu.Lock()
	defer idx.orderMu.Unlock()
	sorted := idx.order.Sorted()
	i := 0
	defer func() {
		idx.order.Drop(i)
	}()
	for n := len(sorted); i < n; i++ {
		hash := sorted[i].(checksum.Hash)
		finals, ok := idx.completeFinals(hash)
		if !ok {
			return
		}
		entries := finals.entries
		num := func(i int) int {
			return entries[i].num
		}
		sort.Slice(entries, func(i, j int) bool {
			return num(i) < num(j)
		})
		err = writeEntries(idx.w, entries)
		if err != nil {
			return
		}
	}
	return
}

func (idx *indexProc) completeFinals(hash checksum.Hash) (f *finals, ok bool) {
	f, ok = idx.getFinals(hash)
	if !ok {
		return
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	ok = f.complete
	return
}

func (idx *indexProc) setOrder(c *scat.Chunk) {
	idx.orderMu.Lock()
	defer idx.orderMu.Unlock()
	idx.order.Add(c.Num(), c.Hash())
}

func writeEntries(w io.Writer, entries []indexEntry) (err error) {
	for _, entry := range entries {
		_, err = index.Write(w, entry.hash, entry.targetSize)
		if err != nil {
			return
		}
	}
	return
}

var IndexUnproc Proc = ChunkIterFunc(indexUnprocess)

func indexUnprocess(c *scat.Chunk) scat.ChunkIter {
	return index.NewScanner(c.Num(), c.Data().Reader())
}
