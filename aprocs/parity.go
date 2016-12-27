package aprocs

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

func NewParity(ndata, nparity int) (p ProcUnprocer, err error) {
	enc, err := reedsolomon.New(ndata, nparity)
	p = &parity{
		enc:     enc,
		ndata:   ndata,
		nshards: ndata + nparity,
	}
	return
}

func (p *parity) Proc() Proc {
	return ProcFunc(p.process)
}

func (p *parity) Unproc() Proc {
	return InplaceProcFunc(p.unprocess)
}

func (p *parity) process(c *ss.Chunk) <-chan Res {
	ch := make(chan Res)
	shards, err := p.split(c)
	go func() {
		defer close(ch)
		if err != nil {
			ch <- Res{Chunk: c, Err: err}
			return
		}
		for i, shard := range shards {
			chunk := &ss.Chunk{
				Num:  c.Num*p.nshards + i,
				Data: shard,
			}
			ch <- Res{Chunk: chunk}
		}
	}()
	return ch
}

func (p *parity) split(c *ss.Chunk) (shards [][]byte, err error) {
	shards, err = p.enc.Split(c.Data)
	if err != nil {
		return
	}
	err = p.enc.Encode(shards)
	return
}

func (p *parity) unprocess(c *ss.Chunk) (err error) {
	data, err := p.join(c)
	c.Data = data
	return
}

func (p *parity) join(c *ss.Chunk) (joined []byte, err error) {
	// Shard chunks
	chunks, err := getGroup(c, p.nshards)
	if err != nil {
		return
	}

	// Shards slice
	out := bytes.NewBuffer(make([]byte, 0, len(c.Data)*p.ndata))
	shards := make([][]byte, len(chunks))
	mustReconstruct := false
	for i, c := range chunks {
		ok, err := getIntegrityCheck(c)
		if err != nil {
			return nil, err
		}
		if !ok {
			mustReconstruct = true
			continue
		}
		shards[i] = c.Data
	}

	// Reconstruct invalid shards
	if mustReconstruct {
		err = p.enc.Reconstruct(shards)
		if err != nil {
			return
		}
	}

	// Verify integrity
	ok, err := p.enc.Verify(shards)
	if err == nil && !ok {
		err = errors.New("verification failed")
	}
	if err != nil {
		return
	}

	// Join data shards, trim trailing padding
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

func getIntegrityCheck(c *ss.Chunk) (bool, error) {
	err, ok := c.GetMeta("err").(error)
	if !ok {
		return true, nil
	}
	if err == ErrIntegrityCheckFailed {
		return false, nil
	}
	return false, err
}
