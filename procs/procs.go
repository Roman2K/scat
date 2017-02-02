package procs

import (
	"errors"
	"fmt"

	"gitlab.com/Roman2K/scat"
	"gitlab.com/Roman2K/scat/concur"
)

var (
	ErrShort           = errors.New("missing final chunks")
	ErrUnreturnedSlots = errors.New("unreturned slots left")
)

var Nop Proc

func init() {
	Nop = InplaceFunc(func(*scat.Chunk) error { return nil })
}

type MissingDataError struct {
	Err error
}

var _ error = MissingDataError{}

func (err MissingDataError) Error() string {
	return fmt.Sprintf("missing data: %v", err.Err)
}

type Proc interface {
	Process(*scat.Chunk) <-chan Res
	Finish() error
}

type EndProc interface {
	ProcessFinal(*scat.Chunk, *scat.Chunk) error
	ProcessEnd(*scat.Chunk) error
}

type ErrProc interface {
	ProcessErr(*scat.Chunk, error) <-chan Res
}

type WrapperProc interface {
	Proc
	Underlying() Proc
}

func underlying(p Proc) Proc {
	for {
		w, ok := p.(WrapperProc)
		if !ok {
			break
		}
		p = w.Underlying()
	}
	return p
}

type Res struct {
	Chunk *scat.Chunk
	Err   error
}

type Procer interface {
	Proc() Proc
}

type Unprocer interface {
	Unproc() Proc
}

type ProcUnprocer interface {
	Procer
	Unprocer
}

type DynProcer interface {
	Procs(*scat.Chunk) ([]Proc, error)
	Finish() error
}

func Process(proc Proc, chunk *scat.Chunk) error {
	defer proc.Finish()
	for res := range proc.Process(chunk) {
		if res.Err != nil {
			return res.Err
		}
	}
	return proc.Finish()
}

func finishFuncs(procs []Proc) (fns concur.Funcs) {
	fns = make(concur.Funcs, len(procs))
	for i, p := range procs {
		fns[i] = p.Finish
	}
	return
}
