package aprocs_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	ss "secsplit"
	"secsplit/aprocs"
)

func TestMutex(t *testing.T) {
	processed := []int{}
	proc := aprocs.InplaceProcFunc(func(c *ss.Chunk) error {
		processed = append(processed, c.Num)
		defer func() {
			processed = append(processed, c.Num)
		}()
		time.Sleep(c.GetMeta("testDelay").(time.Duration))
		return nil
	})
	mutex := aprocs.NewMutex(proc)
	wg := sync.WaitGroup{}
	wg.Add(2)
	process := func(num int, delay time.Duration) {
		c := &ss.Chunk{Num: num}
		c.SetMeta("testDelay", delay)
		mutex.Process(c)
	}
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		process(0, 0*time.Millisecond)
	}()
	go func() {
		defer wg.Done()
		process(1, 20*time.Millisecond)
	}()
	start := time.Now()
	wg.Wait()
	elapsed := time.Now().Sub(start)
	assert.True(t, elapsed > 20*time.Millisecond)
	assert.True(t, elapsed < 25*time.Millisecond)
	assert.Equal(t, []int{1, 1, 0, 0}, processed)
}
