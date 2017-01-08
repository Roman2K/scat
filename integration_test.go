package scat_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/klauspost/reedsolomon"
	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/aprocs"
	"scat/checksum"
	"scat/index"
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
	indexw io.Writer, in []scat.Chunk, ndata, nparity int, store aprocs.Proc,
) (
	err error,
) {
	parity, err := aprocs.NewParity(ndata, nparity)
	if err != nil {
		return
	}
	chain := aprocs.NewChain([]aprocs.Proc{
		aprocs.ChecksumProc,
		aprocs.NewIndex(indexw),
		parity.Proc(),
		aprocs.NewCompress().Proc(),
		aprocs.ChecksumProc,
		store,
	})
	defer chain.Finish()
	return processFinish(chain, &testutil.SliceIter{S: in})
}

func doJoin(
	w io.Writer, indexr io.Reader, ndata, nparity int, store aprocs.Proc,
) (
	err error,
) {
	scan := index.NewScanner(indexr)
	parity, err := aprocs.NewParity(ndata, nparity)
	if err != nil {
		return
	}
	chain := aprocs.NewChain([]aprocs.Proc{
		store,
		aprocs.ChecksumUnproc,
		aprocs.NewCompress().Unproc(),
		aprocs.NewGroup(ndata + nparity),
		parity.Unproc(),
		aprocs.NewWriterTo(w),
	})
	defer chain.Finish()
	return processFinish(chain, scan)
}

func processFinish(proc aprocs.Proc, iter scat.ChunkIter) (err error) {
	err = aprocs.Process(proc, iter)
	if err != nil {
		return
	}
	return proc.Finish()
}

type memStore map[checksum.Hash][]byte

func (ms memStore) Proc() aprocs.Proc {
	return aprocs.InplaceFunc(ms.process)
}

func (ms memStore) Unproc() aprocs.Proc {
	return aprocs.ChunkFunc(ms.unprocess)
}

func (ms memStore) process(c scat.Chunk) error {
	ms[c.Hash()] = append([]byte{}, c.Data()...)
	return nil
}

func (ms memStore) unprocess(c scat.Chunk) (scat.Chunk, error) {
	return c.WithData(ms[c.Hash()]), nil
}
