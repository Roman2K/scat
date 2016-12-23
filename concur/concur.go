package concur

import "sync"

type Funcs []func() error

func (fns Funcs) FirstErr() (err error) {
	errs := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(len(fns))
	call := func(fn func() error) {
		go func() {
			defer wg.Done()
			errs <- fn()
		}()
	}
	for _, fn := range fns {
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
