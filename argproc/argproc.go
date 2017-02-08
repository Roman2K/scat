package argproc

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"gitlab.com/Roman2K/scat"
	ap "gitlab.com/Roman2K/scat/argparse"
	"gitlab.com/Roman2K/scat/procs"
	"gitlab.com/Roman2K/scat/stats"
	"gitlab.com/Roman2K/scat/stores"
	"gitlab.com/Roman2K/scat/stores/quota"
	storestripe "gitlab.com/Roman2K/scat/stores/stripe"
	"gitlab.com/Roman2K/scat/stripe"
	"gitlab.com/Roman2K/scat/tmpdedup"
)

func New(tmp *tmpdedup.Dir, stats *stats.Statsd) ap.Parser {
	return builder{tmp, stats}.argProc()
}

func NewArgChain(argProc ap.Parser) ap.Parser {
	return ap.ArgFilter{
		Parser: ap.ArgVariadic{argProc},
		Filter: func(val interface{}) (interface{}, error) {
			slice := procSlice(val.([]interface{}))
			if len(slice) == 1 {
				return slice[0], nil
			}
			return procs.Chain(slice), nil
		},
	}
}

func procSlice(args []interface{}) (slice []procs.Proc) {
	slice = make([]procs.Proc, len(args))
	for i, p := range args {
		slice[i] = p.(procs.Proc)
	}
	return
}

type builder struct {
	tmp   *tmpdedup.Dir
	stats *stats.Statsd
}

func (b builder) argProc() ap.Parser {
	var (
		argProc    = ap.ArgFn{}                // procs.Procs
		argStore   = b.newArgStore()           // stores.Stores
		argDynProc = b.newArgDynProc(argStore) // procs.DynProcers
	)

	update(argProc, b.newArgProc(argProc, argDynProc, argStore))

	for k, v := range argStore {
		argProc[k] = newArgStoreProc(v, getProc)
		argProc["u"+k] = newArgStoreProc(v, getUnproc)
	}

	if b.stats != nil {
		for k, v := range argProc {
			argProc[k] = b.newArgStatsProc(v, k)
		}
	}

	return argProc
}

func newArgStoreProc(argStore ap.Parser, getProc getProcFn) ap.Parser {
	return ap.ArgFilter{
		Parser: argStore,
		Filter: func(val interface{}) (interface{}, error) {
			store := val.(stores.Store)
			return getProc(store), nil
		},
	}
}

func (b builder) newArgStatsProc(argProc ap.Parser, id interface{}) ap.Parser {
	return ap.ArgFilter{
		Parser: argProc,
		Filter: func(val interface{}) (interface{}, error) {
			proc := val.(procs.Proc)
			return stats.Proc{b.stats, id, proc}, nil
		},
	}
}

func (b builder) newArgProc(argProc, argDynp, argStore ap.Parser) ap.ArgFn {
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
			Run: func([]interface{}) (interface{}, error) {
				return procs.Split, nil
			},
		},
		"split2": ap.ArgLambda{
			Args: ap.Args{ap.ArgBytes, ap.ArgBytes},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					min = uintBytes(args[0])
					max = uintBytes(args[1])
				)
				return procs.NewSplitSize(min, max), nil
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
		"chain": ap.ArgLambda{
			Args: ap.ArgVariadic{argProc},
			Run: func(args []interface{}) (interface{}, error) {
				return procs.Chain(procSlice(args)), nil
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
			Args: ap.ArgVariadic{b.newArgCopier(argStore, getUnproc)},
			Run: func(args []interface{}) (interface{}, error) {
				copiers := make([]stores.Copier, len(args))
				for i, icp := range args {
					copiers[i] = icp.(stores.Copier)
				}
				return stores.NewMultiReader(copiers)
			},
		},
		"parity":  newArgParity(getProc),
		"uparity": newArgParity(getUnproc),
		"gzip":    newArgGzip(getProc),
		"ugzip":   newArgGzip(getUnproc),
		"sort": ap.ArgLambda{
			Run: func([]interface{}) (interface{}, error) {
				return &procs.Sort{}, nil
			},
		},
		"write": ap.ArgLambda{
			Args: ap.Args{ap.ArgStr},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					path = args[0].(string)
				)
				w, err := openOut(path)
				return procs.WriterTo{w}, err
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
		"cmd": newArgCmdProc(func(fn procs.CmdFunc) procs.Proc {
			return fn
		}),
		"cmdin": newArgCmdProc(func(fn procs.CmdFunc) procs.Proc {
			return procs.CmdInFunc(fn)
		}),
		"cmdout": newArgCmdProc(func(fn procs.CmdFunc) procs.Proc {
			return procs.CmdOutFunc(fn)
		}),
	}
}

