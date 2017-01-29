package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const url = "https://github.com/Roman2K/scat#usage"

type cmdArgs struct {
	seedPath string
	procStr  string
	stats    bool
	version  bool
}

func (a *cmdArgs) Parse(args []string) {
	name := "<command>"
	if len(args) > 0 {
		name, args = args[0], args[1:]
	}
	fl := flag.NewFlagSet(name, flag.ContinueOnError)
	fl.BoolVar(&a.stats, "stats", false, "print stats: data rates, quotas, etc.")
	fl.BoolVar(&a.version, "version", false, "print version and exit")
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
		fmt.Fprintf(w, "\t\tex: split chain[checksum index[-] cat[my_dir]]\n")
		fmt.Fprintf(w, "\t\tex: uindex ucat[my_dir] uchecksum join[-]\n")
		fmt.Fprintln(w)
		fmt.Fprintf(w, "options:\n")
		fl.SetOutput(w)
		defer fl.SetOutput(ioutil.Discard)
		fl.PrintDefaults()
		fmt.Fprintln(w)
		fmt.Fprintf(w, "see %s\n", url)
	}
	err := fl.Parse(args)
	if err != nil || (fl.NArg() != 2 && !a.version) {
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
