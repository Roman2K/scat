package aprocs

import ss "secsplit"

type mutex struct {
	proc Proc
	lock chan struct{}
}

func NewMutex(proc Proc) Proc {
	lock := make(chan struct{}, 1)
	lock <- struct{}{}
	return mutex{
		proc: proc,
		lock: lock,
	}
}

func (m mutex) Process(c *ss.Chunk) <-chan Res {
	<-m.lock
	defer func() { m.lock <- struct{}{} }()
	return m.proc.Process(c)
}

func (m mutex) Finish() error {
	return m.proc.Finish()
}
