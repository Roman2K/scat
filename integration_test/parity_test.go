package integration_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/klauspost/reedsolomon"
	assert "github.com/stretchr/testify/require"

	"scat"
	"scat/checksum"
	"scat/procs"
	"scat/stores"
)

func TestParityCorruptNone(t *testing.T) {
	testParity(t, corruptNone)
}

func TestParityCorruptNoneNonRecoverable(t *testing.T) {
	testParity(t, corruptNone|corruptNonRecoverable)
}

func TestParityCorruptRecoverable(t *testing.T) {
	testParity(t, corruptRecoverable)
}

func TestParityCorruptNonRecoverable(t *testing.T) {
	testParity(t, corruptNonRecoverable)
}

func TestParityCorruptMissingData(t *testing.T) {
	testParity(t, corruptMissingData)
}

func TestParityCorruptMissingDataNonRecoverable(t *testing.T) {
	testParity(t, corruptMissingData|corruptNonRecoverable)
}

type corruption uint

const (
	corruptNone corruption = 1 << iota
	corruptRecoverable
	corruptNonRecoverable
	corruptMissingData
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
	store := stores.NewMem()

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

	corrupt := func(n int, modify func(checksum.Hash)) {
		hashes := store.Hashes()
		assert.True(t, len(hashes) >= n)
		for _, h := range hashes[:n] {
			modify(h)
		}
	}
	tamperWithData := func(h checksum.Hash) {
		store.Set(h, append(store.Get(h), 'x'))
	}
	deleteData := func(h checksum.Hash) {
		store.Delete(h)
	}

	storeUnproc := store.Unproc()
	nonRecoverableErr := reedsolomon.ErrTooFewShards

	switch cor {
	case corruptNone:
	case corruptNone | corruptNonRecoverable:
		someErr := errors.New("some err")
		nonRecoverableErr = someErr
		storeUnproc = procs.Filter{
			Proc: storeUnproc,
			Filter: func(res procs.Res) procs.Res {
				res.Err = someErr
				return res
			},
		}
	case corruptRecoverable:
		corrupt(nparity, tamperWithData)
	case corruptNonRecoverable:
		corrupt(nparity+1, tamperWithData)
	case corruptMissingData:
		corrupt(nparity, deleteData)
	case corruptMissingData | corruptNonRecoverable:
		corrupt(nparity+1, deleteData)
	default:
		panic("unhandled corruption type")
	}

	seed = scat.NewChunk(0, scat.NewReaderData(indexBuf))
	outBuf := &bytes.Buffer{}
	chain = procs.Chain{
		procs.IndexUnproc,
		storeUnproc,
		procs.ChecksumUnproc,
		procs.Gzip{}.Unproc(),
		procs.NewGroup(ndata + nparity),
		parity.Unproc(),
		procs.NewWriterTo(outBuf),
	}
	err = procs.Process(chain, seed)

	if cor&corruptNonRecoverable != 0 {
		assert.Equal(t, nonRecoverableErr, err)
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, inputStr, outBuf.String())
}
