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
	return ppool.Finish()
}

func cmdJoin() (err error) {
	in, out := os.Stdin, os.Stdout

	parity, err := procs.Parity(ndata, nparity)
	if err != nil {
		return
	}

	outIter := procs.Iter()
	ppool := procs.NewPool(4, procs.NewChain([]procs.Proc{
		(&procs.LocalStore{"out"}).Unproc(),
		procs.Checksum{}.Unproc(),
		(&procs.Compress{}).Unproc(),
		procs.Group(ndata + nparity),
		parity.Unproc(),
		outIter,
	}))
	outChain := procs.NewChain([]procs.Proc{
		&procs.Sort{},
		procs.WriteTo(out),
	})

	process := func() (err error) {
		defer ppool.Finish()
		scan := indexscan.NewScanner(in)
		err = procs.ProcessAsync(ppool, scan)
		if err != nil {
			return
		}
		return ppool.Finish()
	}

	processOut := func() (err error) {
		defer outChain.Finish()
		err = procs.Process(outChain, outIter)
		if err != nil {
			return
		}
		return outChain.Finish()
	}

	return concur.Funcs{processOut, process}.FirstErr()
}
