package stripe

type RR struct {
	Items []interface{}
	cur   int
}

func (rr *RR) Next() interface{} {
	n := len(rr.Items)
	if rr.cur > n-1 {
		return nil
	}
	item := rr.Items[rr.cur]
	rr.cur = (rr.cur + 1) % n
	return item
}

func (rr *RR) Reset() {
	rr.cur = 0
}
