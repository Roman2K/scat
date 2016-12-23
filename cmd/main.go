package main

import (
	"errors"
	"fmt"
	"os"

	"secsplit/concur"
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
	splitter := split.NewSplitter(os.Stdin)
	index := procs.NewIndex(os.Stdout)
	parity, err := procs.Parity(ndata, nparity)
	if err != nil {
		return
	}
	chain := procs.NewChain([]procs.Proc{
		procs.Checksum{}.Proc(),
		procs.Size,
		index,
		// procs.NewDedup(),
		parity.Proc(),
		(&procs.Compress{}).Proc(),
		procs.Checksum{}.Proc(),
		(&procs.LocalStore{"out"}).Proc(),
	})
	ppool := procs.NewPool(4, chain)
	defer ppool.Finish()
	err = procs.ProcessAsync(ppool, splitter)
	if err != nil {
		return
	}
	return chain.Finish()
}

func cmdJoin() (err error) {
	scan := indexscan.NewScanner(os.Stdin)
	out := procs.WriteTo(os.Stdout)
	parity, err := procs.Parity(ndata, nparity)
	if err != nil {
		return
	}
	outIter := procs.Iter()
	chain := procs.NewChain([]procs.Proc{
		(&procs.LocalStore{"out"}).Unproc(),
		procs.Checksum{}.Unproc(),
		(&procs.Compress{}).Unproc(),
		procs.Group(ndata + nparity),
		parity.Unproc(),
		outIter,
	})
	ppool := procs.NewPool(8, chain)
	defer ppool.Finish()
	process := func() error {
		defer chain.Finish()
		return procs.ProcessAsync(ppool, scan)
	}

	output := func() error {
		chain := procs.NewChain([]procs.Proc{
			&procs.Sort{},
			out,
		})
		return procs.Process(chain, outIter)
	}

	return concur.FirstErr(output, process)
}
