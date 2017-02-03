package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"gitlab.com/Roman2K/scat"
	"gitlab.com/Roman2K/scat/ansirefresh"
	"gitlab.com/Roman2K/scat/argparse"
	"gitlab.com/Roman2K/scat/argproc"
	"gitlab.com/Roman2K/scat/procs"
	"gitlab.com/Roman2K/scat/stats"
	"gitlab.com/Roman2K/scat/tmpdedup"
)

//go:generate ../tools/genversion VERSION _version.go version.go

const url = "https://gitlab.com/Roman2K/scat#usage"

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		if exit, ok := err.(*exec.ExitError); ok {
			fmt.Fprintf(os.Stderr, "stderr=%q\n", exit.Stderr)
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
	seed := scat.NewChunk(0, scat.NewReaderData(os.Stdin))

	return procs.Process(proc, seed)
}

type cmdArgs struct {
	procStr string
	stats   bool
	version bool
}

func (a *cmdArgs) Parse(args []string) {
	name := "<command>"
	if len(args) > 0 {
		name, args = args[0], args[1:]
	}
	fl := flag.NewFlagSet(name, flag.ContinueOnError)
	fl.BoolVar(&a.stats, "stats", false, "print stats: data rates, quotas, etc.")
	fl.BoolVar(&a.version, "version", false, "show version")
	fl.SetOutput(ioutil.Discard)
	usage := func(w io.Writer) {
		fmt.Fprintf(w, "usage: %s [options] <proc>\n", name)
		fmt.Fprintln(w)
		fmt.Fprintf(w, "\t<proc>\tproc string\n")
		fmt.Fprintf(w, "\t\tsee %s\n", url)
		fmt.Fprintln(w)
		fmt.Fprintf(w, "options:\n")
		fl.SetOutput(w)
		defer fl.SetOutput(ioutil.Discard)
		fl.PrintDefaults()
		fmt.Fprintln(w)
		fmt.Fprintf(w, "see %s\n", url)
	}
	err := fl.Parse(args)
	if err != nil || (fl.NArg() != 1 && !a.version) {
		w, code := os.Stderr, 2
		if err == flag.ErrHelp {
			w, code = os.Stdout, 0
		}
		usage(w)
		os.Exit(code)
	}
	a.procStr = fl.Arg(0)
}
