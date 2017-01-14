package main

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"

	humanize "github.com/dustin/go-humanize"

	"scat/ansirefresh"
	"scat/cpprocs"
	"scat/cpprocs/mincopies"
	"scat/cpprocs/quota"
	"scat/index"
	"scat/procs"
	"scat/split"
	"scat/stats"
	"scat/tmpdedup"
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

func catRemotes(*tmpdedup.Dir) []remote {
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
}

func driveRemotes(tmp *tmpdedup.Dir) []remote {
	return []remote{
		{"drive", cpprocs.NewRclone("drive:tmp", tmp), 7 * humanize.GiByte},
		{"drive2", cpprocs.NewRclone("drive2:tmp", tmp), 14 * humanize.GiByte},
	}
}

func remotes(tmp *tmpdedup.Dir) []remote {
	// return catRemotes(tmp)
	return driveRemotes(tmp)
}

func quotaMan(statsd *stats.Statsd, tmp *tmpdedup.Dir) (qman quota.Man) {
	qman = quota.NewMan()
	qman.OnUse = func(res quota.Res, use, max uint64) {
		cnt := statsd.Counter(res.Id())
		cnt.Quota.Use = use
		cnt.Quota.Max = max
	}
	for _, r := range remotes(tmp) {
		id := r.name
		cnt := statsd.Counter(id)
		cnt.Quota.Max = r.quota
		proc := stats.NewProc(statsd, r.name, r.lsp.Proc())
		lser := quotaInitReport{r.lsp, cnt}
		copier := cpprocs.NewCopier(id, lser, proc)
		qman.AddResQuota(copier, r.quota)
	}
	return
}

type quotaInitReport struct {
	lser cpprocs.Lister
	cnt  *stats.Counter
}

func (r quotaInitReport) Ls() ([]cpprocs.LsEntry, error) {
	r.cnt.Quota.Init = true
	defer func() {
		r.cnt.Quota.Init = false
	}()
	return r.lser.Ls()
}

func readers(statsd *stats.Statsd, tmp *tmpdedup.Dir) (cps []cpprocs.Copier) {
	rems := remotes(tmp)
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
		// w := ansirefresh.NewWriter(ioutil.Discard)
		t := ansirefresh.NewWriteTicker(w, statsd, 500*time.Millisecond)
		defer t.Stop()
	}

	parity, err := procs.NewParity(ndata, nparity)
	if err != nil {
		return
	}

	tmp, err := tmpdedup.TempDir("")
	if err != nil {
		return
	}
	defer tmp.Finish()

	minCopies, err := mincopies.New(2, quotaMan(statsd, tmp))
	if err != nil {
		return
	}

	chain := procs.NewBacklog(10, procs.Chain{
		stats.NewProc(statsd, "checksum",
			procs.ChecksumProc,
		),
		stats.NewProc(statsd, "index",
			procs.NewIndex(os.Stdout),
		),
		stats.NewProc(statsd, "parity",
			parity.Proc(),
		),
		stats.NewProc(statsd, "compress",
			procs.NewGzip().Proc(),
		),
		stats.NewProc(statsd, "checksum2",
			procs.ChecksumProc,
		),
		procs.NewConcur(10, minCopies),
	})
	defer chain.Finish()

	splitter := split.NewSplitter(os.Stdin)
	err = procs.Process(chain, splitter)
	if err != nil {
		return
	}
	return chain.Finish()
}

func cmdJoin() (err error) {
	statsd := stats.New()
	{
		w := ansirefresh.NewWriter(os.Stderr)
		t := ansirefresh.NewWriteTicker(w, statsd, 500*time.Millisecond)
		defer t.Stop()
	}

	parity, err := procs.NewParity(ndata, nparity)
	if err != nil {
		return
	}

	tmp, err := tmpdedup.TempDir("")
	if err != nil {
		return
	}
	defer tmp.Finish()

	mrd, err := cpprocs.NewMultiReader(readers(statsd, tmp))
	if err != nil {
		return
	}

	chain := procs.NewBacklog(2, procs.Chain{
		stats.NewProc(statsd, "mrd",
			mrd,
		),
		stats.NewProc(statsd, "checksum",
			procs.ChecksumUnproc,
		),
		stats.NewProc(statsd, "compress",
			procs.NewGzip().Unproc(),
		),
		stats.NewProc(statsd, "group",
			procs.NewGroup(ndata+nparity),
		),
		stats.NewProc(statsd, "parity",
			parity.Unproc(),
		),
		procs.NewMutex(procs.Chain{
			stats.NewProc(statsd, "sort",
				procs.NewSort(),
			),
			stats.NewProc(statsd, "writerto",
				procs.NewWriterTo(os.Stdout),
			),
		}),
	})
	defer chain.Finish()

	scan := index.NewScanner(os.Stdin)
	err = procs.Process(chain, scan)
	if err != nil {
		return
	}
	return chain.Finish()
}
