package procs

import (
	"github.com/Roman2K/scat"
	"github.com/Roman2K/scat/slots"
)

type backlog struct {
	proc  Proc
	slots slots.Slots
}

func NewBacklog(nslots int, proc Proc) Proc {
	return backlog{
		proc:  proc,
		slots: slots.New(nslots),
	}
}

func (bl backlog) Process(c *scat.Chunk) <-chan Res {
	bl.slots.Take()
	ch := bl.proc.Process(c)
	out := make(chan Res)
	go func() {
		defer bl.slots.Release()
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
		return ErrUnreturnedSlots
	}
	return
}
