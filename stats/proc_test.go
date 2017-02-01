package stats_test

import (
	"github.com/Roman2K/scat/procs"
	"github.com/Roman2K/scat/stats"
	"github.com/Roman2K/scat/testutil"
	"testing"
)

func TestProcFinish(t *testing.T) {
	statsd := stats.New()
	testutil.TestFinishErrForward(t, func(proc procs.Proc) testutil.Finisher {
		return stats.Proc{statsd, nil, proc}
	})
}
