package main

import (
	"errors"
	"fmt"
	"os"

	"secsplit/aprocs"
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

	// parity, err := aprocs.NewParity(ndata, nparity)
	// if err != nil {
	// 	return
	// }

	ppool := aprocs.NewPool(1, aprocs.NewChain([]aprocs.Proc{
		procs.A(procs.Checksum{}.Proc()),
		procs.A(procs.Size),
		aprocs.NewIndex(out),
		// parity.Proc(),
		// procs.A((&procs.Compress{}).Proc()),
		// procs.A(procs.Checksum{}.Proc()),
		procs.A((&procs.LocalStore{"out"}).Proc()),
	}))
	defer ppool.Finish()

	splitter := split.NewSplitter(in)
	err = aprocs.Process(ppool, splitter)
	if err != nil {
		return
	}
	return ppool.Finish()
}

func cmdJoin() (err error) {
	in, out := os.Stdin, os.Stdout

	// parity, err := aprocs.NewParity(ndata, nparity)
	// if err != nil {
	// 	return
	// }

	// procChain := aprocs.NewPool(1, aprocs.NewChain([]aprocs.Proc{
	// 	procs.A((&procs.LocalStore{"out"}).Unproc()),
	// 	procs.A(procs.Checksum{}.Unproc()),
	// 	// procs.A((&procs.Compress{}).Unproc()),
	// 	// aprocs.NewGroup(ndata + nparity),
	// 	// parity.Unproc(),
	// }))
	// outChain := aprocs.NewChain([]aprocs.Proc{
	// 	procs.A(&procs.Sort{}),
	// 	procs.A(procs.WriteTo(out)),
	// })
	// chain := aprocs.NewChain([]aprocs.Proc{
	// 	procChain,
	// 	aprocs.NewMutex(outChain),
	// })
	// defer chain.Finish()

	chain := aprocs.NewChain([]aprocs.Proc{
		procs.A((&procs.LocalStore{"out"}).Unproc()),
		procs.A(procs.Checksum{}.Unproc()),
		aprocs.NewWriterTo(out),
	})

	// chain := procs.A(procs.NewChain([]procs.Proc{
	// 	(&procs.LocalStore{"out"}).Unproc(),
	// 	procs.Checksum{}.Unproc(),
	// 	&procs.Sort{},
	// 	procs.WriteTo(out),
	// }))

	scan := indexscan.NewScanner(in)
	err = aprocs.Process(chain, scan)
	if err != nil {
		return
	}
	return chain.Finish()
}
