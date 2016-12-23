package procs_test

import (
	"testing"

	assert "github.com/stretchr/testify/require"

	"secsplit/procs"
)

func TestIterFinish(t *testing.T) {
	iter := procs.Iter()
	err := iter.Finish()
	assert.NoError(t, err)

	// idempotence
	err = iter.Finish()
	assert.NoError(t, err)
}
