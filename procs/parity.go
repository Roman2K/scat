package procs

import (
	"bytes"
	"errors"

	"github.com/klauspost/reedsolomon"

	"scat"
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
	return ChunkFunc(p.unprocess)
}

func (p *parity) process(c scat.Chunk) <-chan Res {
	ch := make(chan Res)
	shards, err := p.split(c)
	go func() {
		defer close(ch)
		if err != nil {
			ch <- Res{Chunk: c, Err: err}
			return
		}
		for i, shard := range shards {
			chunk := scat.NewChunk(c.Num()*p.nshards+i, shard)
			ch <- Res{Chunk: chunk}
		}
	}()
	return ch
}

func (p *parity) split(c scat.Chunk) (shards [][]byte, err error) {
	shards, err = p.enc.Split(c.Data())
	if err != nil {
		return
	}
	err = p.enc.Encode(shards)
	return
}

func (p *parity) unprocess(c scat.Chunk) (new scat.Chunk, err error) {
	data, err := p.join(c)
	new = c.WithData(data)
	return
}

func (p *parity) join(c scat.Chunk) (joined []byte, err error) {
	// Shard chunks
	chunks, err := getGroup(c, p.nshards)
	if err != nil {
		return
	}

	// Shards slice
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
		shards[i] = c.Data()
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
	out := bytes.NewBuffer(make([]byte, 0, len(c.Data())*p.ndata))
	err = p.enc.Join(out, shards, c.TargetSize())
	joined = out.Bytes()
	return
}

func getGroup(c scat.Chunk, size int) (chunks []scat.Chunk, err error) {
	chunks, ok := c.Meta().Get("group").([]scat.Chunk)
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

func getIntegrityCheck(c scat.Chunk) (bool, error) {
	err, ok := c.Meta().Get("err").(error)
	if !ok {
		return true, nil
	}
	if err == ErrIntegrityCheckFailed {
		return false, nil
	}
	return false, err
}
