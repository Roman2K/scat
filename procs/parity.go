package procs

import (
	"bytes"
	"errors"

	"github.com/klauspost/reedsolomon"

	ss "secsplit"
)

type parity struct {
	enc            reedsolomon.Encoder
	ndata, nshards int
}

func Parity(ndata, nparity int) (p *parity, err error) {
	enc, err := reedsolomon.New(ndata, nparity)
	p = &parity{
		enc:     enc,
		ndata:   ndata,
		nshards: ndata + nparity,
	}
	return
}

func (p *parity) Proc() Proc {
	return p
}

func (p *parity) Unproc() Proc {
	return inplaceProcFunc(p.unprocess)
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

func (p *parity) unprocess(c *ss.Chunk) (err error) {
	data, err := p.join(c)
	c.Data = data
	return
}

func (p *parity) join(c *ss.Chunk) (joined []byte, err error) {
	chunks, err := getGroup(c, p.nshards)
	if err != nil {
		return
	}

	// TODO verify + reconstruct
	// TODO size
	c.Size = 6

	out := bytes.NewBuffer(make([]byte, 0, len(c.Data)*p.ndata))
	shards := make([][]byte, len(chunks))
	for i, c := range chunks {
		shards[i] = c.Data
	}
	err = p.enc.Join(out, shards, c.Size)
	joined = out.Bytes()
	return
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
