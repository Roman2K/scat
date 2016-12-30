package aprocs_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	"secsplit/aprocs"
)

func TestPoolFinish(t *testing.T) {
	testErr := errors.New("test")
	proc := finishErrProc{testErr}
	ppool := aprocs.NewPool(1, proc)

	// returns procs's err
	err := ppool.Finish()
	assert.Equal(t, testErr, err)

	// idempotence
	err = ppool.Finish()
	assert.Equal(t, testErr, err)
}
