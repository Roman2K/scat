package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"

	humanize "github.com/dustin/go-humanize"

	"secsplit/ansirefresh"
	"secsplit/aprocs"
	"secsplit/cpprocs"
	"secsplit/cpprocs/mincopies"
	"secsplit/cpprocs/quota"
	"secsplit/index"
	"secsplit/procs"
	"secsplit/split"
	"secsplit/stats"
	"secsplit/tmpdedup"
)

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		if exit, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(os.Stderr, "stderr: %s\n", exit.Stderr)
		}
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

type remote struct {
	name  string
	lsp   cpprocs.LsProcUnprocer
	quota uint64
}

func remotes() []remote {
	cat := func(n int, quota uint64) remote {
		name := fmt.Sprintf("cat%d", n)
		lsp := cpprocs.NewCat("/Users/roman/tmp/" + name)
		return remote{name, lsp, quota}
	}
	return []remote{
		cat(1, 10*humanize.MiByte),
		cat(2, quota.Unlimited),
		cat(3, quota.Unlimited),
	}

	// addCopier("drive",
	// 	cpprocs.NewRclone("drive:tmp", tmp),
	// 	7*humanize.GiByte,
	// )
	// addCopier("drive2",
	// 	cpprocs.NewRclone("drive2:tmp", tmp),
	// 	14*humanize.GiByte,
	// )
}

func quotaMan(statsd *stats.Statsd) (qman quota.Man) {
	qman = quota.NewMan()
	for _, r := range remotes() {
		proc := stats.NewProc(statsd, r.name, r.lsp.Proc())
		copier := cpprocs.NewCopier(r.name, r.lsp, proc)
		qman.AddResQuota(copier, r.quota)
	}
	return
}

func readers(statsd *stats.Statsd) (cps []cpprocs.Copier) {
	rems := remotes()
	cps = make([]cpprocs.Copier, len(rems))
	for i, r := range rems {
		proc := stats.NewProc(statsd, r.name, r.lsp.Unproc())
		cps[i] = cpprocs.NewCopier(r.name, r.lsp, proc)
	}
	return
}

const (
	ndata   = 2
	nparity = 1
)

func cmdSplit() (err error) {
	statsd := stats.New()
	{
		w := ansirefresh.NewWriter(os.Stderr)
		t := ansirefresh.NewWriteTicker(w, statsd, 250*time.Millisecond)
		defer t.Stop()
	}

	parity, err := aprocs.NewParity(ndata, nparity)
	if err != nil {
		return
	}

	tmp, err := tmpdedup.TempDir("")
	if err != nil {
		return
	}
	defer tmp.Finish()

	minCopies, err := mincopies.New(2, quotaMan(statsd))
	if err != nil {
		return
	}

	chain := aprocs.NewBacklog(10, aprocs.NewChain([]aprocs.Proc{
		stats.NewProc(statsd, "checksum",
			procs.A(procs.Checksum{}.Proc()),
		),
		stats.NewProc(statsd, "size",
			procs.A(procs.Size),
		),
		stats.NewProc(statsd, "index",
			aprocs.NewIndex(os.Stdout),
		),
		stats.NewProc(statsd, "parity",
			parity.Proc(),
		),
		stats.NewProc(statsd, "compress",
			procs.A((&procs.Compress{}).Proc()),
		),
		stats.NewProc(statsd, "checksum2",
			procs.A(procs.Checksum{}.Proc()),
		),
		aprocs.NewConcur(10, minCopies),
	}))
	defer chain.Finish()

	splitter := split.NewSplitter(os.Stdin)
	err = aprocs.Process(chain, splitter)
	if err != nil {
		return
	}
	return chain.Finish()
}

func cmdJoin() (err error) {
	statsd := stats.New()
	{
		w := ansirefresh.NewWriter(os.Stderr)
		t := ansirefresh.NewWriteTicker(w, statsd, 250*time.Millisecond)
		defer t.Stop()
	}

	parity, err := aprocs.NewParity(ndata, nparity)
	if err != nil {
		return
	}

	mrd, err := cpprocs.NewMultiReader(readers(statsd))
	if err != nil {
		return
	}

	chain := aprocs.NewBacklog(2, aprocs.NewChain([]aprocs.Proc{
		stats.NewProc(statsd, "mrd",
			mrd,
		),
		stats.NewProc(statsd, "checksum",
			procs.A(procs.Checksum{}.Unproc()),
		),
		stats.NewProc(statsd, "compress",
			procs.A((&procs.Compress{}).Unproc()),
		),
		stats.NewProc(statsd, "group",
			aprocs.NewGroup(ndata+nparity),
		),
		stats.NewProc(statsd, "parity",
			parity.Unproc(),
		),
		aprocs.NewMutex(aprocs.NewChain([]aprocs.Proc{
			stats.NewProc(statsd, "sort",
				aprocs.NewSort(),
			),
			stats.NewProc(statsd, "writerto",
				aprocs.NewWriterTo(os.Stdout),
			),
		})),
	}))

	scan := index.NewScanner(os.Stdin)
	err = aprocs.Process(chain, scan)
	if err != nil {
		return
	}
	return chain.Finish()
}
