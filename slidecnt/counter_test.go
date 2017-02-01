package slidecnt

import (
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

func TestCounter(t *testing.T) {
	origGet := getNow
	defer func() {
		getNow = origGet
	}()

	var now time.Time
	getNow = func() time.Time {
		return now
	}
	c := Counter{Window: 2 * time.Second}
	avgPerSec := func() uint64 {
		return c.AvgRate(1 * time.Second)
	}

	now = time.Now()
	assert.Equal(t, int(0), int(c.AvgRate(0)))
	assert.Equal(t, int(0), int(avgPerSec()))

	// 0s(1000)
	c.Add(1000)
	assert.Equal(t, int(0), int(avgPerSec()))

	// 0.5s(1000)
	now = now.Add(500 * time.Millisecond)
	assert.Equal(t, int(2000), int(avgPerSec()))

	// 1s(2000)
	now = now.Add(500 * time.Millisecond)
	c.Add(1000)
	assert.Equal(t, int(2000), int(avgPerSec()))

	// 2s(2000)
	now = now.Add(1 * time.Second)
	assert.Equal(t, int(1000), int(avgPerSec()))
	assert.Equal(t, 2, cap(c.ticks))

	// 3s(6000) - 1s(2000)
	// => 2s(4000)
	now = now.Add(1 * time.Second)
	c.Add(3000)
	assert.Equal(t, int(2000), int(avgPerSec()))
	assert.Equal(t, 2, cap(c.ticks))

	// 3s(4000) - 1s(1000)
	// => 2s(3000)
	now = now.Add(1 * time.Second)
	assert.Equal(t, int(3000), int(avgPerSec()))
	assert.Equal(t, 2, cap(c.ticks))

	// 3s(3000) - 1s(0)
	// => 2s(3000)
	now = now.Add(1 * time.Second)
	assert.Equal(t, int(1500), int(avgPerSec()))
	assert.Equal(t, 2, cap(c.ticks))

	// 3s(3000) - 1s(3000)
	// => 2s(0)
	now = now.Add(1 * time.Second)
	assert.Equal(t, int(0), int(avgPerSec()))
	assert.Equal(t, 2, cap(c.ticks))
}
