package stripe

import (
	"sort"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestStripe(t *testing.T) {
	origSortItems := sortItems
	defer func() {
		sortItems = origSortItems
	}()
	sortItems = func(items []Item) {
		str := func(i int) string {
			return items[i].(string)
		}
		sort.Slice(items, func(i, j int) bool {
			return str(i) < str(j)
		})
	}

	seq := &RR{Items: []interface{}{"a", "b", "c"}}
	dests := NewLocs("a", "b", "c")

	s := S{}
	res, err := s.Stripe(dests, seq, 0, 1)
	assert.NoError(t, err)
	assert.Equal(t, S{}, res)

	seq.Reset()

	s = S{
		"chunk1": nil,
	}
	res, err = s.Stripe(dests, seq, 0, 1)
	assert.NoError(t, err)
	assert.Equal(t, S{
		"chunk1": NewLocs("a"),
	}, res)

	seq.Reset()

	s = S{
		"chunk1": NewLocs("a"),
	}
	res, err = s.Stripe(dests, seq, 0, 1)
	assert.NoError(t, err)
	assert.Equal(t, S{
		"chunk1": NewLocs(),
	}, res)

	seq.Reset()

	s = S{
		"chunk1": NewLocs("x"),
	}
	res, err = s.Stripe(dests, seq, 0, 1)
	assert.NoError(t, err)
	assert.Equal(t, S{
		"chunk1": NewLocs("a"),
	}, res)

	seq.Reset()

	s = S{
		"chunk1": nil,
		"chunk2": nil,
	}
	res, err = s.Stripe(dests, seq, 0, 2)
	assert.NoError(t, err)
	assert.Equal(t, S{
		"chunk1": NewLocs("a", "b"),
		"chunk2": NewLocs("c", "a"),
	}, res)

	seq.Reset()

	s = S{
		"chunk1": NewLocs("a"),
		"chunk2": nil,
	}
	res, err = s.Stripe(dests, seq, 0, 3)
	assert.NoError(t, err)
	assert.Equal(t, S{
		"chunk1": NewLocs("b", "c"),
		"chunk2": NewLocs("a", "b", "c"),
	}, res)

	seq.Reset()

	s = S{
		"chunk1": NewLocs("a"),
	}
	res, err = s.Stripe(dests, seq, 0, 4)
	short, ok := err.(ShortError)
	assert.True(t, ok)
	assert.Equal(t, 4, short.Min)
	assert.Equal(t, 3, short.Avail)

	seq.Reset()

	s = S{
		"chunk1": NewLocs("a"),
	}
	dests2 := NewLocs("a", "c", "d")
	res, err = s.Stripe(dests2, seq, 0, 3)
	short, ok = err.(ShortError)
	assert.True(t, ok)
	assert.Equal(t, 3, short.Min)
	assert.Equal(t, 2, short.Avail)

	seq.Reset()

	s = S{
		"chunk1": NewLocs("a"),
		"chunk2": NewLocs("b"),
	}
	res, err = s.Stripe(dests, seq, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, S{
		"chunk1": NewLocs("c"),
		"chunk2": NewLocs("c"),
	}, res)

	seq.Reset()

	s = S{
		"chunk1": NewLocs("b", "c"),
		"chunk2": NewLocs("b", "c"),
	}
	res, err = s.Stripe(dests, seq, 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, S{
		"chunk1": NewLocs(),
		"chunk2": NewLocs("a"),
	}, res)

	seq.Reset()

	s = S{
		"chunk1": NewLocs("c"),
		"chunk2": NewLocs("b"),
	}
	res, err = s.Stripe(dests, seq, 1, 1)
	assert.NoError(t, err)
	assert.Equal(t, S{
		"chunk1": NewLocs(),
		"chunk2": NewLocs(),
	}, res)

	seq.Reset()

	s = S{
		"chunk1": NewLocs("a"),
		"chunk2": NewLocs("b"),
	}
	res, err = s.Stripe(dests, seq, 2, 2)
	short, ok = err.(ShortError)
	assert.True(t, ok)
	assert.Equal(t, 2, short.Distinct)
	assert.Equal(t, 2, short.Min)
	assert.Equal(t, 1, short.Avail)

	seq.Reset()

	s = S{
		"chunk1": NewLocs("a"),
		"chunk2": NewLocs("b"),
	}
	dests2 = NewLocs("a", "b", "c", "d")
	seq2 := &RR{Items: []interface{}{"a", "b", "c", "d"}}
	res, err = s.Stripe(dests2, seq2, 2, 2)
	assert.NoError(t, err)
	assert.Equal(t, S{
		"chunk1": NewLocs("c"),
		"chunk2": NewLocs("d"),
	}, res)
}
