package secsplit_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/klauspost/reedsolomon"
	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/indexscan"
	"secsplit/procs"
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
	input := [][]byte{[]byte(inputStr)}

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
	indexw io.Writer, in [][]byte, ndata, nparity int, store procs.Proc,
) (
	err error,
) {
	parity, err := procs.Parity(ndata, nparity)
	if err != nil {
		return
	}
	chain := procs.NewChain([]procs.Proc{
		procs.Checksum{}.Proc(),
		procs.Size,
		procs.NewIndex(indexw),
		parity.Proc(),
		(&procs.Compress{}).Proc(),
		procs.Checksum{}.Proc(),
		store,
	})
	for i, b := range in {
		chunk := &ss.Chunk{Num: i, Data: b}
		err = chain.Process(chunk).Err
		if err != nil {
			return
		}
	}
	return chain.Finish()
}

func doJoin(
	w io.Writer, indexr io.Reader, ndata, nparity int, store procs.Proc,
) (
	err error,
) {
	scan := indexscan.NewScanner(indexr)
	parity, err := procs.Parity(ndata, nparity)
	if err != nil {
		return
	}
	chain := procs.NewChain([]procs.Proc{
		store,
		procs.Checksum{}.Unproc(),
		(&procs.Compress{}).Unproc(),
		procs.Group(ndata + nparity),
		parity.Unproc(),
		procs.WriteTo(w),
	})
	for scan.Next() {
		err = chain.Process(scan.Chunk()).Err
		if err != nil {
			return
		}
	}
	return scan.Err()
}
