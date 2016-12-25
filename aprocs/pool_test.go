package aprocs_test

import (
	"errors"
	"testing"

	assert "github.com/stretchr/testify/require"

	ss "secsplit"
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

type finishErrProc struct {
	err error
}

func (p finishErrProc) Process(*ss.Chunk) <-chan aprocs.Res {
	ch := make(chan aprocs.Res)
	close(ch)
	return ch
}

func (p finishErrProc) Finish() error {
	return p.err
}
