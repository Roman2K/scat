package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"scat/cpprocs"
	"scat/cpprocs/mincopies"
	"scat/cpprocs/quota"
	"scat/procs"
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
		return cmdIndexSplit(args[1:])
	}
	return fmt.Errorf("unknown cmd: %s", cmd)
}

func cmdIndexSplit(args []string) error {
	const sconcur = "10"
	if len(args) < 2 {
		return errors.New("usage: <split args> [<proc> ...] <mincopies args>")
	}
	args = []string{"-", "" +
		"chain[" +
		"  split[" + args[0] + "]" +
		"  backlog[" + sconcur + " chain[" +
		"    checksum" +
		"    index[-]" +
		"    " + strings.Join(args[1:len(args)-1], " ") +
		"    checksum" +
		"    concur[" + sconcur + " mincopies[" + args[len(args)-1] + "]]" +
		"  ]]" +
		"]",
	}
	return cmdSplit(args)
}

func cmdSplit(args []string) (err error) {
	if len(args) < 2 {
		return errors.New("usage: <seed> <proc>")
	}
	proc, err := parseProc(args[1])
	if err != nil {
		return
	}
	seedrd, err := openIn(args[0])
	if err != nil {
		return
	}
	defer seedrd.Close()
	return nil
}

func openIn(path string) (io.ReadCloser, error) {
	if path == "-" {
		return ioutil.NopCloser(os.Stdin), nil
	}
	return os.Open(path)
}

func openOut(path string) (io.WriteCloser, error) {
	if path == "-" {
		return nopWriteCloser{os.Stdout}, nil
	}
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
}

type nopWriteCloser struct {
	io.Writer
}

func (nopWriteCloser) Close() error {
	return nil
}

func procParsers(tmp *tmpdedup.Dir) map[string]parser {
	return map[string]parser{
		"chain": parser{
			args: []arg{argVariadic{argProc}},
			build: func(procs []interface{}) (interface{}, error) {
				chain := make(procs.Chain, len(procs))
				for i, p := range procs {
					chain[i] = p
				}
				return chain, nil
			},
		},
		"split": parser{
			args: []arg{argBytes, argBytes},
			build: func(args []interface{}) (interface{}, error) {
				return procs.NewSplit(args[0].(uint64), args[1].(uint64)), nil
			},
		},
		"backlog": parser{
			args: []arg{argInt, argProc},
			build: func(args []interface{}) (interface{}, error) {
				return procs.NewBacklog(args[0].(int), args[1].(procs.Proc)), nil
			},
		},
		"checksum": parser{
			build: func([]interface{}) (interface{}, error) {
				return procs.ChecksumProc, nil
			},
		},
		"index": parser{
			args: []arg{argString},
			build: func(args []interface{}) (interface{}, error) {
				w, err := openOut(args[0].(string))
				if err != nil {
					return nil, err
				}
				return procs.NewIndexProc(w), nil
			},
		},
		"parity": parser{
			args: []arg{argInt, argInt},
			build: func(args []interface{}) (interface{}, error) {
				parity, err := procs.NewParity(args[0].(int), args[1].(int))
				if err != nil {
					return nil, err
				}
				return parity.Proc(), nil
			},
		},
		"gzip": parser{
			build: func([]interface{}) (interface{}, error) {
				return procs.NewGzip().Proc(), nil
			},
		},
		"concur": parser{
			args: []arg{argInt, argDynProcer},
			build: func(args []interface{}) (interface{}, error) {
				return procs.NewConcur(args[0].(int), args[1].(procs.DynProcer)), nil
			},
		},
		"mincopies": parser{
			args: []arg{argInt, argVariadic{argQuotaRes}},
			build: func(args []interface{}) (interface{}, error) {
				qman := quota.NewMan()
				for _, ires := range args[1:] {
					res := ires.(quotaRes)
					qman.AddResQuota(res.copier, res.max)
				}
				return mincopies.New(args[0].(int), qman)
			},
		},
		"quota": parser{
			args: []arg{argBytes, argCopier},
			build: func(args []interface{}) (interface{}, error) {
				res := quotaRes{
					max:    args[0].(uint64),
					copier: args[1].(cpprocs.Copier),
				}
				return res, nil
			},
		},
		"copier": parser{
			args: []arg{argString, argLsProc},
			build: func(args []interface{}) (interface{}, error) {
				lsp := args[1].(lsProc)
				return cpprocs.NewCopier(args[0].(string), lsp, lsp)
			},
		},
		"rclone": parser{
			args: []arg{argString},
			build: func(args []interface{}) (interface{}, error) {
				rclone := cpprocs.NewRclone(args[0].(string), tmp)
				return lsProc{rclone, rclone.Proc()}, nil
			},
		},
		"cat": parser{
			args: []arg{argString},
			build: func(args []interface{}) (interface{}, error) {
				cat := cpprocs.NewCopier(args[0].(string))
				return lsProc{cat, cat.Proc()}, nil
			},
		},
	}
}

type lsProc struct {
	cpprocs.Lister
	procs.Proc
}

type quotaRes struct {
	copier cpprocs.Copier
	max    uint64
}

type parser struct {
	args  []arg
	build func([]interface{}) (interface{}, error)
}
