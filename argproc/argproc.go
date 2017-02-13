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

var chainBrackets = ap.Brackets{'{', '}'}

func New(tmp *tmpdedup.Dir, stats *stats.Statsd) ap.Parser {
	argProc := builder{tmp, stats}.argProc()
	return ap.ArgFilter{
		Parser: ap.ArgPiped{Arg: argProc, Nest: chainBrackets},
		Filter: func(val interface{}) (interface{}, error) {
			return newChain(val.([]interface{})), nil
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

type builder struct {
	tmp   *tmpdedup.Dir
	stats *stats.Statsd
}

func (b builder) argProc() ap.Parser {
	var (
		argProc    = make(ap.ArgOr, 2)         // procs.Procs
		argStore   = b.newArgStore()           // stores.Stores
		argDynProc = b.newArgDynProc(argStore) // procs.DynProcers
	)

	argChain := ap.ArgLambda{
		Open:  chainBrackets.Open,
		Close: chainBrackets.Close,
		Args:  ap.ArgPiped{Arg: argProc, Nest: chainBrackets},
		Run: func(args []interface{}) (interface{}, error) {
			return newChain(args), nil
		},
	}

	procFns := b.newArgProc(argProc, argDynProc, argStore)
	for k, v := range argStore {
		procFns[k] = newArgStoreProc(v, getProc)
		procFns["u"+k] = newArgStoreProc(v, getUnproc)
	}
	if b.stats != nil {
		for k, v := range procFns {
			procFns[k] = b.newArgStatsProc(v, k)
		}
	}

	argProc[0] = argChain
	argProc[1] = procFns
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
	newS := func(min, excl int, iress []interface{}) (procs.DynProcer, error) {
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
		cfg := stripe.Config{Min: min, Excl: excl}
		return storestripe.New(cfg, qman)
	}
	argQuota := b.newArgQuota(b.newArgCopier(argStore, getProc))
	return ap.ArgFn{
		"mincopies": ap.ArgLambda{
			Args: ap.Args{
				ap.ArgInt,
				ap.ArgVariadic{argQuota},
			},
			Run: func(args []interface{}) (interface{}, error) {
				const excl = 0
				var (
					min   = args[0].(int)
					iress = args[1].([]interface{})
				)
				return newS(min, excl, iress)
			},
		},
		"stripe": ap.ArgLambda{
			Args: ap.Args{
				ap.ArgInt,
				ap.ArgInt,
				ap.ArgVariadic{argQuota},
			},
			Run: func(args []interface{}) (interface{}, error) {
				var (
					min   = args[0].(int)
					excl  = args[1].(int)
					iress = args[2].([]interface{})
				)
				return newS(min, excl, iress)
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
	return ap.ArgPair{
		Left:  ap.ArgStr,
		Right: argStore,
		Run: func(iid, istore interface{}) (interface{}, error) {
			var (
				id    = iid.(string)
				store = istore.(stores.Store)
			)
			var (
				lser stores.Lister = store
				proc procs.Proc    = getProc(store)
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
	argQuotaMax := ap.ArgPair{
		Left:  argCopier,
		Right: ap.ArgBytes,
		Run: func(icp, ibytes interface{}) (interface{}, error) {
			qr := quotaRes{
				copier: icp.(stores.Copier),
				max:    ibytes.(uint64),
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
