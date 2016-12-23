package concur

import "sync"

func FirstErr(funcs ...func() error) (err error) {
	errs := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(len(funcs))
	call := func(fn func() error) {
		go func() {
			defer wg.Done()
			errs <- fn()
		}()
	}
	for _, fn := range funcs {
		go call(fn)
	}
	go func() {
		defer close(errs)
		wg.Wait()
	}()
	for e := range errs {
		if e != nil && err == nil {
			err = e
		}
	}
	return
}
