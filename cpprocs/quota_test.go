package cpprocs_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"secsplit/cpprocs"
)

func TestQuotaMan(t *testing.T) {
	a := cpprocs.NewCopier("a", nil)
	b := cpprocs.NewCopier("b", nil)
	ids := func(copiers []cpprocs.Copier) (res []string) {
		res = []string{}
		for _, cp := range copiers {
			res = append(res, cp.Id().(string))
		}
		return
	}
	man := make(cpprocs.QuotaMan)

	// none
	assert.Equal(t, []string{}, ids(man.Copiers(4)))

	// a) quota = unlimited, used = 0
	man.AddCopier(a, nil)
	assert.Equal(t, []string{"a"}, ids(man.Copiers(4)))

	// a) quota = unlimited, used = 1
	man.AddCopier(a, []cpprocs.LsEntry{{Size: 1}})
	assert.Equal(t, []string{"a"}, ids(man.Copiers(4)))

	// a) quota = 100, used = 2
	a.SetQuota(100)
	man.AddUse(a, 1)
	assert.Equal(t, []string{"a"}, ids(man.Copiers(97)))
	assert.Equal(t, []string{}, ids(man.Copiers(98)))

	// a) quota = 100, used = 100
	man.AddUse(a, 98)
	assert.Equal(t, []string{}, ids(man.Copiers(0)))
	assert.Equal(t, []string{}, ids(man.Copiers(1)))

	// a) quota = 100, used = 101
	man.AddUse(a, 1)
	assert.Equal(t, []string{}, ids(man.Copiers(0)))

	// b)
	man.AddCopier(b, []cpprocs.LsEntry{{Size: 1}})
	assert.Equal(t, []string{"b"}, ids(man.Copiers(4)))

	// b) deleted
	man.Delete(b)
	assert.Equal(t, []string{}, ids(man.Copiers(4)))
	assert.Equal(t, []string{}, ids(man.Copiers(0)))

	// b) deleted, used = 1
	man.AddUse(b, 1)
	assert.Equal(t, []string{}, ids(man.Copiers(0)))
}
