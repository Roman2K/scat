package quota_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"secsplit/cpprocs/quota"
)

func TestMan(t *testing.T) {
	man := quota.NewMan()
	a := resource("a")
	b := resource("b")

	ids := func(ress []quota.Res) (strs []string) {
		strs = []string{}
		for _, res := range ress {
			strs = append(strs, string(res.(resource)))
		}
		return
	}

	// none
	assert.Equal(t, []string{}, ids(man.Resources(4)))

	// a) quota = unlimited, used = 0
	man.AddRes(a)
	assert.Equal(t, []string{"a"}, ids(man.Resources(4)))

	// a) quota = unlimited, used = 1
	man.AddRes(a)
	man.AddUse(a, 1)
	assert.Equal(t, []string{"a"}, ids(man.Resources(4)))

	// a) quota = 100, used = 2
	man.AddResQuota(a, 100)
	man.AddUse(a, 1)
	assert.Equal(t, []string{"a"}, ids(man.Resources(97)))
	assert.Equal(t, []string{}, ids(man.Resources(98)))

	// a) quota = 100, used = 100
	man.AddUse(a, 98)
	assert.Equal(t, []string{}, ids(man.Resources(0)))
	assert.Equal(t, []string{}, ids(man.Resources(1)))

	// a) quota = 100, used = 101
	man.AddUse(a, 1)
	assert.Equal(t, []string{}, ids(man.Resources(0)))

	// b)
	man.AddRes(b)
	man.AddUse(b, 1)
	assert.Equal(t, []string{"b"}, ids(man.Resources(4)))

	// b) deleted
	man.Delete(b)
	assert.Equal(t, []string{}, ids(man.Resources(4)))
	assert.Equal(t, []string{}, ids(man.Resources(0)))

	// b) deleted, used = 1
	man.AddUse(b, 1)
	assert.Equal(t, []string{}, ids(man.Resources(0)))
}

type resource string

func (res resource) Id() interface{} {
	return res
}
