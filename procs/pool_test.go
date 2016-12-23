package procs

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestPoolFinish(t *testing.T) {
	testErr := errors.New("test")
	proc := finishErrProc{testErr}
	ppool := NewPool(1, proc)

	// returns procs's err
	err := ppool.Finish()
	assert.Equal(t, testErr, err)

	// idempotence
	err = ppool.Finish()
	assert.Equal(t, testErr, err)
}
