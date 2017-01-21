package argproc

import (
	"fmt"
	"io"
	"os"

	ap "scat/argparse"
	"scat/cpprocs"
	"scat/cpprocs/mincopies"
	"scat/cpprocs/quota"
	"scat/procs"
	"scat/stats"
	"scat/tmpdedup"
)

func New(tmp *tmpdedup.Dir, stats *stats.Statsd) ap.Parser {
	return builder{tmp, stats}.argProc()
}

func NewArgChain(argProc ap.Parser) ap.Parser {
	return ap.ArgFilter{
		Parser: ap.ArgVariadic{argProc},
		Filter: func(val interface{}) (interface{}, error) {
			chain := newChain(val.([]interface{}))
			if len(chain) == 1 {
				if c, ok := chain[0].(procs.Chain); ok {
					return c, nil
				}
			}
			return chain, nil
		},
	}
}

type builder struct {
	tmp   *tmpdedup.Dir
	stats *stats.Statsd
}

func (b builder) argProc() ap.Parser {
	var (
		argProc    = ap.ArgFn{}              // procs.Procs
		argCpp     = b.newArgCpp()           // cpprocs.LsProcUnprocers
		argDynProc = b.newArgDynProc(argCpp) // procs.DynProcers
	)

	update(argProc, b.newArgProc(argProc, argDynProc, argCpp))

	return argProc
}

func (b builder) newArgProc(argProc, argDynp, argCpp ap.Parser) ap.ArgFn {
	return ap.ArgFn{
		"checksum": ap.ArgLambda{
			Run: func([]interface{}) (interface{}, error) {
				return procs.ChecksumProc, nil
			},
		},
		"uchecksum": ap.ArgLambda{
			Run: func([]interface{}) (interface{}, error) {
				return procs.ChecksumUnproc, nil
			},
		},
		"index": ap.ArgLambda{
			Args: ap.Args{ap.ArgStr},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					path = args[0].(string)
				)
				w, err := openOut(path)
				return procs.NewIndexProc(w), err
			},
		},
		"uindex": ap.ArgLambda{
			Run: func([]interface{}) (interface{}, error) {
				return procs.IndexUnproc, nil
			},
		},
		"split": ap.ArgLambda{
			Args: ap.Args{ap.ArgBytes, ap.ArgBytes},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					min = uintBytes(args[0])
					max = uintBytes(args[1])
				)
				return procs.NewSplit(min, max), nil
			},
		},
		"backlog": ap.ArgLambda{
			Args: ap.Args{ap.ArgInt, argProc},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					nslots = args[0].(int)
					proc   = args[1].(procs.Proc)
				)
				return procs.NewBacklog(nslots, proc), nil
			},
		},
		"pool": ap.ArgLambda{
			Args: ap.Args{ap.ArgInt, argProc},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					size = args[0].(int)
					proc = args[1].(procs.Proc)
				)
				return procs.NewPool(size, proc), nil
			},
		},
		"chain": ap.ArgLambda{
			Args: ap.ArgVariadic{argProc},
			Run: func(args []interface{}) (interface{}, error) {
				return newChain(args), nil
			},
		},
		"concur": ap.ArgLambda{
			Args: ap.Args{ap.ArgInt, argDynp},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					max  = args[0].(int)
					dynp = args[1].(procs.DynProcer)
				)
				return procs.NewConcur(max, dynp), nil
			},
		},
		"multireader": ap.ArgLambda{
			Args: ap.ArgVariadic{b.newArgCopier(argCpp, getUnproc)},
			Run: func(args []interface{}) (interface{}, error) {
				copiers := make([]cpprocs.Copier, len(args))
				for i, icp := range args {
					copiers[i] = icp.(cpprocs.Copier)
				}
				return cpprocs.NewMultiReader(copiers)
			},
		},
		"parity":  newArgParity(getProc),
		"uparity": newArgParity(getUnproc),
		"gzip":    newArgGzip(getProc),
		"ugzip":   newArgGzip(getUnproc),
		"mutex": ap.ArgLambda{
			Args: ap.Args{argProc},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					proc = args[0].(procs.Proc)
				)
				return procs.NewMutex(proc), nil
			},
		},
		"sort": ap.ArgLambda{
			Run: func([]interface{}) (interface{}, error) {
				return procs.NewSort(), nil
			},
		},
		"writerTo": ap.ArgLambda{
			Args: ap.Args{ap.ArgStr},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					path = args[0].(string)
				)
				w, err := openOut(path)
				return procs.NewWriterTo(w), err
			},
		},
		"join": ap.ArgLambda{
			Args: ap.Args{ap.ArgStr},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					path = args[0].(string)
				)
				w, err := openOut(path)
				return procs.NewJoin(w), err
			},
		},
		"group": ap.ArgLambda{
			Args: ap.Args{ap.ArgInt},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					size = args[0].(int)
				)
				return procs.NewGroup(size), nil
			},
		},
	}
}

