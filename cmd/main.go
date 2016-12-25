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

	ppool := aprocs.NewPool(4, procs.A(procs.NewChain([]procs.Proc{
		procs.Checksum{}.Proc(),
		procs.Size,
		procs.NewIndex(out),
		(&procs.LocalStore{"out"}).Proc(),
	})))
	defer ppool.Finish()

	// ppool := aprocs.NewPool(8, aprocs.NewChain([]aprocs.Proc{
	// 	procs.A(procs.Checksum{}.Proc()),
	// 	procs.A(procs.Size),
	// 	procs.A(procs.NewIndex(out)),
	// 	// (&procs.LocalStore{"out"}).Proc(),
	// }))
	// defer ppool.Finish()

	splitter := split.NewSplitter(in)
	err = aprocs.Process(ppool, splitter)
	if err != nil {
		return
	}
	return ppool.Finish()
}

func cmdJoin() (err error) {
	in, out := os.Stdin, os.Stdout

	procChain := aprocs.NewPool(4, procs.A(procs.NewChain([]procs.Proc{
		(&procs.LocalStore{"out"}).Unproc(),
		procs.Checksum{}.Unproc(),
	})))
	outChain := procs.A(procs.NewChain([]procs.Proc{
		&procs.Sort{},
		procs.WriteTo(out),
	}))
	chain := aprocs.NewChain([]aprocs.Proc{
		procChain,
		aprocs.NewMutex(outChain),
	})
	defer chain.Finish()

	scan := indexscan.NewScanner(in)
	err = aprocs.Process(chain, scan)
	if err != nil {
		return
	}
	return chain.Finish()
}
