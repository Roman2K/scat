// Manage rolling series of values sorted by a position
package seriessort

type Series struct {
	values []interface{}
	start  int
}

// Add an element at position i. As values are dropped, adding new values in
// their place is not supported. So, i must be >= min position.
func (s *Series) Add(i int, val interface{}) {
	i -= s.start
	if i < 0 {
		panic("position lower than minimum")
	}
	if minLen := i + 1; len(s.values) < minLen {
		if cap(s.values) < minLen {
			resized := make([]interface{}, minLen, i*2+1)
			copy(resized, s.values)
			s.values = resized
		}
		s.values = s.values[:minLen]
	}
	s.values[i] = val
}

// Removes at most count values from the beginning of the series to free up
// memory.
func (s *Series) Drop(count int) {
	if count < 0 {
		count = 0
	}
	if n := len(s.values); count > n {
		count = n
	}
	s.values = s.values[count:]
	s.start += count
}

// Sorted values from the beginning up to and excluding an unset value.
func (s Series) Sorted() (values []interface{}) {
	i := 0
	for n := len(s.values); i < n; i++ {
		if s.values[i] == nil {
			break
		}
	}
	values = make([]interface{}, i)
	copy(values, s.values[:i])
	return
}

func (s Series) Len() int {
	return len(s.values)
}
