package procs_test

import (
	"testing"

	"github.com/Roman2K/scat/procs"
	"github.com/Roman2K/scat/testutil"
)

func TestPoolFinish(t *testing.T) {
	testutil.TestFinishErrForward(t, func(proc procs.Proc) testutil.Finisher {
		return procs.NewPool(1, proc)
	})
}
