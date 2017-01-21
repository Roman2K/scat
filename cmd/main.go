package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"scat"
	"time"

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

	name, args := os.Args[0], os.Args[1:]
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.SetOutput(ioutil.Discard)
	usage := func(w io.Writer) {
		fmt.Fprintf(w, "usage: %s [options] <seed> <proc>\n", name)
		fmt.Fprintln(w)
		fmt.Fprintf(w, "\t<seed>\tpath to data of seed chunk\n")
		fmt.Fprintln(w)
		fmt.Fprintf(w, "\t\tex: -\n")
		fmt.Fprintf(w, "\t\tex: path/to/file\n")
		fmt.Fprintln(w)
		fmt.Fprintf(w, "\t<proc>\tproc string\n")
		fmt.Fprintln(w)
		fmt.Fprintf(w, "\t\tex: chain[gzip writerTo[-]]\n")
		fmt.Fprintf(w, "\t\tex: gzip writerTo[-]\n")
		fmt.Fprintln(w)
		fmt.Fprintf(w, "options:\n")
		flags.SetOutput(w)
		defer flags.SetOutput(ioutil.Discard)
		flags.PrintDefaults()
	}
	err = flags.Parse(args)
	if err != nil || flags.NArg() != 2 {
		w, code := os.Stderr, 2
		if err == flag.ErrHelp {
			w, code = os.Stdout, 0
		}
		usage(w)
		os.Exit(code)
	}
	var (
		seedPath = flags.Arg(0)
		procStr  = flags.Arg(1)
	)

	tmp, err := tmpdedup.TempDir("")
	if err != nil {
		return
	}
	defer tmp.Finish()

	statsd := stats.New()
	{
		w := ansirefresh.NewWriter(os.Stderr)
		// w := ansirefresh.NewWriter(ioutil.Discard)
		t := ansirefresh.NewWriteTicker(w, statsd, 500*time.Millisecond)
		defer t.Stop()
	}

	argProc := argproc.NewArgChain(argproc.New(tmp, statsd))
	res, _, err := argparse.Args{argProc}.Parse(procStr)
	if err != nil {
		return
	}

	proc := res.([]interface{})[0].(procs.Proc)
	seedrd, err := openIn(seedPath)
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
