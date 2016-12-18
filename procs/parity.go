package procs

import (
	"errors"

	"github.com/klauspost/reedsolomon"

	ss "secsplit"
)

type parity struct {
	enc     reedsolomon.Encoder
	NShards int
}

func Parity(ndata, nparity int) (*parity, error) {
	enc, err := reedsolomon.New(ndata, nparity)
	return &parity{enc: enc, NShards: ndata + nparity}, err
}

func (p *parity) Proc() Proc {
	return p
}

func (p *parity) Unproc() Proc {
	return procFunc(p.unprocess)
}

func (p *parity) Process(c *ss.Chunk) Res {
	shards, err := p.enc.Split(c.Data)
	if err != nil {
		return Res{Err: err}
	}
	chunks := make([]*ss.Chunk, len(shards))
	for i, shard := range shards {
		chunks[i] = &ss.Chunk{Data: shard}
	}
	return Res{Chunks: chunks}
}

func (p *parity) unprocess(c *ss.Chunk) Res {
	chunks, err := getGroup(c, p.NShards)
	if err != nil {
		return Res{Err: err}
	}
	_ = chunks
	return Res{Err: errors.New("TODO size for enc.Join")}
}

func getGroup(c *ss.Chunk, size int) (chunks []*ss.Chunk, err error) {
	chunks, ok := c.GetMeta("group").([]*ss.Chunk)
	if !ok {
		err = errors.New("missing group")
		return
	}
	if len(chunks) != size {
		err = errors.New("invalid group size")
		return
	}
	return
}
