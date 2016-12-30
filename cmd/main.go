package main

import (
	"errors"
	"fmt"
	"os"

	"secsplit/aprocs"
	"secsplit/cpprocs"
	"secsplit/cpprocs/mincopies"
	"secsplit/indexscan"
	"secsplit/procs"
	"secsplit/split"
)

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func start() error {
	args := os.Args[1:]
	if len(args) != 1 {
		return errors.New("usage: split|join")
	}
	cmd := args[0]
	switch cmd {
	case "split":
		return cmdSplit()
	case "join":
		return cmdJoin()
	}
	return fmt.Errorf("unknown cmd: %s", cmd)
}

const (
	ndata   = 2
	nparity = 1
)

func cmdSplit() (err error) {
	in, out := os.Stdin, os.Stdout

	parity, err := aprocs.NewParity(ndata, nparity)
	if err != nil {
		return
	}

	minCopies, err := mincopies.New(2, []cpprocs.Proc{
		cpprocs.NewCommand("cat", cpprocs.NewCat("/Users/roman/tmp/cat")),
	})
	if err != nil {
		return
	}

	chain := aprocs.NewBacklog(2, aprocs.NewChain([]aprocs.Proc{
		procs.A(procs.Checksum{}.Proc()),
		procs.A(procs.Size),
		aprocs.NewIndex(out),
		parity.Proc(),
		procs.A((&procs.Compress{}).Proc()),
		procs.A(procs.Checksum{}.Proc()),
		// procs.A((&procs.LocalStore{"out"}).Proc()),
		aprocs.NewConcur(4, minCopies),
	}))
	defer chain.Finish()

	splitter := split.NewSplitter(in)
	err = aprocs.Process(chain, splitter)
	if err != nil {
		return
	}
	return chain.Finish()
}

func cmdJoin() (err error) {
	in, out := os.Stdin, os.Stdout

	parity, err := aprocs.NewParity(ndata, nparity)
	if err != nil {
		return
	}

	chain := aprocs.NewBacklog(3, aprocs.NewChain([]aprocs.Proc{
		procs.A((&procs.LocalStore{"out"}).Unproc()),
		aprocs.NewPool(2, aprocs.NewChain([]aprocs.Proc{
			procs.A(procs.Checksum{}.Unproc()),
			procs.A((&procs.Compress{}).Unproc()),
			aprocs.NewGroup(ndata + nparity),
			parity.Unproc(),
		})),
		aprocs.NewMutex(aprocs.NewChain([]aprocs.Proc{
			aprocs.NewSort(),
			aprocs.NewWriterTo(out),
		})),
	}))

	scan := indexscan.NewScanner(in)
	err = aprocs.Process(chain, scan)
	if err != nil {
		return
	}
	return chain.Finish()
}
