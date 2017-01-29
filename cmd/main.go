package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"scat"
	"scat/ansirefresh"
	"scat/argparse"
	"scat/argproc"
	"scat/procs"
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

func start() (err error) {
	rand.Seed(time.Now().UnixNano())

	args := cmdArgs{}
	args.Parse(os.Args)

	if args.version {
		fmt.Println(version)
		return
	}

	tmp, err := tmpdedup.TempDir("")
	if err != nil {
		return
	}
	defer tmp.Finish()

	var statsd *stats.Statsd
	if args.stats {
		statsd = stats.New()
		{
			w := ansirefresh.NewWriter(os.Stderr)
			// w := ansirefresh.NewWriter(ioutil.Discard)
			t := ansirefresh.NewWriteTicker(w, statsd, 500*time.Millisecond)
			defer t.Stop()
		}
	}

	argProc := argproc.NewArgChain(argproc.New(tmp, statsd))
	res, _, err := argparse.Args{argProc}.Parse(args.procStr)
	if err != nil {
		return
	}

	proc := res.([]interface{})[0].(procs.Proc)
	seedrd, err := openIn(args.seedPath)
	if err != nil {
		return
	}
	defer seedrd.Close()

	seed := scat.NewChunk(0, scat.NewReaderData(seedrd))
	return procs.Process(proc, seed)
}

func openIn(path string) (io.ReadCloser, error) {
	if path == "-" {
		return ioutil.NopCloser(os.Stdin), nil
	}
	return os.Open(path)
}
