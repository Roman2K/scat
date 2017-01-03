package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"secsplit/aprocs"
	"secsplit/cpprocs"
	"secsplit/cpprocs/mincopies"
	"secsplit/index"
	"secsplit/procs"
	"secsplit/split"
	"secsplit/stats"
)

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func start() error {
	rand.Seed(time.Now().UnixNano())
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

	log := stats.NewLog(os.Stderr, 250*time.Millisecond)
	// log := stats.NewLog(ioutil.Discard, 250*time.Millisecond)
	// lsLog := stats.NewLsLog(os.Stderr, 250*time.Millisecond)

	parity, err := aprocs.NewParity(ndata, nparity)
	if err != nil {
		return
	}

	cat1 := cpprocs.NewCat("/Users/roman/tmp/cat1")
	cat1cp := cpprocs.NewCopier("cat1", cat1,
		stats.NewProc(log, "cat1", cpprocs.NewCommand(cat1)),
	)
	cat2 := cpprocs.NewCat("/Users/roman/tmp/cat2")
	cat2cp := cpprocs.NewCopier("cat2", cat2,
		stats.NewProc(log, "cat2", cpprocs.NewCommand(cat2)),
	)
	cat3 := cpprocs.NewCat("/Users/roman/tmp/cat3")
	cat3cp := cpprocs.NewCopier("cat3", cat3,
		stats.NewProc(log, "cat3", cpprocs.NewCommand(cat3)),
	)
	minCopies, err := mincopies.New(2, []cpprocs.Copier{cat1cp, cat2cp, cat3cp})
	if err != nil {
		return
	}
	// lsLog.Finish()

	chain := aprocs.NewBacklog(4, aprocs.NewChain([]aprocs.Proc{
		stats.NewProc(log, "checksum",
			procs.A(procs.Checksum{}.Proc()),
		),
		stats.NewProc(log, "size",
			procs.A(procs.Size),
		),
		stats.NewProc(log, "index",
			aprocs.NewIndex(out),
		),
		stats.NewProc(log, "parity",
			parity.Proc(),
		),
		stats.NewProc(log, "compress",
			procs.A((&procs.Compress{}).Proc()),
		),
		stats.NewProc(log, "checksum2",
			procs.A(procs.Checksum{}.Proc()),
		),
		// stats.NewProc(log, "localstore",
		// 	procs.A((&procs.LocalStore{"out"}).Proc()),
		// ),
		aprocs.NewConcur(2, minCopies),
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

	scan := index.NewScanner(in)
	err = aprocs.Process(chain, scan)
	if err != nil {
		return
	}
	return chain.Finish()
}
