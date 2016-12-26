package aprocs

import (
	"sync"

	ss "secsplit"
)

func Process(proc Proc, iter ss.ChunkIterator) (err error) {
	errors := make(chan error)
	chunks := make(chan *ss.Chunk)
	done := make(chan struct{})

	go func() {
		defer close(errors)
		wg := sync.WaitGroup{}
		for c := range chunks {
			ch := proc.Process(c)
			wg.Add(1)
			go func() {
				defer wg.Done()
				res := <-ch
				errors <- res.Err
			}()
		}
		wg.Wait()
	}()

	go func() {
		defer close(chunks)
		for iter.Next() {
			c := iter.Chunk()
			select {
			case chunks <- c:
			case <-done:
				return
			}
		}
	}()

	collect := func() (err error) {
		defer close(done)
		for err = range errors {
			if err != nil {
				return
			}
		}
		return
	}

	err = collect()
	for range errors {
	}
	if err != nil {
		return
	}

	return iter.Err()
}
