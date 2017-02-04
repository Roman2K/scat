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
	for it, locs := range s {
		items = append(items, it)
		got := make(Locs, len(locs))
		for loc, _ := range locs {
			if _, ok := dests[loc]; !ok {
				continue
			}
			got[loc] = struct{}{}
		}
		exist[it] = got
	}
	sortItems(items)
	res := make(S, len(items))
	for _, it := range items {
		locs, ok := exist[it]
		if !ok {
			panic("invalid item")
		}
		missing := min - len(locs)
		if missing <= 0 {
			continue
		}
		newLocs := make(Locs, missing)
		for i := 0; i < missing; i++ {
			seen := make(Locs, len(dests))
			for {
				new := seq.Next()
				if _, ok := seen[new]; ok {
					err := ShortError{
						Distinct: distinct, Min: min,
						Missing: missing, Avail: i,
					}
					return nil, err
				}
				seen[new] = struct{}{}
				if _, ok := dests[new]; !ok {
					continue
				}
				if _, ok := locs[new]; ok {
					continue
				}
				locs[new] = struct{}{}
				if len(locs) <= distinct && !exist.exclusive(it) {
					delete(locs, new)
					continue
				}
				newLocs[new] = struct{}{}
				break
			}
		}
		res[it] = newLocs
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
	Distinct, Min, Missing, Avail int
}

func (e ShortError) Error() string {
	return fmt.Sprintf("not enough target locations for"+
		" distinct=%d min=%d missing=%d avail=%d",
		e.Distinct, e.Min, e.Missing, e.Avail,
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
