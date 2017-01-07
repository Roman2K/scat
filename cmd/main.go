package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"time"

	humanize "github.com/dustin/go-humanize"

	ss "secsplit"
	"secsplit/aprocs"
	"secsplit/cpprocs"
	"secsplit/cpprocs/mincopies"
	"secsplit/index"
	"secsplit/procs"
	"secsplit/split"
	"secsplit/stats"
	"secsplit/tmpdedup"

	_ "net/http/pprof"
)

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

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

const (
	ndata   = 2
	nparity = 1
)

func cmdSplit() (err error) {
	in, out := os.Stdin, os.Stdout

	log := stats.NewLog(os.Stderr, 250*time.Millisecond)
	// log := stats.NewLog(ioutil.Discard, 250*time.Millisecond)

	parity, err := aprocs.NewParity(ndata, nparity)
	if err != nil {
		return
	}

	tmp, err := tmpdedup.TempDir("")
	if err != nil {
		return
	}
	defer tmp.Finish()

	copiers := copierList{}
	copiers.Add(log, "drive",
		cpprocs.NewRclone("drive:tmp", tmp),
		7*humanize.GiByte,
	)
	copiers.Add(log, "drive2",
		cpprocs.NewRclone("drive2:tmp", tmp),
		14*humanize.GiByte,
	)

	minCopies, err := mincopies.New(2, copiers)
	if err != nil {
		return
	}

	chain := aprocs.NewBacklog(10, aprocs.NewChain([]aprocs.Proc{
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
		aprocs.InplaceProcFunc(func(*ss.Chunk) error {
			time.Sleep(200 * time.Millisecond)
			return nil
		}),
		aprocs.NewConcur(10, minCopies),
		// stats.NewProc(log, "drive",
		// 	aprocs.NewPool(3, drive),
		// ),
	}))
	defer chain.Finish()

	splitter := split.NewSplitter(in)
	err = aprocs.Process(chain, splitter)
	if err != nil {
		return
	}
	return chain.Finish()
}

type copierList []cpprocs.Copier

func (cl *copierList) Add(
	log *stats.Log, id string, lsp cpprocs.LsProcUnprocer, quota uint64,
) {
	proc := stats.NewProc(log, id, lsp.Proc())
	copier := cpprocs.NewCopier(id, lsp, proc)
	copier.SetQuota(quota)
	*cl = append(*cl, copier)
}

func cmdJoin() (err error) {
	in, out := os.Stdin, os.Stdout

	log := stats.NewLog(os.Stderr, 250*time.Millisecond)

	parity, err := aprocs.NewParity(ndata, nparity)
	if err != nil {
		return
	}

	// cats := make([]cpprocs.Reader, 3)
	// for i, n := 0, len(cats); i < n; i++ {
	// 	id := fmt.Sprintf("cat%d", i+1)
	// 	cat := cpprocs.NewCat("/Users/roman/tmp/" + id)
	// 	proc := stats.NewProc(log, id, cat.Proc())
	// 	cats[i] = cpprocs.NewReader(id, cat, proc)
	// }

	// mrd, err := cpprocs.NewMultiReader(cats)
	// if err != nil {
	// 	return
	// }

	chain := aprocs.NewBacklog(2, aprocs.NewChain([]aprocs.Proc{
		// stats.NewProc(log, "localstore",
		// 	procs.A((&procs.LocalStore{"out"}).Unproc()),
		// ),
		// stats.NewProc(log, "cats",
		// 	mrd,
		// ),
		stats.NewProc(log, "checksum",
			procs.A(procs.Checksum{}.Unproc()),
		),
		stats.NewProc(log, "compress",
			procs.A((&procs.Compress{}).Unproc()),
		),
		stats.NewProc(log, "group",
			aprocs.NewGroup(ndata+nparity),
		),
		stats.NewProc(log, "parity",
			parity.Unproc(),
		),
		aprocs.NewMutex(aprocs.NewChain([]aprocs.Proc{
			stats.NewProc(log, "sort",
				aprocs.NewSort(),
			),
			stats.NewProc(log, "writerto",
				aprocs.NewWriterTo(out),
			),
		})),
	}))

	scan := index.NewScanner(in)
	err = aprocs.Process(chain, scan)
	if err != nil {
		return
	}
	return chain.Finish()
}
