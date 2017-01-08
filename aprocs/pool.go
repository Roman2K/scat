package aprocs

import (
	"sync"

	"scat"
)

type pool struct {
	proc   Proc
	wg     *sync.WaitGroup
	tasks  chan<- task
	closed bool
}

type task struct {
	chunk scat.Chunk
	ch    chan<- Res
}

func (t task) sendProcessed(proc Proc) {
	defer close(t.ch)
	for res := range proc.Process(t.chunk) {
		t.ch <- res
	}
}

func NewPool(size int, proc Proc) Proc {
	tasks := make(chan task)
	wg := &sync.WaitGroup{}
	wg.Add(size)
	for i := 0; i < size; i++ {
		go func() {
			defer wg.Done()
			for task := range tasks {
				task.sendProcessed(proc)
			}
		}()
	}
	return &pool{
		proc:  proc,
		wg:    wg,
		tasks: tasks,
	}
}

func (p *pool) Process(c scat.Chunk) <-chan Res {
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
	return p.proc.Finish()
}
