package stripe

import "fmt"

type S map[Item]Locs
type Item interface{}
type Locs map[Loc]struct{}
type Loc interface{}

type Seq interface {
	Next() interface{}
}

// var for tests
var sortItems = func([]Item) {}

func (s S) Stripe(dests Locs, seq Seq, distinct, min int) (S, error) {
	items := make([]Item, 0, len(s))
	exist := make(S, len(s))
	prios := make(map[Loc]uint)
	for it, locs := range s {
		items = append(items, it)
		got := make(Locs, len(locs))
		for loc, _ := range locs {
			if _, ok := dests[loc]; !ok {
				continue
			}
			got[loc] = struct{}{}
			prios[loc]++
		}
		exist[it] = got
	}
	sortItems(items)
	res := make(S, len(items))
	for _, it := range items {
		got, ok := exist[it]
		if !ok {
			panic("invalid item")
		}
		newLocs := make(Locs, min)
		res[it] = newLocs
		old := make([]Loc, 0, len(got))
		for loc := range got {
			old = append(old, loc)
		}
		seen := make(Locs, len(dests))
		next := func() (Loc, error) {
			if len(old) > 0 {
				new := old[0]
				old = old[1:]
				return new, nil
			}
			new := seq.Next()
			if _, ok := dests[new]; !ok {
				return nil, nil
			}
			if _, ok := got[new]; !ok {
				if prio, ok := prios[new]; ok && prio > 0 {
					prios[new]--
					return nil, nil
				}
			}
			if _, ok := seen[new]; ok {
				err := ShortError{
					Distinct: distinct,
					Min:      min,
					Avail:    len(newLocs),
				}
				return nil, err
			}
			seen[new] = struct{}{}
			return new, nil
		}
		for len(newLocs) < min {
			for {
				new, err := next()
				if err != nil {
					return nil, err
				}
				if new == nil {
					continue
				}
				if _, ok := newLocs[new]; ok {
					continue
				}
				newLocs[new] = struct{}{}
				if len(newLocs) <= distinct && !res.exclusive(it) {
					delete(newLocs, new)
					continue
				}
				break
			}
		}
	}
	for it, new := range res {
		for loc := range exist[it] {
			delete(new, loc)
		}
	}
	return res, nil
}

func (s S) exclusive(it Item) bool {
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

type ShortError struct {
	Distinct, Min, Avail int
}

func (e ShortError) Error() string {
	return fmt.Sprintf("not enough target locations for"+
		" distinct=%d min=%d avail=%d",
		e.Distinct, e.Min, e.Avail,
	)
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

func NewLocs(locs ...Loc) (res Locs) {
	res = make(Locs, len(locs))
	for _, loc := range locs {
		res[loc] = struct{}{}
	}
	return
}