func (b builder) newArgDynProc(argStore ap.Parser) ap.ArgFn {
	return ap.ArgFn{
		"stripe": ap.ArgLambda{
			Args: ap.Args{
				ap.ArgInt,
				ap.ArgInt,
				ap.ArgVariadic{b.newArgQuota(b.newArgCopier(argStore, getProc))},
			},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					distinct = args[0].(int)
					min      = args[1].(int)
					iress    = args[2].([]interface{})
				)
				qman := quota.NewMan()
				if b.stats != nil {
					qman.OnUse = func(res quota.Res, use, max uint64) {
						cnt := b.stats.Counter(res.Id())
						cnt.Quota.Use = use
						cnt.Quota.Max = max
					}
				}
				for _, ires := range iress {
					res := ires.(quotaRes)
					qman.AddResQuota(res.copier, res.max)
				}
				cfg := stripe.Config{
					Distinct: distinct,
					Min:      min,
				}
				return storestripe.New(cfg, qman)
			},
		},
	}
}

func (b builder) newArgStore() ap.ArgFn {
	newDir := func(args []interface{}) stores.Dir {
		var (
			path    = args[0].(string)
			nesting = args[1].([]interface{})
		)
		part := make(stores.StrPart, len(nesting))
		for i, n := range nesting {
			part[i] = n.(int)
		}
		return stores.Dir{path, part}
	}
	return ap.ArgFn{
		"rclone": ap.ArgLambda{
			Args: ap.Args{ap.ArgStr},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					remote = args[0].(string)
				)
				return stores.Rclone{remote, b.tmp}, nil
			},
		},
		"cp": ap.ArgLambda{
			Args: ap.Args{ap.ArgStr, ap.ArgVariadic{ap.ArgInt}},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					dir = newDir(args)
				)
				return stores.Cp(dir), nil
			},
		},
		"scp": ap.ArgLambda{
			Args: ap.Args{ap.ArgStr, ap.ArgStr, ap.ArgVariadic{ap.ArgInt}},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					host = args[0].(string)
					dir  = newDir(args[1:])
				)
				return stores.NewScp(host, dir), nil
			},
		},
	}
}

func (b builder) newArgCopier(argStore ap.Parser, getProc getProcFn) ap.Parser {
	return ap.ArgLambda{
		Args: ap.Args{ap.ArgStr, argStore},
		Run: func(args []interface{}) (interface{}, error) {
			var (
				id  = args[0].(string)
				lsp = args[1].(stores.Store)
			)
			var (
				lser stores.Lister = lsp
				proc procs.Proc    = getProc(lsp)
			)
			if b.stats != nil {
				lser = quotaInitReport{
					lser:       lser,
					getCounter: func() *stats.Counter { return b.stats.Counter(id) },
				}
				proc = stats.Proc{b.stats, id, proc}
			}
			return stores.Copier{id, lser, proc}, nil
		},
	}
}

func (b builder) newArgQuota(argCopier ap.Parser) ap.Parser {
	argQuotaMax := ap.ArgLambda{
		Args: ap.Args{argCopier, ap.ArgBytes},
		Run: func(args []interface{}) (interface{}, error) {
			qr := quotaRes{
				copier: args[0].(stores.Copier),
				max:    args[1].(uint64),
			}
			return qr, nil
		},
	}
	argRes := ap.ArgFilter{
		Parser: ap.ArgOr{argQuotaMax, argCopier},
		Filter: func(val interface{}) (interface{}, error) {
			if cp, ok := val.(stores.Copier); ok {
				val = quotaRes{
					copier: cp,
					max:    quota.Unlimited,
				}
			}
			return val, nil
		},
	}
	if b.stats == nil {
		return argRes
	}
	return ap.ArgFilter{
		Parser: argRes,
		Filter: func(val interface{}) (interface{}, error) {
			res := val.(quotaRes)
			cnt := b.stats.Counter(res.copier.Id())
			cnt.Quota.Max = res.max
			return val, nil
		},
	}
}

type quotaInitReport struct {
	lser       stores.Lister
	getCounter func() *stats.Counter
}

func (r quotaInitReport) Ls() ([]stores.LsEntry, error) {
	cnt := r.getCounter()
	cnt.Quota.Init = true
	defer func() {
		cnt.Quota.Init = false
	}()
	return r.lser.Ls()
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
			return getProc(procs.Gzip{}), nil
		},
	}
}

func newArgCmdProc(getProc func(procs.CmdFunc) procs.Proc) ap.Parser {
	return ap.ArgLambda{
		Args: ap.Args{ap.ArgStr, ap.ArgVariadic{ap.ArgStr}},
		Run: func(args []interface{}) (interface{}, error) {
			var (
				name     = args[0].(string)
				icmdArgs = args[1].([]interface{})
			)
			cmdArgs := make([]string, len(icmdArgs))
			for i, a := range icmdArgs {
				cmdArgs[i] = a.(string)
			}
			newCmd := func(*scat.Chunk) (*exec.Cmd, error) {
				return exec.Command(name, cmdArgs...), nil
			}
			fn := procs.CmdFunc(newCmd)
			return getProc(fn), nil
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
	copier stores.Copier
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
