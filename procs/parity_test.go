package procs

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestParityChunkNum(t *testing.T) {
	const (
		ndata   = 2
		nparity = 1
		nshards = ndata + nparity
	)
	parity, err := Parity(ndata, nparity)
	assert.NoError(t, err)
	testChunkNums(t, parity.Proc(), 2)
	testChunkNums(t, NewChain([]Proc{
		Group(nshards),
		parity.Unproc(),
	}), nshards*2)
}
