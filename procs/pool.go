package procs

import (
	"sync"

	ss "secsplit"
)

type pool struct {
	proc   Proc
	wg     *sync.WaitGroup
	tasks  chan task
	closed bool
}

type task struct {
	chunk *ss.Chunk
	ch    chan<- Res
}

func NewPool(size int, proc Proc) *pool {
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
	return &pool{
		proc:  proc,
		wg:    wg,
		tasks: tasks,
	}
}

func (p *pool) Process(c *ss.Chunk) <-chan Res {
	ch := make(chan Res)
	p.tasks <- task{chunk: c, ch: ch}
	return ch
}

func (p *pool) Finish() error {
	if !p.closed {
		close(p.tasks)
		p.wg.Wait()
		p.closed = true
	}
	if f, ok := p.proc.(Finisher); ok {
		return f.Finish()
	}
	return nil
}
