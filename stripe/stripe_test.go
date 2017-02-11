package stripe

import (
	"bufio"
	"fmt"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestStripe(t *testing.T) {
	origSortItems := sortItems
	defer func() {
		sortItems = origSortItems
	}()
	sortItems = func(items []item) {
		str := func(i int) string {
			return items[i].(string)
		}
		sort.Slice(items, func(i, j int) bool {
			return str(i) < str(j)
		})
	}

	test(t, `
		// nothing to do
		distinct=0 min=0 a,b _
		chunk1 () []

		// empty dests but nothing to do
		distinct=0 min=0 [] a,b
		chunk1 () []

		// empty dests
		distinct=0 min=1 [] a,b
		chunk1 () .
		err=ErrShort

		// empty seq
		distinct=0 min=1 a,b []
		chunk1 () .
		err=ErrShort

		distinct=0 min=1 a,b _
		chunk1 () a
		chunk2 () b
		chunk3 () a

		// ignore locs outside of dests
		distinct=0 min=1 a,b a,b,XXX
		chunk1 () a
		chunk2 () b
		chunk3 () a

		// reuse old locs first
		distinct=0 min=1 a,b,c _
		chunk1 (b) []
		chunk2 () a
		chunk3 () c

		// ignore old locs outside of dests
		distinct=0 min=1 a,b,c _
		chunk1 (XXX) a
		chunk2 () b
		chunk3 () c

		distinct=0 min=2 a,b,c _
		chunk1 () a,b
		chunk2 () c,a
		chunk3 () b,c

		// spread to less used first
		distinct=0 min=2 a,b,c _
		chunk1 (a,b) []
		chunk2 (a) c
		chunk3 () c,b

		distinct=1 min=1 a,b,c _
		chunk1 () a
		chunk2 () b
		chunk3 () c
		chunk4 () a
		chunk5 () b
		chunk6 () a

		distinct=1 min=1 a,b,c _
		chunk1 (b) []
		chunk2 (b) []
		chunk3 () a

		distinct=2 min=1 a,b,c _
		chunk1 (b) []
		chunk2 (b) c
		chunk3 () a

		distinct=1 min=2 a,b,c _
		chunk1 () a,b
		chunk2 () c .
		err=ErrShort

		distinct=1 min=2 a,b,c,d _
		chunk1 () a,b
		chunk2 () c,d
		chunk3 () a,b
		chunk4 () a,b

		distinct=2 min=1 a,b,c _
		chunk1 () a
		chunk2 () b
		chunk3 () c
		chunk4 () a
		chunk5 () a

		distinct=2 min=2 a,b,c,d _
		chunk1 () a,b
		chunk2 () c,d
		chunk3 () .
		err=ErrShort
	`)
}

func test(t *testing.T, spec string) {
	const (
		empty  = "[]"
		idem   = "_"
		strSep = ","
	)
	var (
		commentRe = regexp.MustCompile(`^//`)
		configRe  = regexp.MustCompile(`^distinct=(\d+) min=(\d+) (.+) (.+)?$`)
		itemRe    = regexp.MustCompile(`^(.+) \((.*)\) (.+)?$`)
		errRe     = regexp.MustCompile(`^err=(.+)$`)
		errors    = map[string]error{
			"ErrShort": ErrShort,
		}
		split = func(s string) []string {
			switch s {
			case "", empty:
				return nil
			}
			return strings.Split(s, strSep)
		}
		locStrings = func(s string) (locs Locs) {
			parts := split(s)
			locs = make(Locs, len(parts))
			for _, s := range parts {
				locs.Add(s)
			}
			return
		}
		seqStrings = func(s string) Seq {
			parts := split(s)
			items := make([]interface{}, len(parts))
			for i, s := range parts {
				items[i] = s
			}
			return &RR{Items: items}
		}
	)
	_, _, lineNr, ok := runtime.Caller(1)
	assert.True(t, ok)
	lineNr -= strings.Count(spec, "\n")
	scan := bufio.NewScanner(strings.NewReader(spec))
	for {
		var (
			subLine     = 0
			blank       = 0
			distinct    = -1
			min         = -1
			seq         Seq
			dests       Locs
			s           = make(S)
			expected    = make(S)
			expectedErr error
		)
		run := func() {
			fmt.Printf("running test at line %d\n", lineNr)
			fmt.Printf("  distinct=%d\n", distinct)
			fmt.Printf("  min=%d\n", min)
			fmt.Printf("  seq=%v\n", seq)
			fmt.Printf("  dests=%v\n", dests)
			fmt.Printf("  s=%v\n", s)
			fmt.Printf("  expected=%v\n", expected)
			fmt.Printf("  expectedErr=%v\n", expectedErr)
			res, err := s.Stripe(dests, seq, distinct, min)
			if expectedErr != nil {
				assert.Equal(t, expectedErr, err, fmt.Sprintf("returned %v", res))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expected, res)
			}
			fmt.Printf("  OK\n")
		}
		for scan.Scan() {
			subLine++
			line := strings.TrimSpace(scan.Text())
			if len(line) == 0 {
				blank++
				if blank > 0 && distinct > -1 {
					run()
					break
				}
				continue
			}
			blank = 0
			if commentRe.MatchString(line) {
				continue
			}
			if m := configRe.FindStringSubmatch(line); m != nil {
				distinctS, minS, destsS, seqS := m[1], m[2], m[3], m[4]
				var err error
				distinct, err = strconv.Atoi(distinctS)
				assert.NoError(t, err)
				min, err = strconv.Atoi(minS)
				assert.NoError(t, err)
				dests = locStrings(destsS)
				if seqS == idem {
					seqS = destsS
				}
				seq = seqStrings(seqS)
				continue
			}
			if m := itemRe.FindStringSubmatch(line); m != nil {
				item := m[1]
				_, ok := s[item]
				assert.False(t, ok)
				s[item] = locStrings(m[2])
				expected[item] = locStrings(m[3])
				continue
			}
			if m := errRe.FindStringSubmatch(line); m != nil {
				err, ok := errors[m[1]]
				assert.True(t, ok)
				expectedErr = err
				continue
			}
			t.Fatalf("invalid line: %q", line)
		}
		if subLine <= 0 {
			break
		}
		lineNr += subLine
	}
	err := scan.Err()
	assert.NoError(t, err)
}
