package secsplit_test

import (
	"bytes"
	"io"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
	"secsplit/indexscan"
	"secsplit/procs"
)

func TestParity(t *testing.T) {
	const (
		ndata    = 2
		nparity  = 1
		inputStr = "hello"
	)
	indexBuf := &bytes.Buffer{}
	outputBuf := &bytes.Buffer{}
	store := procs.MemStore()
	input := [][]byte{[]byte(inputStr)}
	err := doSplit(indexBuf, input, ndata, nparity, store.Proc())
	assert.NoError(t, err)
	err = doJoin(outputBuf, indexBuf, ndata, nparity, store.Unproc())
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
		procs.NewDedup(),
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

	// TODO proc pool, respect order from index iterator
	for scan.Next() {
		err = chain.Process(scan.Chunk()).Err
		if err != nil {
			return
		}
	}
	return scan.Err()
}
