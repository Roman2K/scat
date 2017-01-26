package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type cmdArgs struct {
	seedPath string
	procStr  string
	stats    bool
}

func (a *cmdArgs) Parse(args []string) {
	name := "<command>"
	if len(args) > 0 {
		name, args = args[0], args[1:]
	}
	fl := flag.NewFlagSet(name, flag.ContinueOnError)
	fl.BoolVar(&a.stats, "stats", true, "proc stats: data rate, quota, etc.")
	fl.SetOutput(ioutil.Discard)
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
		fmt.Fprintf(w, "\t\tex: chain[gzip join[-]]\n")
		fmt.Fprintf(w, "\t\tex: gzip join[-]\n")
		fmt.Fprintln(w)
		fmt.Fprintf(w, "options:\n")
		fl.SetOutput(w)
		defer fl.SetOutput(ioutil.Discard)
		fl.PrintDefaults()
	}
	err := fl.Parse(args)
	if err != nil || fl.NArg() != 2 {
		w, code := os.Stderr, 2
		if err == flag.ErrHelp {
			w, code = os.Stdout, 0
		}
		usage(w)
		os.Exit(code)
	}
	a.seedPath = fl.Arg(0)
	a.procStr = fl.Arg(1)
}
