package procs

import (
	"sync"

	ss "secsplit"
	"secsplit/checksum"
)

type dedup struct {
	seen   map[checksum.Hash]struct{}
	seenMu sync.Mutex
}

func NewDedup() Proc {
	return &dedup{
		seen: make(map[checksum.Hash]struct{}),
	}
}

func (d *dedup) Process(c *ss.Chunk) Res {
	return inplaceProcFunc(d.process).Process(c)
}

func (d *dedup) process(c *ss.Chunk) error {
	if d.getSeen(c.Hash) {
		c.Dup = true
	} else {
		d.setSeen(c.Hash)
	}
	return nil
}

func (d *dedup) getSeen(hash checksum.Hash) (ok bool) {
	d.seenMu.Lock()
	defer d.seenMu.Unlock()
	_, ok = d.seen[hash]
	return
}

func (d *dedup) setSeen(hash checksum.Hash) {
	d.seenMu.Lock()
	defer d.seenMu.Unlock()
	d.seen[hash] = struct{}{}
}
