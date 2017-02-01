package stats_test

import (
	"scat/procs"
	"scat/stats"
	"scat/testutil"
	"testing"
)

func TestProcFinish(t *testing.T) {
	statsd := stats.New()
	testutil.TestFinishErrForward(t, func(proc procs.Proc) testutil.Finisher {
		return stats.Proc{statsd, nil, proc}
	})
}
