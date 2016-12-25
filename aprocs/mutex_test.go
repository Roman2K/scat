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
		time.Sleep(c.GetMeta("delay").(time.Duration))
		processed = append(processed, c.Num)
		return nil
	})
	mutex := aprocs.NewMutex(proc)
	wg := sync.WaitGroup{}
	wg.Add(2)
	process := func(delay time.Duration, num int) {
		defer wg.Done()
		c := &ss.Chunk{Num: num}
		c.SetMeta("delay", delay)
		mutex.Process(c)
	}
	go process(100, 0)
	go process(0, 1)
	wg.Wait()
	assert.Equal(t, []int{1, 0}, processed)
}
