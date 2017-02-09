package stripe

import "errors"

var ErrShort = errors.New("not enough target locations to satisfy requirements")

type S map[item]Locs
type item interface{}
type Locs map[loc]struct{}
type loc interface{}

func NewLocs(locs ...loc) (res Locs) {
	res = make(Locs, len(locs))
	for _, loc := range locs {
		res[loc] = struct{}{}
	}
	return
}

func (locs Locs) Add(loc loc) {
	locs[loc] = struct{}{}
}

type Seq interface {
	Next() interface{}
}

// var for tests
var sortItems = func([]item) {}

func (s S) Stripe(dests Locs, seq Seq, distinct, min int) (S, error) {
	items := make([]item, 0, len(s))
	prios := make(map[loc]int)
	for it, got := range s {
		items = append(items, it)
		for loc := range got {
			prios[loc]++
		}
	}
	sortItems(items)
	res := make(S, len(items))
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
		newLocs := make(Locs, min)
		res[it] = newLocs
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
			newLocs.Add(new)
			if len(newLocs) <= distinct && !res.exclusive(it) {
				delete(newLocs, new)
				continue
			}
		}
	}
	for it, got := range s {
		for old := range got {
			delete(res[it], old)
		}
	}
	return res, nil
}

func (s S) exclusive(it item) bool {
	locs, ok := s[it]
	if !ok {
		return true
	}
	for it2, otherLocs := range s {
		if it2 == it {
			continue
		}
		a, b := locs, otherLocs
		if len(b) < len(a) {
			a, b = b, a
		}
		for loc := range a {
			if _, ok := b[loc]; ok {
				return false
			}
		}
	}
	return true
}

type Striper interface {
	Stripe(S, Locs, Seq) (S, error)
}

type Config struct {
	Distinct, Min int
}

var _ Striper = Config{}

func (cfg Config) Stripe(s S, locs Locs, seq Seq) (S, error) {
	return s.Stripe(locs, seq, cfg.Distinct, cfg.Min)
}
