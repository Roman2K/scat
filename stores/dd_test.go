package stores_test

import (
	"scat/stores"
	"testing"
)

func TestDd(t *testing.T) {
	dirStoreTest(func(dir stores.Dir) stores.Store {
		return stores.Dd{Dir: dir}
	}).test(t)
}
