package stripe

import (
	"errors"
	"sync"

	"gitlab.com/Roman2K/scat"
	"gitlab.com/Roman2K/scat/checksum"
	"gitlab.com/Roman2K/scat/concur"
	"gitlab.com/Roman2K/scat/procs"
	"gitlab.com/Roman2K/scat/stores"
	"gitlab.com/Roman2K/scat/stores/copies"
	"gitlab.com/Roman2K/scat/stores/quota"
	"gitlab.com/Roman2K/scat/stripe"
)

type stripeP struct {
	cfg    stripe.Striper
	qman   *quota.Man
	reg    *copies.Reg
	seq    stripe.Seq
	seqMu  sync.Mutex
	finish func() error
}

func New(cfg stripe.Striper, qman *quota.Man) (procs.DynProcer, error) {
	reg := copies.NewReg()
	ress := copiersRes(qman.Resources(0))
	ids := ress.ids()
	rrItems := make([]interface{}, len(ids))
	for i, id := range ids {
		rrItems[i] = id
	}
	seq := &stripe.RR{Items: rrItems}
	ml := stores.MultiLister(ress.listers())
	err := ml.AddEntriesTo([]stores.LsEntryAdder{
		stores.QuotaEntryAdder{Qman: qman},
		stores.CopiesEntryAdder{Reg: reg},
	})
	dynp := &stripeP{
		cfg:    cfg,
		qman:   qman,
		reg:    reg,
		seq:    seq,
		finish: ress.finishFuncs().FirstErr,
	}
	return dynp, err
}

func (sp *stripeP) Procs(chunk *scat.Chunk) ([]procs.Proc, error) {
	type chunkInfo struct {
		chunk    *scat.Chunk
		quotaUse uint64
	}
	group, ok := procs.GetGroup(chunk)
	if !ok {
		group = []*scat.Chunk{chunk}
	}
	chunks := map[checksum.Hash]chunkInfo{}
	quotaUse := uint64(0)
	for _, c := range group {
		if err, ok := procs.GetGroupErr(c); ok && err != nil {
			return nil, err
		}
		qUse, err := calcQuotaUse(c.Data())
		if err != nil {
			return nil, err
		}
		inf := chunkInfo{
			chunk:    c,
			quotaUse: qUse,
		}
		chunks[c.Hash()] = inf
		quotaUse += qUse
	}
	curStripe := make(stripe.S, len(chunks))
	for hash := range chunks {
		copies := sp.reg.List(hash)
		copies.Mu.Lock()
		owners := copies.Owners()
		locs := make(stripe.Locs, len(owners))
		for _, o := range owners {
			locs.Add(o.Id())
		}
		curStripe[hash] = locs
	}
	all := copiersRes(sp.qman.Resources(quotaUse)).copiersById()
	dests := make(stripe.Locs, len(all))
	for _, cp := range all {
		dests.Add(cp.Id())
	}
	sp.seqMu.Lock()
	newStripe, err := sp.cfg.Stripe(curStripe, dests, sp.seq)
	sp.seqMu.Unlock()
	if err != nil {
		return nil, err
	}
	nprocs := 0
	for _, locs := range newStripe {
		nprocs += len(locs)
	}
	cpProcs := make([]procs.Proc, 1, nprocs+1)
	{
		proc := make(sliceProc, 0, len(chunks))
		for _, ci := range chunks {
			proc = append(proc, ci.chunk)
		}
		cpProcs[0] = proc
	}
	for item, locs := range newStripe {
		hash := item.(checksum.Hash)
		ci, ok := chunks[hash]
		if !ok {
			panic("unknown chunk hash")
		}
		copies := sp.reg.List(hash)
		cProcs := make([]procs.Proc, 0, len(locs))
		wg := sync.WaitGroup{}
		wg.Add(cap(cProcs))
		go func() {
			defer copies.Mu.Unlock()
			wg.Wait()
		}()
		for id := range locs {
			copier, ok := all[id]
			if !ok {
				panic("unknown copier ID")
			}
			var proc procs.Proc = copier
			proc = chunkArgProc{proc, ci.chunk}
			proc = procs.DiscardChunks{proc}
			proc = procs.OnEnd{proc, func(err error) {
				defer wg.Done()
				if err != nil {
					sp.qman.Delete(copier)
					return
				}
				copies.Add(copier)
				sp.qman.AddUse(copier, ci.quotaUse)
			}}
			cProcs = append(cProcs, proc)
		}
		cpProcs = append(cpProcs, cProcs...)
	}
	return cpProcs, nil
}

func calcQuotaUse(d scat.Data) (uint64, error) {
	sz, ok := d.(scat.Sizer)
	if !ok {
		return 0, errors.New("sized-data required for calculating data use")
	}
	return uint64(sz.Size()), nil
}

func (sp *stripeP) Finish() error {
	return sp.finish()
}

type sliceProc []*scat.Chunk

func (s sliceProc) Process(*scat.Chunk) <-chan procs.Res {
	ch := make(chan procs.Res, len(s))
	defer close(ch)
	for _, c := range s {
		ch <- procs.Res{Chunk: c}
	}
	return ch
}

func (s sliceProc) Finish() error {
	return nil
}

type chunkArgProc struct {
	procs.Proc
	chunk *scat.Chunk
}

func (p chunkArgProc) Process(*scat.Chunk) <-chan procs.Res {
	return p.Proc.Process(p.chunk)
}

type copiersRes []quota.Res

func (ress copiersRes) listers() (lsers []stores.Lister) {
	lsers = make([]stores.Lister, len(ress))
	for i, res := range ress {
		lsers[i] = res.(stores.Lister)
	}
	return
}

func (ress copiersRes) ids() (ids []interface{}) {
	ids = make([]interface{}, len(ress))
	for i, res := range ress {
		ids[i] = res.Id()
	}
	return
}

func (ress copiersRes) finishFuncs() (fns concur.Funcs) {
	fns = make(concur.Funcs, len(ress))
	for i, res := range ress {
		fns[i] = res.(procs.Proc).Finish
	}
	return
}

func (ress copiersRes) copiersById() map[interface{}]stores.Copier {
	cps := make(map[interface{}]stores.Copier, len(ress))
	for _, res := range ress {
		cps[res.Id()] = res.(stores.Copier)
	}
	return cps
}
