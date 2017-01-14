package procs_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	"scat/procs"
	"scat/testutil"
)

func TestPoolFinish(t *testing.T) {
	testErr := errors.New("test")
	proc := testutil.FinishErrProc{Err: testErr}
	ppool := procs.NewPool(1, proc)

	// returns procs's err
	err := ppool.Finish()
	assert.Equal(t, testErr, err)

	// idempotence
	err = ppool.Finish()
	assert.Equal(t, testErr, err)
}
