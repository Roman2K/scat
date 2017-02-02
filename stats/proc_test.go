package stats_test

import (
	"testing"

	"gitlab.com/Roman2K/scat/procs"
	"gitlab.com/Roman2K/scat/stats"
	"gitlab.com/Roman2K/scat/testutil"
)

func TestProcFinish(t *testing.T) {
	statsd := stats.New()
	testutil.TestFinishErrForward(t, func(proc procs.Proc) testutil.Finisher {
		return stats.Proc{statsd, nil, proc}
	})
}
