package scat_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/klauspost/reedsolomon"
	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/checksum"
	"scat/index"
	"scat/procs"
	"scat/testutil"
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

	indexBuf := &bytes.Buffer{}
	outputBuf := &bytes.Buffer{}
	store := memStore{}
	input := []scat.Chunk{scat.NewChunk(0, []byte(inputStr))}

	err := doSplit(indexBuf, input, ndata, nparity, store.Proc())
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

	err = doJoin(outputBuf, indexBuf, ndata, nparity, store.Unproc())
	if cor == corruptNonRecoverable {
		assert.Equal(t, reedsolomon.ErrTooFewShards, err)
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, inputStr, outputBuf.String())
}

func doSplit(
	indexw io.Writer, in []scat.Chunk, ndata, nparity int, store procs.Proc,
) (
	err error,
) {
	parity, err := procs.NewParity(ndata, nparity)
	if err != nil {
		return
	}
	chain := procs.Chain{
		procs.ChecksumProc,
		procs.NewIndex(indexw),
		parity.Proc(),
		procs.NewGzip().Proc(),
		procs.ChecksumProc,
		store,
	}
	defer chain.Finish()
	return processFinish(chain, &testutil.SliceIter{S: in})
}

func doJoin(
	w io.Writer, indexr io.Reader, ndata, nparity int, store procs.Proc,
) (
	err error,
) {
	scan := index.NewScanner(indexr)
	parity, err := procs.NewParity(ndata, nparity)
	if err != nil {
		return
	}
	chain := procs.Chain{
		store,
		procs.ChecksumUnproc,
		procs.NewGzip().Unproc(),
		procs.NewGroup(ndata + nparity),
		parity.Unproc(),
		procs.NewWriterTo(w),
	}
	defer chain.Finish()
	return processFinish(chain, scan)
}

func processFinish(proc procs.Proc, iter scat.ChunkIter) (err error) {
	err = procs.Process(proc, iter)
	if err != nil {
		return
	}
	return proc.Finish()
}

type memStore map[checksum.Hash][]byte

func (ms memStore) Proc() procs.Proc {
	return procs.InplaceFunc(ms.process)
}

func (ms memStore) Unproc() procs.Proc {
	return procs.ChunkFunc(ms.unprocess)
}

func (ms memStore) process(c scat.Chunk) error {
	ms[c.Hash()] = append([]byte{}, c.Data()...)
	return nil
}

func (ms memStore) unprocess(c scat.Chunk) (scat.Chunk, error) {
	return c.WithData(ms[c.Hash()]), nil
}