func (b builder) newArgDynProc(argCpp ap.Parser) ap.ArgFn {
	return ap.ArgFn{
		"mincopies": ap.ArgLambda{
			Args: ap.Args{
				ap.ArgInt,
				ap.ArgVariadic{newArgQuota(b.newArgCopier(argCpp, getProc))},
			},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					min   = args[0].(int)
					iress = args[1].([]interface{})
				)
				qman := quota.NewMan()
				for _, ires := range iress {
					res := ires.(quotaRes)
					qman.AddResQuota(res.copier, res.max)
				}
				return mincopies.New(min, qman)
			},
		},
	}
}

func (b builder) newArgCpp() ap.ArgFn {
	return ap.ArgFn{
		"rclone": ap.ArgLambda{
			Args: ap.Args{ap.ArgStr},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					remote = args[0].(string)
				)
				return cpprocs.NewRclone(remote, b.tmp), nil
			},
		},
		"cat": ap.ArgLambda{
			Args: ap.Args{ap.ArgStr},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					dir = args[0].(string)
				)
				return cpprocs.NewCat(dir), nil
			},
		},
	}
}

func (b builder) newArgCopier(argCpp ap.Parser, getProc getProcFn) ap.Parser {
	return ap.ArgLambda{
		Args: ap.Args{ap.ArgStr, argCpp},
		Run: func(args []interface{}) (interface{}, error) {
			var (
				id  = args[0].(string)
				lsp = args[1].(cpprocs.LsProcUnprocer)
			)
			return cpprocs.NewCopier(id, lsp, getProc(lsp)), nil
		},
	}
}

func newChain(args []interface{}) (chain procs.Chain) {
	chain = make(procs.Chain, len(args))
	for i, p := range args {
		chain[i] = p.(procs.Proc)
	}
	return
}

func newArgParity(getProc getProcFn) ap.Parser {
	return ap.ArgLambda{
		Args: ap.Args{ap.ArgInt, ap.ArgInt},
		Run: func(args []interface{}) (interface{}, error) {
			var (
				ndata   = args[0].(int)
				nparity = args[1].(int)
			)
			parity, err := procs.NewParity(ndata, nparity)
			if err != nil {
				return nil, err
			}
			return getProc(parity), nil
		},
	}
}

func newArgGzip(getProc getProcFn) ap.Parser {
	return ap.ArgLambda{
		Run: func([]interface{}) (interface{}, error) {
			return getProc(procs.NewGzip()), nil
		},
	}
}

func newArgQuota(argCopier ap.Parser) ap.Parser {
	argQuotaMax := ap.ArgLambda{
		Args: ap.Args{argCopier, ap.ArgBytes},
		Run: func(args []interface{}) (interface{}, error) {
			qr := quotaRes{
				copier: args[0].(cpprocs.Copier),
				max:    args[1].(uint64),
			}
			return qr, nil
		},
	}
	return ap.ArgFilter{
		Parser: ap.ArgOr{argQuotaMax, argCopier},
		Filter: func(val interface{}) (interface{}, error) {
			if cp, ok := val.(cpprocs.Copier); ok {
				val = quotaRes{
					copier: cp,
					max:    quota.Unlimited,
				}
			}
			return val, nil
		},
	}
}

type getProcFn func(procs.ProcUnprocer) procs.Proc

func getProc(p procs.ProcUnprocer) procs.Proc {
	return p.Proc()
}

func getUnproc(p procs.ProcUnprocer) procs.Proc {
	return p.Unproc()
}

func update(dst, src ap.ArgFn) {
	for k, v := range src {
		dst[k] = v
	}
}

type quotaRes struct {
	max    uint64
	copier cpprocs.Copier
}

func uintBytes(val interface{}) uint {
	const uintMax = ^uint(0)
	i := val.(uint64)
	if i > uint64(uintMax) {
		panic(fmt.Errorf("uint64 greater than uint max: %d", i))
	}
	return uint(i)
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
