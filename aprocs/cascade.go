package aprocs

import ss "secsplit"

type Cascade []Proc

var _ Proc = Cascade{}

func (casc Cascade) Process(c *ss.Chunk) <-chan Res {
	out := make(chan Res)
	go func() {
		defer close(out)
		buf := []Res{}
		for _, proc := range casc {
			ch := proc.Process(c)
			buf = buf[:0]
			err := false
			for res := range ch {
				buf = append(buf, res)
				if res.Err != nil {
					err = true
					break
				}
			}
			for range ch {
			}
			if !err {
				break
			}
		}
		for _, res := range buf {
			out <- res
		}
	}()
	return out
}

func (casc Cascade) Finish() error {
	return finishFuncs([]Proc(casc)).FirstErr()
}
