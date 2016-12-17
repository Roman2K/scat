package procs

import (
	"sync"

	ss "secsplit"
)

type pool struct {
	wg    *sync.WaitGroup
	tasks chan task
}

type task struct {
	chunk *ss.Chunk
	ch    chan<- Res
}

func NewPool(size int, proc Proc) AsyncProcFinisher {
	tasks := make(chan task)
	wg := &sync.WaitGroup{}
	wg.Add(size)
	for i := 0; i < size; i++ {
		go func() {
			defer wg.Done()
			for task := range tasks {
				task.ch <- proc.Process(task.chunk)
			}
		}()
	}
	return pool{wg: wg, tasks: tasks}
}

func (p pool) Process(c *ss.Chunk) <-chan Res {
	ch := make(chan Res)
	p.tasks <- task{chunk: c, ch: ch}
	return ch
}

func (p pool) Finish() error {
	close(p.tasks)
	p.wg.Wait()
	return nil
}
