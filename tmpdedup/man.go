package tmpdedup

import (
	"errors"
	"os"
	"sync"
)

type Man struct {
	created   created
	createdMu sync.Mutex
	wg        sync.WaitGroup
}

type created map[string]*sync.WaitGroup

func NewMan() *Man {
	return &Man{
		created: make(created),
		wg:      sync.WaitGroup{},
	}
}

func (man *Man) Get(path string, create func() error) (
	wg *sync.WaitGroup, err error,
) {
	man.createdMu.Lock()
	defer man.createdMu.Unlock()
	wg, ok := man.created[path]
	if ok {
		wg.Add(1)
		return
	}
	err = create()
	if err != nil {
		return
	}
	wg = &sync.WaitGroup{}
	wg.Add(1)
	man.wg.Add(1)
	go func() {
		wg.Wait()
		err := man.remove(path)
		if err != nil {
			panic(err)
		}
		man.wg.Done()
	}()
	man.created[path] = wg
	return
}

func (man *Man) remove(path string) (err error) {
	man.createdMu.Lock()
	defer man.createdMu.Unlock()
	if _, ok := man.created[path]; !ok {
		err = errors.New("file not supposed to be created")
		return
	}
	err = os.Remove(path)
	if err != nil {
		return
	}
	delete(man.created, path)
	return
}

func (man *Man) Len() int {
	man.createdMu.Lock()
	defer man.createdMu.Unlock()
	return len(man.created)
}

func (man *Man) Wait() {
	man.wg.Wait()
}
