package procs

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestChainFinish(t *testing.T) {
	testErr := errors.New("test")
	proc := finishErrProc{testErr}
	chain := NewChain([]Proc{proc})

	// returns proc's err
	err := chain.Finish()
	assert.Equal(t, testErr, err)

	// idempotence
	err = chain.Finish()
	assert.Equal(t, testErr, err)
}
