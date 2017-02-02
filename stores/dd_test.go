package stores_test

import (
	"testing"

	"gitlab.com/Roman2K/scat/stores"
)

func TestDd(t *testing.T) {
	dirStoreTest(func(dir stores.Dir) stores.Store {
		return stores.Dd{Dir: dir}
	}).run(t)
}
