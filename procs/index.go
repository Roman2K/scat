package procs

import (
	"fmt"
	"io"
	"sync"

	ss "secsplit"
	"secsplit/checksum"
	"secsplit/indexscan"
)

type index struct {
	w        io.Writer
	order    []*checksum.Hash
	orderMu  sync.Mutex
	finals   map[checksum.Hash][]indexEntry
	finalsMu sync.Mutex
}

type indexEntry struct {
	hash *checksum.Hash
	size int
}

func NewIndex(w io.Writer) ProcFinisher {
	return &index{
		w:      w,
		finals: make(map[checksum.Hash][]indexEntry),
	}
}

func (i *index) Process(c *ss.Chunk) Res {
	return inplaceProcFunc(i.process).Process(c)
}

func (i *index) process(c *ss.Chunk) error {
	i.setOrder(c.Hash, c.Num)
	return nil
}

func (i *index) end(c *ss.Chunk, finals []*ss.Chunk) {
	i.finalsMu.Lock()
	defer i.finalsMu.Unlock()
	// We can't just check for finals[hash] and return if present because chunks
	// are potentially processed out of order. So, for example, when deduping,
	// a duplicate might land here before the first occurrence, registering no
	// finals. However, we want to register the finals of the first occurrence
	// only.
	if len(finals) < len(i.finals[c.Hash]) {
		return
	}
	entries := make([]indexEntry, len(finals))
	for i, fc := range finals {
		entries[i] = indexEntry{hash: &fc.Hash, size: c.Size}
	}
	i.finals[c.Hash] = entries
}

func (i *index) setOrder(hash checksum.Hash, num int) {
	i.orderMu.Lock()
	defer i.orderMu.Unlock()
	if minLen := num + 1; len(i.order) < minLen {
		if cap(i.order) < minLen {
			resized := make([]*checksum.Hash, minLen, num*2+1)
			copy(resized, i.order)
			i.order = resized
		}
		i.order = i.order[:minLen]
	}
	i.order[num] = &hash
}

func (i *index) Finish() (err error) {
	for num, hash := range i.order {
		if hash == nil {
			return fmt.Errorf("missing chunk %d", num)
		}
	}
	for _, hash := range i.order {
		for _, entry := range i.getFinals(*hash) {
			_, err = indexscan.Write(i.w, *entry.hash, entry.size)
			if err != nil {
				return
			}
		}
	}
	return
}

func (i *index) getFinals(hash checksum.Hash) []indexEntry {
	i.finalsMu.Lock()
	defer i.finalsMu.Unlock()
	return i.finals[hash]
}
