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
	chunks, err := p.split(c)
	return Res{Chunks: chunks, Err: err}
}

func (p *parity) unprocess(c *ss.Chunk) (err error) {
	data, err := p.join(c)
	c.Data = data
	return
}

func (p *parity) split(c *ss.Chunk) (chunks []*ss.Chunk, err error) {
	shards, err := p.enc.Split(c.Data)
	if err != nil {
		return
	}
	err = p.enc.Encode(shards)
	if err != nil {
		return
	}
	chunks = make([]*ss.Chunk, len(shards))
	for i, shard := range shards {
		chunks[i] = &ss.Chunk{
			Num:  c.Num*p.nshards + i,
			Data: shard,
		}
	}
	return
}

func (p *parity) join(c *ss.Chunk) (joined []byte, err error) {
	chunks, err := getGroup(c, p.nshards)
	if err != nil {
		return
	}

	out := bytes.NewBuffer(make([]byte, 0, len(c.Data)*p.ndata))
	shards := make([][]byte, len(chunks))
	mustReconstruct := false
	for i, c := range chunks {
		check, ok := c.GetMeta("integrityCheck").(bool)
		if !ok {
			err = errors.New("missing integrityCheck")
			return
		}
		if !check {
			mustReconstruct = true
			continue
		}
		shards[i] = c.Data
	}

	if mustReconstruct {
		err = p.enc.Reconstruct(shards)
		if err != nil {
			return
		}
	}

	ok, err := p.enc.Verify(shards)
	if err == nil && !ok {
		err = errors.New("verification failed")
	}
	if err != nil {
		return
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
