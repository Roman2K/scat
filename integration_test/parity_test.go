package integration_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/klauspost/reedsolomon"
	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/checksum"
	"scat/procs"
)

func TestParityCorruptNone(t *testing.T) {
	testParity(t, corruptNone)
}

func TestParityCorruptRecoverable(t *testing.T) {
	testParity(t, corruptRecoverable)
}

func TestParityCorruptNonRecoverable(t *testing.T) {
	testParity(t, corruptNonRecoverable)
}

type corruption int

const (
	corruptNone corruption = iota
	corruptRecoverable
	corruptNonRecoverable
)

func testParity(t *testing.T, cor corruption) {
	const (
		ndata    = 2
		nparity  = 1
		inputStr = "hello"
	)

	parity, err := procs.NewParity(ndata, nparity)
	assert.NoError(t, err)

	indexBuf := &bytes.Buffer{}
	store := memStore{}

	// split
	seed := scat.NewChunk(0, scat.BytesData(inputStr))
	seed.SetTargetSize(len(inputStr))
	chain := procs.Chain{
		procs.ChecksumProc,
		procs.NewIndexProc(indexBuf),
		parity.Proc(),
		procs.Gzip{}.Proc(),
		procs.ChecksumProc,
		store.Proc(),
	}
	err = procs.Process(chain, seed)
	assert.NoError(t, err)

	corrupt := func(n int) {
		i := 0
		for hash, data := range store {
			if i >= n {
				break
			}
			store[hash] = append(data, 'x')
			i++
		}
	}

	switch cor {
	case corruptNone:
	case corruptRecoverable:
		corrupt(nparity)
	case corruptNonRecoverable:
		corrupt(nparity + 1)
	default:
		panic("unhandled corruption type")
	}

	seed = scat.NewChunk(0, scat.NewReaderData(indexBuf))
	outBuf := &bytes.Buffer{}
	chain = procs.Chain{
		procs.IndexUnproc,
		store.Unproc(),
		procs.ChecksumUnproc,
		procs.Gzip{}.Unproc(),
		procs.NewGroup(ndata + nparity),
		parity.Unproc(),
		procs.NewWriterTo(outBuf),
	}
	err = procs.Process(chain, seed)

	if cor == corruptNonRecoverable {
		assert.Equal(t, reedsolomon.ErrTooFewShards, err)
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, inputStr, outBuf.String())
}

type memStore map[checksum.Hash]scat.BytesData

func (ms memStore) Proc() procs.Proc {
	return procs.InplaceFunc(ms.process)
}

func (ms memStore) Unproc() procs.Proc {
	return procs.ChunkFunc(ms.unprocess)
}

func (ms memStore) process(c *scat.Chunk) (err error) {
	buf, err := ioutil.ReadAll(c.Data().Reader())
	if err != nil {
		return
	}
	ms[c.Hash()] = scat.BytesData(buf)
	return
}

func (ms memStore) unprocess(c *scat.Chunk) (*scat.Chunk, error) {
	return c.WithData(ms[c.Hash()]), nil
}
