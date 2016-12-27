package secsplit_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/klauspost/reedsolomon"
	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/indexscan"
	"secsplit/procs"
	"secsplit/testhelp"
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
	store := procs.MemStore{}
	input := []*ss.Chunk{&ss.Chunk{Num: 0, Data: []byte(inputStr)}}

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
	indexw io.Writer, in []*ss.Chunk, ndata, nparity int, store procs.Proc,
) (
	err error,
) {
	parity, err := aprocs.NewParity(ndata, nparity)
	if err != nil {
		return
	}
	chain := aprocs.NewChain([]aprocs.Proc{
		procs.A(procs.Checksum{}.Proc()),
		procs.A(procs.Size),
		aprocs.NewIndex(indexw),
		parity.Proc(),
		procs.A((&procs.Compress{}).Proc()),
		procs.A(procs.Checksum{}.Proc()),
		procs.A(store),
	})
	defer chain.Finish()
	return processFinish(chain, &testhelp.SliceIter{S: in})
}

func doJoin(
	w io.Writer, indexr io.Reader, ndata, nparity int, store procs.Proc,
) (
	err error,
) {
	scan := indexscan.NewScanner(indexr)
	parity, err := aprocs.NewParity(ndata, nparity)
	if err != nil {
		return
	}
	chain := aprocs.NewChain([]aprocs.Proc{
		procs.A(store),
		procs.A(procs.Checksum{}.Unproc()),
		procs.A((&procs.Compress{}).Unproc()),
		aprocs.NewGroup(ndata + nparity),
		parity.Unproc(),
		procs.A(procs.WriteTo(w)),
	})
	defer chain.Finish()
	return processFinish(chain, scan)
}

func processFinish(proc aprocs.Proc, iter ss.ChunkIterator) (err error) {
	err = aprocs.Process(proc, iter)
	if err != nil {
		return
	}
	return proc.Finish()
}
