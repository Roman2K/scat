package concur

import "sync"

type Funcs []func() error

func (fns Funcs) FirstErr() (err error) {
	errs := make(chan error)
	wg := sync.WaitGroup{}
	wg.Add(len(fns))
	for i := range fns {
		fn := fns[i]
		go func() {
			defer wg.Done()
			errs <- fn()
		}()
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
