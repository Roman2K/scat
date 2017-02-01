// Sliding time-window rate counter
package slidecnt

import "time"

type Counter struct {
	Window time.Duration
	ticks  []tick
}

type tick struct {
	time time.Time
	num  uint64
}

var getNow = func() time.Time {
	return time.Now()
}

func (c *Counter) Add(num uint64) {
	now := getNow()
	c.prune(now)
	c.ticks = append(c.ticks, tick{time: now, num: num})
}

func (c *Counter) prune(now time.Time) {
	minTime := now.Add(-c.Window)
	i := 0
	for n := len(c.ticks); i < n; i++ {
		time := c.ticks[i].time
		if time.After(minTime) || time == minTime {
			break
		}
	}
	c.ticks = append(c.ticks[:0], c.ticks[i:]...)
}

func (c *Counter) AvgRate(unit time.Duration) uint64 {
	if unit == 0 {
		return 0
	}
	now := getNow()
	c.prune(now)
	if len(c.ticks) == 0 {
		return 0
	}
	elapsed := now.Sub(c.ticks[0].time)
	timeDivisor := float64(elapsed) / float64(unit)
	if timeDivisor == 0 {
		return 0
	}
	var sum uint64
	for _, t := range c.ticks {
		sum += t.num
	}
	return uint64(float64(sum) / timeDivisor)
}
