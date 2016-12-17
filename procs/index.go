package procs

import (
	"fmt"
	"io"
	"sync"

	ss "secsplit"
	"secsplit/checksum"
)

type index struct {
	w       io.Writer
	order   []*checksum.Hash
	orderMu sync.Mutex
}

func NewIndex(w io.Writer) ProcFinisher {
	return &index{w: w}
}

func (i *index) Process(c *ss.Chunk) Res {
	return Res{Chunks: []*ss.Chunk{c}}
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
		_, err = checksum.Write(i.w, *hash)
		if err != nil {
			return
		}
	}
	return nil
}
