package concur_test

import (
	"errors"
	"secsplit/concur"
	"testing"
	"time"

	assert "github.com/stretchr/testify/require"
)

func TestFuncs(t *testing.T) {
	const delay = 20 * time.Millisecond
	someErr := errors.New("some err")
	a := func() error { time.Sleep(delay); return nil }
	b := func() error { time.Sleep(delay); return someErr }
	start := time.Now()
	err := concur.Funcs{a, b}.FirstErr()
	elapsed := time.Now().Sub(start)
	assert.Equal(t, someErr, err)
	assert.True(t, elapsed > delay)
	assert.True(t, elapsed < delay+delay/2)
}
