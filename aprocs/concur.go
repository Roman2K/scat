package aprocs

import (
	"sync"

	"scat"
	"scat/slots"
)

type concurProc struct {
	slots slots.Slots
	dynp  DynProcer
}

func NewConcur(max int, dynp DynProcer) Proc {
	return concurProc{
		slots: slots.New(max),
		dynp:  dynp,
	}
}

func (concp concurProc) Process(c scat.Chunk) <-chan Res {
	procs, err := concp.dynp.Procs(c)
	if err != nil {
		ch := make(chan Res, 1)
		ch <- Res{Chunk: c, Err: err}
		close(ch)
		return ch
	}
	out := make(chan Res)
	wg := sync.WaitGroup{}
	wg.Add(len(procs))
	go func() {
		defer close(out)
		wg.Wait()
	}()
	sendProcessed := func(proc Proc) {
		defer concp.slots.Release()
		defer wg.Done()
		ch := proc.Process(c)
		for res := range ch {
			out <- res
		}
	}
	go func() {
		for _, proc := range procs {
			concp.slots.Take()
			go sendProcessed(proc)
		}
	}()
	return out
}

func (concp concurProc) Finish() (err error) {
	err = concp.dynp.Finish()
	if err != nil {
		return
	}
	if len(concp.slots) < cap(concp.slots) {
		return ErrUnreturnedSlots
	}
	return
}
