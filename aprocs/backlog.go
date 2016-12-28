package aprocs

import (
	"errors"

	ss "secsplit"
)

type backlog struct {
	proc  Proc
	slots chan struct{}
}

func NewBacklog(nslots int, proc Proc) Proc {
	slots := make(chan struct{}, nslots)
	for i, n := 0, cap(slots); i < n; i++ {
		slots <- struct{}{}
	}
	return backlog{
		proc:  proc,
		slots: slots,
	}
}

func (bl backlog) Process(c *ss.Chunk) <-chan Res {
	<-bl.slots
	out := make(chan Res)
	ch := bl.proc.Process(c)
	go func() {
		defer func() { bl.slots <- struct{}{} }()
		defer close(out)
		for res := range ch {
			out <- res
		}
	}()
	return out
}

func (bl backlog) Finish() (err error) {
	err = bl.proc.Finish()
	if err != nil {
		return
	}
	if len(bl.slots) < cap(bl.slots) {
		return errors.New("unreturned slots left")
	}
	return
}

func NewMutex(proc Proc) Proc {
	return NewBacklog(1, proc)
}
