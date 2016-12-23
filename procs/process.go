package procs

import (
	"sync"

	ss "secsplit"
)

func Process(proc Proc, it ss.ChunkIterator) (err error) {
	for it.Next() {
		err = proc.Process(it.Chunk()).Err
		if err != nil {
			return
		}
	}
	return it.Err()
}

func ProcessAsync(proc AsyncProc, it ss.ChunkIterator) error {
	chunks := make(chan *ss.Chunk)
	results := make(chan error)
	done := make(chan struct{})
	resultSends := sync.WaitGroup{}

	resultSends.Add(1)
	go func() {
		defer resultSends.Done()
		defer close(chunks)
		for it.Next() {
			select {
			case chunks <- it.Chunk():
			case <-done:
				return
			}
		}
		results <- it.Err()
	}()

	resultSends.Add(1)
	go func() {
		defer resultSends.Done()
		for c := range chunks {
			ch := proc.Process(c)
			resultSends.Add(1)
			go func() {
				defer resultSends.Done()
				res := <-ch
				results <- res.Err
			}()
		}
	}()

	go func() {
		defer close(results)
		resultSends.Wait()
	}()

	collect := func() error {
		defer close(done)
		for err := range results {
			if err != nil {
				return err
			}
		}
		return nil
	}

	err := collect()
	for range results {
	}
	return err
}
