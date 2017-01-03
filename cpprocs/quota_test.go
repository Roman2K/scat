package cpprocs_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"secsplit/cpprocs"
)

func TestQuotaMan(t *testing.T) {
	a := cpprocs.NewCopier("a", nil, nil)
	b := cpprocs.NewCopier("b", nil, nil)
	ids := func(copiers []cpprocs.Copier) (res []string) {
		res = []string{}
		for _, cp := range copiers {
			res = append(res, cp.Id().(string))
		}
		return
	}
	man := cpprocs.NewQuotaMan()
	assert.Equal(t, []string{}, ids(man.Copiers()))

	man.AddCopier(a, nil)
	assert.Equal(t, []string{"a"}, ids(man.Copiers()))

	man.AddCopier(a, []cpprocs.LsEntry{{Size: 1}})
	assert.Equal(t, []string{"a"}, ids(man.Copiers()))

	a.SetQuota(100)
	man.AddCopier(a, []cpprocs.LsEntry{{Size: 1}})
	assert.Equal(t, []string{"a"}, ids(man.Copiers()))

	man.AddCopier(a, []cpprocs.LsEntry{{Size: 98}})
	assert.Equal(t, []string{}, ids(man.Copiers()))

	man.AddCopier(a, []cpprocs.LsEntry{{Size: 1}})
	assert.Equal(t, []string{}, ids(man.Copiers()))

	man.AddCopier(b, []cpprocs.LsEntry{{Size: 1}})
	assert.Equal(t, []string{"b"}, ids(man.Copiers()))

	man.Delete(b)
	assert.Equal(t, []string{}, ids(man.Copiers()))

	man.AddCopier(b, []cpprocs.LsEntry{{Size: 1}})
	assert.Equal(t, []string{}, ids(man.Copiers()))
}
