package stripe

import "errors"

var ErrShort = errors.New("not enough target locations to satisfy requirements")

type S map[item]Locs
type item interface{}
type Locs map[loc]struct{}
type loc interface{}

func (locs Locs) Add(loc loc) {
	locs[loc] = struct{}{}
}

type Seq interface {
	Next() interface{}
}

// var for tests
var sortItems = func([]item) {}

func (s S) Stripe(dests Locs, seq Seq, excl, min int) (S, error) {
	items := make([]item, 0, len(s))
	prios := make(map[loc]int)
	for it, got := range s {
		items = append(items, it)
		for loc := range got {
			if _, ok := dests[loc]; !ok {
				continue
			}
			prios[loc]++
		}
	}
	sortItems(items)
	res := make(S, len(items))
	for pass := 0; pass < 2; pass++ {
		for _, it := range items {
			got, ok := s[it]
			if !ok {
				panic("unknown item")
			}
			old := make(Locs, len(got))
			for k, v := range got {
				old[k] = v
			}
			seen := make(Locs, len(dests))
			next := func() (loc, error) {
				for loc := range old {
					delete(old, loc)
					return loc, nil
				}
				new := seq.Next()
				if _, ok := seen[new]; ok {
					delete(prios, new)
					if len(prios) > 0 {
						return nil, nil
					}
					return nil, ErrShort
				}
				seen.Add(new)
				return new, nil
			}
			newLocs, ok := res[it]
			if !ok {
				newLocs = make(Locs, min)
				res[it] = newLocs
			}
			for len(newLocs) < min {
				new, err := next()
				if err != nil {
					return nil, err
				}
				if new == nil {
					continue
				}
				if _, ok := got[new]; !ok {
					if prio, ok := prios[new]; ok {
						if prio > 0 {
							prios[new]--
							delete(seen, new)
							continue
						}
						delete(prios, new)
					}
				}
				if _, ok := newLocs[new]; ok {
					continue
				}
				if _, ok := dests[new]; !ok {
					continue
				}
				nexcl := excl
				if max := len(res); nexcl > max {
					nexcl = max
				}
				newLocs.Add(new)
				if res.exclusives() < nexcl {
					delete(newLocs, new)
					if pass == 0 {
						break
					}
				}
			}
		}
	}
	for it, got := range s {
		new := res[it]
		for old := range got {
			delete(new, old)
		}
	}
	return res, nil
}

func (s S) exclusives() (count int) {
	counts := map[loc]uint{}
	for _, locs := range s {
		for loc := range locs {
			counts[loc]++
		}
	}
	for _, locs := range s {
		excl := true
		for loc := range locs {
			if counts[loc] > 1 {
				excl = false
				break
			}
		}
		if excl {
			count++
		}
	}
	return
}

type Striper interface {
	Stripe(S, Locs, Seq) (S, error)
}

type Config struct {
	Excl, Min int
}

var _ Striper = Config{}

func (cfg Config) Stripe(s S, locs Locs, seq Seq) (S, error) {
	return s.Stripe(locs, seq, cfg.Excl, cfg.Min)
}
